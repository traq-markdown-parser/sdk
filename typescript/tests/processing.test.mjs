import test from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { createRuntime, processors, presets } from "../../dist/index.js";
const bytes = await readFile(
  new URL("../../dist/parser.wasm", import.meta.url),
);

test("Rust processing preserves all 787 notification expectations", async (t) => {
  const runtime = await createRuntime(bytes);
  t.after(() => runtime.dispose());
  const processor = runtime.createProcessor(processors.traq.v1, {
    origin: "https://q.example.test",
  });
  const fixtures = JSON.parse(
    await readFile(
      new URL(
        "../../tests/fixtures/processing-notifications.json",
        import.meta.url,
      ),
    ),
  );
  assert.equal(fixtures.length, 787);
  for (const fixture of fixtures) {
    const output = processor.process(fixture.source);
    assert.equal(output.notificationText, fixture.notification, fixture.name);
    assert.deepEqual(Object.keys(output).sort(), [
      "attachments",
      "citations",
      "notificationText",
      "plainText",
      "references",
    ]);
  }
});

test("processors share compilation with parsers and have independent configuration and lifetime", async () => {
  const runtime = await createRuntime(bytes);
  try {
    const options = { origin: "https://q.example.test" };
    const first = runtime.createProcessor(processors.traq.v1, options);
    options.origin = "changed";
    const plain = runtime.createProcessor(processors.traq.v1, { origin: "" });
    const parser = runtime.createParser(presets.traq.v1);
    const source =
      "https://q.example.test/files/00000000-0000-0000-0000-000000000001";
    assert.equal(first.process(source).notificationText, "[添付ファイル]");
    assert.equal(plain.process(source).notificationText, source);
    for (const [preset, options] of [
      ["missing", { origin: "" }],
      [processors.traq.v1, { origin: "x".repeat(2049) }],
      [processors.traq.v1, { extra: true }],
    ])
      assert.throws(() => runtime.createProcessor(preset, options));
    for (const source of [
      "\ud800",
      "x".repeat(65537),
      "!!".repeat(100) + "deep" + "!!".repeat(100),
    ])
      assert.throws(() => first.process(source));
    const id = "00000000-0000-0000-0000-000000000001";
    const user = `!{"type":"user","id":"${id}","raw":"@alice"}`;
    const result = first.process(`${user} !!${user}!! \`${user}\``);
    assert.deepEqual(result.references, {
      mentions: [id, id],
      groupMentions: [],
      channelLinks: [],
      embeddings: [{raw:"@alice",type:"user",id}, {raw:"@alice",type:"user",id}],
    });
    assert.equal(result.notificationText, `@alice ██████ ${user}`);
    first.dispose();
    first.dispose();
    assert.throws(() => first.process("closed"), /disposed/);
    assert.equal(plain.process("still alive").notificationText, "still alive");
    assert.equal(parser.parse("still alive").source, "still alive");
    runtime.dispose();
    runtime.dispose();
    assert.throws(() => plain.process("closed"), /disposed/);
    assert.throws(() => parser.parse("closed"), /disposed/);
    assert.throws(
      () => runtime.createProcessor(processors.traq.v1, { origin: "" }),
      /disposed/,
    );
    assert.deepEqual(result.references.mentions, [id, id]);
  } finally {
    runtime.dispose();
  }
});
