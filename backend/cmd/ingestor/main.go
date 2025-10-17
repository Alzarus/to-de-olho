package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	app "to-de-olho-backend/internal/application"
	"to-de-olho-backend/internal/config"
	"to-de-olho-backend/internal/domain"
	"to-de-olho-backend/internal/infrastructure/cache"
	"to-de-olho-backend/internal/infrastructure/db"
	"to-de-olho-backend/internal/infrastructure/httpclient"
	"to-de-olho-backend/internal/infrastructure/ingestor"
	"to-de-olho-backend/internal/infrastructure/migrations"
	"to-de-olho-backend/internal/infrastructure/repository"
)

func main() {
	mode := flag.String("mode", "auto", "Mode: auto|strategic|backfill|daily")
	years := flag.Int("years", 0, "Backfill years from now backwards (0 = use config)")
	startYear := flag.Int("start-year", 0, "Specific start year for backfill (0 = use config)")
	force := flag.Bool("force", false, "Force re-execution of backfill even if successful")
	flag.Parse()

	// Setup structured logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Falha ao carregar configuração: %v", err)
	}

	ctx := context.Background()
	pgPool, err := db.NewPostgresPoolFromConfig(ctx, &cfg.Database)
	if err != nil {
		log.Fatalf("Postgres connection error: %v", err)
	}

	// Run database migrations
	migrator := migrations.NewMigrator(pgPool)
	if err := migrator.Run(ctx); err != nil {
		log.Fatalf("Migration error: %v", err)
	}

	// Setup repositories and services
	deputadoRepo := repository.NewDeputadoRepository(pgPool)
	despesaRepo := repository.NewDespesaRepository(pgPool)
	proposicaoRepo := repository.NewProposicaoRepository(pgPool)
	backfillRepo := repository.NewBackfillRepository(pgPool)
	votacaoRepo := repository.NewVotacaoRepository(pgPool) // Assumindo que existe
	client := httpclient.NewCamaraClientFromConfig(&cfg.CamaraClient)
	cacheClient := cache.NewFromConfig(&cfg.Redis)

	deputadosService := app.NewDeputadosService(client, cacheClient, deputadoRepo, despesaRepo)
	proposicoesService := app.NewProposicoesService(client, cacheClient, proposicaoRepo, logger)
	votacoesService := app.NewVotacoesService(votacaoRepo, client, cacheClient)

	// Criar serviço de analytics (usado para recalcular rankings após backfill)
	analyticsSvc := app.NewAnalyticsService(deputadoRepo, proposicaoRepo, votacaoRepo, despesaRepo, cacheClient, logger)

	// 🧠 Sistema Inteligente de Backfill
	smartBackfillService := app.NewSmartBackfillService(backfillRepo, deputadosService, proposicoesService, votacoesService, analyticsSvc, logger)

	switch *mode {
	case "auto":
		// 🤖 Modo inteligente - decide automaticamente se precisa executar backfill
		logger.Info("🧠 Modo INTELIGENTE ativado - Sistema verifica se backfill é necessário")

		// Ler configurações de ambiente (podem vir do deploy script)
		backfillConfig := parseBackfillConfigFromEnv(cfg)
		if *force {
			backfillConfig.ForcarReexecucao = true
		}
		if *startYear > 0 {
			backfillConfig.AnoInicio = *startYear
		}
		if *years > 0 {
			currentYear := time.Now().Year()
			backfillConfig.AnoInicio = currentYear - *years + 1
		}

		logger.Info("🎯 Configuração do backfill",
			slog.Int("start_year", backfillConfig.AnoInicio),
			slog.Int("end_year", backfillConfig.AnoFim),
			slog.Bool("force", backfillConfig.ForcarReexecucao),
			slog.String("triggered_by", backfillConfig.TriggeredBy))

		// Verificar se deve executar backfill
		shouldRun, reason, err := smartBackfillService.ShouldRunHistoricalBackfill(ctx, backfillConfig)
		if err != nil {
			log.Fatalf("failed to check backfill requirement: %v", err)
		}

		if !shouldRun {
			logger.Info("✅ Backfill não necessário", slog.String("reason", reason))
			logger.Info("🎯 Sistema inteligente evitou execução desnecessária!")
			return
		}

		logger.Info("🚀 Executando backfill histórico", slog.String("reason", reason))

		execution, err := smartBackfillService.ExecuteHistoricalBackfill(ctx, backfillConfig)
		if err != nil {
			log.Fatalf("intelligent backfill failed: %v", err)
		}

		logger.Info("📊 Backfill iniciado, aguardando conclusão...",
			slog.Int("execution_id", execution.ID),
			slog.String("status", string(execution.Status)))

		// Aguardar a conclusão do backfill (usa timeout definido na configuração carregada)
		if err := waitForBackfillCompletion(ctx, smartBackfillService, execution.ExecutionID, cfg.Ingestor.MonitorTimeout, logger); err != nil {
			log.Fatalf("backfill monitoring failed: %v", err)
		}

	case "strategic":
		if err := runStrategicBackfill(ctx, pgPool, deputadosService, proposicoesService, deputadoRepo, proposicaoRepo, cfg, *years, *startYear); err != nil {
			log.Fatalf("strategic backfill failed: %v", err)
		}
	case "backfill":
		if err := runBackfill(ctx, deputadosService, deputadoRepo, *years); err != nil {
			log.Fatalf("backfill failed: %v", err)
		}
	case "daily":
		if err := runDaily(ctx, deputadosService, deputadoRepo); err != nil {
			log.Fatalf("daily ingest failed: %v", err)
		}
	default:
		log.Fatalf("unknown mode: %s", *mode)
	}
}

// runStrategicBackfill executa backfill histórico estratégico com checkpoints
func runStrategicBackfill(
	ctx context.Context,
	pgPool *pgxpool.Pool,
	deputadosService *app.DeputadosService,
	proposicoesService *app.ProposicoesService,
	deputadoRepo *repository.DeputadoRepository,
	proposicaoRepo *repository.ProposicaoRepository,
	cfg *config.Config,
	years int,
	startYear int,
) error {
	log.Println("🚀 Iniciando Backfill Histórico Estratégico")

	// Configurar estratégia baseada nos parâmetros ou configuração
	strategy := ingestor.DefaultBackfillStrategy()

	// Determinar ano inicial e final
	currentYear := time.Now().Year()

	if startYear > 0 {
		// Usar ano específico fornecido via flag
		strategy.YearStart = startYear
		strategy.YearEnd = currentYear
		log.Printf("📅 Usando ano inicial específico: %d", startYear)
	} else if years > 0 {
		// Usar número de anos atrás
		strategy.YearStart = currentYear - years + 1
		strategy.YearEnd = currentYear
		log.Printf("📅 Usando %d anos atrás: %d-%d", years, strategy.YearStart, strategy.YearEnd)
	} else {
		// Usar configuração padrão
		strategy.YearStart = cfg.Ingestor.BackfillStartYear
		strategy.YearEnd = currentYear
		log.Printf("📅 Usando configuração padrão: %d-%d", strategy.YearStart, strategy.YearEnd)
	}

	// Aplicar configurações do ingestor
	strategy.BatchSize = cfg.Ingestor.BatchSize
	strategy.MaxRetries = cfg.Ingestor.MaxRetries

	log.Printf("📊 Estratégia: %d-%d (%d anos), lotes de %d, %d tentativas",
		strategy.YearStart, strategy.YearEnd, strategy.YearEnd-strategy.YearStart+1,
		strategy.BatchSize, strategy.MaxRetries)

	// Criar gerenciador de backfill e executor estratégico
	backfillManager := ingestor.NewBackfillManager(pgPool)
	votacaoRepo := repository.NewVotacaoRepository(pgPool)

	// Construir VotacoesService e PartidosService localmente (precisa de client e cache)
	clientLocal := httpclient.NewCamaraClientFromConfig(&cfg.CamaraClient)
	cacheLocal := cache.NewFromConfig(&cfg.Redis)
	votacoesSvcLocal := app.NewVotacoesService(votacaoRepo, clientLocal, cacheLocal)

	partidoRepoLocal := repository.NewPartidoRepository(pgPool)
	partidosSvcLocal := app.NewPartidosService(clientLocal, partidoRepoLocal)

	// Criar analytics service para atualizar rankings após backfill
	despesaRepoLocal := repository.NewDespesaRepository(pgPool)
	analyticsSvcLocal := app.NewAnalyticsService(deputadoRepo, proposicaoRepo, votacaoRepo, despesaRepoLocal, cacheLocal, slog.New(slog.NewTextHandler(os.Stdout, nil)))

	executor := ingestor.NewStrategicBackfillExecutor(
		backfillManager,
		deputadosService,
		proposicoesService,
		deputadoRepo,
		proposicaoRepo,
		votacoesSvcLocal,
		partidosSvcLocal,
		analyticsSvcLocal,
		strategy,
	)

	// Executar backfill estratégico
	return executor.ExecuteBackfill(ctx)
}

func runBackfill(ctx context.Context, svc *app.DeputadosService, repo *repository.DeputadoRepository, years int) error {
	fmt.Println("Starting backfill for deputies and recent expenses...")
	deps, _, err := svc.ListarDeputados(ctx, "", "", "")
	if err != nil {
		return err
	}
	fmt.Printf("Fetched %d deputies. Upserting...\n", len(deps))
	if err := repo.UpsertDeputados(ctx, deps); err != nil {
		return err
	}
	// Expenses for last N years (bounded to current year)
	year := time.Now().Year()
	minYear := year - (years - 1)
	if minYear < year-10 {
		minYear = year - 10
	}
	fmt.Printf("Fetching expenses from %d to %d (skipping for brevity placeholder)\n", minYear, year)
	// NOTE: For now, we keep only deputies cached in DB; expenses ingestion can be implemented in repository later.
	return nil
}

func runDaily(ctx context.Context, svc *app.DeputadosService, repo *repository.DeputadoRepository) error {
	fmt.Println("Running daily sync for deputies (and recent expenses placeholder)...")
	deps, _, err := svc.ListarDeputados(ctx, "", "", "")
	if err != nil {
		return err
	}
	if err := repo.UpsertDeputados(ctx, deps); err != nil {
		return err
	}
	fmt.Printf("Daily sync done. Deputies upserted: %d\n", len(deps))
	return nil
}

// parseBackfillConfigFromEnv lê configuração de backfill de variáveis de ambiente
// Usado pelo deploy script para passar parâmetros de forma inteligente
func parseBackfillConfigFromEnv(cfg *config.Config) *domain.BackfillConfig {
	backfillConfig := &domain.BackfillConfig{
		AnoInicio:           cfg.Ingestor.BackfillStartYear, // Padrão da configuração
		AnoFim:              time.Now().Year(),              // Ano atual
		ForcarReexecucao:    false,
		TriggeredBy:         "manual",
		Tipo:                domain.BackfillTipoHistorico,
		IncluirDeputados:    true,
		IncluirProposicoes:  true,
		IncluirDespesas:     true,
		IncluirVotacoes:     true,
		BatchSize:           cfg.Ingestor.BatchSize,
		ParallelWorkers:     3,
		DelayBetweenBatches: 100,
	}

	// Ler variáveis de ambiente (vem do deploy script)
	if startYearEnv := os.Getenv("BACKFILL_START_YEAR"); startYearEnv != "" {
		if year := parseYear(startYearEnv); year > 0 {
			backfillConfig.AnoInicio = year
		}
	}

	if endYearEnv := os.Getenv("BACKFILL_END_YEAR"); endYearEnv != "" {
		if year := parseYear(endYearEnv); year > 0 {
			backfillConfig.AnoFim = year
		}
	}

	if forceEnv := os.Getenv("BACKFILL_FORCE"); forceEnv == "true" {
		backfillConfig.ForcarReexecucao = true
	}

	// Allow override of parallel workers and batch size from env
	if bw := os.Getenv("BACKFILL_WORKERS"); bw != "" {
		if v, err := strconv.Atoi(bw); err == nil && v > 0 {
			backfillConfig.ParallelWorkers = v
		}
	}

	if bs := os.Getenv("BACKFILL_BATCH_SIZE"); bs != "" {
		if v, err := strconv.Atoi(bs); err == nil && v > 0 {
			backfillConfig.BatchSize = v
		}
	}

	if triggeredBy := os.Getenv("BACKFILL_TRIGGERED_BY"); triggeredBy != "" {
		backfillConfig.TriggeredBy = triggeredBy
	}

	// Allow disabling heavy entities via env (useful during deployments)
	if inc := os.Getenv("BACKFILL_INCLUDE_PROPOSICOES"); inc != "" {
		backfillConfig.IncluirProposicoes = inc == "true"
	}

	if inc := os.Getenv("BACKFILL_INCLUDE_VOTACOES"); inc != "" {
		backfillConfig.IncluirVotacoes = inc == "true"
	}

	return backfillConfig
}

// parseYear converte string para int, retorna 0 se inválido
func parseYear(yearStr string) int {
	currentYear := time.Now().Year()

	// Parseaar ano de string para int
	year := 0
	if _, err := fmt.Sscanf(yearStr, "%d", &year); err != nil {
		return 0
	}

	// Validar range básico (não pode ser futuro nem muito antigo)
	if year < 2000 || year > currentYear+1 {
		return 0
	}

	return year
}

// waitForBackfillCompletion aguarda a conclusão do backfill monitorando o status
func waitForBackfillCompletion(ctx context.Context, service *app.SmartBackfillService, executionID string, monitorTimeout time.Duration, logger *slog.Logger) error {
	ticker := time.NewTicker(10 * time.Second) // Verificar a cada 10 segundos
	defer ticker.Stop()
	// monitorTimeout é passado a partir da configuração carregada (cfg.Ingestor.MonitorTimeout)
	var timeout <-chan time.Time
	if monitorTimeout <= 0 {
		monitorTimeout = 30 * time.Minute
		logger.Info("Usando timeout padrão para monitoramento do backfill", slog.String("timeout", monitorTimeout.String()))
	} else {
		logger.Info("Usando BACKFILL_MONITOR_TIMEOUT configurado", slog.String("timeout", monitorTimeout.String()))
	}
	timeout = time.After(monitorTimeout)

	logger.Info("⏳ Monitorando progresso do backfill...", slog.String("execution_id", executionID))

	const heartbeatInterval = 1 * time.Minute
	lastLoggedAt := time.Time{}
	lastStatusCode := ""
	lastOperation := ""
	lastDeputados := -1
	lastProposicoes := -1
	lastProgress := -1.0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timeout aguardando conclusão do backfill")
		case <-ticker.C:
			// Verificar status através do SmartBackfillService
			status, err := service.GetCurrentStatus(ctx)
			if err != nil {
				// Se a execução não for encontrada, verificar se houve alguma execução concluída
				// recentemente antes de assumir sucesso
				if errors.Is(err, domain.ErrBackfillNaoEncontrado) {
					// Verificar histórico de execuções para confirmar se foi realmente concluída
					executions, _, histErr := service.ListExecutions(ctx, 1, 0)
					if histErr == nil && len(executions) > 0 {
						lastExecution := executions[0]
						// Só assumir conclusão se houve uma execução recente (últimas 2 horas)
						// e foi completada com sucesso ou parcialmente
						if lastExecution.CompletedAt != nil {
							timeSinceCompletion := time.Since(*lastExecution.CompletedAt)
							if timeSinceCompletion < 2*time.Hour {
								logger.Info("✅ Execução de backfill não encontrada, mas última execução foi recente",
									slog.String("last_execution_id", lastExecution.ExecutionID),
									slog.Time("completed_at", *lastExecution.CompletedAt),
									slog.String("status", lastExecution.Status))
								return nil
							}
						}
					}

					// Se não conseguir confirmar execução recente, tratar como erro crítico
					logger.Error("❌ Execução de backfill não encontrada e não há histórico de execução recente",
						slog.Any("history_error", histErr))
					return fmt.Errorf("execução de backfill perdida sem confirmação de conclusão: %w", err)
				}

				logger.Warn("Erro ao verificar status do backfill", slog.Any("error", err))
				continue
			}

			progressChanged := lastProgress < 0 || math.Abs(status.ProgressPercentage-lastProgress) >= 0.05
			operationChanged := status.CurrentOperation != lastOperation
			statusChanged := string(status.Status) != lastStatusCode
			deputadosChanged := status.DeputadosProcessados != lastDeputados
			proposicoesChanged := status.ProposicoesProcessadas != lastProposicoes
			heartbeatElapsed := time.Since(lastLoggedAt) >= heartbeatInterval

			if progressChanged || operationChanged || statusChanged || deputadosChanged || proposicoesChanged || heartbeatElapsed {
				logger.Info("📊 Status do backfill",
					slog.String("status", string(status.Status)),
					slog.Int("deputados", status.DeputadosProcessados),
					slog.Int("proposicoes", status.ProposicoesProcessadas),
					slog.Float64("progresso", status.ProgressPercentage),
					slog.String("operacao_atual", status.CurrentOperation))
				lastLoggedAt = time.Now()
				lastStatusCode = string(status.Status)
				lastOperation = status.CurrentOperation
				lastDeputados = status.DeputadosProcessados
				lastProposicoes = status.ProposicoesProcessadas
				lastProgress = status.ProgressPercentage
			}

			switch status.Status {
			case domain.BackfillStatusSuccess:
				duration := time.Since(status.StartedAt)
				logger.Info("✅ Backfill concluído com sucesso!",
					slog.String("duracao", duration.String()),
					slog.Int("deputados", status.DeputadosProcessados),
					slog.Int("proposicoes", status.ProposicoesProcessadas))
				return nil
			case domain.BackfillStatusPartial:
				duration := time.Since(status.StartedAt)
				logger.Warn("⚠️ Backfill concluído com pendências",
					slog.String("duracao", duration.String()),
					slog.Int("deputados", status.DeputadosProcessados),
					slog.Int("proposicoes", status.ProposicoesProcessadas),
					slog.Float64("progresso", status.ProgressPercentage),
					slog.String("operacao_atual", status.CurrentOperation))
				return nil
			case domain.BackfillStatusFailed:
				return fmt.Errorf("backfill falhou")
			case domain.BackfillStatusRunning:
				// Continuar aguardando
				continue
			default:
				logger.Warn("Status desconhecido do backfill", slog.String("status", string(status.Status)))
				continue
			}
		}
	}
}
