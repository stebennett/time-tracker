package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

//go:embed *.sql
var migrationsFS embed.FS

// MigrationType represents the type of migration
type MigrationType string

const (
	SQLMigration MigrationType = "sql"
	GoMigration  MigrationType = "go"
)

// Migration represents a database migration
type Migration struct {
	Version  int
	Type     MigrationType
	UpSQL    string
	DownSQL  string
	UpFunc   func(*sql.Tx) error
	DownFunc func(*sql.Tx) error
}

// MigrationStatus represents the status of a migration
type MigrationStatus struct {
	Version int
	Applied bool
	Dirty   bool
}

// Global registry for Go migrations
var goMigrationRegistry = make(map[int]Migration)

// RegisterGoMigration registers a Go migration by version
func RegisterGoMigration(version int, up, down func(*sql.Tx) error) {
	goMigrationRegistry[version] = Migration{
		Version:  version,
		Type:     GoMigration,
		UpFunc:   up,
		DownFunc: down,
	}
}

// RunMigrations executes all pending migrations in order
func RunMigrations(db *sql.DB) error {
	// Get database file path for backup
	dbPath, err := getDatabasePath(db)
	if err != nil {
		return fmt.Errorf("failed to get database path: %w", err)
	}

	// Create backup before starting migrations
	backupPath, err := createBackup(dbPath)
	if err != nil {
		return fmt.Errorf("failed to create database backup: %w", err)
	}

	// Clean up backup on success, restore on failure
	defer func() {
		if err != nil {
			// Migration failed, restore from backup
			if restoreErr := restoreFromBackup(dbPath, backupPath); restoreErr != nil {
				fmt.Printf("Warning: failed to restore database from backup: %v\n", restoreErr)
			} else {
				fmt.Printf("Database restored from backup: %s\n", backupPath)
			}
		} else {
			// Migration succeeded, clean up backup and any existing failed backups
			if cleanupErr := cleanupBackups(dbPath); cleanupErr != nil {
				fmt.Printf("Warning: failed to cleanup backups: %v\n", cleanupErr)
			}
		}
	}()

	// Create migrations table if it doesn't exist
	if err = createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Check for dirty database state
	if err = checkDirtyDatabase(db); err != nil {
		return err
	}

	// Load all migrations (SQL and Go) into a single sorted list
	migrations, err := loadAllMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Get applied migrations
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Apply migrations in order
	for _, migration := range migrations {
		fmt.Printf("Applying migration version %d (type: %s)\n", migration.Version, migration.Type)
		if !applied[migration.Version] {
			if err = applyMigration(db, migration); err != nil {
				// Mark migration as failed (dirty state)
				if markErr := markMigrationFailed(db, migration.Version); markErr != nil {
					return fmt.Errorf("failed to mark migration %d as failed: %w (original error: %w)", migration.Version, markErr, err)
				}
				return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
			}
		}
	}

	return nil
}

func getDatabasePath(db *sql.DB) (string, error) {
	// For SQLite, we need to get the database path from the connection
	// This is a simplified approach - in practice, the database path should be passed in
	// or stored in the connection context

	// Try to get the database path from PRAGMA database_list
	rows, err := db.Query("PRAGMA database_list")
	if err != nil {
		return "", fmt.Errorf("unable to query database list: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var seq, name, file string
		if err := rows.Scan(&seq, &name, &file); err != nil {
			continue
		}
		if name == "main" && file != "" {
			return file, nil
		}
	}

	// If we can't get the path, return empty string for in-memory databases
	return "", nil
}

func createBackup(dbPath string) (string, error) {
	// Skip backup for in-memory databases
	if dbPath == "" || dbPath == ":memory:" {
		return "", nil
	}

	// Create backup filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupPath := dbPath + ".backup." + timestamp

	// Copy the database file
	source, err := os.Open(dbPath)
	if err != nil {
		return "", fmt.Errorf("failed to open source database: %w", err)
	}
	defer source.Close()

	dest, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	if err != nil {
		return "", fmt.Errorf("failed to copy database to backup: %w", err)
	}

	fmt.Printf("Database backup created: %s\n", backupPath)
	return backupPath, nil
}

func restoreFromBackup(dbPath, backupPath string) error {
	// Skip restore for in-memory databases or if no backup was created
	if dbPath == "" || dbPath == ":memory:" || backupPath == "" {
		return nil
	}

	// Close any open connections to the database
	// Note: This is a limitation - we can't restore while connections are open
	// In practice, this would require closing the database connection first

	// Copy backup back to original location
	source, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer source.Close()

	dest, err := os.Create(dbPath)
	if err != nil {
		return fmt.Errorf("failed to create restored database: %w", err)
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	if err != nil {
		return fmt.Errorf("failed to copy backup to database: %w", err)
	}

	return nil
}

func cleanupBackups(dbPath string) error {
	// Skip cleanup for in-memory databases
	if dbPath == "" || dbPath == ":memory:" {
		return nil
	}

	// Find all backup files for this database
	backupPattern := dbPath + ".backup.*"
	backupFiles, err := filepath.Glob(backupPattern)
	if err != nil {
		return fmt.Errorf("failed to find backup files: %w", err)
	}

	// Remove all backup files
	for _, backupFile := range backupFiles {
		if err := os.Remove(backupFile); err != nil {
			return fmt.Errorf("failed to remove backup file %s: %w", backupFile, err)
		}
		fmt.Printf("Removed backup file: %s\n", backupFile)
	}

	return nil
}

func checkDirtyDatabase(db *sql.DB) error {
	rows, err := db.Query("SELECT version FROM migrations WHERE dirty = TRUE")
	if err != nil {
		return fmt.Errorf("failed to check for dirty migrations: %w", err)
	}
	defer rows.Close()

	var dirtyVersions []int
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("failed to scan dirty migration version: %w", err)
		}
		dirtyVersions = append(dirtyVersions, version)
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating dirty migrations: %w", err)
	}

	if len(dirtyVersions) > 0 {
		return fmt.Errorf("database is in a dirty state due to failed migration(s): %v. Please resolve the failed migration(s) before running new migrations", dirtyVersions)
	}

	return nil
}

func createMigrationsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS migrations (
		version INTEGER PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		dirty BOOLEAN DEFAULT FALSE
	)`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	// Add dirty column to existing migrations table if it doesn't exist
	_, err = db.Exec("ALTER TABLE migrations ADD COLUMN dirty BOOLEAN DEFAULT FALSE")
	// Ignore error if column already exists
	return nil
}

func loadAllMigrations() ([]Migration, error) {
	var migrations []Migration

	// Load SQL migrations
	sqlMigrations, err := loadSQLMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to load SQL migrations: %w", err)
	}
	migrations = append(migrations, sqlMigrations...)

	// Load Go migrations
	goMigrations := loadGoMigrations()
	migrations = append(migrations, goMigrations...)

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func loadSQLMigrations() ([]Migration, error) {
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
			Type:    SQLMigration,
			UpSQL:   string(upSQL),
			DownSQL: string(downSQL),
		})
	}

	return migrations, nil
}

func loadGoMigrations() []Migration {
	var migrations []Migration
	for _, m := range goMigrationRegistry {
		migrations = append(migrations, m)
	}
	return migrations
}

func getAppliedMigrations(db *sql.DB) (map[int]bool, error) {
	rows, err := db.Query("SELECT version FROM migrations WHERE dirty = FALSE")
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

	// Apply the migration based on its type
	switch migration.Type {
	case SQLMigration:
		if _, err := tx.Exec(migration.UpSQL); err != nil {
			tx.Rollback()
			return err
		}
	case GoMigration:
		if err := migration.UpFunc(tx); err != nil {
			tx.Rollback()
			return err
		}
	default:
		tx.Rollback()
		return fmt.Errorf("unknown migration type: %s", migration.Type)
	}

	// Mark migration as applied
	if _, err := tx.Exec("INSERT INTO migrations (version, dirty) VALUES (?, FALSE)", migration.Version); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func markMigrationFailed(db *sql.DB, version int) error {
	// Insert or update migration record as failed (dirty)
	_, err := db.Exec(`
		INSERT OR REPLACE INTO migrations (version, dirty) 
		VALUES (?, TRUE)
	`, version)
	return err
}

func extractVersion(filename string) int {
	var version int
	fmt.Sscanf(filename, "%d_", &version)
	return version
}
