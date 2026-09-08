// Code generated from Rust node payload types. DO NOT EDIT.
package commonmark

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const ThematicBreakName = "markdown_commonmark_contracts::nodes::ThematicBreak"

type ThematicBreak struct {
	Marker string `json:"marker"`
}

func decodeThematicBreak(raw json.RawMessage) (any, error) {
	var value ThematicBreak
	if err := ast.DecodeFields(raw, &value, []string{"marker"}, nil, nil); err != nil {
		return nil, err
	}

	return value, nil
}
