package votacao

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/pedroalmeida/to-de-olho/internal/senador"
	senadoapi "github.com/pedroalmeida/to-de-olho/pkg/senado"
)

// SyncService gerencia sincronizacao de votacoes
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

// SyncFromAPI busca votacoes da API para todos os senadores
func (s *SyncService) SyncFromAPI(ctx context.Context) error {
	slog.Info("iniciando sync de votacoes")

	// Buscar todos os senadores
	senadores, err := s.senadorRepo.FindAll()
	if err != nil {
		return err
	}

	var totalVotacoes, totalSenadores int

	for _, sen := range senadores {
		// A API retorna sessoes de votacao
		sessoes, err := s.client.ListarVotacoesParlamentar(ctx, sen.CodigoParlamentar)
		if err != nil {
			slog.Warn("falha ao buscar votacoes", "senador", sen.Nome, "error", err)
			continue
		}

		// Processar cada sessao
		for _, sessao := range sessoes {
			// Encontrar o voto do senador nesta sessao
			var siglaVoto string
			for _, voto := range sessao.Votos {
				if voto.CodigoParlamentar == sen.CodigoParlamentar {
					siglaVoto = voto.SiglaVoto
					break
				}
			}

			if siglaVoto == "" {
				continue // Senador nao votou nesta sessao
			}

			votacao := s.convertSessaoToVotacao(sessao, sen.ID, siglaVoto)
			if err := s.repo.Upsert(&votacao); err != nil {
				slog.Warn("falha ao salvar votacao", "senador", sen.ID, "error", err)
				continue
			}
			totalVotacoes++
		}

		totalSenadores++
		slog.Debug("votacoes sincronizadas", "senador", sen.Nome, "sessoes", len(sessoes))
	}

	slog.Info("sync de votacoes concluido", "senadores", totalSenadores, "votacoes", totalVotacoes)
	return nil
}

// SyncSenador busca votacoes de um senador especifico
func (s *SyncService) SyncSenador(ctx context.Context, senadorID int) (int, error) {
	sen, err := s.senadorRepo.FindByID(senadorID)
	if err != nil {
		return 0, err
	}

	sessoes, err := s.client.ListarVotacoesParlamentar(ctx, sen.CodigoParlamentar)
	if err != nil {
		return 0, err
	}

	var count int
	for _, sessao := range sessoes {
		var siglaVoto string
		for _, voto := range sessao.Votos {
			if voto.CodigoParlamentar == sen.CodigoParlamentar {
				siglaVoto = voto.SiglaVoto
				break
			}
		}

		if siglaVoto == "" {
			continue
		}

		votacao := s.convertSessaoToVotacao(sessao, sen.ID, siglaVoto)
		if err := s.repo.Upsert(&votacao); err != nil {
			continue
		}
		count++
	}

	return count, nil
}

// convertSessaoToVotacao converte uma sessao de votacao para modelo interno
func (s *SyncService) convertSessaoToVotacao(sessao senadoapi.VotacaoSessaoAPI, senadorID int, siglaVoto string) Votacao {
	var data time.Time
	if sessao.DataSessao != "" {
		// Formato: YYYY-MM-DD ou DD/MM/YYYY
		if t, err := time.Parse("2006-01-02", sessao.DataSessao); err == nil {
			data = t
		} else if t, err := time.Parse("02/01/2006", sessao.DataSessao); err == nil {
			data = t
		}
	}

	return Votacao{
		SenadorID:        senadorID,
		SessaoID:         strconv.Itoa(sessao.CodigoSessao) + "_" + strconv.Itoa(sessao.Ano),
		CodigoSessao:     strconv.Itoa(sessao.CodigoSessao),
		Data:             data,
		Voto:             siglaVoto,
		DescricaoVotacao: sessao.DescricaoVotacao,
		Materia:          "",
	}
}
