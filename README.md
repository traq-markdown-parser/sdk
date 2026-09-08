# traQ Markdown SDK

Rust の Markdown パーサーを、一つの WebAssembly バイナリとして TypeScript と Go から呼び出す SDK です。文法・AST の契約・検証は Rust が所有します。

- npm: **`@traq-markdown-parser/ts`**
- Go module: **`github.com/traq-markdown-parser/sdk/go`**
- Wasm: ABI **3** / AST **4**

文法の実装は [core](https://github.com/traq-markdown-parser/core)、[commonmark](https://github.com/traq-markdown-parser/commonmark)、[trap](https://github.com/traq-markdown-parser/trap) にあります。このリポジトリで配布するプリセットを選び、Wasm と対応する型を生成します。

## ビルド

Node.js 24 以降、Go 1.25 以降、rustup が必要です。Rust と Wasm target は `rust-toolchain.toml` で固定しています。

```sh
npm ci
npm run build
npm run examples
```

他のリポジトリの checkout は不要です。Cargo が固定した Git revision を取得します。Wasm・JavaScript・型定義は `dist/` に出力します。Rust から生成する TypeScript / Go のソースと、対応する Rust ビルド ID はソース管理します。バイナリと SDK は同じソース・固定依存から生成した組を配布してください。

まだレジストリへ公開していません。TypeScript は `npm pack` で作ったアーカイブを利用できます。

## TypeScript

```ts
import { createRuntime, presets } from '@traq-markdown-parser/ts'

const runtime = await createRuntime(wasmBytes)
try {
  const parser = runtime.createParser(presets.traq.v1)
  const document = parser.parse('**hello** :stamp:')
  const inline = parser.parseInline('**hello**')
} finally {
  runtime.dispose()
}
```

`wasmBytes` は `Uint8Array` です。Node.js は `@traq-markdown-parser/ts/parser.wasm` を `readFile` で読み、ブラウザーは `new Uint8Array(await response.arrayBuffer())` を渡します。Runtime と Parser は再利用できます。同じ Runtime から異なるプリセットの Parser も作成できます。

ノード型は判別可能な union です。文法別の payload 型と任意利用の guard は `/commonmark/nodes`・`/generic/nodes`・`/trap/nodes`、全体の一覧は `/nodes` から利用できます。

## Go

```go
import markdown "github.com/traq-markdown-parser/sdk/go"

runtime, err := markdown.NewRuntime(ctx, wasmBytes)
if err != nil { return err }
defer runtime.Close(ctx)

parser, err := runtime.NewParser(ctx, markdown.PresetTraQV1)
if err != nil { return err }
defer parser.Close(ctx)
document, err := parser.Parse(ctx, "**hello** :stamp:")
```

`Parse` / `ParseInline` は `*markdown.Document` を返します。`Node.Data` の具体的な型も Rust から生成します。Runtime は Wasm のコンパイル結果を共有します。一つの Parser の呼び出しは直列化し、並列実行には同じ Runtime から Parser を複数作ります。`Parser.Close` はその Parser だけ、`Runtime.Close` は配下の全 Parser を解放します。

## 文法の変更

TypeScript の `presets.commonmark` / `presets.traq.v1`、Go の `PresetCommonMark` / `PresetTraQV1` は Rust が公開するプリセットから生成します。独自の文法は Rust で組み立て、配布層からプリセットとして公開して再ビルドします。ホスト API はプリセットの選択、解析、解放に絞っています。

手書きの実装は TypeScript の `index.ts` と任意の payload guard 用の `validation.ts`、Go の `parser.go` です。文法ビルダー、Plugin / Rule のミラー、文法ハンドル、worker pool は持ちません。

HTML / CSS は [traq-markdown-it](https://github.com/traPtitech/traq-markdown-it) が担当します。

[API と実装](docs/implementation.md)、[実行例](examples/README.md)、[開発と検証](CONTRIBUTING.md) を参照してください。
