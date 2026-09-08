import { readFile } from "node:fs/promises";
import path from "node:path";
import { javascript } from "./javascript.mjs";
import { quoted as q } from "./schema.mjs";

export async function typescriptFiles(manifest, input) {
  const entries = Object.entries(manifest.nodes).map(([key, node]) => [
    key,
    node.schema,
  ]);
  const groups = Map.groupBy(entries, ([key]) => manifest.nodes[key].group);
  const files = new Map();
  for (const [group, entries] of groups) {
    const types = new Map();
    async function payload(name) {
      if (types.has(name)) return;
      const source = await readFile(path.join(input, name + ".ts"), "utf8");
      types.set(
        name,
        source
          .replace(/^\/\/[^\n]*\n/gm, "")
          .replace(/^import type .*;\r?\n/gm, "")
          .trim(),
      );
      for (const match of source.matchAll(/from ["']\.\/([^"']+)\.js["']/g))
        await payload(match[1]);
    }

    for (const [, schema] of entries) await payload(schema.title);

    files.set(
      `typescript/generated/${group}.ts`,
      "// Generated from Rust contracts. Do not edit.\n" +
        [...types.values()].join("\n") +
        "\nexport type NodeKind =\n" +
        entries
          .map(([key, s]) => ` | { kind: ${q(key)}; data: ${s.title} }`)
          .join("\n") +
        ";\n" +
        javascript(entries, "../validation.js"),
    );
  }
  const owners = [...groups.keys()];
  files.set(
    "typescript/generated/nodes.ts",
    "// Generated from Rust contracts. Do not edit.\n" +
      owners.map((g) => `import * as ${g} from './${g}.js';`).join("\n") +
      "\n" +
      `export type NodeKind = ${owners.map((g) => `${g}.NodeKind`).join(" | ")};\n` +
      "export type Node<AllowUnknown extends boolean = false> =\n" +
      " (NodeKind | (AllowUnknown extends true ? {kind:string;data:unknown} : never)) &\n" +
      " {span:{start:number;end:number};children?:Node<AllowUnknown>[]};\n" +
      "export type Document<AllowUnknown extends boolean = false> = {source:string;children:Node<AllowUnknown>[]};\n" +
      `export const names = Object.freeze({${owners.map((g) => `...${g}.names`).join(",")}});\n` +
      `export const nodes: ReadonlyMap<string,(data:unknown)=>boolean> = new Map([${owners.map((g) => `...${g}.nodes`).join(",")}]);\n` +
      "export function isKnownNode(node:Node<true>):node is Node<true> & NodeKind {return nodes.get(node.kind)?.(node.data) ?? false;}\n" +
      (await readFile(path.join(input, "ParseError.ts"), "utf8")).replace(
        /^\/\/[^\n]*\n/gm,
        "",
      ),
  );
  return files;
}
