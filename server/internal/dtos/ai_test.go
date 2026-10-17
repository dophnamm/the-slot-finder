package dtos

import (
	"encoding/json"
	"strings"
	"testing"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/models"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDraftSummaryResponse(t *testing.T) {
	got := NewDraftSummaryResponse(models.DraftSummary{Summary: "All staffed.", Source: constants.SummarySourceTemplate})
	raw, err := json.Marshal(got)
	require.NoError(t, err)
	assert.JSONEq(t, `{"summary":"All staffed.","source":"template"}`, string(raw))
}

func TestIntakeRequestValidation(t *testing.T) {
	terms := func(n, size int) []string {
		out := make([]string, n)
		for i := range out {
			out[i] = strings.Repeat("a", size)
		}
		return out
	}
	tests := []struct {
		name    string
		req     IntakeRequest
		wantErr bool
	}{
		{name: "valid", req: IntakeRequest{Text: "brief", MaskTerms: terms(10, 100)}},
		{name: "empty text", req: IntakeRequest{Text: "  "}, wantErr: true},
		{name: "max text", req: IntakeRequest{Text: strings.Repeat("a", constants.IntakeMaxTextChars)}},
		{name: "text too long", req: IntakeRequest{Text: strings.Repeat("a", constants.IntakeMaxTextChars+1)}, wantErr: true},
		{name: "too many terms", req: IntakeRequest{Text: "x", MaskTerms: terms(11, 1)}, wantErr: true},
		{name: "term too long", req: IntakeRequest{Text: "x", MaskTerms: terms(1, 101)}, wantErr: true},
	}
	v := validator.New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.req.Normalize()
			err := v.Struct(tt.req)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
