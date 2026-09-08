import { readFile } from "node:fs/promises";
import { createRuntime, presets, processors } from "@traq-markdown-parser/traq";
const wasm = await readFile(
  new URL(import.meta.resolve("@traq-markdown-parser/traq/parser.wasm")),
);
const runtime = await createRuntime(wasm);
try {
  const parser = runtime.createParser(presets.traq.v1);
  console.log(JSON.stringify(parser.parse("**hello** :stamp: $x$"), null, 2));
  console.log(JSON.stringify(parser.parseInline("$x$"), null, 2));
  const processor = runtime.createProcessor(processors.traq.v1, {origin: "https://q.example.test"});
  console.log(JSON.stringify(processor.process("**hello** !!secret!!"), null, 2));
} finally {
  runtime.dispose();
}
