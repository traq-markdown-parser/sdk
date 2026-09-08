import { readFile } from "node:fs/promises";
import path from "node:path";
import { javascript } from "./javascript.mjs";
import { typeName, quoted as q } from "./schema.mjs";
import { contractsDirectory, relativeImport } from "./layout.mjs";

const header = "// Generated from Rust node contracts. Do not edit.\n";
const union = (entries) => entries.map(([key, schema]) =>
  `  | { kind: ${q(key)}; data: ${schema.title} }`).join("\n");

export async function typescriptFiles(manifest, input) {
  const entries = Object.entries(manifest.nodes).map(([key, node]) => [key, node.schema]);
  const groups = Map.groupBy(entries, ([key]) => manifest.nodes[key].group);
  const used = new Set();
  for (const [, schema] of entries) {
    const name = typeName(schema).toLowerCase();
    if (used.has(name)) throw new Error("Duplicate generated payload type: " + name);
    used.add(name);
  }
  const files = new Map();
  for (const [group, entries] of groups) {
    const directory = contractsDirectory(group), types = new Set();
    async function payload(name) {
      if (types.has(name)) return;
      types.add(name);
      const source = await readFile(path.join(input, name + ".ts"), "utf8");
      files.set(directory + "/" + name + ".ts", source);
      for (const match of source.matchAll(/from ["']\.\/([^"']+)\.js["']/g))
        await payload(match[1]);
    }
    for (const [, schema] of entries) await payload(schema.title);
    files.set(directory + "/nodes.ts", header +
      entries.map(([, s]) => `import type { ${s.title} } from './${s.title}.js';`).join("\n") + "\n" +
      [...types].map((name) => `export type { ${name} } from './${name}.js';`).join("\n") + "\n" +
      "export type NodeKind =\n" + union(entries) + ";\n" + javascript(entries,
      relativeImport(directory, "typescript/src/core/validation/fields.js")));
  }
  const directory = "typescript/src/parser/generated";
  files.set(directory + "/nodes.ts", header +
    "import type { Node } from './Node.js';\nimport type { NodeKind } from './NodeKind.js';\n" +
    [...groups.keys()].map((group) => `import * as ${group} from ${q(relativeImport(directory, contractsDirectory(group) + "/nodes.js"))};`).join("\n") + "\n" +
    "export const names = Object.freeze({" + [...groups.keys()].map((g) => `...${g}.names`).join(",") + "});\n" +
    "const validators = new Map([" + [...groups.keys()].map((g) => `...${g}.nodes`).join(",") + "]);\n" +
    "export const nodes: ReadonlyMap<string, (data: unknown) => boolean> = new Map(validators);\n" +
    "export function isKnownNode(node: Node<true>): node is Node<true> & NodeKind { return validators.get(node.kind)?.(node.data) ?? false; }\n");
  files.set(directory + "/contracts.json", JSON.stringify(manifest, null, 2) + "\n");
  files.set("typescript/src/core/parser/ParseError.ts", await readFile(path.join(input, "ParseError.ts"), "utf8"));
  return files;
}
