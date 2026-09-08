package ast

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
)

// The wire JSON is decoded once. Tree validation never reparses a parent subtree.
func objectFields(value any, required, optional, nullable []string) (map[string]any, error) {
	fields, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected object")
	}
	for _, name := range required {
		if _, ok := fields[name]; !ok {
			return nil, fmt.Errorf("missing field %s", name)
		}
	}
	for name, value := range fields {
		if !slices.Contains(required, name) && !slices.Contains(optional, name) {
			return nil, fmt.Errorf("unknown field %s", name)
		}
		if value == nil && !slices.Contains(nullable, name) {
			return nil, fmt.Errorf("null field %s", name)
		}
	}
	return fields, nil
}

func number(value any) (uint32, error) {
	n, ok := value.(json.Number)
	if !ok {
		return 0, fmt.Errorf("expected unsigned integer")
	}
	u, err := strconv.ParseUint(string(n), 10, 32)
	return uint32(u), err
}

func readNode(value any) (Node, json.RawMessage, []any, error) {
	var node Node
	fields, err := objectFields(value, []string{"kind", "span", "data"}, []string{"children"}, nil)
	if err != nil {
		return node, nil, nil, err
	}
	kind, ok := fields["kind"].(string)
	if !ok || kind == "" {
		return node, nil, nil, fmt.Errorf("invalid node kind")
	}
	node.Kind = kind
	span, err := objectFields(fields["span"], []string{"start", "end"}, nil, nil)
	if err != nil {
		return node, nil, nil, err
	}
	if node.Span.Start, err = number(span["start"]); err != nil {
		return node, nil, nil, err
	}
	if node.Span.End, err = number(span["end"]); err != nil {
		return node, nil, nil, err
	}
	if _, ok := fields["data"].(map[string]any); !ok {
		return node, nil, nil, fmt.Errorf("expected payload object")
	}
	payload, err := json.Marshal(fields["data"])
	if err != nil {
		return node, nil, nil, err
	}
	var children []any
	if value, present := fields["children"]; present {
		children, ok = value.([]any)
		if !ok {
			return node, nil, nil, fmt.Errorf("invalid children")
		}
	}
	return node, payload, children, nil
}
