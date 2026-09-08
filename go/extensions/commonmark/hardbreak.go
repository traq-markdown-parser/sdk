// Code generated from Rust node payload types. DO NOT EDIT.
package commonmark

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const HardbreakName = "markdown_commonmark_contracts::nodes::Hardbreak"

type Hardbreak struct {
}

func decodeHardbreak(raw json.RawMessage) (any, error) {
	var value Hardbreak
	if err := ast.DecodeFields(raw, &value, nil, nil, nil); err != nil {
		return nil, err
	}

	return value, nil
}
