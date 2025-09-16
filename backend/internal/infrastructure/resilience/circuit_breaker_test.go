package resilience

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxFailures:      2,
		ResetTimeout:     100 * time.Millisecond,
		SuccessThreshold: 1,
		Timeout:          50 * time.Millisecond,
	}
	cb := NewCircuitBreaker(config)

	// Estado inicial deve ser Closed
	if cb.GetState() != StateClosed {
		t.Errorf("Expected initial state to be CLOSED, got %v", cb.GetState())
	}

	// Simular falhas para abrir circuito
	failOperation := func(ctx context.Context) error {
		return errors.New("simulated failure")
	}

	// Primeira falha - circuito ainda fechado
	err := cb.Execute(context.Background(), failOperation)
	if err == nil {
		t.Error("Expected error from failing operation")
	}
	if cb.GetState() != StateClosed {
		t.Errorf("Expected state to be CLOSED after first failure, got %v", cb.GetState())
	}

	// Segunda falha - circuito deve abrir
	err = cb.Execute(context.Background(), failOperation)
	if err == nil {
		t.Error("Expected error from failing operation")
	}
	if cb.GetState() != StateOpen {
		t.Errorf("Expected state to be OPEN after max failures, got %v", cb.GetState())
	}

	// Tentativa durante circuito aberto - deve falhar imediatamente
	err = cb.Execute(context.Background(), failOperation)
	if err != ErrCircuitBreakerOpen {
		t.Errorf("Expected ErrCircuitBreakerOpen, got %v", err)
	}
}

func TestCircuitBreaker_HalfOpenRecovery(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxFailures:      1,
		ResetTimeout:     10 * time.Millisecond,
		SuccessThreshold: 1,
		Timeout:          50 * time.Millisecond,
	}
	cb := NewCircuitBreaker(config)

	// Forçar abertura do circuito
	failOperation := func(ctx context.Context) error {
		return errors.New("simulated failure")
	}
	cb.Execute(context.Background(), failOperation)

	if cb.GetState() != StateOpen {
		t.Errorf("Expected state to be OPEN, got %v", cb.GetState())
	}

	// Aguardar reset timeout
	time.Sleep(15 * time.Millisecond)

	// Operação de sucesso deve fechar o circuito
	successOperation := func(ctx context.Context) error {
		return nil
	}

	err := cb.Execute(context.Background(), successOperation)
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}

	if cb.GetState() != StateClosed {
		t.Errorf("Expected state to be CLOSED after success, got %v", cb.GetState())
	}
}

func TestCircuitBreaker_Timeout(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxFailures:      1,
		ResetTimeout:     100 * time.Millisecond,
		SuccessThreshold: 1,
		Timeout:          10 * time.Millisecond, // Timeout muito baixo
	}
	cb := NewCircuitBreaker(config)

	// Operação que demora mais que o timeout
	slowOperation := func(ctx context.Context) error {
		select {
		case <-time.After(50 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	err := cb.Execute(context.Background(), slowOperation)
	if err == nil {
		t.Error("Expected timeout error")
	}

	// Circuito deve abrir após timeout
	if cb.GetState() != StateOpen {
		t.Errorf("Expected state to be OPEN after timeout, got %v", cb.GetState())
	}
}

func TestCircuitBreaker_Metrics(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	cb := NewCircuitBreaker(config)

	metrics := cb.GetMetrics()

	// Verificar se todas as métricas estão presentes
	expectedKeys := []string{
		"state", "failures", "successes", "last_failure",
		"next_retry", "max_failures", "reset_timeout", "success_threshold",
	}

	for _, key := range expectedKeys {
		if _, exists := metrics[key]; !exists {
			t.Errorf("Expected metric key %s not found", key)
		}
	}

	// Estado inicial
	if metrics["state"] != StateClosed {
		t.Errorf("Expected initial state to be CLOSED in metrics")
	}

	if metrics["failures"] != 0 {
		t.Errorf("Expected initial failures to be 0")
	}
}

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxFailures:      5,
		ResetTimeout:     100 * time.Millisecond,
		SuccessThreshold: 2,
		Timeout:          50 * time.Millisecond,
	}
	cb := NewCircuitBreaker(config)

	// Teste de concorrência
	done := make(chan bool, 10)

	operation := func(ctx context.Context) error {
		time.Sleep(1 * time.Millisecond)
		return nil
	}

	// Lançar múltiplas goroutines
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			for j := 0; j < 10; j++ {
				cb.Execute(context.Background(), operation)
			}
		}()
	}

	// Aguardar conclusão
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verificar que não houve panic e circuito está funcional
	if cb.GetState() != StateClosed {
		t.Errorf("Expected state to be CLOSED after concurrent success operations")
	}
}

func BenchmarkCircuitBreaker_Execute(b *testing.B) {
	config := DefaultCircuitBreakerConfig()
	cb := NewCircuitBreaker(config)

	operation := func(ctx context.Context) error {
		return nil
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cb.Execute(context.Background(), operation)
		}
	})
}
