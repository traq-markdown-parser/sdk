import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { readFile } from "node:fs/promises";
import { createParser, presets } from "@traq-markdown-parser/ts";
import { names } from "@traq-markdown-parser/ts/trap/nodes";
const bytes = await readFile(
  new URL(import.meta.resolve("@traq-markdown-parser/ts/parser.wasm")),
);
const parser = await createParser(bytes, presets.traq.v1);
assert.equal(createHash('sha256').update(bytes).digest('hex'), process.argv[2]);
try {
  assert.equal(parser.parseInline(":stamp:").children[0].kind, names.Stamp);
} finally {
  parser.dispose();
}
console.log("Packed TypeScript consumer passed");
