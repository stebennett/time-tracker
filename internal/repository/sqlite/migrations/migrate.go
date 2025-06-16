package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strings"
)

//go:embed *.sql
var migrationsFS embed.FS

// Migration represents a database migration
type Migration struct {
	Version int
	Up      string
	Down    string
}

// RunMigrations executes all pending migrations
func RunMigrations(db *sql.DB) error {
	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get all migration files
	migrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Get applied migrations
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Apply pending migrations
	for _, migration := range migrations {
		if !applied[migration.Version] {
			if err := applyMigration(db, migration); err != nil {
				return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
			}
		}
	}

	return nil
}

func createMigrationsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS migrations (
		version INTEGER PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`
	_, err := db.Exec(query)
	return err
}

func loadMigrations() ([]Migration, error) {
	entries, err := migrationsFS.ReadDir(".")
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}

		version := extractVersion(entry.Name())
		if version == 0 {
			continue
		}

		upSQL, err := migrationsFS.ReadFile(entry.Name())
		if err != nil {
			return nil, err
		}

		downFile := strings.Replace(entry.Name(), ".up.sql", ".down.sql", 1)
		downSQL, err := migrationsFS.ReadFile(downFile)
		if err != nil {
			return nil, err
		}

		migrations = append(migrations, Migration{
			Version: version,
			Up:      string(upSQL),
			Down:    string(downSQL),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func getAppliedMigrations(db *sql.DB) (map[int]bool, error) {
	rows, err := db.Query("SELECT version FROM migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}
	return applied, rows.Err()
}

func applyMigration(db *sql.DB, migration Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(migration.Up); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.Exec("INSERT INTO migrations (version) VALUES (?)", migration.Version); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func extractVersion(filename string) int {
	var version int
	fmt.Sscanf(filename, "%d_", &version)
	return version
} 