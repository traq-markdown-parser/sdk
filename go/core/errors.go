package core

import (
	"encoding/json"
	"fmt"

	"github.com/traq-markdown-parser/sdk/go/ast"
)

type Error struct {
	Code     string `json:"code"`
	Resource string `json:"resource,omitempty"`
}

func (e *Error) Error() string {
	if e.Resource != "" {
		return e.Code + ": " + e.Resource
	}
	return e.Code
}

type BuildError struct {
	Code    string `json:"code"`
	Element string `json:"element,omitempty"`
	Scope   string `json:"scope,omitempty"`
	Name    string `json:"name,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

func (e *BuildError) Error() string {
	switch e.Code {
	case "duplicate", "missing":
		return e.Code + ": " + e.Element
	case "duplicate_name":
		return "duplicate name " + e.Name + " in " + e.Scope
	default:
		return e.Code + ": " + e.Reason
	}
}
func decodeParseError(raw json.RawMessage) error {
	var failure Error
	if err := json.Unmarshal(raw, &failure); err != nil {
		return err
	}
	fields, ok := map[string][]string{"invalid_utf8": {}, "internal_error": {}, "resource_limit": {"resource"}}[failure.Code]
	if !ok {
		return fmt.Errorf("invalid parse error")
	}
	if err := ast.DecodeFields(raw, &failure, append([]string{"code"}, fields...), nil, nil); err != nil {
		return err
	}
	return &failure
}
func decodeBuildError(raw json.RawMessage) error {
	var failure BuildError
	if err := json.Unmarshal(raw, &failure); err != nil {
		return err
	}
	fields, ok := map[string][]string{"duplicate": {"element"}, "missing": {"element"}, "duplicate_name": {"scope", "name"}, "invalid_definition": {"reason"}}[failure.Code]
	if !ok {
		return fmt.Errorf("invalid build error")
	}
	if err := ast.DecodeFields(raw, &failure, append([]string{"code"}, fields...), nil, nil); err != nil {
		return err
	}
	return &failure
}
