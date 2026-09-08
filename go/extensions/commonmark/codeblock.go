// Code generated from Rust node payload types. DO NOT EDIT.
package commonmark

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const CodeBlockName = "markdown_commonmark_contracts::nodes::CodeBlock"

type CodeBlock struct {
	Fenced  bool   `json:"fenced"`
	Info    string `json:"info"`
	Literal string `json:"literal"`
}

func decodeCodeBlock(raw json.RawMessage) (any, error) {
	var value CodeBlock
	if err := ast.DecodeFields(raw, &value, []string{"fenced", "info", "literal"}, nil, nil); err != nil {
		return nil, err
	}

	return value, nil
}
