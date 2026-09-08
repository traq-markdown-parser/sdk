# traQ Markdown SDK

Rust の Markdown 文法を WebAssembly として配布し、TypeScript と Go から同じ AST を利用するための SDK です。

- npm: **`@traq-markdown-parser/ts`**
- Go module: **`github.com/traq-markdown-parser/sdk/go`**
- Wasm: ABI **2** / AST **4**

文法の実装は [core](https://github.com/traq-markdown-parser/core)、[commonmark](https://github.com/traq-markdown-parser/commonmark)、[trap](https://github.com/traq-markdown-parser/trap) が所有します。このリポジトリは収録する文法・ノード契約を選択し、Wasm と対応する bindings を一緒に生成・検証します。

## ビルド

Node.js 24 以降、Go 1.25 以降、rustup が必要です。Rust と Wasm target は `rust-toolchain.toml` で固定しています。Windows の PowerShell でも実行できます。

```sh
npm ci
npm run build
npm run examples
```

他のリポジトリの checkout は不要です。Cargo が manifest に固定された Git revision を取得します。Wasm と JavaScript / 型定義は `dist/` に生成し、ソース管理には含めません。Rust 由来の TypeScript 契約型・検証コードと Go binding は生成済みソースも管理します。

パッケージはまだレジストリへ公開していません。ローカルでは `npm pack` で生成したアーカイブを利用できます。

## TypeScript

```ts
import { loadRuntime } from '@traq-markdown-parser/ts'

const runtime = await loadRuntime(wasmBytes)
try {
  const parser = runtime.parser(runtime.presets.traq.v1)
  try {
    const document = parser.parse('**hello** :stamp:')
    const inline = parser.parseInline('**hello**')
  } finally {
    parser.dispose()
  }
} finally {
  runtime.dispose()
}
```

Node.js では `@traq-markdown-parser/ts/parser.wasm` を `readFile` で読み、ブラウザーでは配布した Wasm URL を `fetch` して `ArrayBuffer` を渡します。Runtime と Parser はメッセージごとに作り直さず再利用できます。

既知ノードは判別可能な union です。文法別の payload 型と guard は `/commonmark/nodes`・`/generic/nodes`・`/trap/nodes`、全体の一覧は `/nodes` から利用できます。`/definitions` は parser と外部 renderer が共有する Plugin 宣言です。

HTML / CSS は別パッケージの [traq-markdown-it](https://github.com/traPtitech/traq-markdown-it) が担当します。bindings には HTML renderer や markdown-it への依存はありません。

## 文法と処理

`runtime.presets.commonmark` は CommonMark、`runtime.presets.traq.v1` は traQ V1 です。`toBuilder()` から plugin を追加・削除・並べ替えた独立した文法を作れます。preset は最初の Parser 作成時にコンパイルされ、独自構成の `build()` はその場で検証・構築されます。

文法版と保存済みメッセージの対応は利用側で管理します。原文を適切な preset で再解析する運用を想定し、永続 AST の互換層は設けません。

[API と所有権](docs/implementation.md)、[Rust / Go / TypeScript の実行例](examples/README.md)、[開発と検証](CONTRIBUTING.md) を参照してください。
