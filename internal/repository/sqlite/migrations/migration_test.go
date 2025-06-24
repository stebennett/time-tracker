package migrations

import (
	"database/sql"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"time-tracker/internal/logging"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestRFC3339Migration(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE time_entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		start_time TEXT,
		end_time TEXT
	)`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO time_entries (start_time, end_time) VALUES
		('2025-06-23 11:47:24.890799237 +0100 BST m=+0.002409088', NULL),
		('2025-06-23 11:20:10.149658307 +0100 BST', NULL),
		('2025-06-23 11:20:10', NULL),
		('2025-06-23T11:20:10+01:00', NULL)
	`)
	require.NoError(t, err)

	// Show before values
	logging.Debugln("Before migration:")
	rows, err := db.Query("SELECT id, start_time FROM time_entries ORDER BY id")
	require.NoError(t, err)
	for rows.Next() {
		var id int64
		var startTime string
		require.NoError(t, rows.Scan(&id, &startTime))
		logging.Debugf("  ID %d: %s\n", id, startTime)
	}
	rows.Close()

	// Apply the Go migration using a transaction
	tx, err := db.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	err = Up_000003_update_time_entries_to_rfc3339(tx)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	// Show after values
	logging.Debugln("After migration:")
	rows, err = db.Query("SELECT id, start_time FROM time_entries ORDER BY id")
	require.NoError(t, err)
	defer rows.Close()

	rfc3339re := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?([+-]\d{2}:\d{2}|Z)$`)
	for rows.Next() {
		var id int64
		var startTime string
		require.NoError(t, rows.Scan(&id, &startTime))
		logging.Debugf("  ID %d: %s\n", id, startTime)
		require.Truef(t, rfc3339re.MatchString(startTime), "not RFC3339: %s", startTime)
	}
}

func TestRunMigrations_DirtyDatabase(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create migrations table
	_, err = db.Exec(`
		CREATE TABLE migrations (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			dirty BOOLEAN DEFAULT FALSE
		)
	`)
	if err != nil {
		t.Fatalf("failed to create migrations table: %v", err)
	}

	// Mark a migration as dirty
	_, err = db.Exec("INSERT INTO migrations (version, dirty) VALUES (1, TRUE)")
	if err != nil {
		t.Fatalf("failed to insert dirty migration: %v", err)
	}

	// Try to run migrations - should fail due to dirty state
	err = RunMigrations(db)
	if err == nil {
		t.Fatal("expected RunMigrations to fail on dirty database, but it succeeded")
	}

	// Check that the error message contains the expected information
	if !strings.Contains(err.Error(), "database is in a dirty state") {
		t.Errorf("expected error to mention dirty state, got: %v", err)
	}

	if !strings.Contains(err.Error(), "failed migration(s): [1]") {
		t.Errorf("expected error to mention failed migration version 1, got: %v", err)
	}
}

func TestRunMigrations_BackupAndRestore(t *testing.T) {
	// Create a temporary database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create some initial data
	_, err = db.Exec(`CREATE TABLE test_data (id INTEGER PRIMARY KEY, value TEXT)`)
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}

	_, err = db.Exec(`INSERT INTO test_data (value) VALUES ('original data')`)
	if err != nil {
		t.Fatalf("failed to insert test data: %v", err)
	}

	// Verify initial data exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_data").Scan(&count)
	if err != nil {
		t.Fatalf("failed to count test data: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row, got %d", count)
	}

	// Run migrations - this should create a backup
	err = RunMigrations(db)
	if err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	// Check that backup files were cleaned up (successful migration)
	backupFiles, err := filepath.Glob(dbPath + ".backup.*")
	if err != nil {
		t.Fatalf("failed to check for backup files: %v", err)
	}
	if len(backupFiles) > 0 {
		t.Errorf("expected no backup files after successful migration, found: %v", backupFiles)
	}

	// Verify the original data is still intact
	err = db.QueryRow("SELECT COUNT(*) FROM test_data").Scan(&count)
	if err != nil {
		t.Fatalf("failed to count test data after migration: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row after migration, got %d", count)
	}
}
