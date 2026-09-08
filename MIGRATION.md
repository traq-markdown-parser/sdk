# Repository split

Extracted from [the original parser repository](https://github.com/traq-markdown-parser/core/tree/131fd3e9be86aa342b29666f9f31602d8b53bda4), including the subsequent TypeScript and renderer separation work.

The core repository preserves the original Git history. The other repositories start with their owned source trees. Package names and Rust module paths are preserved, so moving files does not change generated AST type keys.

## traQ distribution naming (2026-09-08)

The repository `sdk` is now `traq`: it owns traQ's grammar and processing
composition together with its Wasm and language bindings. Extension components
live in `trap-extension` (formerly `trap`). The dependency direction is
`traq -> core / commonmark / trap-extension`.

| Previous | Current |
| --- | --- |
| `traq-markdown-parser/sdk` | `traq-markdown-parser/traq` |
| `@traq-markdown-parser/ts` | `@traq-markdown-parser/traq` |
| `github.com/traq-markdown-parser/sdk/go` | `github.com/traq-markdown-parser/traq/go` |
| `traq-markdown-parser/trap` | `traq-markdown-parser/trap-extension` |

Update imports, package dependencies and local checkout paths, then rebuild the
Wasm and bindings together. These packages remain unpublished; no compatibility
alias package is provided. Rust package names and AST type keys remain unchanged.

`crates/traq`, `crates/traq-processing` and the text composition example moved
here from the extension repository at `bf220c640d23f39e33fa0a52147cdf3bd32998ff`.
The Wasm build ID also covers the source and manifests of those local crates.
