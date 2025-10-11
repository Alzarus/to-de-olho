package domain

import (
	"context"
	"time"
)

type votacoesChunkProgressKey struct{}

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
