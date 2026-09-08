# Examples

Run `npm ci` and `npm run build` at the repository root first.

| Language | Source | Command |
| --- | --- | --- |
| Rust | [main.rs](rust/src/main.rs) | `npm run example:rust` |
| Go | [main.go](go/main.go) | `npm run example:go` |
| TypeScript | [main.mts](typescript/main.mts) | `npm run example:ts` |

`npm run examples` runs all three. Each example parses with traQ V1 and constructs an independent grammar without the math extension. They print the AST and explicitly release host resources. Rust uses RAII for cleanup.

Rust resolves the pinned `markdown-traq` Git dependency. Go uses a local `replace` for this repository's module. TypeScript imports this package's public exports. No registry publication or application checkout is required.

The Go example accepts `-wasm /path/to/parser.wasm`. TypeScript uses `readFile` in Node.js; browser applications can pass the result of `fetch(wasmUrl).then(response => response.arrayBuffer())` to the same `loadRuntime` API.

HTML rendering examples live in [traq-markdown-it](https://github.com/traPtitech/traq-markdown-it). Native notification and extraction examples live in [trap](https://github.com/traq-markdown-parser/trap/tree/main/crates/traq-processing/examples).
