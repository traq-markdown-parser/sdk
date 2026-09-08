import {processors, type ProcessOutput} from "@traq-markdown-parser/ts";
import {
  createRuntime,
  presets,
  isKnownNode,
  type Node,
  type Document,
} from "@traq-markdown-parser/ts";
import { names, type ReferenceData } from "@traq-markdown-parser/ts/trap/nodes";
const runtime = await createRuntime(new Uint8Array());
const parser = runtime.createParser(presets.traq.v1);
const document: Document = parser.parse("text");
for (const node of document.children)
  if (node.kind === names.Reference) {
    const reference: ReferenceData = node.data;
    const id: string = reference.id;
    void id;
  }
declare const unknownNode: Node<true>;
if (isKnownNode(unknownNode) && unknownNode.kind === names.Reference) {
  const id: string = unknownNode.data.id;
  void id;
}
// @ts-expect-error Only Rust-exported presets are accepted
runtime.createParser("not-exported");
const processor = runtime.createProcessor(processors.traq.v1, {origin:""});
const output: ProcessOutput = processor.process("hello");
const mentions: string[] = output.references.mentions;
void mentions;
// @ts-expect-error Processor has no host AST parsing API
processor.parse("x");
// @ts-expect-error Options are generated from Rust
runtime.createProcessor(processors.traq.v1, {origin:4});
runtime.dispose();
