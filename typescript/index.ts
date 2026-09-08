import { buildId, inputBytes } from "./generated/artifact.js";
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

export interface Runtime {
  createParser(preset: Preset): Parser;
  dispose(): void;
}

/** Compile Wasm once; each Parser owns an independent instance and Rust preset. */
export async function createRuntime(bytes: Uint8Array): Promise<Runtime> {
  let module: WebAssembly.Module | undefined = await WebAssembly.compile(
    new Uint8Array(bytes).buffer,
  );
  const parsers = new Set<Parser>();
  return Object.freeze({
    createParser,
    dispose() {
      module = undefined;
      for (const parser of parsers) parser.dispose();
    },
  });

  function createParser(preset: Preset): Parser {
    if (!module) throw new Error("Runtime is disposed");
    let wasm: Wasm | undefined = new WebAssembly.Instance(module, {})
      .exports as Wasm;
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
      let result: { document: Document; configured?: string; error?: unknown };
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
      // Rust validates the AST before encoding; the build ID pairs its types.
      if (result.error)
        throw new Error("Markdown: " + JSON.stringify(result.error), {
          cause: result.error,
        });
      if (operation === "configure" && result.configured !== buildId)
        throw new Error("Wasm does not match this SDK build");
      return result.document;
    }
    call("configure", preset);
    const parser = Object.freeze({
      parse: (source: string) => call("parse", source),
      parseInline: (source: string) => call("parse", source, 1),
      dispose() {
        wasm = undefined;
        parsers.delete(parser);
      },
    });
    parsers.add(parser);
    return parser;
  }
}
