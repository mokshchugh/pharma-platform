package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"pharma-platform/internal/postgres"
	"pharma-platform/internal/questdb"
)

func MigratePostgres(ctx context.Context, client *postgres.Client, schemaDir, seedDir string, doSeed bool) error {
	db := client.DB()

	if err := runSQLFiles(db, schemaDir); err != nil {
		return fmt.Errorf("postgres schema migration: %w", err)
	}

	if doSeed {
		var count int
		if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM machines").Scan(&count); err != nil {
			return fmt.Errorf("check machines table: %w", err)
		}

		if count == 0 {
			if err := runSQLFiles(db, seedDir); err != nil {
				return fmt.Errorf("postgres seed: %w", err)
			}
		}
	}

	return nil
}

func MigrateQuestDB(ctx context.Context, client *questdb.Client, dir string) error {
	return client.MigrateDir(ctx, dir)
}

func runSQLFiles(db *sql.DB, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read dir %s: %w", dir, err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		stmt := strings.TrimSpace(string(content))
		if stmt == "" {
			continue
		}

		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("exec %s: %w", entry.Name(), err)
		}
	}

	return nil
}
