package application

import (
	"context"
	"fmt"
	"sync"
	"time"

	"to-de-olho-backend/internal/domain"
)

type stubCamaraClient struct {
	deputados         []domain.Deputado
	despesas          map[string][]domain.Despesa
	fetchDeputadosErr error
	fetchDespesasErr  map[string]error
}

func (c *stubCamaraClient) FetchDeputados(ctx context.Context, partido, uf, nome string) ([]domain.Deputado, error) {
	if c.fetchDeputadosErr != nil {
		return nil, c.fetchDeputadosErr
	}
	deps := make([]domain.Deputado, len(c.deputados))
	copy(deps, c.deputados)
	return deps, nil
}

func (c *stubCamaraClient) FetchDeputadoByID(ctx context.Context, id string) (*domain.Deputado, error) {
	for i := range c.deputados {
		if fmt.Sprintf("%d", c.deputados[i].ID) == id {
			dep := c.deputados[i]
			return &dep, nil
		}
	}
	return nil, domain.ErrDeputadoNaoEncontrado
}

func (c *stubCamaraClient) FetchDespesas(ctx context.Context, deputadoID, ano string) ([]domain.Despesa, error) {
	if c.fetchDespesasErr != nil {
		if err, ok := c.fetchDespesasErr[deputadoID+":"+ano]; ok {
			return nil, err
		}
	}
	if c.despesas == nil {
		return []domain.Despesa{}, nil
	}
	key := deputadoID + ":" + ano
	items := c.despesas[key]
	if len(items) == 0 {
		return []domain.Despesa{}, nil
	}
	out := make([]domain.Despesa, len(items))
	copy(out, items)
	return out, nil
}

type stubCache struct {
	mu    sync.RWMutex
	store map[string]string
}

func newStubCache() *stubCache {
	return &stubCache{store: make(map[string]string)}
}

func (c *stubCache) Get(ctx context.Context, key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.store[key]
	return value, ok
}

func (c *stubCache) Set(ctx context.Context, key, value string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = value
}

type stubDeputadoRepo struct{}

func (r *stubDeputadoRepo) UpsertDeputados(ctx context.Context, deps []domain.Deputado) error {
	return nil
}

func (r *stubDeputadoRepo) ListFromCache(ctx context.Context, limit int) ([]domain.Deputado, error) {
	return nil, nil
}

type trackingDespesaRepository struct {
	mu          sync.Mutex
	stored      map[string][]domain.Despesa
	upsertCount int
	lastID      string
	upsertErr   error
}

func newTrackingDespesaRepository() *trackingDespesaRepository {
	return &trackingDespesaRepository{stored: make(map[string][]domain.Despesa)}
}

func (r *trackingDespesaRepository) UpsertDespesas(ctx context.Context, deputadoID int, ano int, despesas []domain.Despesa) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.upsertErr != nil {
		return r.upsertErr
	}
	key := fmt.Sprintf("%d:%d", deputadoID, ano)
	copyList := make([]domain.Despesa, len(despesas))
	copy(copyList, despesas)
	r.stored[key] = copyList
	r.upsertCount++
	r.lastID = key
	return nil
}

func (r *trackingDespesaRepository) ListDespesasByDeputadoAno(ctx context.Context, deputadoID int, ano int) ([]domain.Despesa, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.stored == nil {
		return []domain.Despesa{}, nil
	}
	key := fmt.Sprintf("%d:%d", deputadoID, ano)
	despesas := r.stored[key]
	if len(despesas) == 0 {
		return []domain.Despesa{}, nil
	}
	copyList := make([]domain.Despesa, len(despesas))
	copy(copyList, despesas)
	return copyList, nil
}
