// Code generated from Rust node payload types. DO NOT EDIT.
package generic

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const MarkName = "markdown_generic_contracts::mark::MarkData"

type Mark struct {
}

func decodeMark(raw json.RawMessage) (any, error) {
	var value Mark
	if err := ast.DecodeFields(raw, &value, nil, nil, nil); err != nil {
		return nil, err
	}

	return value, nil
}
