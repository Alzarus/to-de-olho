package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWithRetry_SuccessOnFirstAttempt(t *testing.T) {
	calls := 0
	err := WithRetry(context.Background(), 3, "test-op", func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("esperava sucesso, obteve erro: %v", err)
	}
	if calls != 1 {
		t.Fatalf("esperava 1 chamada, obteve %d", calls)
	}
}

func TestWithRetry_SuccessAfterRetries(t *testing.T) {
	calls := 0
	err := WithRetry(context.Background(), 3, "test-op", func() error {
		calls++
		if calls < 3 {
			return errors.New("erro temporario")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("esperava sucesso apos retries, obteve erro: %v", err)
	}
	if calls != 3 {
		t.Fatalf("esperava 3 chamadas, obteve %d", calls)
	}
}

func TestWithRetry_AllAttemptsFail(t *testing.T) {
	calls := 0
	err := WithRetry(context.Background(), 2, "test-op", func() error {
		calls++
		return errors.New("erro permanente")
	})
	if err == nil {
		t.Fatal("esperava erro, obteve sucesso")
	}
	if calls != 2 {
		t.Fatalf("esperava 2 chamadas, obteve %d", calls)
	}
}

func TestWithRetry_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()
	err := WithRetry(ctx, 5, "test-op", func() error {
		calls++
		return errors.New("erro")
	})
	if err == nil {
		t.Fatal("esperava erro por cancelamento")
	}
	if calls >= 5 {
		t.Fatalf("esperava menos que 5 chamadas (cancelado), obteve %d", calls)
	}
}
