// Code generated from Rust node payload types. DO NOT EDIT.
package commonmark

import (
	"encoding/json"
	"fmt"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const LinkName = "markdown_commonmark_contracts::nodes::Link"

type Link struct {
	Destination string  `json:"destination"`
	Form        string  `json:"form"`
	Title       *string `json:"title"`
}

func decodeLink(raw json.RawMessage) (any, error) {
	var value Link
	if err := ast.DecodeFields(raw, &value, []string{"destination", "form", "title"}, nil, []string{"title"}); err != nil {
		return nil, err
	}
	if value.Form != "explicit" && value.Form != "autolink" && value.Form != "linkify" {
		return nil, fmt.Errorf("invalid form")
	}
	return value, nil
}
