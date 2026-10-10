package main

import (
	"testing"

	"the-slot-finder/server/internal/config"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func TestAppOptions_Validate(t *testing.T) {
	require.NoError(t, fx.ValidateApp(appOptions(&config.Config{})...))
}
