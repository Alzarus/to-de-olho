package comissao

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Alzarus/to-de-olho/pkg/retry"

	"github.com/Alzarus/to-de-olho/internal/senador"
	senadoapi "github.com/Alzarus/to-de-olho/pkg/senado"
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

// SyncFromAPI busca as participacoes em comissoes de todos os senadores em
// exercicio. Cada senador tem ate 3 tentativas; se algum falhar em todas,
// devolve erro com os nomes, depois de processar os demais.
func (s *SyncService) SyncFromAPI(ctx context.Context) error {
	slog.Info("iniciando sync de comissoes")

	senadores, err := s.senadorRepo.FindAll(false)
	if err != nil {
		return err
	}

	var total int
	var falhas []string
	for _, sen := range senadores {
		var count int
		err := retry.WithRetry(ctx, 3, "comissoes "+sen.Nome, func() error {
			var err error
			count, err = s.syncSenador(ctx, sen.ID, sen.CodigoParlamentar)
			return err
		})
		if err != nil {
			slog.Error("comissoes do senador nao sincronizadas", "senador", sen.Nome, "error", err)
			falhas = append(falhas, sen.Nome)
			continue
		}
		total += count
	}

	slog.Info("sync de comissoes concluido", "senadores", len(senadores)-len(falhas), "falhas", len(falhas), "participacoes", total)
	if len(falhas) > 0 {
		return fmt.Errorf("comissoes de %d senadores nao sincronizadas: %s", len(falhas), strings.Join(falhas, ", "))
	}
	return nil
}

// SyncSenador busca comissoes de um senador especifico
func (s *SyncService) SyncSenador(ctx context.Context, senadorID int) (int, error) {
	sen, err := s.senadorRepo.FindByID(senadorID)
	if err != nil {
		return 0, err
	}
	return s.syncSenador(ctx, sen.ID, sen.CodigoParlamentar)
}

func (s *SyncService) syncSenador(ctx context.Context, senadorID, codigoParlamentar int) (int, error) {
	comissoesAPI, err := s.client.ListarComissoesParlamentar(ctx, codigoParlamentar)
	if err != nil {
		return 0, err
	}
	participacoes := make([]ComissaoMembro, 0, len(comissoesAPI))
	for _, c := range comissoesAPI {
		participacoes = append(participacoes, s.convertToModel(c, senadorID))
	}
	if err := s.repo.SubstituirDoSenador(senadorID, participacoes); err != nil {
		return 0, fmt.Errorf("falha ao gravar comissoes: %w", err)
	}
	return len(participacoes), nil
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
