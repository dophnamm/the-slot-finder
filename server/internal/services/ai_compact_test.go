package services

import (
	"encoding/json"
	"strings"
	"testing"

	"the-slot-finder/server/internal/constants"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompactJSON(t *testing.T) {
	long := `{"people":[` + strings.Repeat(`{"name":"Ann","note":null,"hours":10},`, 40) + `{"name":"Zed","note":"","hours":1}],"week":"2026-W42"}`
	tests := []struct {
		name    string
		in      string
		limit   int
		check   func(t *testing.T, out string)
		wantCut bool
	}{
		{name: "fits unchanged", in: `{"a":null,"b":1}`, limit: 100,
			check: func(t *testing.T, out string) { assert.Equal(t, `{"a":null,"b":1}`, out) }},
		{name: "nulls and empty values dropped", in: `{"a":null,"b":"","c":[],"d":{},"e":0,"f":"x"}`, limit: 15,
			check: func(t *testing.T, out string) { assert.JSONEq(t, `{"e":0,"f":"x"}`, out) }},
		{name: "arrays trimmed with a marker", in: long, limit: 300,
			check: func(t *testing.T, out string) {
				assert.LessOrEqual(t, len(out), 300)
				var m map[string]any
				require.NoError(t, json.Unmarshal([]byte(out), &m))
				people := m["people"].([]any)
				assert.NotEmpty(t, people)
				assert.EqualValues(t, 41-len(people), m["people"+aiOmittedSuffix])
				assert.Equal(t, "2026-W42", m["week"])
			}},
		{name: "top-level list is wrapped", in: `[` + strings.Repeat(`{"n":1},`, 50) + `{"n":2}]`, limit: 80,
			check: func(t *testing.T, out string) {
				var m map[string]any
				require.NoError(t, json.Unmarshal([]byte(out), &m))
				assert.Contains(t, m, aiOmittedSuffix)
				assert.Contains(t, m, aiItemsKey)
			}},
		{name: "falls back to a cut with the marker", in: `{"text":"` + strings.Repeat("x", 200) + `"}`, limit: 50, wantCut: true,
			check: func(t *testing.T, out string) { assert.Len(t, out, 50+len(constants.LLMToolCutMarker)) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := compactJSON([]byte(tt.in), tt.limit)
			assert.Equal(t, tt.wantCut, strings.HasSuffix(out, constants.LLMToolCutMarker))
			tt.check(t, out)
		})
	}
}
