// Code generated from Rust node payload types. DO NOT EDIT.
package generic

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const StrikethroughName = "markdown_generic_contracts::strikethrough::StrikethroughData"

type Strikethrough struct {
}

func decodeStrikethrough(raw json.RawMessage) (any, error) {
	var value Strikethrough
	if err := ast.DecodeFields(raw, &value, nil, nil, nil); err != nil {
		return nil, err
	}

	return value, nil
}
