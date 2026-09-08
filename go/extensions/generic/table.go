// Code generated from Rust node payload types. DO NOT EDIT.
package generic

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const TableName = "markdown_generic_contracts::table::TableData"

type Table struct {
}

func decodeTable(raw json.RawMessage) (any, error) {
	var value Table
	if err := ast.DecodeFields(raw, &value, nil, nil, nil); err != nil {
		return nil, err
	}

	return value, nil
}
