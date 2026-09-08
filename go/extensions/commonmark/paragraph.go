// Code generated from Rust node payload types. DO NOT EDIT.
package commonmark

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const ParagraphName = "markdown_commonmark_contracts::nodes::Paragraph"

type Paragraph struct {
}

func decodeParagraph(raw json.RawMessage) (any, error) {
	var value Paragraph
	if err := ast.DecodeFields(raw, &value, nil, nil, nil); err != nil {
		return nil, err
	}

	return value, nil
}
