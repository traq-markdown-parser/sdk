// Generated from Rust node contracts. Do not edit.
import type { Blockquote } from './Blockquote.js';
import type { CodeBlock } from './CodeBlock.js';
import type { Emphasis } from './Emphasis.js';
import type { Hardbreak } from './Hardbreak.js';
import type { Heading } from './Heading.js';
import type { HtmlBlock } from './HtmlBlock.js';
import type { HtmlInline } from './HtmlInline.js';
import type { Image } from './Image.js';
import type { InlineCode } from './InlineCode.js';
import type { Link } from './Link.js';
import type { List } from './List.js';
import type { ListItem } from './ListItem.js';
import type { Paragraph } from './Paragraph.js';
import type { Softbreak } from './Softbreak.js';
import type { Strong } from './Strong.js';
import type { Text } from './Text.js';
import type { ThematicBreak } from './ThematicBreak.js';
export type { Blockquote } from './Blockquote.js';
export type { CodeBlock } from './CodeBlock.js';
export type { Emphasis } from './Emphasis.js';
export type { Hardbreak } from './Hardbreak.js';
export type { Heading } from './Heading.js';
export type { HtmlBlock } from './HtmlBlock.js';
export type { HtmlInline } from './HtmlInline.js';
export type { Image } from './Image.js';
export type { InlineCode } from './InlineCode.js';
export type { Link } from './Link.js';
export type { LinkForm } from './LinkForm.js';
export type { List } from './List.js';
export type { ListItem } from './ListItem.js';
export type { Paragraph } from './Paragraph.js';
export type { Softbreak } from './Softbreak.js';
export type { Strong } from './Strong.js';
export type { Text } from './Text.js';
export type { ThematicBreak } from './ThematicBreak.js';
export type NodeKind =
  | { kind: "markdown_commonmark_contracts::nodes::Blockquote"; data: Blockquote }
  | { kind: "markdown_commonmark_contracts::nodes::CodeBlock"; data: CodeBlock }
  | { kind: "markdown_commonmark_contracts::nodes::Emphasis"; data: Emphasis }
  | { kind: "markdown_commonmark_contracts::nodes::Hardbreak"; data: Hardbreak }
  | { kind: "markdown_commonmark_contracts::nodes::Heading"; data: Heading }
  | { kind: "markdown_commonmark_contracts::nodes::HtmlBlock"; data: HtmlBlock }
  | { kind: "markdown_commonmark_contracts::nodes::HtmlInline"; data: HtmlInline }
  | { kind: "markdown_commonmark_contracts::nodes::Image"; data: Image }
  | { kind: "markdown_commonmark_contracts::nodes::InlineCode"; data: InlineCode }
  | { kind: "markdown_commonmark_contracts::nodes::Link"; data: Link }
  | { kind: "markdown_commonmark_contracts::nodes::List"; data: List }
  | { kind: "markdown_commonmark_contracts::nodes::ListItem"; data: ListItem }
  | { kind: "markdown_commonmark_contracts::nodes::Paragraph"; data: Paragraph }
  | { kind: "markdown_commonmark_contracts::nodes::Softbreak"; data: Softbreak }
  | { kind: "markdown_commonmark_contracts::nodes::Strong"; data: Strong }
  | { kind: "markdown_commonmark_contracts::nodes::Text"; data: Text }
  | { kind: "markdown_commonmark_contracts::nodes::ThematicBreak"; data: ThematicBreak };
// Generated from Rust node payload types. Do not edit.
import {fields,string,boolean,nullable,oneOf} from "../../core/validation/fields.js"
export const names = Object.freeze({"Blockquote":"markdown_commonmark_contracts::nodes::Blockquote","CodeBlock":"markdown_commonmark_contracts::nodes::CodeBlock","Emphasis":"markdown_commonmark_contracts::nodes::Emphasis","Hardbreak":"markdown_commonmark_contracts::nodes::Hardbreak","Heading":"markdown_commonmark_contracts::nodes::Heading","HtmlBlock":"markdown_commonmark_contracts::nodes::HtmlBlock","HtmlInline":"markdown_commonmark_contracts::nodes::HtmlInline","Image":"markdown_commonmark_contracts::nodes::Image","InlineCode":"markdown_commonmark_contracts::nodes::InlineCode","Link":"markdown_commonmark_contracts::nodes::Link","List":"markdown_commonmark_contracts::nodes::List","ListItem":"markdown_commonmark_contracts::nodes::ListItem","Paragraph":"markdown_commonmark_contracts::nodes::Paragraph","Softbreak":"markdown_commonmark_contracts::nodes::Softbreak","Strong":"markdown_commonmark_contracts::nodes::Strong","Text":"markdown_commonmark_contracts::nodes::Text","ThematicBreak":"markdown_commonmark_contracts::nodes::ThematicBreak"} as const)
const validators = new Map<string, (data: unknown) => boolean>([
  ["markdown_commonmark_contracts::nodes::Blockquote",value => fields(value,{},{})],
  ["markdown_commonmark_contracts::nodes::CodeBlock",value => fields(value,{"fenced":boolean,"info":string,"literal":string},{})],
  ["markdown_commonmark_contracts::nodes::Emphasis",value => fields(value,{},{})],
  ["markdown_commonmark_contracts::nodes::Hardbreak",value => fields(value,{},{})],
  ["markdown_commonmark_contracts::nodes::Heading",value => fields(value,{"level":(value) => typeof value === "number" && Number.isInteger(value) && value >= 0 && value <= 255},{})],
  ["markdown_commonmark_contracts::nodes::HtmlBlock",value => fields(value,{"literal":string},{})],
  ["markdown_commonmark_contracts::nodes::HtmlInline",value => fields(value,{"literal":string},{})],
  ["markdown_commonmark_contracts::nodes::Image",value => fields(value,{"destination":string,"label_source":string,"title":nullable(string)},{})],
  ["markdown_commonmark_contracts::nodes::InlineCode",value => fields(value,{"literal":string},{})],
  ["markdown_commonmark_contracts::nodes::Link",value => fields(value,{"destination":string,"form":oneOf("explicit","autolink","linkify"),"title":nullable(string)},{})],
  ["markdown_commonmark_contracts::nodes::List",value => fields(value,{"ordered":boolean,"start":(value) => typeof value === "number" && Number.isInteger(value) && value >= 0 && value <= 4294967295,"tight":boolean},{})],
  ["markdown_commonmark_contracts::nodes::ListItem",value => fields(value,{"marker":string},{})],
  ["markdown_commonmark_contracts::nodes::Paragraph",value => fields(value,{},{})],
  ["markdown_commonmark_contracts::nodes::Softbreak",value => fields(value,{},{})],
  ["markdown_commonmark_contracts::nodes::Strong",value => fields(value,{},{})],
  ["markdown_commonmark_contracts::nodes::Text",value => fields(value,{"value":string},{})],
  ["markdown_commonmark_contracts::nodes::ThematicBreak",value => fields(value,{"marker":string},{})],
])
export const nodes: ReadonlyMap<string, (data: unknown) => boolean> = new Map(validators)
// Check this payload only; children can still contain unknown nodes.
export function isKnownNode<T extends {kind: string; data: unknown}>(node: T): node is T & NodeKind {
  return validators.get(node.kind)?.(node.data) ?? false
}
