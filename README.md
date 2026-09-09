# traQ Markdown

traQ 向けの Markdown 文法・通知処理の構成と、Rust・WebAssembly・Go・TypeScript 向けの配布を所有します。文法・AST の契約・検証は Rust が所有します。

- npm: **`@traq-markdown-parser/traq`**
- Go module: **`github.com/traq-markdown-parser/traq/go`**
- Wasm: ABI **3** / AST **4**

文法の実装は [core](https://github.com/traq-markdown-parser/core)、[commonmark](https://github.com/traq-markdown-parser/commonmark)、[trap-extension](https://github.com/traq-markdown-parser/trap-extension) にあります。このリポジトリで配布するプリセットを選び、Wasm と対応する型を生成します。

## 構成と責任

| 場所 | 責任 |
| --- | --- |
| `crates/grammar` | CommonMark・汎用拡張・traP 拡張を選択し、文法プリセットを構成 |
| `crates/processing` | 通知の表示方針と参照抽出のプリセットを構成 |
| `crates/processor` | 一度の解析から通知テキストと参照一覧を生成 |
| `crates/wasm` | 配布する文法・ノード型・処理 API を Wasm として公開 |
| `go`・`typescript`・`scripts/contracts` | この配布物に対応する bindings と型生成 |

共通機構は core、CommonMark と汎用拡張は commonmark、traP 固有の拡張部品は
trap-extension にあります。このリポジトリがそれらに依存し、traQ 向けに組み合わせます。
下位の部品はこの配布物に依存しません。bindings はここで選んだ型とプリセットに対応します。

## ビルド

Node.js 24 以降、Go 1.25 以降、rustup が必要です。Rust と Wasm target は `rust-toolchain.toml` で固定しています。

```sh
npm ci
npm run build
npm run examples
```

Rust の依存は Cargo が固定した Git revision を取得します。TypeScript のローカルビルドでは、同じ親ディレクトリに core・commonmark・trap-extension を置き、依存順に `npm ci` と `npm run build` を実行しておきます。Wasm・JavaScript・型定義は `dist/` に出力します。Rust から生成する TypeScript / Go のソースと、対応する Rust ビルド ID はソース管理します。バイナリと SDK は同じソース・固定依存から生成した組を配布してください。

まだレジストリへ公開していません。TypeScript は4パッケージの `npm pack` アーカイブを利用できます。

## TypeScript

```ts
import { createRuntime, presets } from '@traq-markdown-parser/traq'

const runtime = await createRuntime(wasmBytes)
try {
  const parser = runtime.createParser(presets.traq.v1)
  const document = parser.parse('**hello** :stamp:')
  const inline = parser.parseInline('**hello**')
} finally {
  runtime.dispose()
}
```

`wasmBytes` は `Uint8Array` です。Node.js は `@traq-markdown-parser/traq/parser.wasm` を `readFile` で読み、ブラウザーは `new Uint8Array(await response.arrayBuffer())` を渡します。Runtime と Parser は再利用できます。同じ Runtime から異なるプリセットの Parser も作成できます。

ノード型は判別可能な union です。文法別の payload 型と任意利用の guard は `@traq-markdown-parser/commonmark/nodes`・`@traq-markdown-parser/commonmark/generic/nodes`・`@traq-markdown-parser/trap-extension/nodes`、配布物全体の一覧は `@traq-markdown-parser/traq/nodes` から利用できます。

## Go

```go
import markdown "github.com/traq-markdown-parser/traq/go"

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

Wasm のホスト実装は TypeScript の `index.ts` と Go の `parser.go` です。共通 AST 型と payload guard の検証部品は core、構文の生成型はそれぞれのリポジトリが所有します。文法ビルダー、Plugin / Rule のミラー、文法ハンドル、worker pool は持ちません。

HTML 描画の共通基盤は core、構文別の描画は commonmark と trap-extension が担当します。traQ の描画構成・inline preview・CSS はこのリポジトリの `typescript/renderer` が所有します。

[API と実装](docs/implementation.md)、[実行例](examples/README.md)、[開発と検証](CONTRIBUTING.md) を参照してください。

## 通知テキストと参照抽出

`Processor` は原文を一度だけ Rust AST に解析し、その AST を通知用レンダラーと参照抽出器が借用します。ホストとの間では最終結果だけを渡します。

```ts
import { createRuntime, processors } from '@traq-markdown-parser/traq'

const runtime = await createRuntime(wasmBytes)
try {
  const processor = runtime.createProcessor(processors.traq.v1, {
    origin: 'https://q.example.test',
  })
  const { notificationText, references } = processor.process('**hello** !!secret!!')
  // notificationText: "hello ██████"
  // references: { mentions: [], groupMentions: [], channelLinks: [] }
} finally {
  runtime.dispose()
}
```

```go
processor, err := runtime.NewProcessor(ctx, markdown.ProcessorPresetTraQV1,
    markdown.ProcessorOptions{Origin: "https://q.example.test"})
if err != nil { return err }
defer processor.Close(ctx)
result, err := processor.Process(ctx, "**hello** !!secret!!")
```

Parser と Processor は同じ Runtime のコンパイル結果を共有し、独立した instance と設定を持ちます。Processor のライフサイクルと Go の直列化・キャンセル規則は Parser と同じです。プリセット・設定・結果の型は Rust から生成します。ネイティブ Rust では `traq-markdown-processor` crate の `Processor` を使います。

通知は spoiler をマスクし、空白を正規化した一行のテキストです。参照はユーザー・グループ・チャンネルの UUID を種類別に返し、文書順・重複・spoiler 内の参照を保持します。コード内の文字列は参照として抽出しません。`origin` は通知中の traQ 添付・引用 URL の表示判定用です。添付・引用 ID の抽出と Bot 用 PlainText はこの API の対象外で、アプリ側の方針として残ります。

## TypeScript / HTML rendering

TypeScript の実装は各リポジトリの責務に合わせて配置しています。

| npm package | 責務 |
| --- | --- |
| `@traq-markdown-parser/core` | 共通 AST 型、HTML handler・Plugin・PresetBuilder、契約検証と生成の基盤 |
| `@traq-markdown-parser/commonmark` | CommonMark・汎用拡張の生成ノード型と HTML 描画 |
| `@traq-markdown-parser/trap-extension` | traP の生成ノード型・参照・スタンプ等の HTML 描画 |
| `@traq-markdown-parser/traq` | Wasm / Go / TypeScript 配布、traQ の描画構成・preview・CSS |

ローカル開発では4リポジトリを同じ親ディレクトリに置き、core → commonmark → trap-extension → traq の順に `npm install`・`npm run build` を実行します。npm パッケージはまだ未公開です。配布検証は traq の `npm run check:package` で4パッケージを pack し、独立した consumer で実行します。

AST の共通形は core の `typescript/ast.ts` に一度だけ定義し、traq の生成 bindings はそれを構文の union で特殊化します。構文の payload は Rust を正として生成し、commonmark と trap-extension の `npm run generate:bindings` でそれぞれの契約 crate から再生成できます。

HTML API は `/renderer` サブパスです。traQ は `@traq-markdown-parser/traq/renderer/v1` の `messageRenderer`、CSS は `@traq-markdown-parser/traq/index.css` を利用します。

### HTML の利用例

```ts
import { createRuntime, presets } from "@traq-markdown-parser/traq";
import { messageRenderer } from "@traq-markdown-parser/traq/renderer/v1";
import "@traq-markdown-parser/traq/index.css";

const runtime = await createRuntime(wasmBytes);
const parser = runtime.createParser(presets.traq.v1);
const view = messageRenderer({ origin: "https://q.example.test" });
const { renderedText, embeddings } = view.render(parser.parse(source));
runtime.dispose();
```

Wasm の起動は利用側が明示的に行います。`/renderer` を import しても Wasm runtime は読み込みません。描画のカスタマイズは `@traq-markdown-parser/core/renderer` の `Plugin`・`PresetBuilder` と、各構文の `/renderer` を利用します。

## Corpus comparison

Use `npm run corpus:collect`, `npm run corpus:compare`, and `npm run corpus:report` to collect messages and generate offline HTML/MHTML difference reports. See [Corpus comparison](docs/CORPUS.md) for inputs, options, and output files.
