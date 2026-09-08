// Code generated from Rust node payload types. DO NOT EDIT.
package trap

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const StampName = "markdown_trap_contracts::stamp::StampData"

type Stamp struct {
	Literal string `json:"literal"`
}

func decodeStamp(raw json.RawMessage) (any, error) {
	var value Stamp
	if err := ast.DecodeFields(raw, &value, []string{"literal"}, nil, nil); err != nil {
		return nil, err
	}

	return value, nil
}
