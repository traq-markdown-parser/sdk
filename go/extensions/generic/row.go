// Code generated from Rust node payload types. DO NOT EDIT.
package generic

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const RowName = "markdown_generic_contracts::table::RowData"

type Row struct {
	Header bool `json:"header"`
}

func decodeRow(raw json.RawMessage) (any, error) {
	var value Row
	if err := ast.DecodeFields(raw, &value, []string{"header"}, nil, nil); err != nil {
		return nil, err
	}

	return value, nil
}
