// Code generated from Rust node payload types. DO NOT EDIT.
package commonmark

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const StrongName = "markdown_commonmark_contracts::nodes::Strong"

type Strong struct {
}

func decodeStrong(raw json.RawMessage) (any, error) {
	var value Strong
	if err := ast.DecodeFields(raw, &value, nil, nil, nil); err != nil {
		return nil, err
	}

	return value, nil
}
