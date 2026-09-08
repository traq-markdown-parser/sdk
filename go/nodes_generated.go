// Code generated from Rust contracts. DO NOT EDIT.
package markdown

import (
	"encoding/json"
	"fmt"
)

type Span struct {
	Start uint32 `json:"start"`
	End   uint32 `json:"end"`
}
type Document struct {
	Source   string `json:"source"`
	Children []Node `json:"children"`
}
type Payload interface{ nodePayload() }
type Node struct {
	Kind     string  `json:"kind"`
	Span     Span    `json:"span"`
	Data     Payload `json:"data"`
	Children []Node  `json:"children,omitempty"`
}

const BlockquoteName = "markdown_commonmark_contracts::nodes::Blockquote"

type Blockquote struct {
}

func (*Blockquote) nodePayload() {}

const CodeBlockName = "markdown_commonmark_contracts::nodes::CodeBlock"

type CodeBlock struct {
	Fenced  bool   `json:"fenced"`
	Info    string `json:"info"`
	Literal string `json:"literal"`
}

func (*CodeBlock) nodePayload() {}

const EmphasisName = "markdown_commonmark_contracts::nodes::Emphasis"

type Emphasis struct {
}

func (*Emphasis) nodePayload() {}

const HardbreakName = "markdown_commonmark_contracts::nodes::Hardbreak"

type Hardbreak struct {
}

func (*Hardbreak) nodePayload() {}

const HeadingName = "markdown_commonmark_contracts::nodes::Heading"

type Heading struct {
	Level uint8 `json:"level"`
}

func (*Heading) nodePayload() {}

const HtmlBlockName = "markdown_commonmark_contracts::nodes::HtmlBlock"

type HtmlBlock struct {
	Literal string `json:"literal"`
}

func (*HtmlBlock) nodePayload() {}

const HtmlInlineName = "markdown_commonmark_contracts::nodes::HtmlInline"

type HtmlInline struct {
	Literal string `json:"literal"`
}

func (*HtmlInline) nodePayload() {}

const ImageName = "markdown_commonmark_contracts::nodes::Image"

type Image struct {
	Destination string  `json:"destination"`
	LabelSource string  `json:"label_source"`
	Title       *string `json:"title"`
}

func (*Image) nodePayload() {}

const InlineCodeName = "markdown_commonmark_contracts::nodes::InlineCode"

type InlineCode struct {
	Literal string `json:"literal"`
}

func (*InlineCode) nodePayload() {}

const LinkName = "markdown_commonmark_contracts::nodes::Link"

type Link struct {
	Destination string  `json:"destination"`
	Form        string  `json:"form"`
	Title       *string `json:"title"`
}

func (*Link) nodePayload() {}

const ListName = "markdown_commonmark_contracts::nodes::List"

type List struct {
	Ordered bool   `json:"ordered"`
	Start   uint32 `json:"start"`
	Tight   bool   `json:"tight"`
}

func (*List) nodePayload() {}

const ListItemName = "markdown_commonmark_contracts::nodes::ListItem"

type ListItem struct {
	Marker string `json:"marker"`
}

func (*ListItem) nodePayload() {}

const ParagraphName = "markdown_commonmark_contracts::nodes::Paragraph"

type Paragraph struct {
}

func (*Paragraph) nodePayload() {}

const SoftbreakName = "markdown_commonmark_contracts::nodes::Softbreak"

type Softbreak struct {
}

func (*Softbreak) nodePayload() {}

const StrongName = "markdown_commonmark_contracts::nodes::Strong"

type Strong struct {
}

func (*Strong) nodePayload() {}

const TextName = "markdown_commonmark_contracts::nodes::Text"

type Text struct {
	Value string `json:"value"`
}

func (*Text) nodePayload() {}

const ThematicBreakName = "markdown_commonmark_contracts::nodes::ThematicBreak"

type ThematicBreak struct {
	Marker string `json:"marker"`
}

func (*ThematicBreak) nodePayload() {}

const MarkName = "markdown_generic_contracts::mark::MarkData"

type Mark struct {
}

func (*Mark) nodePayload() {}

const BlockMathName = "markdown_generic_contracts::math::BlockMathData"

type BlockMath struct {
	Tex string `json:"tex"`
}

func (*BlockMath) nodePayload() {}

const InlineMathName = "markdown_generic_contracts::math::InlineMathData"

type InlineMath struct {
	Tex string `json:"tex"`
}

func (*InlineMath) nodePayload() {}

const StrikethroughName = "markdown_generic_contracts::strikethrough::StrikethroughData"

type Strikethrough struct {
}

func (*Strikethrough) nodePayload() {}

const CellName = "markdown_generic_contracts::table::CellData"

type Cell struct {
	Alignment *string `json:"alignment"`
}

func (*Cell) nodePayload() {}

const RowName = "markdown_generic_contracts::table::RowData"

type Row struct {
	Header bool `json:"header"`
}

func (*Row) nodePayload() {}

const TableName = "markdown_generic_contracts::table::TableData"

type Table struct {
}

func (*Table) nodePayload() {}

const BlankLineName = "markdown_trap_contracts::compat::BlankLineData"

type BlankLine struct {
}

func (*BlankLine) nodePayload() {}

const ReferenceName = "markdown_trap_contracts::reference::ReferenceData"

type Reference struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

func (*Reference) nodePayload() {}

const SpoilerName = "markdown_trap_contracts::spoiler::SpoilerData"

type Spoiler struct {
}

func (*Spoiler) nodePayload() {}

const StampName = "markdown_trap_contracts::stamp::StampData"

type Stamp struct {
	Literal string `json:"literal"`
}

func (*Stamp) nodePayload() {}
func (n *Node) UnmarshalJSON(raw []byte) error {
	var wire struct {
		Kind     string          `json:"kind"`
		Span     Span            `json:"span"`
		Data     json.RawMessage `json:"data"`
		Children []Node          `json:"children"`
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		return err
	}
	var payload Payload
	switch wire.Kind {
	case BlockquoteName:
		payload = &Blockquote{}
	case CodeBlockName:
		payload = &CodeBlock{}
	case EmphasisName:
		payload = &Emphasis{}
	case HardbreakName:
		payload = &Hardbreak{}
	case HeadingName:
		payload = &Heading{}
	case HtmlBlockName:
		payload = &HtmlBlock{}
	case HtmlInlineName:
		payload = &HtmlInline{}
	case ImageName:
		payload = &Image{}
	case InlineCodeName:
		payload = &InlineCode{}
	case LinkName:
		payload = &Link{}
	case ListName:
		payload = &List{}
	case ListItemName:
		payload = &ListItem{}
	case ParagraphName:
		payload = &Paragraph{}
	case SoftbreakName:
		payload = &Softbreak{}
	case StrongName:
		payload = &Strong{}
	case TextName:
		payload = &Text{}
	case ThematicBreakName:
		payload = &ThematicBreak{}
	case MarkName:
		payload = &Mark{}
	case BlockMathName:
		payload = &BlockMath{}
	case InlineMathName:
		payload = &InlineMath{}
	case StrikethroughName:
		payload = &Strikethrough{}
	case CellName:
		payload = &Cell{}
	case RowName:
		payload = &Row{}
	case TableName:
		payload = &Table{}
	case BlankLineName:
		payload = &BlankLine{}
	case ReferenceName:
		payload = &Reference{}
	case SpoilerName:
		payload = &Spoiler{}
	case StampName:
		payload = &Stamp{}
	default:
		return fmt.Errorf("unsupported Rust node: %s", wire.Kind)
	}
	if err := json.Unmarshal(wire.Data, payload); err != nil {
		return err
	}
	*n = Node{wire.Kind, wire.Span, payload, wire.Children}
	return nil
}
