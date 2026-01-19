package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/Alzarus/to-de-olho/internal/ceaps"
	"github.com/Alzarus/to-de-olho/internal/emenda"
	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/internal/votacao"
)

// Scheduler gerencia tarefas agendadas
type Scheduler struct {
	senadorSync  *senador.SyncService
	votacaoSync  *votacao.SyncService
	ceapsSync    *ceaps.SyncService
	emendaSync   *emenda.SyncService
}

// NewScheduler cria um novo scheduler
func NewScheduler(
	senadorSync *senador.SyncService,
	votacaoSync *votacao.SyncService,
	ceapsSync *ceaps.SyncService,
	emendaSync *emenda.SyncService,
) *Scheduler {
	return &Scheduler{
		senadorSync:  senadorSync,
		votacaoSync:  votacaoSync,
		ceapsSync:    ceapsSync,
		emendaSync:   emendaSync,
	}
}

// Start inicia o loop de agendamento em background
func (s *Scheduler) Start(ctx context.Context) {
	slog.Info("iniciando scheduler")

	// Tickers para cada tarefa
	// Em producao, usariamos Cloud Scheduler chamando endpoints HTTP,
	// mas este scheduler in-process serve para ambientes self-hosted ou dev.
	
	// Sync Diario (Votacoes, Dados Cadastrais) - 24h
	dailyTicker := time.NewTicker(24 * time.Hour)
	
	// Sync Semanal (CEAPS, Emendas) - 168h
	weeklyTicker := time.NewTicker(168 * time.Hour)

	go func() {
		// Executar imediatamente ao iniciar (opcional, cuidado com startup time)
		// s.runDailySync(ctx)

		for {
			select {
			case <-ctx.Done():
				slog.Info("parando scheduler")
				return
			case <-dailyTicker.C:
				s.runDailySync(ctx)
			case <-weeklyTicker.C:
				s.runWeeklySync(ctx)
			}
		}
	}()
}

func (s *Scheduler) runDailySync(ctx context.Context) {
	slog.Info("executando sync diario")
	
	// 1. Senadores (Base)
	// if err := s.senadorSync.Sync(ctx); err != nil {
	// 	slog.Error("falha sync senadores", "error", err)
	// }

	// 2. Votacoes
	if err := s.votacaoSync.SyncFromAPI(ctx); err != nil {
		slog.Error("falha sync votacoes", "error", err)
	}
	
	slog.Info("sync diario finalizado")
}

func (s *Scheduler) runWeeklySync(ctx context.Context) {
	slog.Info("executando sync semanal")

	// 1. CEAPS (Ano atual)
	anoAtual := time.Now().Year()
	if err := s.ceapsSync.SyncFromAPI(ctx, anoAtual); err != nil {
		slog.Error("falha sync ceaps", "error", err)
	}

	// 2. Emendas
	if err := s.emendaSync.SyncAll(ctx, anoAtual); err != nil {
		slog.Error("falha sync emendas", "error", err)
	}

	slog.Info("sync semanal finalizado")
}
