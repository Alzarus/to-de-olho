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

// Start inicia o loop de agendamento em background e verifica necessidade de backfill
func (s *Scheduler) Start(ctx context.Context) {
	slog.Info("iniciando scheduler")

	// Verificar se precisa de backfill inicial em goroutine separada
	go s.runStartupSync(ctx)

	// Tickers para cada tarefa
	// Sync Diario (Votacoes, Dados Cadastrais) - 24h
	dailyTicker := time.NewTicker(24 * time.Hour)

	// Sync Semanal (CEAPS, Emendas) - 168h


	go func() {
		for {
			select {
			case <-ctx.Done():
				slog.Info("parando scheduler")
				return
			case <-dailyTicker.C:
				s.runDailySync(ctx)
			}
		}
	}()
}

// runStartupSync verifica se o DB esta vazio e roda backfill completo
func (s *Scheduler) runStartupSync(ctx context.Context) {
	// Verificar se foi solicitado FORCE_BACKFILL
	forceBackfill := os.Getenv("FORCE_BACKFILL") == "true"
	
	// 1. Verificar se ja existem dados
	count, err := s.senadorRepo.Count()
	if err != nil {
		slog.Error("falha ao verificar contagem de senadores", "error", err)
		return
	}

	if count > 0 && !forceBackfill {
		slog.Info("banco de dados ja populado, pulando backfill inicial", "senadores", count)
		// Opcional: Ainda rodar um calculo de ranking para garantir cache quente
		// s.rankingService.CalcularRanking(ctx, nil)
		return
	}
	
	if forceBackfill {
		slog.Info("FORCE_BACKFILL=true detectado, executando backfill mesmo com dados existentes", "senadores_existentes", count)
	}

	slog.Info("banco de dados vazio detectado. INICIANDO BACKFILL COMPLETO...")

	// 2. Determinar ano de inicio
	anoInicio := 2023
	if envAno := os.Getenv("INICIO_BACKFILL"); envAno != "" {
		if parsed, err := strconv.Atoi(envAno); err == nil {
			anoInicio = parsed
		}
	}
	anoAtual := time.Now().Year()

	slog.Info("configuracao de backfill", "ano_inicio", anoInicio, "ano_fim", anoAtual)

	// 3. Sequencia de Sync
	
	// A. Dados Basicos (Senadores)
	slog.Info("--- PASSO 1/6: SENADORES ---")
	if err := s.senadorSync.SyncFromAPI(ctx); err != nil {
		slog.Error("falha critica no backfill de senadores", "error", err)
		return // Sem senadores nao da pra continuar
	}

	// B. Votacoes (Captura sessoes de todos os anos disponiveis na API)
	slog.Info("--- PASSO 2/6: VOTACOES (LISTA) ---")
	if err := s.votacaoSync.SyncFromAPI(ctx); err != nil {
		slog.Error("falha no backfill de votacoes", "error", err)
	}

	// C. Loop por ano para dados periodicos
	for ano := anoInicio; ano <= anoAtual; ano++ {
		slog.Info("--- PROCESSANDO ANO ---", "ano", ano)

		// Metadata de Votacoes (Ementas, Datas corretas)
		if err := s.votacaoSync.SyncMetadata(ctx, ano); err != nil {
			slog.Error("falha ao sincronizar metadata votacoes", "ano", ano, "error", err)
		}

		// CEAPS (Despesas)
		if err := s.ceapsSync.SyncFromAPI(ctx, ano); err != nil {
			slog.Error("falha ao sincronizar ceaps", "ano", ano, "error", err)
		}

		// Emendas
		if err := s.emendaSync.SyncAll(ctx, ano); err != nil {
			slog.Error("falha ao sincronizar emendas", "ano", ano, "error", err)
		}
	}

	// D. Comissoes (Estado atual/recente)
	slog.Info("--- PASSO 4/6: COMISSOES ---")
	if err := s.comissaoSync.SyncFromAPI(ctx); err != nil {
		slog.Error("falha no backfill de comissoes", "error", err)
	}

	// E. Proposicoes (Historico)
	slog.Info("--- PASSO 5/6: PROPOSICOES ---")
	if err := s.proposicaoSync.SyncFromAPI(ctx); err != nil {
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

func (s *Scheduler) runDailySync(ctx context.Context) {
	slog.Info("executando sync diario integral")

	anoAtual := time.Now().Year()

	// 1. Senadores (Atualizacao cadastral)
	if err := s.senadorSync.SyncFromAPI(ctx); err != nil {
		slog.Error("falha sync senadores", "error", err)
	}

	// 2. Votacoes (Novas sessoes)
	if err := s.votacaoSync.SyncFromAPI(ctx); err != nil {
		slog.Error("falha sync votacoes", "error", err)
	}

	// 3. Metadata do ano atual (para pegar ementas de votacoes recentes)
	if err := s.votacaoSync.SyncMetadata(ctx, anoAtual); err != nil {
		slog.Error("falha sync metadata votacoes", "error", err)
	}

	// 4. CEAPS (Despesas) - MOVIDO DE SEMANAL PARA DIARIO
	if err := s.ceapsSync.SyncFromAPI(ctx, anoAtual); err != nil {
		slog.Error("falha sync ceaps", "error", err)
	}

	// 5. Emendas - MOVIDO DE SEMANAL PARA DIARIO
	if err := s.emendaSync.SyncAll(ctx, anoAtual); err != nil {
		slog.Error("falha sync emendas", "error", err)
	}

	// 6. Comissoes (Mudancas de membros) - MOVIDO DE SEMANAL PARA DIARIO
	if err := s.comissaoSync.SyncFromAPI(ctx); err != nil {
		slog.Error("falha sync comissoes", "error", err)
	}

	// 7. Proposicoes (Novos projetos ou tramitacoes)
	if err := s.proposicaoSync.SyncFromAPI(ctx); err != nil {
		slog.Error("falha sync proposicoes", "error", err)
	}

	// 8. Recalcular Ranking
	s.rankingService.CalcularRanking(ctx, nil)

	slog.Info("sync diario integral finalizado")
}
