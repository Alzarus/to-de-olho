package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Alzarus/to-de-olho/internal/acesso"
	"github.com/Alzarus/to-de-olho/internal/api"
	"github.com/Alzarus/to-de-olho/internal/ceaps"
	"github.com/Alzarus/to-de-olho/internal/comissao"
	"github.com/Alzarus/to-de-olho/internal/emenda"
	"github.com/Alzarus/to-de-olho/internal/gabinete"
	"github.com/Alzarus/to-de-olho/internal/materia"
	"github.com/Alzarus/to-de-olho/internal/proposicao"
	"github.com/Alzarus/to-de-olho/internal/ranking"
	"github.com/Alzarus/to-de-olho/internal/scheduler"
	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/internal/votacao"
	"github.com/Alzarus/to-de-olho/pkg/senado"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strconv"
)

func main() {
	// "server healthcheck" e chamado pelo HEALTHCHECK do Dockerfile: a imagem
	// distroless nao tem shell nem curl, entao o proprio binario faz o GET.
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(healthcheck())
	}

	// Configurar logger estruturado (JSON para Cloud Run)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Carregar .env em ambiente local
	if err := godotenv.Load(); err != nil {
		slog.Warn("arquivo .env nao encontrado (normal em producao se usar vars de ambiente)")
	}

	// Conectar ao banco de dados
	db, err := connectDB()
	if err != nil {
		slog.Error("falha ao conectar ao banco", "error", err)
		os.Exit(1)
	}

	// Auto-migrate das entidades
	if err := db.AutoMigrate(
		&senador.Senador{},
		&senador.Mandato{},
		&ceaps.DespesaCEAPS{},
		&votacao.Votacao{},
		&comissao.ComissaoMembro{},
		&proposicao.Proposicao{},
		&emenda.Emenda{},
		&acesso.Visita{},
		&acesso.Sal{},
		&materia.Materia{},
		&materia.ApelidoCurado{},
	); err != nil {
		slog.Error("falha no auto-migrate", "error", err)
		os.Exit(1)
	}
	// Estrutura de gabinete (numeros agregados: gabinete_recursos, gabinete_beneficios, gabinete_mesa)
	if err := db.AutoMigrate(gabinete.Modelos()...); err != nil {
		slog.Error("falha no auto-migrate", "error", err)
		os.Exit(1)
	}
	// Reaplica a regra de pontuacao vigente (v2.2) nas proposicoes gravadas (idempotente)
	if _, err := proposicao.RecalcularPontuacao(db); err != nil {
		slog.Error("falha ao recalcular pontuacao das proposicoes", "error", err)
		os.Exit(1)
	}
	// Completa colunas novas de votacoes nas linhas antigas (idempotente)
	if err := votacao.PreencherHistorico(db); err != nil {
		slog.Error("falha ao preencher historico de votacoes", "error", err)
		os.Exit(1)
	}
	// Nome popular das materias: curadoria com fonte oficial e codigo_materia
	// das votacoes antigas que o banco consegue derivar (idempotentes)
	if err := materia.SemearApelidosCurados(db); err != nil {
		slog.Error("falha ao gravar apelidos curados", "error", err)
		os.Exit(1)
	}
	if err := votacao.PreencherCodigoMateria(db); err != nil {
		slog.Error("falha ao preencher codigo_materia das votacoes", "error", err)
		os.Exit(1)
	}
	// Fotos gravadas em http:// (conteúdo misto no site em https) passam a
	// https; o sync já grava normalizado (idempotente)
	if err := senador.NormalizarFotosHTTPS(db); err != nil {
		slog.Error("falha ao normalizar fotos dos senadores", "error", err)
		os.Exit(1)
	}

	/*
		// Redis foi removido por questoes de custo no GCP
		// O sistema operara exclusivamente em cima da performance do Postgres
		// e de caches na borda (CDN/Next.js)
	*/

	// Configurar router
	transparenciaKey := os.Getenv("TRANSPARENCIA_API_KEY")
	router := api.SetupRouter(db, transparenciaKey)

	// Criar servidor HTTP
	srv := novoServidor(getPort(), router)

	// --- Inicializar Services para Scheduler (Duplicado do Router por enquanto) ---
	// Repositorios
	senadorRepo := senador.NewRepository(db)
	votacaoRepo := votacao.NewRepository(db)
	ceapsRepo := ceaps.NewRepository(db)
	emendaRepo := emenda.NewRepository(db)
	comissaoRepo := comissao.NewRepository(db)
	proposicaoRepo := proposicao.NewRepository(db)

	// Clients
	legisClient := senado.NewLegisClient()
	admClient := senado.NewAdmClient()

	// Sync Services (Modules)
	senadorSync := senador.NewSyncService(senadorRepo, legisClient)
	votacaoSync := votacao.NewSyncService(votacaoRepo, senadorRepo, legisClient)
	ceapsSync := ceaps.NewSyncService(ceapsRepo, senadorRepo, admClient)
	emendaSync := emenda.NewSyncService(emendaRepo, senadorRepo, transparenciaKey)
	comissaoSync := comissao.NewSyncService(comissaoRepo, senadorRepo, legisClient)
	proposicaoSync := proposicao.NewSyncService(proposicaoRepo, senadorRepo, legisClient)

	// Ranking Service (necessario para recalcular aps sync, suporta redis mas passamos nil)
	rankingService := ranking.NewService(
		senadorRepo,
		proposicaoRepo,
		votacaoRepo,
		ceapsRepo,
		comissaoRepo,
	)

	// Iniciar Scheduler
	sched := scheduler.NewScheduler(
		senadorSync,
		votacaoSync,
		ceapsSync,
		emendaSync,
		comissaoSync,
		proposicaoSync,
		rankingService,
		senadorRepo,
		votacaoRepo,
	).ComMaterias(materia.NewSyncService(db, legisClient))

	// Contexto para o scheduler (cancelado no shutdown)
	ctxSched, cancelSched := context.WithCancel(context.Background())
	defer cancelSched()

	// Gabinete: carga semanal (guarda no sync diario) e no backfill
	sched.SetGabineteSync(gabinete.NewSyncService(gabinete.NewRepository(db), senadorRepo, admClient, legisClient))

	// O ensaio de migração do deploy (scripts/deploy/ensaio-migracao.sh) sobe
	// esta imagem contra uma cópia descartável do banco só para ver se ela
	// chega ao /health. Lá o sync diário não pode rodar: gravaria na cópia e
	// chamaria as APIs do Senado sem motivo.
	if agendadorDesligado() {
		slog.Warn("scheduler desligado por SCHEDULER_DESATIVADO")
	} else {
		sched.Start(ctxSched)
	}

	// Registrar endpoint de sync diario (Cloud Scheduler)
	api.RegisterSchedulerRoutes(router, sched)
	// -----------------------------------------------------------------------------

	// Iniciar servidor em goroutine
	go func() {
		slog.Info("servidor iniciando", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("falha no servidor", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	slog.Info("encerrando servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown forcado", "error", err)
	}

	slog.Info("servidor encerrado")
}

func connectDB() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Default para desenvolvimento local
		dsn = "host=localhost user=postgres password=postgres dbname=todeolho port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true, // Cache de prepared statements
	})
	if err != nil {
		return nil, err
	}

	// Configurar connection pool dinamico para escalabilidade
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	pool := poolDoAmbiente()
	sqlDB.SetMaxOpenConns(pool.abertas)
	sqlDB.SetMaxIdleConns(pool.ociosas)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// configPool é o tamanho do pool de conexões com o Postgres.
type configPool struct {
	abertas, ociosas int
}

// poolDoAmbiente lê DB_MAX_OPEN_CONNS e DB_MAX_IDLE_CONNS.
//
// O padrão antigo (100 abertas, pensado para o Cloud Run) era o
// max_connections inteiro do Postgres 15: sob carga a API sozinha esgotava o
// banco e sobrava zero conexão para o psql de manutenção, o backup e o
// superusuário. 20 abertas dão conta da API (as consultas levam
// milissegundos, e o limite por origem segura a fila antes do banco) e
// deixam 80 livres. 10 ociosas evitam reabrir conexão a cada rajada;
// ConnMaxIdleTime devolve ao banco as que ficarem paradas.
func poolDoAmbiente() configPool {
	p := configPool{abertas: 20, ociosas: 10}
	if v, err := strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNS")); err == nil && v > 0 {
		p.abertas = v
	}
	if v, err := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNS")); err == nil && v >= 0 {
		p.ociosas = v
	}
	// Ociosas acima de abertas não têm efeito; deixa explícito.
	p.ociosas = min(p.ociosas, p.abertas)
	return p
}

// novoServidor monta o http.Server com prazos.
//
// Sem nenhum timeout, um cliente que manda os headers um byte por vez
// (slowloris) segura uma goroutine e um descritor para sempre; poucos milhares
// derrubam a API. A Cloudflare filtra isso na borda, mas a origem precisa se
// defender sozinha para o caso de ser alcançada direto.
//   - ReadHeaderTimeout 10 s: tempo para chegarem os headers.
//   - ReadTimeout 30 s: request inteiro. A API só recebe GET e POST sem corpo.
//   - WriteTimeout 60 s: da leitura ao fim da resposta. A consulta mais
//     pesada (alinhamento, ranking sem cache) leva poucos segundos; os syncs
//     longos tiram o prazo depois de autenticados (api.semPrazoDeEscrita).
//   - IdleTimeout 120 s: keep-alive parado é fechado.
//   - MaxHeaderBytes 64 KiB: o padrão é 1 MiB. Nenhuma URL do site chega
//     perto disso, e query string gigante (?q=, ?tipo=) viraria LIKE e IN
//     enormes no banco.
func novoServidor(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    64 << 10,
	}
}

// healthcheck devolve 0 se a API local responde 200 em /health, 1 caso contrario.
func healthcheck() int {
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://127.0.0.1" + getPort() + "/health")
	if err != nil {
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}

// agendadorDesligado lê SCHEDULER_DESATIVADO ("1" ou "true").
func agendadorDesligado() bool {
	v, _ := strconv.ParseBool(strings.TrimSpace(os.Getenv("SCHEDULER_DESATIVADO")))
	return v
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return ":" + port
}
