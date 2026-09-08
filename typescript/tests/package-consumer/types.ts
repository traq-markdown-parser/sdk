import {
  createParser,
  presets,
  isKnownNode,
  type Node,
  type Document,
} from "@traq-markdown-parser/ts";
import { names, type ReferenceData } from "@traq-markdown-parser/ts/trap/nodes";
const parser = await createParser(new Uint8Array(), presets.traq.v1);
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
await createParser(new Uint8Array(), "not-exported");
parser.dispose();
