import { readFile } from "node:fs/promises";
import path from "node:path";
import { goContract } from "./go.mjs";
import { presetFiles } from "./presets.mjs";

export async function processingFiles(schemas, input) {
  const tree = {};
  for (const name of schemas.ProcessorPreset.enum) {
    const parts = name.split(".");
    let target = tree;
    for (const part of parts.slice(0, -1)) target = target[part] ??= {};
    target[parts.at(-1)] = 0;
  }

  const files = presetFiles(
    tree,
    "ProcessorPreset",
    "processors",
    "processing",
  );
  const types = new Map();
  async function declaration(name) {
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
      await declaration(match[1]);
  }

  let go = "";
  const declarations = new Map();
  function addGo(schema, root = schema) {
    const source = goContract(schema, root);
    const previous = declarations.get(schema.title);
    if (previous !== undefined && previous !== source)
      throw new Error("Conflicting processing type: " + schema.title);
    if (previous === undefined) {
      declarations.set(schema.title, source);
      go += source;
    }
  }

  for (const name of ["ProcessorOptions", "ProcessOutput"]) {
    await declaration(name);
    const schema = schemas[name];
    addGo(schema);
    for (const [name, definition] of Object.entries(schema.$defs ?? {}))
      addGo({ ...definition, title: name }, schema);
  }
  const tsPath = "typescript/generated/processing.ts",
    goPath = "go/processing_generated.go";
  files.set(tsPath, files.get(tsPath) + [...types.values()].join("\n") + "\n");
  files.set(goPath, files.get(goPath) + go);
  return files;
}
