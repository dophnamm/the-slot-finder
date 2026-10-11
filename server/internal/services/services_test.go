package services

import (
	"context"
	"os"
	"testing"

	"the-slot-finder/server/internal/logger"

	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	// logger.Init is only called from cmd/server; tests need a non-nil logger.
	logger.Log = zap.NewNop()
	logger.Sugar = logger.Log.Sugar()
	os.Exit(m.Run())
}

// recordedEvents keeps the change events a service records; fail makes Record return it.
type recordedEvents struct {
	events []ChangeInput
	fail   error
}

func (r *recordedEvents) Record(_ context.Context, in ChangeInput) error {
	if r.fail != nil {
		return r.fail
	}
	r.events = append(r.events, in)
	return nil
}
