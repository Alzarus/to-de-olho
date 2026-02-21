package ceaps

import (
	"context"
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

// SyncFromAPI busca despesas da API e atualiza o banco
func (s *SyncService) SyncFromAPI(ctx context.Context, ano int) error {
	slog.Info("iniciando sync de despesas CEAPS", "ano", ano)

	// Buscar da API Administrativa
	despesasAPI, err := s.client.ListarDespesasCEAPS(ctx, ano)
	if err != nil {
		return err
	}

	slog.Info("despesas recebidas da API", "total", len(despesasAPI))

	// Buscar mapeamento codigo parlamentar -> ID interno
	senadores, _ := s.senadorRepo.FindAll()
	codigoToID := make(map[int]int)
	for _, sen := range senadores {
		codigoToID[sen.CodigoParlamentar] = sen.ID
	}

	// Converter e salvar cada despesa
	var successCount, skipCount int
	for _, d := range despesasAPI {
		senadorID, exists := codigoToID[d.CodSenador]
		if !exists {
			skipCount++
			continue
		}

		despesa := s.convertToDespesa(d, senadorID)
		if err := s.repo.Upsert(&despesa); err != nil {
			slog.Error("falha ao salvar despesa", "senador", d.CodSenador, "error", err)
			continue
		}
		successCount++
	}

	slog.Info("sync de despesas concluido", "salvos", successCount, "ignorados", skipCount, "total", len(despesasAPI))
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
