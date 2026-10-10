package dtos

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptional_UnmarshalJSON(t *testing.T) {
	type payload struct {
		Date Optional[string] `json:"date"`
	}
	tests := []struct {
		name    string
		body    string
		wantSet bool
		want    *string
	}{
		{name: "missing", body: `{}`},
		{name: "null", body: `{"date":null}`, wantSet: true},
		{name: "value", body: `{"date":"2026-10-12"}`, wantSet: true, want: new("2026-10-12")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p payload
			require.NoError(t, json.Unmarshal([]byte(tt.body), &p))
			assert.Equal(t, tt.wantSet, p.Date.Set)
			assert.Equal(t, tt.want, p.Date.Value)
		})
	}

	var p payload
	require.Error(t, json.Unmarshal([]byte(`{"date":12}`), &p))
}
