// Code generated from Rust node payload types. DO NOT EDIT.
package commonmark

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const InlineCodeName = "markdown_commonmark_contracts::nodes::InlineCode"

type InlineCode struct {
	Literal string `json:"literal"`
}

func decodeInlineCode(raw json.RawMessage) (any, error) {
	var value InlineCode
	if err := ast.DecodeFields(raw, &value, []string{"literal"}, nil, nil); err != nil {
		return nil, err
	}

	return value, nil
}
