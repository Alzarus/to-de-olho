package scheduler

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/Alzarus/to-de-olho/internal/ceaps"
	"github.com/Alzarus/to-de-olho/internal/comissao"
	"github.com/Alzarus/to-de-olho/internal/emenda"
	"github.com/Alzarus/to-de-olho/internal/gabinete"
	"github.com/Alzarus/to-de-olho/internal/materia"
	"github.com/Alzarus/to-de-olho/internal/proposicao"
	"github.com/Alzarus/to-de-olho/internal/ranking"
	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/internal/utils"
	"github.com/Alzarus/to-de-olho/internal/votacao"
	"github.com/Alzarus/to-de-olho/pkg/retry"
)

// Scheduler gerencia tarefas agendadas
type Scheduler struct {
	senadorSync    *senador.SyncService
	votacaoSync    *votacao.SyncService
	ceapsSync      *ceaps.SyncService
	emendaSync     *emenda.SyncService
	comissaoSync   *comissao.SyncService
	proposicaoSync *proposicao.SyncService
	rankingService *ranking.Service
	senadorRepo    *senador.Repository
	votacaoRepo    *votacao.Repository
	materiaSync    *materia.SyncService  // opcional: ComMaterias
	gabineteSync   *gabinete.SyncService // opcional (SetGabineteSync)
}

// ComMaterias liga o sync das materias (nome popular e descricao) ao sync
// diario e ao backfill.
func (s *Scheduler) ComMaterias(m *materia.SyncService) *Scheduler {
	s.materiaSync = m
	return s
}

// syncMaterias completa codigo_materia nas votacoes antigas e busca o detalhe
// das materias pendentes. limite <= 0: todas.
func (s *Scheduler) syncMaterias(ctx context.Context, limite int) {
	if s.materiaSync == nil {
		return
	}
	if err := s.votacaoSync.CompletarCodigoMateria(ctx); err != nil {
		slog.Error("falha ao completar codigo_materia das votacoes", "error", err)
	}
	if _, err := s.materiaSync.Sync(ctx, limite); err != nil {
		slog.Error("falha sync materias", "error", err)
	}
}

// intervaloGabinete e a frequencia da carga de gabinete: a fonte muda pouco e
// a rota e lenta, entao o sync diario so roda quando a ultima carga passou disso.
const intervaloGabinete = 7 * 24 * time.Hour

// SetGabineteSync liga a carga de estrutura de gabinete ao scheduler
func (s *Scheduler) SetGabineteSync(g *gabinete.SyncService) {
	s.gabineteSync = g
}

// syncGabinete roda a carga do ano atual (e, em janeiro, do ano anterior, que
// fecha na fonte) quando a guarda semanal permite.
func (s *Scheduler) syncGabinete(ctx context.Context, agora time.Time) {
	if s.gabineteSync == nil {
		return
	}
	anos := []int{agora.Year()}
	if agora.Month() == time.January {
		anos = append(anos, agora.Year()-1)
	}
	for _, ano := range anos {
		precisa, err := s.gabineteSync.PrecisaAtualizar(ano, intervaloGabinete)
		if err != nil {
			slog.Error("falha na guarda do sync de gabinete", "ano", ano, "error", err)
			continue
		}
		if !precisa {
			slog.Info("gabinete atualizado ha menos de uma semana, pulando", "ano", ano)
			continue
		}
		if _, err := s.gabineteSync.SyncAno(ctx, ano); err != nil {
			slog.Error("falha sync gabinete", "ano", ano, "error", err)
		}
	}
	if err := s.gabineteSync.SyncMesa(ctx); err != nil {
		slog.Error("falha sync mesa diretora", "error", err)
	}
}

// NewScheduler cria um novo scheduler
func NewScheduler(
	senadorSync *senador.SyncService,
	votacaoSync *votacao.SyncService,
	ceapsSync *ceaps.SyncService,
	emendaSync *emenda.SyncService,
	comissaoSync *comissao.SyncService,
	proposicaoSync *proposicao.SyncService,
	rankingService *ranking.Service,
	senadorRepo *senador.Repository,
	votacaoRepo *votacao.Repository,
) *Scheduler {
	return &Scheduler{
		senadorSync:    senadorSync,
		votacaoSync:    votacaoSync,
		ceapsSync:      ceapsSync,
		emendaSync:     emendaSync,
		comissaoSync:   comissaoSync,
		proposicaoSync: proposicaoSync,
		rankingService: rankingService,
		senadorRepo:    senadorRepo,
		votacaoRepo:    votacaoRepo,
	}
}

// Start inicia o loop de agendamento em background.
// O backfill nao e mais executado na startup; deve ser disparado
// via HTTP (POST /api/v1/sync/backfill) para que o Cloud Run
// mantenha o container vivo durante a execucao.
func (s *Scheduler) Start(ctx context.Context) {
	slog.Info("iniciando scheduler")

	go func() {
		for {
			proxima := proximaExecucao(time.Now())
			slog.Info("proximo sync diario agendado", "em", proxima.Format(time.RFC3339))
			timer := time.NewTimer(time.Until(proxima))
			select {
			case <-ctx.Done():
				timer.Stop()
				slog.Info("parando scheduler")
				return
			case <-timer.C:
				s.RunDailySync(ctx)
			}
		}
	}()
}

// horaSyncDiarioUTC e a hora fixa do sync diario (06:00 UTC = 03:00 em Brasilia).
// Com um ticker de 24 h contado a partir da subida, cada deploy empurrava o sync
// para o dia seguinte, e com deploys diarios ele nunca rodava.
const horaSyncDiarioUTC = 6

// proximaExecucao devolve o proximo horario fixo do sync diario depois de agora.
func proximaExecucao(agora time.Time) time.Time {
	u := agora.UTC()
	p := time.Date(u.Year(), u.Month(), u.Day(), horaSyncDiarioUTC, 0, 0, 0, time.UTC)
	if !p.After(u) {
		p = p.AddDate(0, 0, 1)
	}
	return p
}

// RunBackfill executa o backfill completo de todos os anos.
// Exportado para ser chamado sincronamente pelo endpoint HTTP,
// garantindo que o Cloud Run mantenha o container vivo.
func (s *Scheduler) RunBackfill(ctx context.Context) {
	forceBackfill := true // Sempre forca quando chamado via HTTP

	// 1. Verificar se ja existem dados
	count, err := s.senadorRepo.Count()
	if err != nil {
		slog.Error("falha ao verificar contagem de senadores", "error", err)
		return
	}

	if count > 0 && !forceBackfill {
		slog.Info("banco de dados ja populado, pulando backfill", "senadores", count)
		return
	}

	slog.Info("INICIANDO BACKFILL COMPLETO via HTTP", "senadores_existentes", count)

	// 2. Determinar ano de inicio
	anoInicio := 2023
	if envAno := os.Getenv("INICIO_BACKFILL"); envAno != "" {
		if parsed, err := strconv.Atoi(envAno); err == nil {
			anoInicio = parsed
		}
	}
	anoAtual := time.Now().Year()

	slog.Info("configuracao de backfill", "ano_inicio", anoInicio, "ano_fim", anoAtual)

	// 3. Sequencia de Sync (com retry em cada passo)

	// A. Dados Basicos (Senadores)
	slog.Info("--- PASSO 1/6: SENADORES ---")
	if err := retry.WithRetry(ctx, 3, "backfill-senadores", func() error {
		return s.senadorSync.SyncFromAPI(ctx)
	}); err != nil {
		slog.Error("falha critica no backfill de senadores", "error", err)
		return // Sem senadores nao da pra continuar
	}
	// periodos de exercicio: base do teto da cota e do piso de tempo (item 8)
	if err := s.senadorSync.SyncExercicios(ctx); err != nil {
		slog.Error("falha no backfill de exercicios", "error", err)
	}

	// B. Votacoes do recorte, por intervalo de datas (upsert: pode repetir)
	slog.Info("--- PASSO 2/6: VOTACOES ---")
	if err := s.votacaoSync.SyncFromAPI(ctx); err != nil {
		slog.Error("falha no backfill de votacoes", "error", err)
	}

	// C. Loop por ano para dados periodicos
	for ano := anoInicio; ano <= anoAtual; ano++ {
		slog.Info("--- PROCESSANDO ANO ---", "ano", ano)
		anoLoop := ano

		// CEAPS (Despesas)
		if err := retry.WithRetry(ctx, 3, "backfill-ceaps", func() error {
			return s.ceapsSync.SyncFromAPI(ctx, anoLoop)
		}); err != nil {
			slog.Error("falha ao sincronizar ceaps", "ano", ano, "error", err)
		}

		// Emendas
		if err := retry.WithRetry(ctx, 3, "backfill-emendas", func() error {
			return s.emendaSync.SyncAll(ctx, anoLoop)
		}); err != nil {
			slog.Error("falha ao sincronizar emendas", "ano", ano, "error", err)
		}
	}

	// D. Comissoes (Estado atual/recente)
	slog.Info("--- PASSO 4/6: COMISSOES ---")
	if err := retry.WithRetry(ctx, 3, "backfill-comissoes", func() error {
		return s.comissaoSync.SyncFromAPI(ctx)
	}); err != nil {
		slog.Error("falha no backfill de comissoes", "error", err)
	}

	// E. Proposicoes (Historico)
	slog.Info("--- PASSO 5/6: PROPOSICOES ---")
	// o retry e por senador, dentro do SyncFromAPI
	if err := s.proposicaoSync.SyncFromAPI(ctx); err != nil {
		slog.Error("falha no backfill de proposicoes", "error", err)
	}

	// Materias (nome popular e descricao): depois de votacoes e proposicoes,
	// que dao os codigos. Nao entra no ranking.
	s.syncMaterias(ctx, 0)
	// Gabinete (numeros agregados): Mesa e todos os anos do recorte
	if s.gabineteSync != nil {
		slog.Info("--- GABINETE ---")
		if err := s.gabineteSync.SyncTodos(ctx, gabinete.AnosDoRecorte(time.Now())); err != nil {
			slog.Error("falha no backfill de gabinete", "error", err)
		}
	}

	// F. Calculo de Ranking Final, so com a carga completa
	slog.Info("--- PASSO 6/6: CALCULANDO RANKING ---")
	if !s.cargaCompleta() {
		return
	}
	s.rankingService.InvalidateCache()
	if _, err := s.rankingService.CalcularRanking(ctx, nil); err != nil {
		slog.Error("falha ao calcular ranking inicial", "error", err)
	} else {
		slog.Info("BACKFILL COMPLETO COM SUCESSO!")
	}
}

// RunDailySync executa o sync diario completo com retry em cada passo.
// Exportado para ser chamado pelo endpoint HTTP do Cloud Scheduler.
func (s *Scheduler) RunDailySync(ctx context.Context) {
	slog.Info("executando sync diario integral")

	anoAtual := time.Now().Year()

	// 1. Senadores (Atualizacao cadastral)
	if err := retry.WithRetry(ctx, 3, "sync-senadores", func() error {
		return s.senadorSync.SyncFromAPI(ctx)
	}); err != nil {
		slog.Error("falha sync senadores", "error", err)
	}
	if err := s.senadorSync.SyncExercicios(ctx); err != nil {
		slog.Error("falha sync exercicios", "error", err)
	}

	// 2. Votacoes dos ultimos 30 dias: 1 chamada por mes, com retry, traz todas
	// as cadeiras de cada votacao. Antes o sync diario so atualizava metadados
	// e nenhuma votacao nova entrava no banco.
	if _, err := s.votacaoSync.SyncRecentes(ctx, 30); err != nil {
		slog.Error("falha sync votacoes recentes", "error", err)
	}

	// 4. CEAPS (Despesas)
	if err := retry.WithRetry(ctx, 3, "sync-ceaps", func() error {
		return s.ceapsSync.SyncFromAPI(ctx, anoAtual)
	}); err != nil {
		slog.Error("falha sync ceaps", "error", err)
	}

	// 5. Emendas
	if err := retry.WithRetry(ctx, 3, "sync-emendas", func() error {
		return s.emendaSync.SyncAll(ctx, anoAtual)
	}); err != nil {
		slog.Error("falha sync emendas", "error", err)
	}

	// 6. Comissoes (Mudancas de membros)
	if err := retry.WithRetry(ctx, 3, "sync-comissoes", func() error {
		return s.comissaoSync.SyncFromAPI(ctx)
	}); err != nil {
		slog.Error("falha sync comissoes", "error", err)
	}

	// 7. Proposicoes (Novos projetos ou tramitacoes)
	if err := s.proposicaoSync.SyncFromAPI(ctx); err != nil {
		slog.Error("falha sync proposicoes", "error", err)
	}

	// 7b. Materias novas ou com detalhe antigo (limite por rodada)
	s.syncMaterias(ctx, materia.PorRodadaDiaria)
	// 7b. Gabinete: semanal (guarda por data da ultima carga)
	s.syncGabinete(ctx, time.Now())

	// 8. Invalidar o cache e recalcular o ranking, so com a carga completa.
	// Carga incompleta mantem o ranking em cache (ate o TTL de 24h); quem
	// ficar sem votos aparece como "dados insuficientes", nunca como 0.
	if !s.cargaCompleta() {
		return
	}
	s.rankingService.InvalidateCache()
	if _, err := s.rankingService.CalcularRanking(ctx, nil); err != nil {
		slog.Error("falha ao recalcular ranking", "error", err)
	}

	slog.Info("sync diario integral finalizado")
}

// cargaCompleta confere que todo senador em exercicio tem ao menos um voto no
// recorte (item 4). Um senador sem votos indica carga que falhou para ele.
//
// Na transicao de legislatura (item 14) a checagem nao se aplica: o ranking
// mostra a legislatura encerrada, e os senadores novos podem nao ter votado.
func (s *Scheduler) cargaCompleta() bool {
	if utils.MandatoEncerrado() {
		slog.Info("transicao de legislatura: checagem de completude dispensada")
		return true
	}
	semVotos, err := s.votacaoSync.SenadoresSemVotos()
	if err != nil {
		slog.Error("falha na checagem de completude das votacoes", "error", err)
		return false
	}
	if len(semVotos) > 0 {
		slog.Error("carga incompleta: senadores em exercicio sem votos no recorte; ranking nao recalculado",
			"total", len(semVotos), "senadores", semVotos)
		return false
	}
	return true
}
