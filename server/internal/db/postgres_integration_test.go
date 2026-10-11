//go:build integration

package db_test

import (
	"testing"

	"the-slot-finder/server/internal/db/dbtest"
)

// TestDatabaseManager_Postgres_Health connects to a real Postgres and checks that it reports healthy.
func TestDatabaseManager_Postgres_Health(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	if status := pdb.HealthCheck(); !status.IsHealthy {
		t.Fatalf("expected database to be healthy, got: %+v", status)
	}
}
