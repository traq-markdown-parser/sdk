// Code generated from Rust node payload types. DO NOT EDIT.
package trap

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const BlankLineName = "markdown_trap_contracts::compat::BlankLineData"

type BlankLine struct {
}

func decodeBlankLine(raw json.RawMessage) (any, error) {
	var value BlankLine
	if err := ast.DecodeFields(raw, &value, nil, nil, nil); err != nil {
		return nil, err
	}

	return value, nil
}
