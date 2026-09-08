// Code generated from Rust node payload types. DO NOT EDIT.
package trap

import "github.com/traq-markdown-parser/sdk/go/ast"

func Registry() ast.Registry {
	return ast.Registry{
		BlankLineName: decodeBlankLine,
		ReferenceName: decodeReference,
		SpoilerName:   decodeSpoiler,
		StampName:     decodeStamp,
	}
}
