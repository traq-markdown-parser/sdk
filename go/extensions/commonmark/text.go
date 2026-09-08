// Code generated from Rust node payload types. DO NOT EDIT.
package commonmark

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const TextName = "markdown_commonmark_contracts::nodes::Text"

type Text struct {
	Value string `json:"value"`
}

func decodeText(raw json.RawMessage) (any, error) {
	var value Text
	if err := ast.DecodeFields(raw, &value, []string{"value"}, nil, nil); err != nil {
		return nil, err
	}

	return value, nil
}
