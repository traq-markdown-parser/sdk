// Code generated from Rust node payload types. DO NOT EDIT.
package commonmark

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const HtmlInlineName = "markdown_commonmark_contracts::nodes::HtmlInline"

type HtmlInline struct {
	Literal string `json:"literal"`
}

func decodeHtmlInline(raw json.RawMessage) (any, error) {
	var value HtmlInline
	if err := ast.DecodeFields(raw, &value, []string{"literal"}, nil, nil); err != nil {
		return nil, err
	}

	return value, nil
}
