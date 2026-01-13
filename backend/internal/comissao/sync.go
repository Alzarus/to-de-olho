package comissao

import (
	"context"
	"log/slog"
	"time"

	"github.com/pedroalmeida/to-de-olho/internal/senador"
	senadoapi "github.com/pedroalmeida/to-de-olho/pkg/senado"
)

// SyncService gerencia sincronizacao de comissoes
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

// SyncFromAPI busca comissoes da API para todos os senadores
func (s *SyncService) SyncFromAPI(ctx context.Context) error {
	slog.Info("iniciando sync de comissoes")

	// Buscar todos os senadores
	senadores, err := s.senadorRepo.FindAll()
	if err != nil {
		return err
	}

	var totalComissoes, totalSenadores int

	for _, sen := range senadores {
		count, err := s.SyncSenador(ctx, sen.ID)
		if err != nil {
			slog.Warn("falha ao buscar comissoes", "senador", sen.Nome, "error", err)
			continue
		}

		totalComissoes += count
		totalSenadores++
		slog.Debug("comissoes sincronizadas", "senador", sen.Nome, "count", count)
	}

	slog.Info("sync de comissoes concluido", "senadores", totalSenadores, "comissoes", totalComissoes)
	return nil
}

// SyncSenador busca comissoes de um senador especifico
func (s *SyncService) SyncSenador(ctx context.Context, senadorID int) (int, error) {
	sen, err := s.senadorRepo.FindByID(senadorID)
	if err != nil {
		return 0, err
	}

	comissoesAPI, err := s.client.ListarComissoesParlamentar(ctx, sen.CodigoParlamentar)
	if err != nil {
		return 0, err
	}

	// Limpar comissoes antigas antes de re-sincronizar
	if err := s.repo.DeleteBySenadorID(senadorID); err != nil {
		slog.Warn("falha ao limpar comissoes antigas", "senador", senadorID, "error", err)
	}

	var count int
	for _, c := range comissoesAPI {
		comissao := s.convertToModel(c, senadorID)
		if err := s.repo.Upsert(&comissao); err != nil {
			slog.Warn("falha ao salvar comissao", "senador", senadorID, "error", err)
			continue
		}
		count++
	}

	return count, nil
}

// convertToModel converte uma comissao da API para modelo interno
func (s *SyncService) convertToModel(api senadoapi.ComissaoAPI, senadorID int) ComissaoMembro {
	var dataInicio, dataFim *time.Time

	if api.DataInicio != "" {
		if t, err := time.Parse("2006-01-02", api.DataInicio); err == nil {
			dataInicio = &t
		}
	}

	if api.DataFim != "" {
		if t, err := time.Parse("2006-01-02", api.DataFim); err == nil {
			dataFim = &t
		}
	}

	return ComissaoMembro{
		SenadorID:             senadorID,
		CodigoComissao:        api.IdentificacaoComissao.CodigoComissao,
		SiglaComissao:         api.IdentificacaoComissao.SiglaComissao,
		NomeComissao:          api.IdentificacaoComissao.NomeComissao,
		SiglaCasaComissao:     api.IdentificacaoComissao.SiglaCasaComissao,
		DescricaoParticipacao: api.DescricaoParticipacao,
		DataInicio:            dataInicio,
		DataFim:               dataFim,
	}
}
