import type { States, Owner } from "./types.js";
export const metadata = new WeakMap<object, States[keyof States]>();
export function data<K extends keyof States>(
  value: object,
  kind: K,
): States[K] {
  const state = metadata.get(value);
  if (!state || state.kind !== kind) throw new TypeError("Expected " + kind);
  return state as States[K];
}
export function owned<K extends "rule" | "grammar">(
  value: object,
  kind: K,
  runtime: Owner,
): States[K] {
  const state = data(value, kind);
  if (state.runtime !== runtime)
    throw new TypeError("Object belongs to a different Runtime");
  return state as States[K];
}
