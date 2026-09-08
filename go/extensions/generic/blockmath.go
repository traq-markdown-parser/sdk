// Code generated from Rust node payload types. DO NOT EDIT.
package generic

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const BlockMathName = "markdown_generic_contracts::math::BlockMathData"

type BlockMath struct {
	Tex string `json:"tex"`
}

func decodeBlockMath(raw json.RawMessage) (any, error) {
	var value BlockMath
	if err := ast.DecodeFields(raw, &value, []string{"tex"}, nil, nil); err != nil {
		return nil, err
	}

	return value, nil
}
