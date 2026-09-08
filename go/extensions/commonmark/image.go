// Code generated from Rust node payload types. DO NOT EDIT.
package commonmark

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const ImageName = "markdown_commonmark_contracts::nodes::Image"

type Image struct {
	Destination string  `json:"destination"`
	LabelSource string  `json:"label_source"`
	Title       *string `json:"title"`
}

func decodeImage(raw json.RawMessage) (any, error) {
	var value Image
	if err := ast.DecodeFields(raw, &value, []string{"destination", "label_source", "title"}, nil, []string{"title"}); err != nil {
		return nil, err
	}

	return value, nil
}
