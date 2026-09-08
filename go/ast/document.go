package ast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"unicode/utf8"
)

func Decode(raw json.RawMessage, registry Registry) (*Document, error) {
	if len(raw) > 1048576 || !utf8.Valid(raw) {
		return nil, fmt.Errorf("invalid document encoding or size")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, fmt.Errorf("trailing JSON")
	}
	fields, err := objectFields(value, []string{"source", "children"}, nil, nil)
	if err != nil {
		return nil, err
	}
	source, ok := fields["source"].(string)
	if !ok || len(source) > 65536 {
		return nil, fmt.Errorf("invalid source")
	}
	children, ok := fields["children"].([]any)
	if !ok {
		return nil, fmt.Errorf("invalid document children")
	}
	count := 0
	var visit func([]any, Span, int) ([]Node, error)
	visit = func(values []any, parent Span, depth int) ([]Node, error) {
		if depth > 64 && len(values) > 0 {
			return nil, fmt.Errorf("depth limit")
		}
		nodes := make([]Node, 0, len(values))
		for _, value := range values {
			count++
			if count > 16384 {
				return nil, fmt.Errorf("node limit")
			}
			node, payload, children, err := readNode(value)
			if err != nil {
				return nil, err
			}
			span := node.Span
			boundary := func(pos uint32) bool {
				return pos == uint32(len(source)) || pos < uint32(len(source)) && utf8.RuneStart(source[pos])
			}
			if span.Start > span.End || span.Start < parent.Start || span.End > parent.End || !boundary(span.Start) || !boundary(span.End) {
				return nil, fmt.Errorf("invalid source span")
			}
			decode := registry[node.Kind]
			if decode == nil {
				return nil, fmt.Errorf("missing node decoder %s", node.Kind)
			}
			if node.Payload, err = decode(payload); err != nil {
				return nil, fmt.Errorf("%s: %w", node.Kind, err)
			}
			if node.Children, err = visit(children, span, depth+1); err != nil {
				return nil, err
			}
			nodes = append(nodes, node)
		}
		return nodes, nil
	}

	nodes, err := visit(children, Span{End: uint32(len(source))}, 1)
	if err != nil {
		return nil, err
	}
	return &Document{Source: source, Children: nodes}, nil
}
