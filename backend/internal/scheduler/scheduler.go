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
	"github.com/Alzarus/to-de-olho/internal/proposicao"
	"github.com/Alzarus/to-de-olho/internal/ranking"
	"github.com/Alzarus/to-de-olho/internal/senador"
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
	senadorRepo    *senador.Repository // Para checar se banco esta vazio
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
	}
}

// Start inicia o loop de agendamento em background.
// O backfill nao e mais executado na startup; deve ser disparado
// via HTTP (POST /api/v1/sync/backfill) para que o Cloud Run
// mantenha o container vivo durante a execucao.
func (s *Scheduler) Start(ctx context.Context) {
	slog.Info("iniciando scheduler")

	dailyTicker := time.NewTicker(24 * time.Hour)

	go func() {
		for {
			select {
			case <-ctx.Done():
				slog.Info("parando scheduler")
				return
			case <-dailyTicker.C:
				s.RunDailySync(ctx)
			}
		}
	}()
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

	// B. Votacoes (Captura sessoes de todos os anos disponiveis na API)
	slog.Info("--- PASSO 2/6: VOTACOES (LISTA) ---")
	if err := retry.WithRetry(ctx, 3, "backfill-votacoes", func() error {
		return s.votacaoSync.SyncFromAPI(ctx)
	}); err != nil {
		slog.Error("falha no backfill de votacoes", "error", err)
	}

	// C. Loop por ano para dados periodicos
	for ano := anoInicio; ano <= anoAtual; ano++ {
		slog.Info("--- PROCESSANDO ANO ---", "ano", ano)

		// Metadata de Votacoes (Ementas, Datas corretas)
		anoLoop := ano
		if err := retry.WithRetry(ctx, 3, "backfill-votacoes-metadata", func() error {
			return s.votacaoSync.SyncMetadata(ctx, anoLoop)
		}); err != nil {
			slog.Error("falha ao sincronizar metadata votacoes", "ano", ano, "error", err)
		}

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
	if err := retry.WithRetry(ctx, 3, "backfill-proposicoes", func() error {
		return s.proposicaoSync.SyncFromAPI(ctx)
	}); err != nil {
		slog.Error("falha no backfill de proposicoes", "error", err)
	}

	// F. Calculo de Ranking Final
	slog.Info("--- PASSO 6/6: CALCULANDO RANKING ---")
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

	// 2. Votacoes (Novas sessoes)
	if err := retry.WithRetry(ctx, 3, "sync-votacoes", func() error {
		return s.votacaoSync.SyncFromAPI(ctx)
	}); err != nil {
		slog.Error("falha sync votacoes", "error", err)
	}

	// 3. Metadata do ano atual (para pegar ementas de votacoes recentes)
	if err := retry.WithRetry(ctx, 3, "sync-votacoes-metadata", func() error {
		return s.votacaoSync.SyncMetadata(ctx, anoAtual)
	}); err != nil {
		slog.Error("falha sync metadata votacoes", "error", err)
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
	if err := retry.WithRetry(ctx, 3, "sync-proposicoes", func() error {
		return s.proposicaoSync.SyncFromAPI(ctx)
	}); err != nil {
		slog.Error("falha sync proposicoes", "error", err)
	}

	// 8. Recalcular Ranking
	s.rankingService.CalcularRanking(ctx, nil)

	slog.Info("sync diario integral finalizado")
}

