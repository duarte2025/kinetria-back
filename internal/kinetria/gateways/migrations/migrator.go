package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strings"
)

//go:embed *.sql
var migrationsFS embed.FS

type Migrator struct {
	db *sql.DB
}

func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

func (m *Migrator) Run(ctx context.Context) error {
	if err := m.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	files, err := migrationsFS.ReadDir(".")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}
	sort.Strings(sqlFiles)

	for _, filename := range sqlFiles {
		if err := m.runMigration(ctx, filename); err != nil {
			return fmt.Errorf("migration %s failed: %w", filename, err)
		}
	}

	return nil
}

func (m *Migrator) createMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`
	_, err := m.db.ExecContext(ctx, query)
	return err
}

func (m *Migrator) runMigration(ctx context.Context, filename string) error {
	var exists bool
	err := m.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE filename = $1)", filename).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	content, err := migrationsFS.ReadFile(filename)
	if err != nil {
		return err
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, string(content)); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (filename) VALUES ($1)", filename); err != nil {
		return err
	}

	return tx.Commit()
}
