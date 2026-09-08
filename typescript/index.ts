import { sha256, inputBytes } from "./generated/artifact.js";
import type { Document } from "./generated/nodes.js";
import type { Preset } from "./generated/presets.js";
export { presets } from "./generated/presets.js";
export { isKnownNode } from "./generated/nodes.js";
export type { Preset } from "./generated/presets.js";
export type {
  Document,
  Node,
  NodeKind,
  ParseError,
} from "./generated/nodes.js";

interface Wasm extends WebAssembly.Exports {
  memory: WebAssembly.Memory;
  input_ptr(length: number): number;
  output_ptr(): number;
  configure(): number;
  parse(mode: number): number;
}
export interface Parser {
  parse(source: string): Document;
  parseInline(source: string): Document;
  dispose(): void;
}

/** The preset and all grammar behavior are defined in Rust. */
export async function createParser(
  bytes: Uint8Array,
  preset: Preset,
): Promise<Parser> {
  const artifact = new Uint8Array(bytes).buffer;
  const digest = new Uint8Array(
    await crypto.subtle.digest("SHA-256", artifact),
  );
  if (
    Array.from(digest, (byte) => byte.toString(16).padStart(2, "0")).join(
      "",
    ) !== sha256
  )
    throw new Error("Wasm does not match this SDK build");
  let wasm: Wasm | undefined = (await WebAssembly.instantiate(artifact, {}))
    .instance.exports as Wasm;
  const encoder = new TextEncoder(),
    decoder = new TextDecoder("utf-8", { fatal: true, ignoreBOM: true });
  function call(
    operation: "configure" | "parse",
    source: string,
    mode = 0,
  ): Document {
    if (!wasm) throw new Error("Parser is disposed");
    if (typeof source !== "string")
      throw new TypeError("Expected source string");
    if (source.length > inputBytes)
      throw new RangeError("Wasm input limit exceeded");
    const input = encoder.encode(source);
    if (input.length > inputBytes)
      throw new RangeError("Wasm input limit exceeded");
    if (decoder.decode(input) !== source)
      throw new TypeError("Source contains an unpaired surrogate");
    const pointer = wasm.input_ptr(input.length);
    if (!pointer) throw new RangeError("Wasm input limit exceeded");
    let result: { document: Document; error?: unknown };
    try {
      new Uint8Array(wasm.memory.buffer, pointer, input.length).set(input);
      const length =
        operation === "configure" ? wasm.configure() : wasm.parse(mode);
      result = JSON.parse(
        decoder.decode(
          new Uint8Array(wasm.memory.buffer, wasm.output_ptr(), length),
        ),
      );
    } catch (error) {
      wasm = undefined;
      throw error;
    }
    // Rust validates the AST before encoding; the digest pins its matching types.
    if (result.error)
      throw new Error("Markdown: " + JSON.stringify(result.error), {
        cause: result.error,
      });
    return result.document;
  }
  call("configure", preset);
  return Object.freeze({
    parse: (source: string) => call("parse", source),
    parseInline: (source: string) => call("parse", source, 1),
    dispose() {
      wasm = undefined;
    },
  });
}
