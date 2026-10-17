package dtos

import (
	"encoding/json"
	"testing"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDraftSummaryResponse(t *testing.T) {
	got := NewDraftSummaryResponse(models.DraftSummary{Summary: "All staffed.", Source: constants.SummarySourceTemplate})
	raw, err := json.Marshal(got)
	require.NoError(t, err)
	assert.JSONEq(t, `{"summary":"All staffed.","source":"template"}`, string(raw))
}
