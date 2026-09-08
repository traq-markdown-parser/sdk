package ast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"slices"
)

// DecodeFields rejects missing, extra and null required fields. Callers explicitly
// list nullable fields; encoding/json alone accepts null into scalar Go fields.
func DecodeFields(raw json.RawMessage, out any, required, optional, nullable []string) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	if fields == nil {
		return fmt.Errorf("expected object")
	}
	for _, name := range required {
		value, ok := fields[name]
		if !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) && !slices.Contains(nullable, name) {
			return fmt.Errorf("missing or null field %s", name)
		}
	}
	for name, value := range fields {
		if !slices.Contains(required, name) && !slices.Contains(optional, name) {
			return fmt.Errorf("unknown field %s", name)
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) && !slices.Contains(nullable, name) {
			return fmt.Errorf("null field %s", name)
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("trailing JSON")
	}
	return nil
}
