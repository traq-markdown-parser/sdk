// Code generated from Rust node payload types. DO NOT EDIT.
package generic

import "github.com/traq-markdown-parser/sdk/go/ast"

func Registry() ast.Registry {
	return ast.Registry{
		MarkName:          decodeMark,
		BlockMathName:     decodeBlockMath,
		InlineMathName:    decodeInlineMath,
		StrikethroughName: decodeStrikethrough,
		CellName:          decodeCell,
		RowName:           decodeRow,
		TableName:         decodeTable,
	}
}
