// Code generated from Rust node payload types. DO NOT EDIT.
package commonmark

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const HtmlBlockName = "markdown_commonmark_contracts::nodes::HtmlBlock"

type HtmlBlock struct {
	Literal string `json:"literal"`
}

func decodeHtmlBlock(raw json.RawMessage) (any, error) {
	var value HtmlBlock
	if err := ast.DecodeFields(raw, &value, []string{"literal"}, nil, nil); err != nil {
		return nil, err
	}

	return value, nil
}
