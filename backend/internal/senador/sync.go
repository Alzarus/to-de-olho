package senador

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/Alzarus/to-de-olho/pkg/senado"
)

// SyncService gerencia sincronizacao de senadores com a API
type SyncService struct {
	repo   *Repository
	client *senado.LegisClient
}

// NewSyncService cria um novo servico de sincronizacao
func NewSyncService(repo *Repository, client *senado.LegisClient) *SyncService {
	return &SyncService{
		repo:   repo,
		client: client,
	}
}

// SyncFromAPI busca senadores da API e atualiza o banco
func (s *SyncService) SyncFromAPI(ctx context.Context) error {
	slog.Info("iniciando sync de senadores")

	// Buscar da API Legislativa
	parlamentares, err := s.client.ListarSenadoresAtuais(ctx)
	if err != nil {
		return err
	}

	slog.Info("senadores recebidos da API", "total", len(parlamentares))

	// Converter e salvar cada senador
	var successCount int
	var activeCodes []int
	for _, p := range parlamentares {
		senador := s.convertToSenador(p)
		if err := s.repo.Upsert(&senador); err != nil {
			slog.Error("falha ao salvar senador", "codigo", senador.CodigoParlamentar, "error", err)
			continue
		}
		activeCodes = append(activeCodes, senador.CodigoParlamentar)
		successCount++
	}

	if len(activeCodes) > 0 {
		if err := s.repo.SetInactive(activeCodes); err != nil {
			slog.Error("falha ao inativar senadores antigos", "error", err)
		}
	}

	slog.Info("sync de senadores concluido", "salvos", successCount, "total", len(parlamentares))
	return nil
}

// convertToSenador converte dados da API para modelo interno
func (s *SyncService) convertToSenador(p senado.ParlamentarAPI) Senador {
	id := p.IdentificacaoParlamentar

	codigo, _ := strconv.Atoi(id.CodigoParlamentar)

	var titular string
	if p.Mandato.Titular != nil {
		titular = p.Mandato.Titular.NomeParlamentar
	}

	return Senador{
		CodigoParlamentar: codigo,
		Nome:              id.NomeParlamentar,
		NomeCompleto:      id.NomeCompletoParlamentar,
		Partido:           id.SiglaPartidoParlamentar,
		UF:                id.UfParlamentar,
		FotoURL:           id.UrlFotoParlamentar,
		Email:             id.EmailParlamentar,
		Cargo:             p.Mandato.DescricaoParticipacao,
		Titular:           titular,
		EmExercicio:       true,
	}
}
