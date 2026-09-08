import type { Owner, GrammarRecord, Grammar, ParserOptions } from "./types.js";
import { metadata, owned } from "./identity.js";
import { makePlugin } from "./plugin.js";
import { GrammarBuilder } from "./builder.js";

export function makeGrammar(runtime: Owner, record: GrammarRecord): Grammar {
  const grammar = {
    plugins: Object.freeze(
      record.snapshot.plugins.map((state) =>
        makePlugin({ ...state, frozen: true }),
      ),
    ),
    toBuilder: () => new GrammarBuilder(runtime, record.snapshot),
    describe: () => record.description,
    dispose() {
      const state = owned(grammar, "grammar", runtime);
      if (!state.disposed) {
        state.disposed = true;
        runtime.release(record);
      }
    },
  };
  metadata.set(grammar, { kind: "grammar", runtime, record, disposed: false });
  return Object.freeze(grammar) as Grammar;
}
export function makeParser(
  runtime: Owner,
  grammar: Grammar,
  options?: ParserOptions,
) {
  const state = owned(grammar, "grammar", runtime);
  if (state.disposed) throw new Error("Grammar is disposed");
  const record = state.record;
  const allowUnknownNodes = options?.allowUnknownNodes ?? false;
  if (typeof allowUnknownNodes !== "boolean")
    throw new TypeError("Expected boolean allowUnknownNodes");
  runtime.prepare(record);
  record.references++;
  let disposed = false;
  const parse = (source: string, mode: number) => {
    if (disposed) throw new Error("Parser is disposed");
    return runtime.parse(record, source, mode, allowUnknownNodes);
  };
  return Object.freeze({
    parse: (source: string) => parse(source, 0),
    parseInline: (source: string) => parse(source, 1),
    dispose() {
      if (!disposed) {
        disposed = true;
        runtime.release(record);
      }
    },
  });
}
