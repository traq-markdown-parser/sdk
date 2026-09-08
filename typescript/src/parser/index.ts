import { createRuntime } from "../core/parser/runtime.js";
import { nodes } from "./generated/nodes.js";
import { catalog } from "./generated/catalog.js";
import type { CatalogViews } from "./generated/catalog.js";
import type { Document } from "./generated/Document.js";
import type {
  Contract,
  Grammar,
  ParserOptions,
  RuntimeOptions,
} from "../core/parser/types.js";
import type { GrammarBuilder } from "../core/parser/builder.js";
export { Plugin } from "../core/parser/plugin.js";
export {
  MarkdownParseError,
  GrammarBuildError,
} from "../core/parser/errors.js";
export { isKnownNode } from "./generated/nodes.js";
export type { Document } from "./generated/Document.js";
export type { Node } from "./generated/Node.js";
export type { NodeKind } from "./generated/NodeKind.js";
export type { ReferenceData } from "../trap/contracts/ReferenceData.js";
export type { GrammarBuilder } from "../core/parser/builder.js";
export type {
  Contract,
  Grammar,
  Rule,
  Phase,
  ParserOptions,
  RuntimeOptions,
  BuildError,
  ParseError,
} from "../core/parser/types.js";

export interface Parser<AllowUnknown extends boolean = false> {
  parse(source: string): Document<AllowUnknown>;
  parseInline(source: string): Document<AllowUnknown>;
  dispose(): void;
}
export interface Runtime extends CatalogViews {
  readonly contract: Contract;
  builder(): GrammarBuilder;
  parser<AllowUnknown extends boolean = false>(
    grammar: Grammar,
    options?: ParserOptions<AllowUnknown>,
  ): Parser<AllowUnknown>;
  dispose(): void;
}

/** Bind the generic Wasm runtime to this distribution's catalog and contracts. */
export async function loadRuntime(
  bytes: BufferSource,
  options: RuntimeOptions = {},
): Promise<Runtime> {
  // The generic runtime verifies the artifact catalog and every node payload.
  return (await createRuntime(
    bytes,
    catalog,
    options.nodes ?? nodes,
  )) as Runtime;
}
