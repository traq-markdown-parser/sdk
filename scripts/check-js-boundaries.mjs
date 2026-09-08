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
