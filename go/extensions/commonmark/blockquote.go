// Code generated from Rust node payload types. DO NOT EDIT.
package commonmark

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const BlockquoteName = "markdown_commonmark_contracts::nodes::Blockquote"

type Blockquote struct {
}

func decodeBlockquote(raw json.RawMessage) (any, error) {
	var value Blockquote
	if err := ast.DecodeFields(raw, &value, nil, nil, nil); err != nil {
		return nil, err
	}

	return value, nil
}
