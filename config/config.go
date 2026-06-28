package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nexbic/platform/pkg/database"
)

type Config struct {
	AppName  string
	AppEnv   string
	LogLevel string

	Server     ServerConfig
	Database   database.Config
	JWT        JWTConfig
	SuperAdmin SuperAdminConfig
}

type SuperAdminConfig struct {
	Email    string
	Password string
}

type ServerConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	CORSOrigins     []string
}

type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	Issuer     string
	Audience   string
}

func Load() *Config {
	cfg := &Config{
		AppName:  getEnv("APP_NAME", "nexbic-db-platform"),
		AppEnv:   getEnv("APP_ENV", "development"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			Port:            getEnvInt("SERVER_PORT", 2121),
			ReadTimeout:     getEnvDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout:    getEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			ShutdownTimeout: getEnvDuration("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second),
			CORSOrigins:     getEnvSlice("CORS_ORIGINS", []string{"http://localhost:5173"}),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", ""),
			AccessTTL:   getEnvDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL:  getEnvDuration("JWT_REFRESH_TTL", 7*24*time.Hour),
			Issuer:      getEnv("JWT_ISSUER", "nexbic-db-platform"),
			Audience:    getEnv("JWT_AUD", "nexbic-db-platform"),
		},
		SuperAdmin: SuperAdminConfig{
			Email:    getEnv("SUPER_ADMIN_EMAIL", ""),
			Password: getEnv("SUPER_ADMIN_PASSWORD", ""),
		},
	}

	parseDatabaseURL(cfg)

	return cfg
}

func parseDatabaseURL(cfg *Config) {
	if ds := getEnv("DATABASE_URL", ""); ds != "" {
		u, err := url.Parse(ds)
		if err != nil {
			return
		}
		if u.User != nil {
			cfg.Database.User = u.User.Username()
			if pw, ok := u.User.Password(); ok {
				cfg.Database.Password = pw
			}
		}
		cfg.Database.Host = u.Hostname()
		cfg.Database.Port, _ = strconv.Atoi(portFromHost(u.Host))
		cfg.Database.DBName = strings.TrimPrefix(u.Path, "/")
		if ssl := u.Query().Get("sslmode"); ssl != "" {
			cfg.Database.SSLMode = ssl
		}
		return
	}
	cfg.Database.Host = getEnv("DB_HOST", "localhost")
	cfg.Database.Port = getEnvInt("DB_PORT", 5432)
	cfg.Database.User = getEnv("DB_USER", "pgadmin")
	cfg.Database.Password = getEnv("DB_PASSWORD", "postgres")
	cfg.Database.DBName = getEnv("DB_NAME", "nexbic_admin")
	cfg.Database.SSLMode = getEnv("DB_SSLMODE", "disable")
	cfg.Database.MaxConns = getEnvInt("DB_MAX_CONNS", 20)
	cfg.Database.MinConns = getEnvInt("DB_MIN_CONNS", 2)
	cfg.Database.MaxConnLifetime = getEnvDuration("DB_MAX_CONN_LIFETIME", 30*time.Minute)
	cfg.Database.MaxConnIdleTime = getEnvDuration("DB_MAX_CONN_IDLE_TIME", 10*time.Minute)
}

func portFromHost(host string) string {
	idx := strings.LastIndex(host, ":")
	if idx >= 0 {
		return host[idx+1:]
	}
	return "0"
}

func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
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
