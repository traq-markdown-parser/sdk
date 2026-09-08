// Code generated from Rust node payload types. DO NOT EDIT.
package commonmark

import "github.com/traq-markdown-parser/sdk/go/ast"

func Registry() ast.Registry {
	return ast.Registry{
		BlockquoteName:    decodeBlockquote,
		CodeBlockName:     decodeCodeBlock,
		EmphasisName:      decodeEmphasis,
		HardbreakName:     decodeHardbreak,
		HeadingName:       decodeHeading,
		HtmlBlockName:     decodeHtmlBlock,
		HtmlInlineName:    decodeHtmlInline,
		ImageName:         decodeImage,
		InlineCodeName:    decodeInlineCode,
		LinkName:          decodeLink,
		ListName:          decodeList,
		ListItemName:      decodeListItem,
		ParagraphName:     decodeParagraph,
		SoftbreakName:     decodeSoftbreak,
		StrongName:        decodeStrong,
		TextName:          decodeText,
		ThematicBreakName: decodeThematicBreak,
	}
}
