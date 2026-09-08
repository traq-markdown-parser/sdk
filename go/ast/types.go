// Package ast decodes a common tree using product-owned payload decoders.
package ast

import "encoding/json"

type Document struct {
	Source   string `json:"source"`
	Children []Node `json:"children"`
}
type Span struct {
	Start uint32 `json:"start"`
	End   uint32 `json:"end"`
}

// Payload contains a generated concrete value, including for CommonMark nodes.
type Node struct {
	Kind     string `json:"kind"`
	Span     Span   `json:"span"`
	Payload  any    `json:"data"`
	Children []Node `json:"children,omitempty"`
}
type Decoder func(json.RawMessage) (any, error)
type Registry map[string]Decoder
