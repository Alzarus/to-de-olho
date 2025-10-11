package config

import (
	"strconv"
	"time"
)

// ⏱️ TimeoutConfig centraliza todas as configurações de timeout do sistema
type TimeoutConfig struct {
	// HTTP Client timeouts
	HTTPRequestTimeout time.Duration
	HTTPClientTimeout  time.Duration

	// Circuit breaker timeouts
	CircuitBreakerResetTimeout time.Duration
	CircuitBreakerOpTimeout    time.Duration

	// Database timeouts
	DatabaseTimeout        time.Duration
	DatabaseConnectTimeout time.Duration

	// Cache timeouts
	CacheReadTimeout  time.Duration
	CacheWriteTimeout time.Duration
	CacheDefaultTTL   time.Duration

	// Rate limiting
	APIRateLimitPerSecond int
	APIBurstLimit         int

	// Retry configurations
	RetryDelay        time.Duration
	MaxRetryAttempts  int
	BackoffMultiplier float64
}

// NewTimeoutConfig cria uma nova instância com valores das variáveis de ambiente
func NewTimeoutConfig() *TimeoutConfig {
	return &TimeoutConfig{
		// HTTP - timeouts mais generosos para API da Câmara (pode ser lenta)
		HTTPRequestTimeout: parseDurationEnv("HTTP_REQUEST_TIMEOUT", "60s"),
		HTTPClientTimeout:  parseDurationEnv("HTTP_CLIENT_TIMEOUT", "90s"),

		// Circuit Breaker - tempos mais conservadores baseados nos logs de timeout
		CircuitBreakerResetTimeout: parseDurationEnv("CIRCUIT_BREAKER_RESET_TIMEOUT", "120s"),
		CircuitBreakerOpTimeout:    parseDurationEnv("CIRCUIT_BREAKER_OP_TIMEOUT", "90s"),

		// Database - timeouts otimizados para PostgreSQL
		DatabaseTimeout:        parseDurationEnv("DATABASE_TIMEOUT", "15s"),
		DatabaseConnectTimeout: parseDurationEnv("DATABASE_CONNECT_TIMEOUT", "5s"),

		// Cache - timeouts para Redis
		CacheReadTimeout:  parseDurationEnv("CACHE_READ_TIMEOUT", "500ms"),
		CacheWriteTimeout: parseDurationEnv("CACHE_WRITE_TIMEOUT", "1s"),
		CacheDefaultTTL:   parseDurationEnv("CACHE_DEFAULT_TTL", "5m"),

		// Rate limiting da API da Câmara mais conservador (baseado nos logs de timeout)
		APIRateLimitPerSecond: parseIntEnv("API_RATE_LIMIT_PER_SECOND", 1), // Mais conservador
		APIBurstLimit:         parseIntEnv("API_BURST_LIMIT", 2),           // Burst menor

		// Retry policy - backoff exponencial mais longo devido aos timeouts
		RetryDelay:        parseDurationEnv("RETRY_DELAY", "10s"),
		MaxRetryAttempts:  parseIntEnv("MAX_RETRY_ATTEMPTS", 2), // Menos tentativas para evitar circuit breaker
		BackoffMultiplier: parseFloatEnv("BACKOFF_MULTIPLIER", 3.0),
	}
}

// parseDurationEnv converte variável de ambiente para time.Duration com fallback
func parseDurationEnv(key, defaultValue string) time.Duration {
	envValue := getEnv(key, defaultValue)
	duration, err := time.ParseDuration(envValue)
	if err != nil {
		// Se não conseguir parsear, usa o valor padrão
		defaultDuration, _ := time.ParseDuration(defaultValue)
		return defaultDuration
	}
	return duration
}

// parseIntEnv converte variável de ambiente para int com fallback
func parseIntEnv(key string, defaultValue int) int {
	envValue := getEnv(key, "")
	if envValue == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(envValue)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// parseFloatEnv converte variável de ambiente para float64 com fallback
func parseFloatEnv(key string, defaultValue float64) float64 {
	envValue := getEnv(key, "")
	if envValue == "" {
		return defaultValue
	}

	floatValue, err := strconv.ParseFloat(envValue, 64)
	if err != nil {
		return defaultValue
	}
	return floatValue
}
