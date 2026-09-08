// Generated from the Rust node contracts. Do not edit.
import type { Span } from './Span.js';
import type { NodeKind } from './NodeKind.js';
export type Node<AllowUnknown extends boolean = false> =
  (NodeKind | (AllowUnknown extends true ? { kind: string; data: unknown } : never)) &
  { span: Span; children?: Node<AllowUnknown>[] };
