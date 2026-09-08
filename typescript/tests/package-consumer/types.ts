import { Plugin as Declaration } from "@traq-markdown-parser/ts/definitions";
import { names } from "@traq-markdown-parser/ts/nodes";
import * as commonNodes from "@traq-markdown-parser/ts/commonmark/nodes";
import {
  MarkdownParseError,
  type ReferenceData,
} from "@traq-markdown-parser/ts";
import {
  loadRuntime,
  Plugin,
  isKnownNode,
  type Node,
  type ParserOptions,
} from "@traq-markdown-parser/ts";

const runtime = await loadRuntime(new Uint8Array());
const grammar = runtime.presets.traq.v1
  .toBuilder()
  .remove(runtime.plugins.generic.math)
  .build();
const core = runtime.parser(grammar);
const generic = Declaration.group("custom");
const plugin = new Plugin(generic.new("math")).add(
  runtime.plugins.generic.math.inlineRules[0],
);
runtime
  .builder()
  .add(runtime.plugins.commonmark.core)
  .add(plugin)
  .build()
  .dispose();
const math = runtime.plugins.generic.math;
// @ts-expect-error Ordering is restricted to the same phase
runtime.builder().before(math.blockRules[0], math.inlineRules[0]);
grammar.dispose();
const doc = core.parse("**check**");
const source: string = doc.source;
const kind: string | undefined = doc.children[0]?.kind;
const error: string = new MarkdownParseError({ code: "internal_error" }).detail
  .code;
const reference: ReferenceData = { type: "user", id: "u", label: "@u" };
for (const parser of [
  core,
  runtime.parser(grammar, { allowUnknownNodes: false }),
]) {
  const node = parser.parseInline("[link](https://example.com)").children[0];
  if (node.kind === names.Link) {
    const destination: string = node.data.destination;
    void destination;
  }
}
function checkOpenNode(node: Node<true>) {
  if (node.kind === names.Link) {
    // @ts-expect-error A kind comparison cannot exclude unknown payloads.
    node.data.destination;
  }
  if (isKnownNode(node) && node.kind === names.Link) {
    const destination: string = node.data.destination;
    void destination;
    for (const child of node.children ?? []) {
      if (child.kind === names.Text) {
        // @ts-expect-error A known parent does not imply known descendants.
        child.data.value;
      }
      if (isKnownNode(child) && child.kind === names.Text) {
        const text: string = child.data.value;
        void text;
      }
    }
  }
}
const open = runtime.parser(grammar, { allowUnknownNodes: true });
const configured: ParserOptions = { allowUnknownNodes: Math.random() > 0.5 };
for (const parser of [open, runtime.parser(grammar, configured)]) {
  for (const result of [parser.parse("text"), parser.parseInline("text")]) {
    checkOpenNode(result.children[0]);
    // @ts-expect-error Unknown nodes cannot be assigned to the closed union.
    const closed: Node = result.children[0];
  }
}
// @ts-expect-error Implementations require a shared declaration.
new Plugin("name");
void [source, kind, error, reference];
core.dispose();
runtime.dispose();

// Owner-specific contracts narrow the same AST without loading a renderer.
import type { ReferenceKind } from "@traq-markdown-parser/ts/trap/nodes";
function ownContract(node: Node<true>) {
  if (commonNodes.isKnownNode(node) && node.kind === commonNodes.names.Link) {
    const destination: string = node.data.destination;
    void destination;
  }
}
const referenceKind: ReferenceKind = "user";
void [ownContract, referenceKind];
