import { readFile } from "node:fs/promises";
import { createParser, presets } from "@traq-markdown-parser/ts";
const wasm = await readFile(
  new URL(import.meta.resolve("@traq-markdown-parser/ts/parser.wasm")),
);
const parser = await createParser(wasm, presets.traq.v1);
try {
  console.log(JSON.stringify(parser.parse("**hello** :stamp: $x$"), null, 2));
  console.log(JSON.stringify(parser.parseInline("$x$"), null, 2));
} finally {
  parser.dispose();
}
