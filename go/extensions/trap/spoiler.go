// Code generated from Rust node payload types. DO NOT EDIT.
package trap

import (
	"encoding/json"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

const SpoilerName = "markdown_trap_contracts::spoiler::SpoilerData"

type Spoiler struct {
}

func decodeSpoiler(raw json.RawMessage) (any, error) {
	var value Spoiler
	if err := ast.DecodeFields(raw, &value, nil, nil, nil); err != nil {
		return nil, err
	}

	return value, nil
}
