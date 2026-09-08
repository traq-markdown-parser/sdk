import assert from "node:assert/strict";
import { readdir, readFile } from "node:fs/promises";
const root = new URL("../", import.meta.url);
const pkg = JSON.parse(await readFile(new URL("package.json", root)));
assert.equal(Object.keys(pkg.dependencies ?? {}).length, 0);
const hand = (await readdir(new URL("typescript/", root)))
  .filter((f) => f.endsWith(".ts"))
  .sort();
assert.deepEqual(hand, ["index.ts", "validation.ts"]);
for (const file of [
  ...hand.map((f) => "typescript/" + f),
  ...(await readdir(new URL("typescript/generated/", root))).map(
    (f) => "typescript/generated/" + f,
  ),
]) {
  const source = await readFile(new URL(file, root), "utf8");
  for (const [, specifier] of source.matchAll(/from ['"]([^'"]+)['"]/g))
    assert(
      specifier.startsWith("."),
      file + ": unexpected external dependency " + specifier,
    );
}
console.log(
  "Thin bindings: two handwritten TypeScript files, no runtime package dependencies",
);

const { execFileSync } = await import("node:child_process");
const dependencies = execFileSync("cargo", ["tree", "--locked", "-p", "traq-markdown-processor", "--edges", "normal"], { cwd: root, encoding: "utf8" });
assert(!dependencies.includes("markdown-codec"), "Native processing must borrow the AST without a codec dependency");
console.log("Native processing: no AST codec dependency");

for (const name of ["traq-markdown-processing", "markdown-trap-text", "markdown-trap-extraction", "markdown-trap-contracts"]) {
  const tree = execFileSync("cargo", ["tree", "--locked", "-p", name, "--edges", "normal"], { cwd: root, encoding: "utf8" });
  assert(!/markdown-parser v|markdown-codec v|traq-markdown-grammar v|traq-markdown-processor v|traq-markdown-wasm v/.test(tree), name + ": lower-level processing must not depend on parsing or distribution");
}
const syntax = execFileSync("cargo", ["tree", "--locked", "-p", "markdown-trap-syntax", "--edges", "normal"], { cwd: root, encoding: "utf8" });
assert(!/traq-markdown-grammar v|traq-markdown-processor v|traq-markdown-wasm v/.test(syntax), "Extension syntax must not depend on the traQ distribution");
console.log("Dependency direction: extension components and processing presets do not depend on the distribution");
