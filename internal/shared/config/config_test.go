package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestServerConfig_Addr(t *testing.T) {
	tests := []struct {
		port int
		want string
	}{
		{8080, ":8080"},
		{9090, ":9090"},
		{0, ":0"},
	}
	for _, tt := range tests {
		sc := ServerConfig{Port: tt.port}
		if got := sc.Addr(); got != tt.want {
			t.Errorf("ServerConfig.Addr() = %v, want %v", got, tt.want)
		}
	}
}

func TestLoad_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
server:
  port: 9090
  mode: release
database:
  path: ./test.db
  wal_mode: false
jwt:
  secret: test-secret-key
  access_expiry: 1h
  refresh_expiry: 168h
rate_limit:
  ip: 50
  auth: 10
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %v, want 9090", cfg.Server.Port)
	}
	if cfg.Server.Mode != "release" {
		t.Errorf("Server.Mode = %v, want release", cfg.Server.Mode)
	}
	if cfg.Database.Path != "./test.db" {
		t.Errorf("Database.Path = %v, want ./test.db", cfg.Database.Path)
	}
	if cfg.Database.WALMode != false {
		t.Errorf("Database.WALMode = %v, want false", cfg.Database.WALMode)
	}
	if cfg.JWT.Secret != "test-secret-key" {
		t.Errorf("JWT.Secret = %v, want test-secret-key", cfg.JWT.Secret)
	}
	if cfg.JWT.AccessExpiry != 1*time.Hour {
		t.Errorf("JWT.AccessExpiry = %v, want 1h", cfg.JWT.AccessExpiry)
	}
	if cfg.JWT.RefreshExpiry != 168*time.Hour {
		t.Errorf("JWT.RefreshExpiry = %v, want 168h", cfg.JWT.RefreshExpiry)
	}
	if cfg.RateLimit.IP != 50 {
		t.Errorf("RateLimit.IP = %v, want 50", cfg.RateLimit.IP)
	}
}

func TestLoad_Defaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := "jwt:\n  secret: test\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %v, want 8080", cfg.Server.Port)
	}
	if cfg.Server.Mode != "debug" {
		t.Errorf("Server.Mode = %v, want debug", cfg.Server.Mode)
	}
	if cfg.Server.ReadTimeout != 30*time.Second {
		t.Errorf("Server.ReadTimeout = %v, want 30s", cfg.Server.ReadTimeout)
	}
	if cfg.Server.WriteTimeout != 30*time.Second {
		t.Errorf("Server.WriteTimeout = %v, want 30s", cfg.Server.WriteTimeout)
	}
	if cfg.Database.WALMode != true {
		t.Errorf("Database.WALMode = %v, want true", cfg.Database.WALMode)
	}
	if cfg.RateLimit.IP != 60 {
		t.Errorf("RateLimit.IP = %v, want 60", cfg.RateLimit.IP)
	}
	if cfg.RateLimit.Auth != 5 {
		t.Errorf("RateLimit.Auth = %v, want 5", cfg.RateLimit.Auth)
	}
}

func TestLoad_MissingSecret(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := "server:\n  port: 8080\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() expected error for missing jwt.secret, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("Load() expected error for missing file, got nil")
	}
}
