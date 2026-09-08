import assert from "node:assert/strict";
import { readdir, readFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import ts from "typescript";

const root = fileURLToPath(new URL("../", import.meta.url));
const packageJson = JSON.parse(await readFile(path.join(root, "package.json")));
assert.equal(packageJson.name, "@traq-markdown-parser/ts");
assert.equal(Object.keys(packageJson.dependencies ?? {}).length, 0, "Bindings have no runtime package dependencies");
assert.equal(packageJson.workspaces, undefined, "JavaScript ships as one npm package");
const files = new Map();
async function collect(directory) {
  for (const entry of await readdir(path.join(root, directory), { withFileTypes: true })) {
    const name = directory + "/" + entry.name;
    if (entry.isDirectory()) await collect(name);
    else {
      assert.notEqual(entry.name, "package.json", "No nested npm packages in src");
      files.set(name, true);
    }
  }
}
await collect("typescript/src");

const contracts = (file) => /\/contracts\//.test(file);
function allowed(from, to) {
  if (from.startsWith("typescript/src/core/")) return to.startsWith("typescript/src/core/");
  if (contracts(from)) {
    const owner = from.slice(0, from.indexOf("/contracts/") + 11);
    return to.startsWith(owner) || to.startsWith("typescript/src/core/validation/");
  }
  if (from.startsWith("typescript/src/parser/"))
    return to.startsWith("typescript/src/parser/") || contracts(to) ||
      ["typescript/src/core/parser/", "typescript/src/core/definitions/", "typescript/src/core/validation/"].some((p) => to.startsWith(p));
  return true;
}
function resolve(from, specifier) {
  const base = path.posix.normalize(path.posix.join(path.posix.dirname(from), specifier));
  return [base, base.replace(/\.js$/, ".ts"), base + ".ts", base + ".d.ts", base + "/index.ts"]
    .find((file) => files.has(file));
}
let imports = 0;
for (const file of files.keys()) {
  if (!/\.(?:[cm]?ts|[cm]?js)$/.test(file)) continue;
  const source = ts.createSourceFile(file, await readFile(path.join(root, file), "utf8"), ts.ScriptTarget.Latest, true);
  function check(specifier) {
    assert(!specifier.startsWith(packageJson.name), `${file}: internal imports must not use the public package barrel`);
    assert(specifier.startsWith("."), `${file}: bindings must not import external packages: ${specifier}`);
    const target = resolve(file, specifier);
    assert(target, `${file}: unresolved import ${specifier}`);
    assert(allowed(file, target), `${file} must not depend on ${target}`);
    imports++;
  }
  function visit(node) {
    if ((ts.isImportDeclaration(node) || ts.isExportDeclaration(node)) &&
        node.moduleSpecifier && ts.isStringLiteral(node.moduleSpecifier))
      check(node.moduleSpecifier.text);
    if (ts.isImportTypeNode(node) && ts.isLiteralTypeNode(node.argument) && ts.isStringLiteral(node.argument.literal))
      check(node.argument.literal.text);
    if (ts.isCallExpression(node) && node.expression.kind === ts.SyntaxKind.ImportKeyword &&
        node.arguments[0] && ts.isStringLiteral(node.arguments[0]))
      check(node.arguments[0].text);
    ts.forEachChild(node, visit);
  }
  visit(source);
}
console.log(`Bindings boundaries passed: ${imports} internal imports, no rendering or runtime package dependencies`);
