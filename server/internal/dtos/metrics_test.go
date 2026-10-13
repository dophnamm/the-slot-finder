package dtos

import (
	"encoding/json"
	"testing"

	"the-slot-finder/server/internal/metrics"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProgressResponse_ActualHours(t *testing.T) {
	actual, burn := 100.0, 25.0
	got := NewProgressResponse(metrics.Progress{ActualHours: &actual, BurnPercent: &burn})
	require.NotNil(t, got.ActualHours)
	assert.InDelta(t, 100, *got.ActualHours, 1e-9)
	require.NotNil(t, got.BurnPercent)
	assert.InDelta(t, 25, *got.BurnPercent, 1e-9)

	body, err := json.Marshal(NewProgressResponse(metrics.Progress{}))
	require.NoError(t, err)
	assert.Contains(t, string(body), `"actual_hours":null`)
	assert.Contains(t, string(body), `"burn_percent":null`)
}
