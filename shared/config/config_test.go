package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	os.Clearenv()
	cfg := Load()

	if cfg.AppName != "nexbic-platform" {
		t.Errorf("expected default app name, got %s", cfg.AppName)
	}

	if cfg.AppEnv != "development" {
		t.Errorf("expected default app env, got %s", cfg.AppEnv)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Clearenv()
	os.Setenv("APP_NAME", "test-app")
	os.Setenv("APP_ENV", "testing")
	os.Setenv("SERVER_PORT", "9999")
	os.Setenv("JWT_SECRET", "test-secret-that-is-long-enough-for-hmac")
	os.Setenv("DB_HOST", "testhost")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")

	cfg := Load()

	if cfg.AppName != "test-app" {
		t.Errorf("expected test-app, got %s", cfg.AppName)
	}

	if cfg.AppEnv != "testing" {
		t.Errorf("expected testing, got %s", cfg.AppEnv)
	}

	if cfg.Server.Port != 9999 {
		t.Errorf("expected port 9999, got %d", cfg.Server.Port)
	}

	if cfg.Database.Host != "testhost" {
		t.Errorf("expected testhost, got %s", cfg.Database.Host)
	}
}

func TestDatabaseURLParsing(t *testing.T) {
	os.Clearenv()
	os.Setenv("DATABASE_URL", "postgres://user:pass@myhost:5432/mydb?sslmode=require")

	cfg := Load()

	if cfg.Database.User != "user" {
		t.Errorf("expected user, got %s", cfg.Database.User)
	}
	if cfg.Database.Password != "pass" {
		t.Errorf("expected pass, got %s", cfg.Database.Password)
	}
	if cfg.Database.Host != "myhost" {
		t.Errorf("expected myhost, got %s", cfg.Database.Host)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("expected 5432, got %d", cfg.Database.Port)
	}
	if cfg.Database.DBName != "mydb" {
		t.Errorf("expected mydb, got %s", cfg.Database.DBName)
	}
	if cfg.Database.SSLMode != "require" {
		t.Errorf("expected require, got %s", cfg.Database.SSLMode)
	}
}
