package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppName  string
	AppEnv   string
	LogLevel string

	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	PostgREST PostgRESTConfig
	Asynq    AsynqConfig
	Monitoring MonitoringConfig
}

type ServerConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	CORSOrigins     []string
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxConns        int
	MinConns        int
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
	HealthCheckPeriod time.Duration
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type JWTConfig struct {
	Secret          string
	AccessTTL      time.Duration
	RefreshTTL     time.Duration
	Issuer          string
	Audience        string
}

type PostgRESTConfig struct {
	URL       string
	AdminURL  string
	Timeout   time.Duration
}

type AsynqConfig struct {
	Concurrency int
	Host        string
	Port        int
	Password    string
	DB          int
}

type MonitoringConfig struct {
	Enabled    bool
	MetricPath string
}

func Load() *Config {
	return &Config{
		AppName:  getEnv("APP_NAME", "nexbic-platform"),
		AppEnv:   getEnv("APP_ENV", "development"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			Port:            getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:     getEnvDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout:    getEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			ShutdownTimeout: getEnvDuration("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second),
			CORSOrigins:     getEnvSlice("CORS_ORIGINS", []string{"*"}),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "api_admin"),
			Password:        getEnv("DB_PASSWORD", ""),
			DBName:          getEnv("DB_NAME", "postgres_api"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxConns:        getEnvInt("DB_MAX_CONNS", 20),
			MinConns:        getEnvInt("DB_MIN_CONNS", 2),
			MaxConnLifetime: getEnvDuration("DB_MAX_CONN_LIFETIME", 30*time.Minute),
			MaxConnIdleTime: getEnvDuration("DB_MAX_CONN_IDLE_TIME", 10*time.Minute),
			HealthCheckPeriod: getEnvDuration("DB_HEALTH_CHECK_PERIOD", 1*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", "change-me-secret"),
			AccessTTL:  getEnvDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL: getEnvDuration("JWT_REFRESH_TTL", 7*24*time.Hour),
			Issuer:      getEnv("JWT_ISSUER", "nexbic-platform"),
			Audience:    getEnv("JWT_AUD", "nexbic-platform"),
		},
		PostgREST: PostgRESTConfig{
			URL:      getEnv("POSTGREST_URL", "http://localhost:3000"),
			AdminURL: getEnv("POSTGREST_ADMIN_URL", "http://localhost:3001"),
			Timeout:  getEnvDuration("POSTGREST_TIMEOUT", 30*time.Second),
		},
		Asynq: AsynqConfig{
			Concurrency: getEnvInt("ASYNQ_CONCURRENCY", 10),
			Host:        getEnv("ASYNQ_HOST", "localhost"),
			Port:        getEnvInt("ASYNQ_PORT", 6379),
			Password:    getEnv("ASYNQ_PASSWORD", ""),
			DB:          getEnvInt("ASYNQ_DB", 1),
		},
		Monitoring: MonitoringConfig{
			Enabled:    getEnvBool("MONITORING_ENABLED", true),
			MetricPath: getEnv("METRIC_PATH", "/metrics"),
		},
	}
}

func (c *Config) DatabaseDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.User, c.Database.Password,
		c.Database.Host, c.Database.Port,
		c.Database.DBName, c.Database.SSLMode)
}

func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

func (c *Config) AsynqAddr() string {
	return fmt.Sprintf("%s:%d", c.Asynq.Host, c.Asynq.Port)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return fallback
}

func getEnvSlice(key string, fallback []string) []string {
	if val := os.Getenv(key); val != "" {
		result := []string{}
		for _, s := range split(val, ",") {
			result = append(result, trim(s))
		}
		return result
	}
	return fallback
}

func split(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep[0] {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	if start <= len(s) {
		result = append(result, s[start:])
	}
	return result
}

func trim(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
