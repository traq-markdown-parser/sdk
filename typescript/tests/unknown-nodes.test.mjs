import test from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { loadRuntime, isKnownNode } from "../../dist/parser/index.js";
import { nodes, names } from "../../dist/parser/generated/nodes.js";
import { validateDocument } from "../../dist/core/parser/validate.js";

test("allowUnknownNodes changes only rejection of unregistered nodes", async (t) => {
  const bytes = await readFile(new URL("../../dist/parser.wasm", import.meta.url));
  const runtime = await loadRuntime(bytes);
  t.after(() => runtime.dispose());
  const strict = runtime.parser(runtime.presets.traq.v1);
  const open = runtime.parser(runtime.presets.traq.v1, { allowUnknownNodes: true });
  const source = "[link](https://example.com) **text** :stamp:";
  for (const method of ["parse", "parseInline"])
    assert.deepEqual(open[method](source), strict[method](source));

  const document = strict.parse(source);
  const link = document.children[0].children[0];
  assert.equal(isKnownNode(link), true);
  const unknown = { kind: "custom::Node", span: link.span, data: { value: "custom" } };
  link.children = [unknown];
  assert.equal(isKnownNode(link), true); // Only the parent payload is known.
  assert.equal(isKnownNode(unknown), false);
  assert.throws(() => validateDocument(document, source, nodes), /Unsupported node/);
  assert.equal(validateDocument(document, source, nodes, true), document);

  link.data.destination = 42;
  assert.equal(isKnownNode(link), false); // A known key alone is insufficient.
  for (const allow of [false, true])
    assert.throws(() => validateDocument(document, source, nodes, allow), /Invalid node/);
});

test("known-node checks use generated validators, independent of custom registrations", () => {
  const node = { kind: names.Text, span: { start: 0, end: 0 }, data: { value: 42 } };
  const original = nodes.get(names.Text);
  try {
    nodes.set(names.Text, () => true);
    assert.equal(isKnownNode(node), false);
  } finally {
    nodes.set(names.Text, original);
  }
  const heading = { kind: names.Heading, span: node.span, data: { level: 7 } };
  assert.equal(isKnownNode(heading), true); // Semantic rules remain in Rust.
});
