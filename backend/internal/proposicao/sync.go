package proposicao

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/pedroalmeida/to-de-olho/internal/senador"
	senadoapi "github.com/pedroalmeida/to-de-olho/pkg/senado"
)

// SyncService gerencia sincronizacao de proposicoes
type SyncService struct {
	repo        *Repository
	senadorRepo *senador.Repository
	client      *senadoapi.LegisClient
}

// NewSyncService cria um novo servico de sincronizacao
func NewSyncService(repo *Repository, senadorRepo *senador.Repository, client *senadoapi.LegisClient) *SyncService {
	return &SyncService{
		repo:        repo,
		senadorRepo: senadorRepo,
		client:      client,
	}
}

// SyncFromAPI busca proposicoes da API para todos os senadores
func (s *SyncService) SyncFromAPI(ctx context.Context) error {
	slog.Info("iniciando sync de proposicoes")

	// Buscar todos os senadores
	senadores, err := s.senadorRepo.FindAll()
	if err != nil {
		return err
	}

	var totalProposicoes, totalSenadores int

	for _, sen := range senadores {
		count, err := s.SyncSenador(ctx, sen.ID)
		if err != nil {
			slog.Warn("falha ao buscar proposicoes", "senador", sen.Nome, "error", err)
			continue
		}

		totalProposicoes += count
		totalSenadores++
		slog.Debug("proposicoes sincronizadas", "senador", sen.Nome, "count", count)
	}

	slog.Info("sync de proposicoes concluido", "senadores", totalSenadores, "proposicoes", totalProposicoes)
	return nil
}

// SyncSenador busca proposicoes de um senador especifico
func (s *SyncService) SyncSenador(ctx context.Context, senadorID int) (int, error) {
	sen, err := s.senadorRepo.FindByID(senadorID)
	if err != nil {
		return 0, err
	}

	proposicoesAPI, err := s.client.ListarProposicoesParlamentar(ctx, sen.CodigoParlamentar)
	if err != nil {
		return 0, err
	}

	// Limpar proposicoes antigas antes de re-sincronizar
	if err := s.repo.DeleteBySenadorID(senadorID); err != nil {
		slog.Warn("falha ao limpar proposicoes antigas", "senador", senadorID, "error", err)
	}

	var count int
	for _, p := range proposicoesAPI {
		proposicao := s.convertToModel(p, senadorID)
		
		// Calcular pontuacao
		proposicao.Pontuacao = proposicao.CalcularPontuacao()
		
		if err := s.repo.Upsert(&proposicao); err != nil {
			slog.Warn("falha ao salvar proposicao", "senador", senadorID, "error", err)
			continue
		}
		count++
	}

	return count, nil
}

// convertToModel converte uma proposicao da API para modelo interno
func (s *SyncService) convertToModel(api senadoapi.MateriaAPI, senadorID int) Proposicao {
	var dataApresentacao *time.Time
	if api.DadosBasicosMateria.DataApresentacao != "" {
		if t, err := time.Parse("2006-01-02", api.DadosBasicosMateria.DataApresentacao); err == nil {
			dataApresentacao = &t
		}
	}

	// Determinar estagio de tramitacao baseado na situacao
	estagio := determinarEstagio(api.SituacaoAtual.Autuacoes.Autuacao.Situacao.DescricaoSituacao)

	ano := 0
	if api.IdentificacaoMateria.AnoMateria != "" {
		ano, _ = strconv.Atoi(api.IdentificacaoMateria.AnoMateria)
	}

	return Proposicao{
		SenadorID:              senadorID,
		CodigoMateria:          api.IdentificacaoMateria.CodigoMateria,
		SiglaSubtipoMateria:    api.IdentificacaoMateria.SiglaSubtipoMateria,
		NumeroMateria:          api.IdentificacaoMateria.NumeroMateria,
		AnoMateria:             ano,
		DescricaoIdentificacao: api.IdentificacaoMateria.DescricaoIdentificacao,
		Ementa:                 api.DadosBasicosMateria.EmentaMateria,
		SituacaoAtual:          api.SituacaoAtual.Autuacoes.Autuacao.Situacao.DescricaoSituacao,
		DataApresentacao:       dataApresentacao,
		EstagioTramitacao:      estagio,
	}
}

// determinarEstagio classifica o estagio de tramitacao baseado na descricao
func determinarEstagio(situacao string) string {
	situacaoLower := strings.ToLower(situacao)

	// Transformado em lei (norma juridica)
	if strings.Contains(situacaoLower, "transformada em norma") ||
		strings.Contains(situacaoLower, "transformado em lei") ||
		strings.Contains(situacaoLower, "lei publicada") {
		return "TransformadoLei"
	}

	// Aprovado em plenario
	if strings.Contains(situacaoLower, "aprovado plenario") ||
		strings.Contains(situacaoLower, "aprovada pelo plenario") ||
		strings.Contains(situacaoLower, "remetida a camara") ||
		strings.Contains(situacaoLower, "sancionada") {
		return "AprovadoPlenario"
	}

	// Aprovado em comissao
	if strings.Contains(situacaoLower, "aprovado comissao") ||
		strings.Contains(situacaoLower, "aprovada pela comissao") ||
		strings.Contains(situacaoLower, "pronto para pauta") {
		return "AprovadoComissao"
	}

	// Em comissao
	if strings.Contains(situacaoLower, "comissao") ||
		strings.Contains(situacaoLower, "relator") ||
		strings.Contains(situacaoLower, "tramit") {
		return "EmComissao"
	}

	// Default: Apresentado
	return "Apresentado"
}
