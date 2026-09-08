import { readFile } from "node:fs/promises";
import { loadRuntime } from "@traq-markdown-parser/ts";

const wasm = await readFile(
  new URL(import.meta.resolve("@traq-markdown-parser/ts/parser.wasm")),
);
const runtime = await loadRuntime(new Uint8Array(wasm));
try {
  const parser = runtime.parser(runtime.presets.traq.v1);
  try {
    const document = parser.parse("**hello** :stamp: $x$");
    console.log(JSON.stringify(document, null, 2));

  } finally {
    parser.dispose();
  }

  const grammar = runtime.presets.traq.v1
    .toBuilder()
    .remove(runtime.plugins.generic.math)
    .build();
  try {
    const withoutMath = runtime.parser(grammar);
    try {
      grammar.dispose(); // The Parser retains the compiled grammar.
      console.log(JSON.stringify(withoutMath.parseInline("$x$"), null, 2));
    } finally {
      withoutMath.dispose();
    }
  } finally {
    grammar.dispose(); // Repeated disposal is safe.
  }
} finally {
  runtime.dispose();
}
