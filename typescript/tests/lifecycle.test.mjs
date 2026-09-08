import { Plugin as Declaration } from "../../dist/core/definitions/index.js";
import test from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import {
  loadRuntime,
  Plugin,
  GrammarBuildError,
} from "../../dist/parser/index.js";
const bytes = await readFile(
  new URL("../../dist/parser.wasm", import.meta.url),
);

test("explicit leases release Wasm grammars and repeated construction plateaus", async () => {
  const instantiate = WebAssembly.instantiate;
  let wasm;
  WebAssembly.instantiate = async (...args) => {
    const result = await instantiate(...args);
    wasm = result.instance.exports;
    return result;
  };
  let runtime;
  try {
    runtime = await loadRuntime(bytes);
  } finally {
    WebAssembly.instantiate = instantiate;
  }
  try {
    const baseline = wasm.grammar_count();
    const g = runtime.presets.traq.v1.toBuilder().build(),
      p = runtime.parser(g);
    assert.equal(wasm.grammar_count(), baseline + 1);
    g.dispose();
    assert.equal(wasm.grammar_count(), baseline + 1);
    p.parse("valid");
    p.dispose();
    assert.equal(wasm.grammar_count(), baseline);
    const sizes = [];
    for (let round = 0; round < 5; round++) {
      for (let index = 0; index < 200; index++) {
        const grammar = runtime.presets.traq.v1.toBuilder().build();
        const parser = runtime.parser(grammar);
        grammar.dispose();
        parser.parse("**text** :stamp: $x$");
        parser.dispose();
        assert.equal(wasm.grammar_count(), baseline);
      }
      sizes.push(wasm.memory.buffer.byteLength);
    }
    assert.equal(sizes.at(-1), sizes[0], "freed allocations must be reused");
    // Failed builds must not allocate a lasting handle or poison the instance.
    assert.throws(
      () =>
        runtime
          .builder()
          .add(new Plugin(new Declaration("same")))
          .add(new Plugin(new Declaration("same")))
          .build(),
      GrammarBuildError,
    );
    assert.equal(wasm.grammar_count(), baseline);
    const held = [];
    for (
      let index = baseline;
      index < runtime.contract.limits.grammars;
      index++
    )
      held.push(runtime.builder().add(runtime.plugins.commonmark.core).build());
    assert.throws(
      () => runtime.builder().add(runtime.plugins.commonmark.core).build(),
      (e) => e.detail?.reason === "live grammar limit exceeded",
    );
    held.forEach((g) => g.dispose());
    assert.equal(wasm.grammar_count(), baseline);
    runtime.builder().add(runtime.plugins.commonmark.core).build().dispose();
  } finally {
    runtime.dispose();
  }
  assert.equal(wasm.grammar_count(), 0);
});
