// Code generated from Rust node payload types. DO NOT EDIT.
package generic

import (
	"encoding/json"
	"fmt"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const CellName = "markdown_generic_contracts::table::CellData"

type Cell struct {
	Alignment *string `json:"alignment"`
}

func decodeCell(raw json.RawMessage) (any, error) {
	var value Cell
	if err := ast.DecodeFields(raw, &value, []string{"alignment"}, nil, []string{"alignment"}); err != nil {
		return nil, err
	}
	if value.Alignment != nil && (*value.Alignment != "left" && *value.Alignment != "center" && *value.Alignment != "right") {
		return nil, fmt.Errorf("invalid alignment")
	}
	return value, nil
}
