// Code generated from Rust node payload types. DO NOT EDIT.
package commonmark

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const ListName = "markdown_commonmark_contracts::nodes::List"

type List struct {
	Ordered bool   `json:"ordered"`
	Start   uint32 `json:"start"`
	Tight   bool   `json:"tight"`
}

func decodeList(raw json.RawMessage) (any, error) {
	var value List
	if err := ast.DecodeFields(raw, &value, []string{"ordered", "start", "tight"}, nil, nil); err != nil {
		return nil, err
	}

	return value, nil
}
