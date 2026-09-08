import { readFile, writeFile } from "node:fs/promises";
import { createHash } from "node:crypto";
import { gzipSync, brotliCompressSync } from "node:zlib";
import { loadRuntime } from "../dist/parser/index.js";
const bytes = await readFile(
  new URL("../dist/parser.wasm", import.meta.url),
);
const core = await loadRuntime(bytes);
const contract = {
  ...core.contract,
  sha256: createHash("sha256").update(bytes).digest("hex"),
  bytes: bytes.length,
  gzipBytes: gzipSync(bytes).length,
  brotliBytes: brotliCompressSync(bytes).length,
};
await writeFile(
  new URL("../dist/contract.json", import.meta.url),
  JSON.stringify(contract, null, 2) + "\n",
);
console.log(`Wasm ABI ${contract.abiVersion} / AST ${contract.astVersion}: ${contract.bytes} bytes, SHA-256 ${contract.sha256}`);
