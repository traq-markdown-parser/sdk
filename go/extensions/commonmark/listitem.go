// Code generated from Rust node payload types. DO NOT EDIT.
package commonmark

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const ListItemName = "markdown_commonmark_contracts::nodes::ListItem"

type ListItem struct {
	Marker string `json:"marker"`
}

func decodeListItem(raw json.RawMessage) (any, error) {
	var value ListItem
	if err := ast.DecodeFields(raw, &value, []string{"marker"}, nil, nil); err != nil {
		return nil, err
	}

	return value, nil
}
