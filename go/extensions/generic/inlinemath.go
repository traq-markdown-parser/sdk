// Code generated from Rust node payload types. DO NOT EDIT.
package generic

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const InlineMathName = "markdown_generic_contracts::math::InlineMathData"

type InlineMath struct {
	Tex string `json:"tex"`
}

func decodeInlineMath(raw json.RawMessage) (any, error) {
	var value InlineMath
	if err := ast.DecodeFields(raw, &value, []string{"tex"}, nil, nil); err != nil {
		return nil, err
	}

	return value, nil
}
