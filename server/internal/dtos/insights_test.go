package dtos

import (
	"encoding/json"
	"testing"

	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHeatmapResponse_ActualHours(t *testing.T) {
	week, err := isoweek.Parse("2026-W42")
	require.NoError(t, err)
	worked := 26.0
	heatmap := models.Heatmap{Weeks: []isoweek.Week{week}, Rows: []models.HeatmapRow{{
		Cells: []models.HeatmapCell{{Week: week, Load: 32, ActualHours: &worked}, {Week: week, Load: 8}},
	}}}

	got := NewHeatmapResponse(heatmap)

	cells := got.Rows[0].Cells
	require.NotNil(t, cells[0].ActualHours)
	assert.InDelta(t, 26, *cells[0].ActualHours, 1e-9)
	body, err := json.Marshal(cells[1])
	require.NoError(t, err)
	assert.Contains(t, string(body), `"actual_hours":null`)
}

func TestNewHeatmapResponse_SustainedOverload(t *testing.T) {
	week, err := isoweek.Parse("2026-W44")
	require.NoError(t, err)
	heatmap := models.Heatmap{Weeks: []isoweek.Week{week}, Rows: []models.HeatmapRow{
		{Sustained: &models.OverloadRun{From: week, Weeks: 3}},
		{},
	}}

	got := NewHeatmapResponse(heatmap)

	require.NotNil(t, got.Rows[0].SustainedOverload)
	assert.Equal(t, OverloadRunResponse{FromWeek: "2026-W44", Weeks: 3}, *got.Rows[0].SustainedOverload)
	body, err := json.Marshal(got.Rows[1])
	require.NoError(t, err)
	assert.Contains(t, string(body), `"sustained_overload":null`)
}
