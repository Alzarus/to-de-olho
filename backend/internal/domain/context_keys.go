package domain

import (
	"context"
	"time"
)

type votacoesChunkProgressKey struct{}
type skipDespesaPersistKey struct{}
type forceDespesaRemoteKey struct{}

// VotacoesChunkProgressCallback representa um callback para acompanhar o avanço por intervalo
// da coleta de votações. Quando `success` é falso, significa que o intervalo foi marcado para retry futuro.
type VotacoesChunkProgressCallback func(start, end time.Time, success bool)

// WithVotacoesChunkProgress retorna um novo contexto contendo o callback para progresso por intervalo.
func WithVotacoesChunkProgress(ctx context.Context, cb VotacoesChunkProgressCallback) context.Context {
	if ctx == nil || cb == nil {
		return ctx
	}
	return context.WithValue(ctx, votacoesChunkProgressKey{}, cb)
}

// GetVotacoesChunkProgress recupera o callback de progresso por intervalo do contexto, se existir.
func GetVotacoesChunkProgress(ctx context.Context) VotacoesChunkProgressCallback {
	if ctx == nil {
		return nil
	}
	if v := ctx.Value(votacoesChunkProgressKey{}); v != nil {
		if cb, ok := v.(VotacoesChunkProgressCallback); ok {
			return cb
		}
	}
	return nil
}

// WithSkipDespesaPersist marca o contexto para impedir que a camada de serviço
// reescreva despesas no repositório – útil quando o próprio chamador fará o upsert.
func WithSkipDespesaPersist(ctx context.Context) context.Context {
	if ctx == nil {
		return context.WithValue(context.Background(), skipDespesaPersistKey{}, true)
	}
	return context.WithValue(ctx, skipDespesaPersistKey{}, true)
}

// ShouldSkipDespesaPersist indica se o serviço deve pular a persistência de despesas.
func ShouldSkipDespesaPersist(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	if v, ok := ctx.Value(skipDespesaPersistKey{}).(bool); ok {
		return v
	}
	return false
}

// WithForceDespesaRemote força o serviço a ignorar hits de cache/banco e ir direto à API.
func WithForceDespesaRemote(ctx context.Context) context.Context {
	if ctx == nil {
		return context.WithValue(context.Background(), forceDespesaRemoteKey{}, true)
	}
	return context.WithValue(ctx, forceDespesaRemoteKey{}, true)
}

// ShouldForceDespesaRemote indica se o serviço deve priorizar a API da Câmara.
func ShouldForceDespesaRemote(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	if v, ok := ctx.Value(forceDespesaRemoteKey{}).(bool); ok {
		return v
	}
	return false
}
