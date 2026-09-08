// Code generated from Rust node payload types. DO NOT EDIT.
package commonmark

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const HeadingName = "markdown_commonmark_contracts::nodes::Heading"

type Heading struct {
	Level uint8 `json:"level"`
}

func decodeHeading(raw json.RawMessage) (any, error) {
	var value Heading
	if err := ast.DecodeFields(raw, &value, []string{"level"}, nil, nil); err != nil {
		return nil, err
	}

	return value, nil
}
