package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLowerTypes(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]any
		want  map[string]any
	}{
		{name: "optional string accepts null, required stays plain",
			input: map[string]any{"type": "OBJECT", "required": []string{"a"}, "properties": map[string]any{
				"a": map[string]any{"type": "STRING"},
				"b": map[string]any{"type": "STRING", "description": "d"},
				"c": map[string]any{"type": "NUMBER"},
			}},
			want: map[string]any{"type": "object", "required": []string{"a"}, "properties": map[string]any{
				"a": map[string]any{"type": "string"},
				"b": map[string]any{"type": []string{"string", "null"}, "description": "d"},
				"c": map[string]any{"type": []string{"number", "null"}},
			}}},
		{name: "no required list makes every property optional",
			input: map[string]any{"type": "OBJECT", "properties": map[string]any{"a": map[string]any{"type": "STRING"}}},
			want: map[string]any{"type": "object", "properties": map[string]any{
				"a": map[string]any{"type": []string{"string", "null"}}}}},
		{name: "nested object follows its own required list",
			input: map[string]any{"type": "OBJECT", "properties": map[string]any{
				"assumptions": map[string]any{"type": "OBJECT", "required": []string{"type"}, "properties": map[string]any{
					"type":  map[string]any{"type": "STRING"},
					"hours": map[string]any{"type": "NUMBER"},
				}}}},
			want: map[string]any{"type": "object", "properties": map[string]any{
				"assumptions": map[string]any{"type": []string{"object", "null"}, "required": []string{"type"}, "properties": map[string]any{
					"type":  map[string]any{"type": "string"},
					"hours": map[string]any{"type": []string{"number", "null"}},
				}}}}},
		{name: "enum property is left alone",
			input: map[string]any{"type": "OBJECT", "properties": map[string]any{
				"kind": map[string]any{"type": "STRING", "enum": []string{"x", "y"}}}},
			want: map[string]any{"type": "object", "properties": map[string]any{
				"kind": map[string]any{"type": "string", "enum": []string{"x", "y"}}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, lowerTypes(tt.input))
		})
	}
}

func TestLowerTypes_DoesNotMutateInput(t *testing.T) {
	input := map[string]any{"type": "OBJECT", "properties": map[string]any{"a": map[string]any{"type": "STRING"}}}
	lowerTypes(input)
	assert.Equal(t, map[string]any{"type": "OBJECT", "properties": map[string]any{"a": map[string]any{"type": "STRING"}}}, input)
}
