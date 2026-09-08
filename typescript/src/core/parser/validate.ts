import type {
  AstDocument,
  Validators,
  ParseError,
  BuildError,
} from "./types.js";
import type { Check } from "../validation/fields.js";
import { fields, string, uint, object } from "../validation/fields.js";

// Missing host support is a configuration error, not corrupt Wasm output.
export class UnsupportedNodeError extends Error {}

export function validateDocument(
  document: unknown,
  source: string,
  nodes: Validators,
  allowUnknown = false,
) {
  if (
    !fields(document, {
      source: (value): value is string => string(value) && value === source,
      children: Array.isArray,
    })
  )
    throw new Error("Invalid document contract");
  const bytes = new TextEncoder().encode(source);
  if (bytes.length > 65536) throw new Error("Source byte limit");
  const boundary = (offset: unknown): offset is number =>
    uint(offset) &&
    offset <= bytes.length &&
    (offset === bytes.length || (bytes[offset] & 0xc0) !== 0x80);
  const pending: [unknown, number, number, number][] = document.children.map(
    (node: unknown) => [node, 0, bytes.length, 1],
  );
  let count = 0;
  while (pending.length) {
    const [node, start, end, depth] = pending.pop()!;
    if (
      ++count > 16384 ||
      depth > 64 ||
      !fields(
        node,
        {
          kind: (value): value is string => string(value) && value.length > 0,
          span: (span) => fields(span, { start: boundary, end: boundary }),
          data: object,
        },
        { children: Array.isArray },
      )
    )
      throw new Error("Invalid AST node contract");
    if (
      node.span.start < start ||
      node.span.end > end ||
      node.span.start > node.span.end
    )
      throw new Error("Invalid AST span");
    const validate = nodes.get(node.kind);
    if (!validate && !allowUnknown)
      throw new UnsupportedNodeError(`Unsupported node: ${node.kind}`);
    if (validate && !validate(node.data))
      throw new Error(`Invalid node: ${node.kind}`);
    for (const child of node.children ?? [])
      pending.push([child, node.span.start, node.span.end, depth + 1]);
  }
  return document as AstDocument;
}

export function validateError(error: unknown): error is ParseError {
  const shapes: Record<string, Record<string, Check>> = {
    invalid_utf8: {},
    internal_error: {},
    resource_limit: { resource: string },
  };
  return (
    object(error) &&
    typeof error.code === "string" &&
    Object.hasOwn(shapes, error.code) &&
    fields(error, { code: string, ...shapes[error.code] })
  );
}

export function validateBuildError(error: unknown): error is BuildError {
  const shapes: Record<string, Record<string, Check>> = {
    duplicate: { element: string },
    missing: { element: string },
    duplicate_name: { scope: string, name: string },
    invalid_definition: { reason: string },
  };
  return (
    object(error) &&
    typeof error.code === "string" &&
    Object.hasOwn(shapes, error.code) &&
    fields(error, { code: string, ...shapes[error.code] })
  );
}
