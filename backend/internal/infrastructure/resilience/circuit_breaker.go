package resilience

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// CircuitBreakerState representa o estado do circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

var (
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
	ErrMaxAttemptsReached = errors.New("max attempts reached")
)

// CircuitBreakerConfig configuração do circuit breaker
type CircuitBreakerConfig struct {
	MaxFailures      int           // Número máximo de falhas antes de abrir
	ResetTimeout     time.Duration // Tempo para tentar fechar o circuito
	SuccessThreshold int           // Sucessos necessários no half-open para fechar
	Timeout          time.Duration // Timeout para operações
}

// DefaultCircuitBreakerConfig configuração padrão otimizada para API da Câmara
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxFailures:      8,                 // Mais tolerante: 8 falhas antes de abrir (vs 5)
		ResetTimeout:     120 * time.Second, // Maior intervalo: 2 minutos para dar tempo da API se recuperar
		SuccessThreshold: 3,                 // 3 sucessos para fechar (mais conservador)
		Timeout:          90 * time.Second,  // Timeout maior para operações lentas da API
	}
}

// CircuitBreaker implementação robusta seguindo padrão Stability Patterns
type CircuitBreaker struct {
	config          CircuitBreakerConfig
	state           CircuitBreakerState
	failures        int
	successes       int
	lastFailureTime time.Time
	nextRetry       time.Time
	mu              sync.RWMutex
}

// NewCircuitBreaker cria um novo circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}
}

// Execute executa uma função com proteção do circuit breaker
func (cb *CircuitBreaker) Execute(ctx context.Context, operation func(ctx context.Context) error) error {
	// Verificar estado atual
	if !cb.allowRequest() {
		return ErrCircuitBreakerOpen
	}

	// Criar contexto com timeout
	opCtx, cancel := context.WithTimeout(ctx, cb.config.Timeout)
	defer cancel()

	// Executar operação
	err := operation(opCtx)

	// Atualizar estado baseado no resultado
	if err != nil {
		cb.recordFailure()
		return fmt.Errorf("circuit breaker operation failed: %w", err)
	}

	cb.recordSuccess()
	return nil
}

// allowRequest verifica se pode executar a requisição
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Verificar se é hora de tentar half-open
		if time.Now().After(cb.nextRetry) {
			cb.state = StateHalfOpen
			cb.successes = 0
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// recordFailure registra uma falha
func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.config.MaxFailures {
			cb.openCircuit()
		}
	case StateHalfOpen:
		cb.openCircuit()
	}
}

// recordSuccess registra um sucesso
func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		cb.failures = 0 // Reset contador de falhas
	case StateHalfOpen:
		cb.successes++
		if cb.successes >= cb.config.SuccessThreshold {
			cb.closeCircuit()
		}
	}
}

// openCircuit abre o circuito
func (cb *CircuitBreaker) openCircuit() {
	cb.state = StateOpen
	cb.nextRetry = time.Now().Add(cb.config.ResetTimeout)
}

// closeCircuit fecha o circuito
func (cb *CircuitBreaker) closeCircuit() {
	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
}

// GetState retorna o estado atual (thread-safe)
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetMetrics retorna métricas do circuit breaker
func (cb *CircuitBreaker) GetMetrics() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":             cb.state,
		"failures":          cb.failures,
		"successes":         cb.successes,
		"last_failure":      cb.lastFailureTime,
		"next_retry":        cb.nextRetry,
		"max_failures":      cb.config.MaxFailures,
		"reset_timeout":     cb.config.ResetTimeout,
		"success_threshold": cb.config.SuccessThreshold,
	}
}

// StateString retorna string do estado para logs
func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// IsCircuitBreakerOpen verifica se o erro informado (ou algum erro encadeado) indica circuito aberto
func IsCircuitBreakerOpen(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, ErrCircuitBreakerOpen) {
		return true
	}

	// Alguns clientes podem encapsular a mensagem sem preservar o sentinel; fazer fallback textual
	return strings.Contains(strings.ToLower(err.Error()), "circuit breaker is open")
}
