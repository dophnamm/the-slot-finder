package handlers

import (
	"os"
	"testing"

	"the-slot-finder/server/internal/logger"

	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	logger.Log = zap.NewNop()
	logger.Sugar = logger.Log.Sugar()
	os.Exit(m.Run())
}
