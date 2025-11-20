package ingestor

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	app "to-de-olho-backend/internal/application"
	"to-de-olho-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type fakeIngestorDB struct {
	execCalls int
}

func (f *fakeIngestorDB) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	f.execCalls++
	return pgconn.CommandTag{}, nil
}

func (f *fakeIngestorDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return nil, fmt.Errorf("not implemented")
}

func (f *fakeIngestorDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return nil
}

type fakeVotacaoRepo struct {
	upserted []*domain.Votacao
}

func (f *fakeVotacaoRepo) CreateVotacao(ctx context.Context, votacao *domain.Votacao) error {
	return nil
}
func (f *fakeVotacaoRepo) GetVotacaoByID(ctx context.Context, id int64) (*domain.Votacao, error) {
	return nil, domain.ErrVotacaoNaoEncontrada
}
func (f *fakeVotacaoRepo) ListVotacoes(ctx context.Context, filtros domain.FiltrosVotacao, pag domain.Pagination) ([]*domain.Votacao, int, error) {
	return nil, 0, nil
}
func (f *fakeVotacaoRepo) UpdateVotacao(ctx context.Context, votacao *domain.Votacao) error {
	return nil
}
func (f *fakeVotacaoRepo) DeleteVotacao(ctx context.Context, id int64) error { return nil }
func (f *fakeVotacaoRepo) CreateVotoDeputado(ctx context.Context, voto *domain.VotoDeputado) error {
	return nil
}
func (f *fakeVotacaoRepo) GetVotosPorVotacao(ctx context.Context, idVotacao int64) ([]*domain.VotoDeputado, error) {
	return nil, nil
}
func (f *fakeVotacaoRepo) GetVotoPorDeputado(ctx context.Context, idVotacao int64, idDeputado int) (*domain.VotoDeputado, error) {
	return nil, domain.ErrVotoDeputadoNaoEncontrado
}
func (f *fakeVotacaoRepo) CreateOrientacaoPartido(ctx context.Context, orientacao *domain.OrientacaoPartido) error {
	return nil
}
func (f *fakeVotacaoRepo) GetOrientacoesPorVotacao(ctx context.Context, idVotacao int64) ([]*domain.OrientacaoPartido, error) {
	return nil, nil
}
func (f *fakeVotacaoRepo) GetVotacaoDetalhada(ctx context.Context, id int64) (*domain.VotacaoDetalhada, error) {
	return nil, domain.ErrVotacaoNaoEncontrada
}
func (f *fakeVotacaoRepo) UpsertVotacao(ctx context.Context, votacao *domain.Votacao) error {
	f.upserted = append(f.upserted, votacao)
	return nil
}
func (f *fakeVotacaoRepo) GetPresencaPorDeputadoAno(ctx context.Context, ano int) ([]domain.PresencaCount, error) {
	return nil, nil
}
func (f *fakeVotacaoRepo) GetRankingDeputadosAggregated(ctx context.Context, ano int) ([]domain.RankingDeputadoVotacao, error) {
	return nil, nil
}
func (f *fakeVotacaoRepo) GetDisciplinaPartidosAggregated(ctx context.Context, ano int) ([]domain.VotacaoPartido, error) {
	return nil, nil
}
func (f *fakeVotacaoRepo) GetVotacaoStatsAggregated(ctx context.Context, ano int) (*domain.VotacaoStats, error) {
	return nil, nil
}

type fakeCamaraClient struct {
	votacoes   []*domain.Votacao
	votos      map[string][]*domain.VotoDeputado
	orientacao map[string][]*domain.OrientacaoPartido
	err        error
}

func (f *fakeCamaraClient) GetVotacoes(ctx context.Context, dataInicio, dataFim time.Time) ([]*domain.Votacao, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.votacoes, nil
}

func (f *fakeCamaraClient) GetVotacao(ctx context.Context, id string) (*domain.Votacao, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeCamaraClient) GetVotosPorVotacao(ctx context.Context, idVotacao string) ([]*domain.VotoDeputado, error) {
	if f.votos != nil {
		if votos, ok := f.votos[idVotacao]; ok {
			return votos, nil
		}
	}
	return nil, nil
}

func (f *fakeCamaraClient) GetOrientacoesPorVotacao(ctx context.Context, idVotacao string) ([]*domain.OrientacaoPartido, error) {
	if f.orientacao != nil {
		if orientacoes, ok := f.orientacao[idVotacao]; ok {
			return orientacoes, nil
		}
	}
	return nil, nil
}

type noopCache struct{}

func (noopCache) Get(ctx context.Context, key string) (string, bool)            { return "", false }
func (noopCache) Set(ctx context.Context, key, value string, ttl time.Duration) {}

func TestExecuteVotacoesBackfillSuccess(t *testing.T) {
	t.Parallel()

	db := &fakeIngestorDB{}
	manager := NewBackfillManagerWithDB(db)

	repo := &fakeVotacaoRepo{}
	client := &fakeCamaraClient{
		votacoes: []*domain.Votacao{{
			ID:       1,
			IDCamara: "abc",
		}},
	}
	cache := noopCache{}

	service := app.NewVotacoesService(repo, client, cache)

	executor := &StrategicBackfillExecutor{
		manager:         manager,
		votacoesService: service,
		strategy:        BackfillStrategy{BatchSize: 50},
	}

	checkpoint := &BackfillCheckpoint{
		Metadata: map[string]interface{}{"year": float64(2024)},
		Progress: BackfillProgress{},
	}

	if err := executor.executeVotacoesBackfill(context.Background(), checkpoint); err != nil {
		t.Fatalf("esperava sucesso, ocorreu erro: %v", err)
	}

	if len(repo.upserted) != 1 {
		t.Fatalf("esperava 1 votação upsertada, obtive %d", len(repo.upserted))
	}

	if checkpoint.Progress.ProcessedItems != 1 {
		t.Fatalf("processedItems = %d, esperado 1", checkpoint.Progress.ProcessedItems)
	}

	if checkpoint.Progress.TotalItems != 1 {
		t.Fatalf("totalItems = %d, esperado 1", checkpoint.Progress.TotalItems)
	}

	if db.execCalls == 0 {
		t.Fatalf("esperava chamada ao UpdateProgress")
	}
}

func TestExecuteVotacoesBackfillError(t *testing.T) {
	t.Parallel()

	db := &fakeIngestorDB{}
	manager := NewBackfillManagerWithDB(db)

	repo := &fakeVotacaoRepo{}
	client := &fakeCamaraClient{err: errors.New("falha API")}
	cache := noopCache{}
	service := app.NewVotacoesService(repo, client, cache)

	executor := &StrategicBackfillExecutor{
		manager:         manager,
		votacoesService: service,
		strategy:        BackfillStrategy{BatchSize: 50},
	}

	checkpoint := &BackfillCheckpoint{
		Metadata: map[string]interface{}{"year": float64(2023)},
		Progress: BackfillProgress{},
	}

	err := executor.executeVotacoesBackfill(context.Background(), checkpoint)
	if err == nil {
		t.Fatalf("esperava erro quando sincronização falha")
	}
	if !strings.Contains(err.Error(), "sincronizar votações") {
		t.Fatalf("mensagem de erro inesperada: %v", err)
	}

	if db.execCalls != 0 {
		t.Fatalf("não deveria atualizar progresso em caso de erro")
	}
}

func TestExecuteVotacoesBackfillMissingYear(t *testing.T) {
	t.Parallel()

	executor := &StrategicBackfillExecutor{}
	checkpoint := &BackfillCheckpoint{Metadata: map[string]interface{}{"foo": 2024}}

	err := executor.executeVotacoesBackfill(context.Background(), checkpoint)
	if err == nil {
		t.Fatalf("esperava erro por metadado ausente")
	}
	if !strings.Contains(err.Error(), "metadado 'year'") {
		t.Fatalf("mensagem de erro inesperada: %v", err)
	}
}
