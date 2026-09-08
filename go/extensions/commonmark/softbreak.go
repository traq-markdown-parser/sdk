// Code generated from Rust node payload types. DO NOT EDIT.
package commonmark

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const SoftbreakName = "markdown_commonmark_contracts::nodes::Softbreak"

type Softbreak struct {
}

func decodeSoftbreak(raw json.RawMessage) (any, error) {
	var value Softbreak
	if err := ast.DecodeFields(raw, &value, nil, nil, nil); err != nil {
		return nil, err
	}

	return value, nil
}
