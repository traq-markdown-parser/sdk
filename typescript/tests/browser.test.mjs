import test from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { loadRuntime, MarkdownParseError } from "../../dist/parser/index.js";
import { validateDocument } from "../../dist/core/parser/validate.js";
import { createRunner } from "../../dist/core/parser/call.js";
import { nodes, names } from "../../dist/parser/generated/nodes.js";
const bytes = await readFile(
  new URL("../../dist/parser.wasm", import.meta.url),
);

test("artifact mismatches fail at initialization and missing decoders at parse", async (t) => {
  await assert.rejects(loadRuntime(new Uint8Array([1, 2, 3])));
  // A valid Wasm module with no parser exports must also be rejected.
  await assert.rejects(
    loadRuntime(new Uint8Array([0, 97, 115, 109, 1, 0, 0, 0])),
  );
  const { instance } = await WebAssembly.instantiate(bytes, {});
  assert.throws(
    () =>
      createRunner({ exports: { ...instance.exports, abi_version: () => 1 } }),
    /Unsupported Wasm ABI/,
  );
  const runtime = await loadRuntime(bytes, { nodes: new Map() });
  t.after(() => runtime.dispose());
  const strict = runtime.parser(runtime.presets.traq.v1);
  assert.throws(() => strict.parse(":stamp:"), /Unsupported.*node/);
  strict.dispose();
  const core = runtime.parser(runtime.presets.traq.v1, {
    allowUnknownNodes: true,
  });
  assert.equal(
    core.parse(":stamp:").children[0].children[0].kind,
    names.Stamp,
  );
});

test("failed requests do not poison later calls; resources remain bounded", async (t) => {
  const runtime = await loadRuntime(bytes);
  t.after(() => runtime.dispose());
  const core = runtime.parser(runtime.presets.traq.v1);
  for (const [source, code] of [
    ["\ud800", "invalid_utf8"],
    ["x".repeat(65537), "resource_limit"],
  ]) {
    assert.throws(
      () => core.parse(source),
      (e) => e instanceof MarkdownParseError && e.detail.code === code,
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
    (e) => e.detail?.resource === "depth",
  );
  for (const source of hostile) {
    try {
      assert.equal(core.parse(source).source, source);
    } catch (e) {
      assert(e instanceof MarkdownParseError);
      assert.equal(e.detail.code, "resource_limit");
    }
    assert.equal(core.parse("after").children[0].children[0].data.value, "after");
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
    (e) => e.detail?.resource === "output_bytes",
  );
  assert.equal(core.parse("after").source, "after");
});

test("linear memory has a 32 MiB maximum and ABI rejects invalid UTF-8", async (t) => {
  const { instance } = await WebAssembly.instantiate(bytes, {}),
    wasm = instance.exports;
  const pointer = wasm.input_ptr(1);
  new Uint8Array(wasm.memory.buffer, pointer, 1)[0] = 255;
  const length = wasm.parse(1, 0);
  assert.deepEqual(
    JSON.parse(
      new TextDecoder().decode(
        new Uint8Array(wasm.memory.buffer, wasm.output_ptr(), length),
      ),
    ),
    { error: { code: "invalid_utf8" } },
  );
  const memory = wasm.memory,
    pages = memory.buffer.byteLength / 65536;
  assert.equal(memory.grow(512 - pages), pages);
  assert.throws(() => memory.grow(1), RangeError);
});

test("common shape, UTF-8 boundaries and plugin payloads are validated", async (t) => {
  const runtime = await loadRuntime(bytes);
  t.after(() => runtime.dispose());
  const core = runtime.parser(runtime.presets.traq.v1);
  const source = '> 日本語\r\n> !{"type":"user","id":"u","raw":"@u"}';
  const doc = core.parse(source);
  const mention = doc.children[0].children[0].children.find(
    (n) => n.kind === names.Reference,
  );
  assert.equal(
    new TextDecoder().decode(
      new TextEncoder()
        .encode(source)
        .slice(mention.span.start, mention.span.end),
    ),
    '!{"type":"user","id":"u","raw":"@u"}',
  );
  for (const change of [
    (n) => {
      n.data.id = 12;
    },
    (n) => {
      n.data.extra = true;
    },
    (n) => {
      delete n.data.label;
    },
    (n) => {
      n.kind = "unknown@1";
    },
    (n) => {
      n.span.start = 3;
    },
    (n) => {
      n.kind = "text";
      n.value = "bad";
    },
  ]) {
    const altered = structuredClone(doc);
    change(
      altered.children[0].children[0].children.find(
        (n) => n.kind === names.Reference,
      ),
    );
    assert.throws(() => validateDocument(altered, source, nodes));
  }
});
