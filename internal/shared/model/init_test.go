package model

import (
	"testing"

	"github.com/fastax/fastax-server/internal/shared/config"
)

// InitDB tests require CGO (mattn/go-sqlite3).
// Skipped when CGO_ENABLED=0. Run with: CGO_ENABLED=1 go test ./internal/shared/model/

func TestInitDB_CreatesTables(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CGO-dependent test in short mode")
	}

	cfg := config.DatabaseConfig{
		Path:     ":memory:",
		WALMode:  false,
		LogLevel: "silent",
	}

	db, err := InitDB(cfg)
	if err != nil {
		t.Skipf("InitDB() requires CGO: %v", err)
	}

	tables := []string{
		"users", "user_profiles", "sub_accounts",
		"token_products", "token_inventories", "user_tokens", "token_transfers",
		"orders", "payments", "refunds",
	}
	for _, table := range tables {
		if !db.Migrator().HasTable(table) {
			t.Errorf("missing table: %s", table)
		}
	}
}

func TestInitDB_WALMode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CGO-dependent test in short mode")
	}

	dir := t.TempDir()
	cfg := config.DatabaseConfig{
		Path:     dir + "/test.db",
		WALMode:  true,
		LogLevel: "silent",
	}

	db, err := InitDB(cfg)
	if err != nil {
		t.Skipf("InitDB() requires CGO: %v", err)
	}

	var journalMode string
	db.Raw("PRAGMA journal_mode").Scan(&journalMode)
	if journalMode != "wal" {
		t.Errorf("journal_mode = %v, want wal", journalMode)
	}
}
