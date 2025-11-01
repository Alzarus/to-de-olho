package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config estrutura principal de configuração
type Config struct {
	Server       ServerConfig
	Database     DatabaseConfig
	Redis        RedisConfig
	CamaraClient CamaraClientConfig
	App          AppConfig
	Ingestor     IngestorConfig
	Timeouts     *TimeoutConfig // Configurações centralizadas de timeout
}

type ServerConfig struct {
	Port            string
	GinMode         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	RateLimit       int
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	DialTimeout  time.Duration
}

type CamaraClientConfig struct {
	BaseURL    string
	Timeout    time.Duration
	RPS        int
	Burst      int
	MaxRetries int
	RetryDelay time.Duration
}

type AppConfig struct {
	Environment string
	LogLevel    string
	Version     string
}

// IngestorConfig configurações específicas do ingestor
type IngestorConfig struct {
	BackfillStartYear int // Ano inicial para backfill histórico
	BatchSize         int // Tamanho dos lotes para processamento
	MaxRetries        int // Máximo de tentativas por lote
	// MonitorTimeout define quanto tempo o monitor do ingestor aguardará a conclusão do backfill
	MonitorTimeout time.Duration
}

// LoadConfig carrega configurações de variáveis de ambiente
func LoadConfig() (*Config, error) {
	// Tentar carregar arquivo .env se existir
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port:            getEnv("PORT", "8080"),
			GinMode:         getEnv("GIN_MODE", "release"),
			ReadTimeout:     getDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout:    getDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			ShutdownTimeout: getDuration("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second),
			RateLimit:       getInt("RATE_LIMIT_RPS", 100),
		},
		Database: DatabaseConfig{
			// Default to docker-compose service name so containers resolve the DB correctly
			Host:            getEnv("POSTGRES_HOST", "postgres"),
			Port:            getEnv("POSTGRES_PORT", "5432"),
			User:            getEnv("POSTGRES_USER", "postgres"),
			Password:        getEnvRequired("POSTGRES_PASSWORD"),
			Database:        getEnv("POSTGRES_DB", "to_de_olho"),
			SSLMode:         getEnv("POSTGRES_SSL_MODE", "disable"),
			MaxConns:        getInt32("POSTGRES_MAX_CONNS", 25),
			MinConns:        getInt32("POSTGRES_MIN_CONNS", 5),
			MaxConnLifetime: getDuration("POSTGRES_MAX_CONN_LIFETIME", 1*time.Hour),
			MaxConnIdleTime: getDuration("POSTGRES_MAX_CONN_IDLE_TIME", 30*time.Minute),
		},
		Redis: RedisConfig{
			// Default to docker-compose service name so containers resolve correctly
			Addr:         getEnv("REDIS_ADDR", "redis:6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getInt("REDIS_DB", 0),
			ReadTimeout:  getDuration("REDIS_READ_TIMEOUT", 500*time.Millisecond),
			WriteTimeout: getDuration("REDIS_WRITE_TIMEOUT", 500*time.Millisecond),
			DialTimeout:  getDuration("REDIS_DIAL_TIMEOUT", 500*time.Millisecond),
		},
		CamaraClient: CamaraClientConfig{
			BaseURL:    getEnv("CAMARA_API_BASE_URL", "https://dadosabertos.camara.leg.br/api/v2"),
			Timeout:    getDuration("CAMARA_CLIENT_TIMEOUT", 30*time.Second),
			RPS:        getInt("CAMARA_CLIENT_RPS", 2),
			Burst:      getInt("CAMARA_CLIENT_BURST", 4),
			MaxRetries: getInt("CAMARA_CLIENT_MAX_RETRIES", 3),
			RetryDelay: getDuration("CAMARA_CLIENT_RETRY_DELAY", 1*time.Second),
		},
		App: AppConfig{
			Environment: getEnv("APP_ENV", "development"),
			LogLevel:    getEnv("LOG_LEVEL", "info"),
			Version:     getEnv("APP_VERSION", "1.0.0"),
		},
		Ingestor: IngestorConfig{
			BackfillStartYear: getIntPrefer("BACKFILL_START_YEAR", "INGESTOR_BACKFILL_START_YEAR", 2025),
			// Aumentar default para evitar páginas muito pequenas no backfill histórico
			BatchSize:  getIntPrefer("BACKFILL_BATCH_SIZE", "INGESTOR_BATCH_SIZE", 100),
			MaxRetries: getIntPrefer("BACKFILL_MAX_RETRIES", "INGESTOR_MAX_RETRIES", 3),
			// Default aumentado para 2h para respeitar .env local quando godotenv não for carregado
			MonitorTimeout: getDuration("BACKFILL_MONITOR_TIMEOUT", 24*time.Hour),
		},
	}

	// Adicionar configurações de timeout centralizadas
	config.Timeouts = NewTimeoutConfig()

	// Validar configurações críticas
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// Validate valida configurações críticas
func (c *Config) Validate() error {
	if c.Database.Password == "" {
		return fmt.Errorf("POSTGRES_PASSWORD is required")
	}

	if c.Server.Port == "" {
		return fmt.Errorf("PORT is required")
	}

	if c.CamaraClient.RPS <= 0 {
		return fmt.Errorf("CAMARA_CLIENT_RPS must be positive")
	}

	return nil
}

// ConnectionString retorna string de conexão PostgreSQL
func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Database, c.SSLMode,
	)
}

// IsDevelopment verifica se está em ambiente de desenvolvimento
func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.App.Environment) == "development"
}

// IsProduction verifica se está em ambiente de produção
func (c *Config) IsProduction() bool {
	return strings.ToLower(c.App.Environment) == "production"
}

// Funções auxiliares para variáveis de ambiente

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return value
}

func getIntPrefer(primary, fallback string, defaultValue int) int {
	if v, ok := getOptionalInt(primary); ok {
		return v
	}
	if fallback != "" {
		if v, ok := getOptionalInt(fallback); ok {
			return v
		}
	}
	return defaultValue
}

func getOptionalInt(key string) (int, bool) {
	value := os.Getenv(key)
	if value == "" {
		return 0, false
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("invalid integer value for %s: %v", key, err)
		return 0, false
	}
	return v, true
}

func getInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Invalid integer value for %s: %s, using default: %d", key, valueStr, defaultValue)
		return defaultValue
	}

	return value
}

// getInt32 retorna um int32 com validação de range para evitar integer overflow
func getInt32(key string, defaultValue int32) int32 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseInt(valueStr, 10, 32)
	if err != nil {
		log.Printf("Invalid int32 value for %s: %s, using default: %d", key, valueStr, defaultValue)
		return defaultValue
	}

	return int32(value)
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		log.Printf("Invalid duration value for %s: %s, using default: %s", key, valueStr, defaultValue)
		return defaultValue
	}

	return value
}

func getBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		log.Printf("Invalid boolean value for %s: %s, using default: %t", key, valueStr, defaultValue)
		return defaultValue
	}

	return value
}
