package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig_Success(t *testing.T) {
	// Setup environment variables
	originalVars := map[string]string{
		"POSTGRES_PASSWORD": os.Getenv("POSTGRES_PASSWORD"),
		"PORT":              os.Getenv("PORT"),
		"GIN_MODE":          os.Getenv("GIN_MODE"),
	}

	// Set test values
	os.Setenv("POSTGRES_PASSWORD", "test_password")
	os.Setenv("PORT", "9000")
	os.Setenv("GIN_MODE", "debug")

	// Cleanup
	defer func() {
		for key, value := range originalVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	config, err := LoadConfig()

	if err != nil {
		t.Fatalf("LoadConfig() deveria ter sucesso, mas falhou: %v", err)
	}

	if config == nil {
		t.Fatal("config não deveria ser nil")
	}

	// Verificar valores configurados
	if config.Database.Password != "test_password" {
		t.Errorf("esperava password 'test_password', obteve '%s'", config.Database.Password)
	}

	if config.Server.Port != "9000" {
		t.Errorf("esperava port '9000', obteve '%s'", config.Server.Port)
	}

	if config.Server.GinMode != "debug" {
		t.Errorf("esperava gin_mode 'debug', obteve '%s'", config.Server.GinMode)
	}
}

func TestLoadConfig_MissingPasswordFails(t *testing.T) {
	// Este teste não pode ser executado porque getEnvRequired() chama log.Fatalf
	// Vamos testar apenas a validação
	config := &Config{
		Database: DatabaseConfig{
			Password: "", // Vazio
		},
		Server: ServerConfig{
			Port: "8080",
		},
		CamaraClient: CamaraClientConfig{
			RPS: 2,
		},
	}

	err := config.Validate()
	if err == nil {
		t.Error("Validate() deveria falhar quando POSTGRES_PASSWORD está vazio")
	}

	expectedMsg := "POSTGRES_PASSWORD is required"
	if err.Error() != expectedMsg {
		t.Errorf("erro esperado: %q, obteve: %q", expectedMsg, err.Error())
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
		errorMsg  string
	}{
		{
			name: "configuração válida",
			config: &Config{
				Database: DatabaseConfig{
					Password: "valid_password",
				},
				Server: ServerConfig{
					Port: "8080",
				},
				CamaraClient: CamaraClientConfig{
					RPS: 2,
				},
			},
			wantError: false,
		},
		{
			name: "password vazio deve falhar",
			config: &Config{
				Database: DatabaseConfig{
					Password: "",
				},
				Server: ServerConfig{
					Port: "8080",
				},
				CamaraClient: CamaraClientConfig{
					RPS: 2,
				},
			},
			wantError: true,
			errorMsg:  "POSTGRES_PASSWORD is required",
		},
		{
			name: "porta vazia deve falhar",
			config: &Config{
				Database: DatabaseConfig{
					Password: "valid_password",
				},
				Server: ServerConfig{
					Port: "",
				},
				CamaraClient: CamaraClientConfig{
					RPS: 2,
				},
			},
			wantError: true,
			errorMsg:  "PORT is required",
		},
		{
			name: "RPS inválido deve falhar",
			config: &Config{
				Database: DatabaseConfig{
					Password: "valid_password",
				},
				Server: ServerConfig{
					Port: "8080",
				},
				CamaraClient: CamaraClientConfig{
					RPS: -1,
				},
			},
			wantError: true,
			errorMsg:  "CAMARA_CLIENT_RPS must be positive",
		},
		{
			name: "RPS zero deve falhar",
			config: &Config{
				Database: DatabaseConfig{
					Password: "valid_password",
				},
				Server: ServerConfig{
					Port: "8080",
				},
				CamaraClient: CamaraClientConfig{
					RPS: 0,
				},
			},
			wantError: true,
			errorMsg:  "CAMARA_CLIENT_RPS must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantError {
				if err == nil {
					t.Error("esperava erro mas não recebeu")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("erro esperado: %q, recebido: %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("não esperava erro, mas recebeu: %v", err)
				}
			}
		})
	}
}

func TestDatabaseConfig_ConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		config   DatabaseConfig
		expected string
	}{
		{
			name: "configuração padrão",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "postgres",
				Password: "senha123",
				Database: "testdb",
				SSLMode:  "disable",
			},
			expected: "postgres://postgres:senha123@localhost:5432/testdb?sslmode=disable",
		},
		{
			name: "configuração com SSL",
			config: DatabaseConfig{
				Host:     "prod-db.example.com",
				Port:     "5432",
				User:     "prod_user",
				Password: "super_secret",
				Database: "production",
				SSLMode:  "require",
			},
			expected: "postgres://prod_user:super_secret@prod-db.example.com:5432/production?sslmode=require",
		},
		{
			name: "configuração com porta customizada",
			config: DatabaseConfig{
				Host:     "127.0.0.1",
				Port:     "15432",
				User:     "dev",
				Password: "dev_pass",
				Database: "dev_db",
				SSLMode:  "prefer",
			},
			expected: "postgres://dev:dev_pass@127.0.0.1:15432/dev_db?sslmode=prefer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.ConnectionString()
			if result != tt.expected {
				t.Errorf("ConnectionString() = %q, esperado %q", result, tt.expected)
			}
		})
	}
}

func TestConfig_EnvironmentChecks(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		isDev       bool
		isProd      bool
	}{
		{
			name:        "desenvolvimento",
			environment: "development",
			isDev:       true,
			isProd:      false,
		},
		{
			name:        "desenvolvimento maiúsculo",
			environment: "DEVELOPMENT",
			isDev:       true,
			isProd:      false,
		},
		{
			name:        "produção",
			environment: "production",
			isDev:       false,
			isProd:      true,
		},
		{
			name:        "produção maiúsculo",
			environment: "PRODUCTION",
			isDev:       false,
			isProd:      true,
		},
		{
			name:        "staging não é dev nem prod",
			environment: "staging",
			isDev:       false,
			isProd:      false,
		},
		{
			name:        "testing não é dev nem prod",
			environment: "testing",
			isDev:       false,
			isProd:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				App: AppConfig{
					Environment: tt.environment,
				},
			}

			if config.IsDevelopment() != tt.isDev {
				t.Errorf("IsDevelopment() = %v, esperado %v", config.IsDevelopment(), tt.isDev)
			}

			if config.IsProduction() != tt.isProd {
				t.Errorf("IsProduction() = %v, esperado %v", config.IsProduction(), tt.isProd)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "usa valor padrão quando env não existe",
			key:          "TEST_NONEXISTENT_KEY",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "usa valor do ambiente quando existe",
			key:          "TEST_EXISTING_KEY",
			defaultValue: "default",
			envValue:     "from_env",
			expected:     "from_env",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Limpar variável de ambiente primeiro
			originalValue := os.Getenv(tt.key)
			os.Unsetenv(tt.key)

			// Definir valor se fornecido
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			}

			// Cleanup
			defer func() {
				if originalValue != "" {
					os.Setenv(tt.key, originalValue)
				} else {
					os.Unsetenv(tt.key)
				}
			}()

			result := getEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnv(%q, %q) = %q, esperado %q", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		expected     int
	}{
		{
			name:         "usa valor padrão quando env não existe",
			key:          "TEST_INT_NONEXISTENT",
			defaultValue: 42,
			envValue:     "",
			expected:     42,
		},
		{
			name:         "converte valor válido do ambiente",
			key:          "TEST_INT_VALID",
			defaultValue: 42,
			envValue:     "123",
			expected:     123,
		},
		{
			name:         "usa padrão para valor inválido",
			key:          "TEST_INT_INVALID",
			defaultValue: 42,
			envValue:     "não_é_número",
			expected:     42,
		},
		{
			name:         "converte número negativo",
			key:          "TEST_INT_NEGATIVE",
			defaultValue: 42,
			envValue:     "-10",
			expected:     -10,
		},
		{
			name:         "converte zero",
			key:          "TEST_INT_ZERO",
			defaultValue: 42,
			envValue:     "0",
			expected:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			originalValue := os.Getenv(tt.key)
			os.Unsetenv(tt.key)

			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			}

			// Cleanup
			defer func() {
				if originalValue != "" {
					os.Setenv(tt.key, originalValue)
				} else {
					os.Unsetenv(tt.key)
				}
			}()

			result := getInt(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getInt(%q, %d) = %d, esperado %d", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetDuration(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue time.Duration
		envValue     string
		expected     time.Duration
	}{
		{
			name:         "usa valor padrão quando env não existe",
			key:          "TEST_DURATION_NONEXISTENT",
			defaultValue: 5 * time.Second,
			envValue:     "",
			expected:     5 * time.Second,
		},
		{
			name:         "converte duration válida",
			key:          "TEST_DURATION_VALID",
			defaultValue: 5 * time.Second,
			envValue:     "10s",
			expected:     10 * time.Second,
		},
		{
			name:         "converte minutos",
			key:          "TEST_DURATION_MINUTES",
			defaultValue: 5 * time.Second,
			envValue:     "2m",
			expected:     2 * time.Minute,
		},
		{
			name:         "usa padrão para valor inválido",
			key:          "TEST_DURATION_INVALID",
			defaultValue: 5 * time.Second,
			envValue:     "não_é_duration",
			expected:     5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			originalValue := os.Getenv(tt.key)
			os.Unsetenv(tt.key)

			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			}

			// Cleanup
			defer func() {
				if originalValue != "" {
					os.Setenv(tt.key, originalValue)
				} else {
					os.Unsetenv(tt.key)
				}
			}()

			result := getDuration(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getDuration(%q, %v) = %v, esperado %v", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		envValue     string
		expected     bool
	}{
		{
			name:         "usa valor padrão quando env não existe",
			key:          "TEST_BOOL_NONEXISTENT",
			defaultValue: true,
			envValue:     "",
			expected:     true,
		},
		{
			name:         "converte true",
			key:          "TEST_BOOL_TRUE",
			defaultValue: false,
			envValue:     "true",
			expected:     true,
		},
		{
			name:         "converte false",
			key:          "TEST_BOOL_FALSE",
			defaultValue: true,
			envValue:     "false",
			expected:     false,
		},
		{
			name:         "converte 1 para true",
			key:          "TEST_BOOL_ONE",
			defaultValue: false,
			envValue:     "1",
			expected:     true,
		},
		{
			name:         "converte 0 para false",
			key:          "TEST_BOOL_ZERO",
			defaultValue: true,
			envValue:     "0",
			expected:     false,
		},
		{
			name:         "usa padrão para valor inválido",
			key:          "TEST_BOOL_INVALID",
			defaultValue: false,
			envValue:     "não_é_bool",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			originalValue := os.Getenv(tt.key)
			os.Unsetenv(tt.key)

			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			}

			// Cleanup
			defer func() {
				if originalValue != "" {
					os.Setenv(tt.key, originalValue)
				} else {
					os.Unsetenv(tt.key)
				}
			}()

			result := getBool(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getBool(%q, %v) = %v, esperado %v", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestLoadConfig_DefaultValues(t *testing.T) {
	// Limpar variáveis de ambiente
	envVars := []string{
		"PORT", "GIN_MODE", "POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_USER",
		"POSTGRES_DB", "POSTGRES_SSL_MODE", "REDIS_ADDR", "CAMARA_API_BASE_URL",
		"APP_ENV", "LOG_LEVEL", "APP_VERSION",
	}

	originalValues := make(map[string]string)
	for _, key := range envVars {
		originalValues[key] = os.Getenv(key)
		os.Unsetenv(key)
	}

	// Definir apenas a password obrigatória
	originalPassword := os.Getenv("POSTGRES_PASSWORD")
	os.Setenv("POSTGRES_PASSWORD", "test_pass_for_defaults")

	// Cleanup
	defer func() {
		for key, value := range originalValues {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
		if originalPassword == "" {
			os.Unsetenv("POSTGRES_PASSWORD")
		} else {
			os.Setenv("POSTGRES_PASSWORD", originalPassword)
		}
	}()

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() falhou: %v", err)
	}

	// Verificar valores padrão
	expectedDefaults := map[string]interface{}{
		"Server.Port":          "8080",
		"Server.GinMode":       "release",
		"Database.Host":        "localhost",
		"Database.Port":        "5432",
		"Database.User":        "postgres",
		"Database.Database":    "to_de_olho",
		"Database.SSLMode":     "disable",
		"Redis.Addr":           "localhost:6379",
		"CamaraClient.BaseURL": "https://dadosabertos.camara.leg.br/api/v2",
		"CamaraClient.RPS":     2,
		"App.Environment":      "development",
		"App.LogLevel":         "info",
		"App.Version":          "1.0.0",
	}

	checks := map[string]interface{}{
		"Server.Port":          config.Server.Port,
		"Server.GinMode":       config.Server.GinMode,
		"Database.Host":        config.Database.Host,
		"Database.Port":        config.Database.Port,
		"Database.User":        config.Database.User,
		"Database.Database":    config.Database.Database,
		"Database.SSLMode":     config.Database.SSLMode,
		"Redis.Addr":           config.Redis.Addr,
		"CamaraClient.BaseURL": config.CamaraClient.BaseURL,
		"CamaraClient.RPS":     config.CamaraClient.RPS,
		"App.Environment":      config.App.Environment,
		"App.LogLevel":         config.App.LogLevel,
		"App.Version":          config.App.Version,
	}

	for key, expected := range expectedDefaults {
		if actual := checks[key]; actual != expected {
			t.Errorf("%s: esperado %v, obteve %v", key, expected, actual)
		}
	}
}
