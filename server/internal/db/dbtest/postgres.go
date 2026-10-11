//go:build integration

// Package dbtest starts a throwaway Postgres for integration tests.
package dbtest

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
	"time"

	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/db"

	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	// Image matches the postgres service in docker-compose.yml.
	Image        = "postgres:18-alpine"
	testDatabase = "testdb"
	testUser     = "testuser"
	testPassword = "testpassword"
)

// NewPostgres starts a Postgres container, connects to it and stops it when the test ends.
// The caller sets up the logger first, because the database manager logs while it connects.
func NewPostgres(t *testing.T) *db.PostgresDB {
	t.Helper()
	ctx := context.Background()
	container, err := tcpostgres.Run(ctx, Image,
		tcpostgres.WithDatabase(testDatabase), tcpostgres.WithUsername(testUser), tcpostgres.WithPassword(testPassword),
		tcpostgres.BasicWaitStrategies())
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)

	cfg := &config.Config{
		DatabaseHost: host, DatabasePort: port.Port(), DatabaseUsername: testUser, DatabasePassword: testPassword,
		DatabaseName: testDatabase, DatabaseMaxOpenConns: 5, DatabaseMaxIdleConns: 5,
		DatabaseConnMaxLifetime: 5 * time.Minute, DatabaseConnMaxIdleTime: time.Minute, DatabaseConnectTimeout: 10 * time.Second,
		DatabaseQueryTimeout: 5 * time.Second, DatabaseHealthTimeout: 3 * time.Second, DatabaseRetryAttempts: 3,
		DatabaseRetryDelay: time.Second, DatabaseSSLMode: "disable", DatabaseTimezone: "UTC",
	}
	pdb := &db.PostgresDB{}
	require.NoError(t, pdb.NewPostgresDB(cfg))
	return pdb
}

// ApplyMigrations runs the Atlas migration files in order, so a test uses the real schema.
func ApplyMigrations(t *testing.T, pdb *db.PostgresDB) {
	t.Helper()
	_, here, _, ok := runtime.Caller(0)
	require.True(t, ok)
	files, err := filepath.Glob(filepath.Join(filepath.Dir(here), "..", "..", "..", "cmd", "migrations", "sql", "*.sql"))
	require.NoError(t, err)
	require.NotEmpty(t, files)
	sort.Strings(files)
	for _, file := range files {
		script, err := os.ReadFile(file)
		require.NoError(t, err)
		require.NoError(t, pdb.Exec(string(script)).Error, filepath.Base(file))
	}
}
