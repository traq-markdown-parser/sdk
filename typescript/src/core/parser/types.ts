import type {
  Plugin as Declaration,
  PluginGroup,
} from "../definitions/index.js";
import type { Plugin } from "./plugin.js";
import type { GrammarBuilder } from "./builder.js";
export type { ParseError } from "./ParseError.js";

export type Phase = "inline" | "block" | "text";
declare const identity: unique symbol;
export interface Rule<P extends Phase = Phase> {
  readonly [identity]: P;
  readonly name: string | null;
  readonly phase: P;
}
export interface Grammar {
  readonly [identity]: "grammar";
  readonly plugins: readonly Plugin[];
  toBuilder(): GrammarBuilder;
  describe(): string;
  dispose(): void;
}
export interface AstNode {
  kind: string;
  data: unknown;
  span: { start: number; end: number };
  children?: AstNode[];
}
export interface AstDocument {
  source: string;
  children: AstNode[];
}
export interface ParserOptions<AllowUnknown extends boolean = boolean> {
  allowUnknownNodes?: AllowUnknown;
}
export type Validators = ReadonlyMap<string, (data: unknown) => boolean>;
export interface RuntimeOptions {
  nodes?: Validators;
}
export type BuildError =
  | { code: "duplicate" | "missing"; element: string }
  | { code: "duplicate_name"; scope: string; name: string }
  | { code: "invalid_definition"; reason: string };
export interface CatalogTree {
  readonly [key: string]: number | CatalogTree;
}
export interface Catalog {
  groups: { name: string; parent: number | null }[];
  rules: { name: string | null; phase: Phase }[];
  plugins: {
    name: string;
    group: number | null;
    rules: number[];
    text: number[];
  }[];
  presets: { plugins: number[]; order: number[]; description: string }[];
  exports: { plugins: CatalogTree; presets: CatalogTree };
}
export interface Contract {
  readonly abiVersion: 2;
  readonly astVersion: 4;
  readonly catalog: unknown;
  readonly limits: Readonly<{
    inputBytes: number;
    outputBytes: number;
    memoryBytes: number;
    grammars: number;
  }>;
}
export interface PluginState {
  kind: "plugin";
  symbol: object;
  declaration: Declaration;
  name: string;
  group: PluginGroup | null;
  rules: Rule[];
  text: { runtime: Owner; index: number }[];
  frozen: boolean;
}
export interface Snapshot {
  plugins: PluginState[];
  order: number[];
}
export interface GrammarRecord {
  snapshot: Snapshot;
  description: string;
  handle: number | null;
  references: number;
}
export interface Owner {
  preset(
    snapshot: Snapshot,
    compiled: { description: string; handle?: number | null },
  ): Grammar;
  build(snapshot: Snapshot): Grammar;
  prepare(record: GrammarRecord): void;
  parse(
    record: GrammarRecord,
    source: string,
    mode: number,
    allowUnknown: boolean,
  ): AstDocument;
  release(record: GrammarRecord): void;
}
export interface States {
  plugin: PluginState;
  rule: {
    kind: "rule";
    runtime: Owner;
    index: number;
    name: string | null;
    phase: Phase;
  };
  grammar: {
    kind: "grammar";
    runtime: Owner;
    record: GrammarRecord;
    disposed: boolean;
  };
}
