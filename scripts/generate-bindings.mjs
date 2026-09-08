import { execFileSync } from "node:child_process";
import { readFile, writeFile, mkdir } from "node:fs/promises";
import path from "node:path";
import { typescriptFiles } from "./contracts/typescript.mjs";
import { goPayload, goRegistry } from "./contracts/go.mjs";
import { catalogFiles } from "./contracts/catalog.mjs";
import { typeName } from "./contracts/schema.mjs";
import { treeFiles } from "./contracts/tree.mjs";
const input = path.resolve(process.argv[2] ?? "target/node-contracts");
const manifest = JSON.parse(await readFile(path.join(input, "contracts.json")));
const entries = Object.entries(manifest.nodes).map(([key, entry]) => [key, entry.schema]);
const files = await typescriptFiles(manifest, input);
for (const [p, s] of treeFiles(entries, manifest.nodes)) files.set(p, s);
const groups = Map.groupBy(entries, ([name]) => manifest.nodes[name].group);
for (const [group, entries] of groups) {
  for (const [name, schema] of entries)
    files.set(
      "go/extensions/" + group + "/" + typeName(schema).toLowerCase() + ".go",
      goPayload(group, name, schema),
    );
  files.set(
    "go/extensions/" + group + "/registry.go",
    goRegistry(group, entries),
  );
}
for (const [p, s] of catalogFiles(manifest.catalog, [...groups.keys()]))
  files.set(p, s);
for (const [p, s] of files) {
  await mkdir(path.dirname(p), { recursive: true });
  await writeFile(p, s);
}
execFileSync(
  "gofmt",
  ["-w", ...[...files.keys()].filter((p) => p.endsWith(".go"))],
  { windowsHide: true },
);
console.log(
  "Generated " +
    files.size +
    " host binding files from " +
    entries.length +
    " Rust payloads",
);
