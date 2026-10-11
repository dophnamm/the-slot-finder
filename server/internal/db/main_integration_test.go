//go:build integration

package db

import (
	"os"
	"testing"

	"the-slot-finder/server/internal/logger"

	"go.uber.org/zap"
)

// TestMain sets a no-op logger: the database manager logs while it connects.
func TestMain(m *testing.M) {
	logger.Log = zap.NewNop()
	logger.Sugar = logger.Log.Sugar()
	os.Exit(m.Run())
}
