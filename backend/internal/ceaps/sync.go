package ceaps

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Alzarus/to-de-olho/internal/senador"
	"github.com/Alzarus/to-de-olho/pkg/senado"
)

// SyncService gerencia sincronizacao de despesas CEAPS
type SyncService struct {
	repo        *Repository
	senadorRepo *senador.Repository
	client      *senado.AdmClient
}

// NewSyncService cria um novo servico de sincronizacao
func NewSyncService(repo *Repository, senadorRepo *senador.Repository, client *senado.AdmClient) *SyncService {
	return &SyncService{
		repo:        repo,
		senadorRepo: senadorRepo,
		client:      client,
	}
}

// SyncFromAPI grava as despesas de um ano como retrato da API: dentro de uma
// transacao, apaga o ano e insere o que a API devolveu (inclusive correcoes e
// exclusoes feitas na origem). Falha na busca nao apaga nada.
func (s *SyncService) SyncFromAPI(ctx context.Context, ano int) error {
	slog.Info("iniciando sync de despesas CEAPS", "ano", ano)

	despesasAPI, err := s.client.ListarDespesasCEAPS(ctx, ano)
	if err != nil {
		return err
	}
	if len(despesasAPI) == 0 {
		return fmt.Errorf("API devolveu 0 despesas para %d", ano)
	}

	senadores, err := s.senadorRepo.FindAll(true)
	if err != nil {
		return err
	}
	codigoToID := make(map[int]int, len(senadores))
	for _, sen := range senadores {
		codigoToID[sen.CodigoParlamentar] = sen.ID
	}

	despesas := make([]DespesaCEAPS, 0, len(despesasAPI))
	vistos := make(map[int]bool, len(despesasAPI))
	var ignorados int
	for _, d := range despesasAPI {
		senadorID, ok := codigoToID[d.CodSenador]
		if !ok {
			ignorados++ // parlamentar fora da tabela senadores
			continue
		}
		if d.ID == 0 || vistos[d.ID] {
			return fmt.Errorf("despesa sem id ou com id repetido na API (%d)", d.ID)
		}
		vistos[d.ID] = true
		despesas = append(despesas, s.convertToDespesa(d, senadorID))
	}

	if err := s.repo.SubstituirAno(ano, despesas); err != nil {
		return fmt.Errorf("falha ao gravar despesas de %d: %w", ano, err)
	}
	slog.Info("sync de despesas concluido", "ano", ano, "salvos", len(despesas), "ignorados", ignorados, "total", len(despesasAPI))
	return nil
}

// convertToDespesa converte dados da API para modelo interno
func (s *SyncService) convertToDespesa(d senado.DespesaCEAPSAPI, senadorID int) DespesaCEAPS {
	var dataEmissao *time.Time
	if d.Data != "" {
		// Tentar formato ISO (YYYY-MM-DD) primeiro, depois o brasileiro (DD/MM/YYYY)
		if t, err := time.Parse("2006-01-02", d.Data); err == nil {
			dataEmissao = &t
		} else if t, err := time.Parse("02/01/2006", d.Data); err == nil {
			dataEmissao = &t
		} else {
			slog.Warn("falha ao formatar data CEAPS", "data_raw", d.Data, "senador_id", senadorID)
		}
	}

	despesa := DespesaCEAPS{
		IDOrigem:    d.ID,
		SenadorID:   senadorID,
		Ano:         d.Ano,
		Mes:         d.Mes,
		TipoDespesa: d.TipoDespesa,
		Fornecedor:  d.Fornecedor,
		CNPJCPF:     d.CNPJCPF,
		Documento:   d.Documento,
		DataEmissao: dataEmissao,
		Valor:       d.ValorReembolso,
	}

	// Calcular valor em centavos para chave de idempotencia
	_ = despesa.BeforeCreate(nil)

	return despesa
}
