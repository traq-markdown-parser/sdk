import test from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { createParser, presets, isKnownNode } from "../../dist/index.js";
import { names } from "../../dist/generated/nodes.js";
const bytes = await readFile(
  new URL("../../dist/parser.wasm", import.meta.url),
);

test("failed requests do not poison later calls; resources remain bounded", async (t) => {
  const core = await createParser(bytes, presets.traq.v1);
  t.after(() => core.dispose());
  for (const [source, code] of [
    ["\ud800", "invalid_utf8"],
    ["x".repeat(65537), "resource_limit"],
  ]) {
    assert.throws(
      () => core.parse(source),
      (e) =>
        code === "invalid_utf8"
          ? e instanceof TypeError
          : e instanceof RangeError,
    );
  }
  assert.throws(() => core.parse(2), TypeError);
  assert.equal(core.parseInline("**x**").children[0].kind, names.Strong);
  assert.equal(core.parse("\ufefftext🦀").source, "\ufefftext🦀");
  const hostile = [
    "[".repeat(5000),
    "> ".repeat(300),
    "!{".repeat(15000),
    "*".repeat(50000),
    ":".repeat(50000),
    ('[a]: "' + "x".repeat(200) + "\n").repeat(200),
    "\t".repeat(65536),
    "> ".repeat(60) + "\t".repeat(60000),
    "- > ".repeat(30) + "\t".repeat(60000),
  ];
  assert.throws(
    () => core.parse("!!".repeat(100) + "deep" + "!!".repeat(100)),
    (e) => e.cause?.resource === "depth",
  );
  for (const source of hostile) {
    try {
      assert.equal(core.parse(source).source, source);
    } catch (e) {
      assert(e instanceof Error);
      assert.equal(e.cause.code, "resource_limit");
    }
    assert.equal(
      core.parse("after").children[0].children[0].data.value,
      "after",
    );
  }
  const columns = 7500;
  assert.throws(
    () =>
      core.parse(
        "|".repeat(columns + 1) +
          "\n" +
          "|-".repeat(columns) +
          "|\n" +
          "|".repeat(columns + 1),
      ),
    (e) => e.cause?.resource === "output_bytes",
  );
  assert.equal(core.parse("after").source, "after");
});

test("artifact pairing, preset selection, disposal and isolated results", async () => {
  for (const bad of [
    new Uint8Array([1, 2, 3]),
    new Uint8Array([0, 97, 115, 109, 1, 0, 0, 0]),
  ])
    await assert.rejects(createParser(bad, presets.traq.v1), /does not match/);
  await assert.rejects(createParser(bytes, "traq.invalid"), /Markdown:/);
  const traq = await createParser(bytes, presets.traq.v1);
  const common = await createParser(bytes, presets.commonmark);
  try {
    const first = traq.parseInline(":stamp:");
    assert.equal(first.children[0].kind, names.Stamp);
    assert(isKnownNode(first.children[0]));
    assert(!isKnownNode({ ...first.children[0], kind: "future::Node" }));
    assert.equal(common.parseInline(":stamp:").children[0].kind, names.Text);
    traq.parse("another");
    assert.equal(first.source, ":stamp:");
    traq.dispose();
    traq.dispose();
    assert.throws(() => traq.parse("closed"), /disposed/);
    assert.equal(common.parse("still open").source, "still open");
  } finally {
    traq.dispose();
    common.dispose();
  }
});

test("raw ABI validates UTF-8, preset and mode; linear memory is bounded", async () => {
  const { instance } = await WebAssembly.instantiate(bytes, {});
  const wasm = instance.exports,
    encoder = new TextEncoder(),
    decoder = new TextDecoder();
  const call = (operation, input, mode = 0) => {
    const encoded = typeof input === "string" ? encoder.encode(input) : input;
    const pointer = wasm.input_ptr(encoded.length);
    new Uint8Array(wasm.memory.buffer, pointer, encoded.length).set(encoded);
    const length =
      operation === "configure" ? wasm.configure() : wasm.parse(mode);
    return JSON.parse(
      decoder.decode(
        new Uint8Array(wasm.memory.buffer, wasm.output_ptr(), length),
      ),
    );
  };
  assert.equal(wasm.abi_version(), 3);
  assert(call("configure", "traq.bad").error);
  assert.deepEqual(call("configure", "traq.v1"), { configured: null });
  assert.deepEqual(call("parse", new Uint8Array([255])), {
    error: { code: "invalid_utf8" },
  });
  assert(call("parse", "x", 99).error);
  assert.equal(call("parse", "ok").document.source, "ok");
  const pages = wasm.memory.buffer.byteLength / 65536;
  assert.equal(wasm.memory.grow(512 - pages), pages);
  assert.throws(() => wasm.memory.grow(1), RangeError);
});

test("Go and TypeScript fixtures retain the Rust AST for blocks and inlines", async () => {
  const parser = await createParser(bytes, presets.traq.v1);
  const commonmark = JSON.parse(
    await readFile(
      new URL("../../tests/fixtures/commonmark-0.31.2.json", import.meta.url),
    ),
  );
  try {
    for (const file of ["traq-v1-commonmark", "traq-v1-extensions"]) {
      const cases = JSON.parse(
        await readFile(
          new URL("../../tests/fixtures/" + file + ".json", import.meta.url),
        ),
      );
      for (const fixture of cases)
        for (const [mode, method] of [
          ["block", "parse"],
          ["inline", "parseInline"],
        ]) {
          const expected = fixture.expected[mode];
          const source =
            fixture.source ??
            commonmark.find((c) => c.example === fixture.example).markdown;
          if ("Ok" in expected)
            assert.deepEqual(
              parser[method](source),
              expected.Ok,
              fixture.name + " " + mode,
            );
          else
            assert.throws(
              () => parser[method](source),
              (e) => assert.deepEqual(e.cause, expected.Err) === undefined,
            );
        }
    }
  } finally {
    parser.dispose();
  }
});

test('Wasm input views respect their byte range and are copied before async work', async () => {
  const padded = Buffer.concat([Buffer.from([99]),bytes,Buffer.from([99])]);
  const view = padded.subarray(1,padded.length-1);
  const pending = createParser(view,presets.traq.v1);
  view.fill(0);
  const parser = await pending;
  try { assert.equal(parser.parse('copied').source,'copied'); }
  finally { parser.dispose(); }
});
