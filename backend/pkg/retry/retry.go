package retry

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// WithRetry executa fn ate maxAttempts vezes com backoff exponencial.
// Respeita cancelamento via contexto. Backoff: 1s, 2s, 4s, 8s...
func WithRetry(ctx context.Context, maxAttempts int, operacao string, fn func() error) error {
	var lastErr error
	for tentativa := 1; tentativa <= maxAttempts; tentativa++ {
		lastErr = fn()
		if lastErr == nil {
			if tentativa > 1 {
				slog.Info("operacao bem-sucedida apos retry",
					"operacao", operacao,
					"tentativa", tentativa,
				)
			}
			return nil
		}

		if tentativa == maxAttempts {
			break
		}

		// Verificar cancelamento antes de aguardar
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelado para %s: %w", operacao, ctx.Err())
		default:
		}

		backoff := time.Duration(1<<uint(tentativa-1)) * time.Second
		if backoff > 30*time.Second {
			backoff = 30 * time.Second
		}

		slog.Warn("operacao falhou, tentando novamente",
			"operacao", operacao,
			"tentativa", tentativa,
			"max_tentativas", maxAttempts,
			"proximo_backoff", backoff,
			"erro", lastErr,
		)

		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelado para %s: %w", operacao, ctx.Err())
		case <-time.After(backoff):
		}
	}

	return fmt.Errorf("todas as %d tentativas falharam para %s: %w", maxAttempts, operacao, lastErr)
}
