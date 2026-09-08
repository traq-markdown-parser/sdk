import test from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { loadRuntime, GrammarBuildError } from "../../dist/parser/index.js";
const bytes = await readFile(
  new URL("../../dist/parser.wasm", import.meta.url),
);

async function observedRuntime(options) {
  const instantiate = WebAssembly.instantiate;
  let wasm;
  WebAssembly.instantiate = async (...args) => {
    const result = await instantiate(...args);
    wasm = result.instance.exports;
    return result;
  };
  try {
    return {
      runtime: await loadRuntime(bytes, options),
      count: () => wasm.grammar_count(),
    };
  } finally {
    WebAssembly.instantiate = instantiate;
  }
}

test("preset metadata stays lazy and parser construction shares one compiled grammar", async () => {
  const { runtime: r, count } = await observedRuntime();
  try {
    const g = r.presets.traq.v1;
    assert.equal(count(), 0);
    const description = g.describe(),
      plugins = g.plugins;
    const builder = g.toBuilder();
    assert.equal(count(), 0);
    const independent = builder.build();
    assert.equal(count(), 1);
    assert.equal(independent.describe(), description);
    assert.equal(independent.plugins.length, plugins.length);
    independent.dispose();
    assert.equal(count(), 0);
    const p = r.parser(g),
      other = r.parser(g);
    assert.equal(count(), 1, "constructors compile once, before any parse");
    g.dispose();
    assert.throws(() => r.parser(g), /disposed/);
    assert.equal(count(), 1);
    assert.deepEqual(p.parse("**text** $x$"), other.parse("**text** $x$"));
    p.dispose();
    assert.equal(count(), 1);
    other.dispose();
    assert.equal(count(), 0);
    r.presets.commonmark.dispose();
    assert.equal(count(), 0, "unused preset has no handle to release");
    assert.throws(() => r.parser(r.presets.commonmark), /disposed/);
  } finally {
    r.dispose();
  }
  assert.equal(count(), 0);
});

test("failed lazy initialization can retry after capacity is released", async () => {
  const { runtime: r, count } = await observedRuntime();
  try {
    const held = Array.from({ length: r.contract.limits.grammars }, () =>
      r.builder().add(r.plugins.commonmark.core).build(),
    );
    assert.throws(() => r.parser(r.presets.traq.v1), GrammarBuildError);
    assert.equal(count(), held.length);
    held.pop().dispose();
    const p = r.parser(r.presets.traq.v1);
    assert.equal(count(), held.length + 1);
    assert.equal(p.parse("after retry").source, "after retry");
    p.dispose();
    r.presets.traq.v1.dispose();
    held.forEach((g) => g.dispose());
    assert.equal(count(), 0);
  } finally {
    r.dispose();
  }
});

test("ownership is checked at construction and missing node codecs at parse", async () => {
  const { runtime: r, count } = await observedRuntime({ nodes: [] });
  const { runtime: foreign, count: foreignCount } = await observedRuntime();
  try {

    assert.throws(() => r.parser(foreign.presets.traq.v1), /runtime/i);
    assert.equal(count(), 0);
    assert.equal(foreignCount(), 0);
    const strict = r.parser(r.presets.traq.v1);
    assert.throws(() => strict.parse("text"), /Unsupported.*node/);
    strict.dispose();
    const p = r.parser(r.presets.traq.v1, { allowUnknownNodes: true });
    assert.equal(p.parse("text").source, "text");
    assert.equal(count(), 1);
    p.dispose();
  } finally {
    r.dispose();
    foreign.dispose();
  }
  assert.equal(count(), 0);
});
