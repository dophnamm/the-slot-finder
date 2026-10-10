package dtos

import (
	"bytes"
	"encoding/json"
)

// jsonNull is the JSON literal that clears an Optional field.
var jsonNull = []byte("null")

// Optional is a PATCH field that tells "not sent" apart from "sent as null".
// Set is true when the key was present; Value is nil when it was null.
type Optional[T any] struct {
	Set   bool
	Value *T
}

// UnmarshalJSON records that the key was present, and its value unless it was null.
func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.Set = true
	if bytes.Equal(bytes.TrimSpace(data), jsonNull) {
		o.Value = nil
		return nil
	}
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	o.Value = &value
	return nil
}
