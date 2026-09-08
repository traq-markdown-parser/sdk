// Code generated from Rust node payload types. DO NOT EDIT.
package commonmark

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const EmphasisName = "markdown_commonmark_contracts::nodes::Emphasis"

type Emphasis struct {
}

func decodeEmphasis(raw json.RawMessage) (any, error) {
	var value Emphasis
	if err := ast.DecodeFields(raw, &value, nil, nil, nil); err != nil {
		return nil, err
	}

	return value, nil
}
