import type {
  Catalog,
  Validators,
  GrammarRecord,
  Snapshot,
  Grammar,
  Owner,
  ParserOptions,
} from "./types.js";
import { createRunner } from "./call.js";
import { GrammarBuilder, composition } from "./builder.js";
import { makeGrammar, makeParser } from "./grammar.js";
import { loadCatalog } from "./catalog.js";

export async function createRuntime(
  bytes: BufferSource,
  expectedCatalog: Catalog,
  validators: Validators,
) {
  const { instance } = await WebAssembly.instantiate(bytes, {});
  const runner = createRunner(instance),
    contract = runner.contract;
  if (JSON.stringify(contract.catalog) !== JSON.stringify(expectedCatalog))
    throw new Error("Wasm catalog does not match this SDK");
  const nodes = new Map(validators);
  for (const [name, validate] of nodes)
    if (typeof name !== "string" || typeof validate !== "function")
      throw new TypeError("Invalid node validator");
  const records = new Set<GrammarRecord>();
  let disposed = false;
  function requireOpen() {
    if (disposed) throw new Error("Runtime is disposed");
  }
  function register(
    snapshot: Snapshot,
    {
      description,
      handle = null,
    }: { description: string; handle?: number | null },
  ): Grammar {
    requireOpen();
    const record: GrammarRecord = {
      snapshot,
      description,
      handle,
      references: 1,
    };
    records.add(record);
    return makeGrammar(owner, record);
  }
  const owner: Owner = {
    preset: register,
    build(snapshot) {
      requireOpen();
      return register(snapshot, runner.build(composition(snapshot)));
    },
    prepare(record) {
      requireOpen();
      if (record.handle !== null) return;
      const compiled = runner.build(composition(record.snapshot));
      if (compiled.description !== record.description) {
        runner.dispose();
        throw new Error("Preset compilation does not match the catalog");
      }
      record.handle = compiled.handle;
    },
    parse(record, source, mode, allowUnknown) {
      requireOpen();
      return runner.parse(record.handle!, source, mode, nodes, allowUnknown);
    },
    release(record) {
      if (--record.references === 0) {
        records.delete(record);
        if (!disposed && record.handle !== null) runner.drop(record.handle);
      }
    },
  };
  let catalog;
  try {
    catalog = loadCatalog(owner, expectedCatalog);
  } catch (error) {
    runner.dispose();
    throw error;
  }
  return Object.freeze({
    contract,
    ...catalog,
    builder() {
      requireOpen();
      return new GrammarBuilder(owner);
    },
    parser(grammar: Grammar, options?: ParserOptions) {
      requireOpen();
      return makeParser(owner, grammar, options);
    },
    dispose() {
      if (disposed) return;
      disposed = true;
      try {
        for (const record of records)
          if (record.handle !== null) runner.drop(record.handle);
      } finally {
        records.clear();
        runner.dispose();
      }
    },
  });
}
