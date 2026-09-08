# 実装と利用

## Rust が所有するもの

文法の構成、解析、ノードの型と意味検証は Rust にあります。文法は trap リポジトリの `crates/traq/src/bindings/mod.rs`、ノード契約は SDK の `crates/wasm/src/node_types.rs` に登録します。Rust の文法ビルダーで独自構成を作り、catalog の preset として公開すると、その名前を TypeScript / Go に生成します。

Rust API は [parser core](https://github.com/traq-markdown-parser/core/blob/main/crates/parser/README.md) と [traQ presets](https://github.com/traq-markdown-parser/trap/blob/main/crates/traq/README.md) を参照してください。ネイティブ Rust では引き続き Plugin / Rule の追加・削除・並べ替えを利用できます。

## TypeScript

```ts
import { createParser, presets } from '@traq-markdown-parser/ts'
import { names } from '@traq-markdown-parser/ts/nodes'

const parser = await createParser(wasmBytes, presets.traq.v1)
try {
  for (const node of parser.parseInline('[資料](https://example.com)').children) {
    if (node.kind === names.Link) console.log(node.data.destination)
  }
} finally {
  parser.dispose()
}
```

生成時のビルド ID と渡された Wasm の初期化応答を照合し、一つの instance に指定の文法を設定します。`parse` と `parseInline` は同期関数です。`dispose` は instance への参照を解放し、以降の解析を拒否します。繰り返し解放できます。

入力は文字列で、64 KiB 以下の UTF-8 に変換できる必要があります。片側だけの surrogate は拒否します。Rust の解析エラーは `Error.cause` に構造化した詳細を保持します。入力エラーや Rust の resource limit は Parser を無効にしません。Wasm trap や通信の破損後は新しい Parser を作成します。

`isKnownNode` は生成された payload guard です。`Node<true>` / `Document<true>` は、レンダラーが独自ノードを扱うための型として使えます。SDK の解析結果は既知ノードの `Document` です。guard は対象ノードの kind と payload のみを確認し、任意の JSON 文書や span、子ノード全体を検証する API ではありません。

## Go

```go
import markdown "github.com/traq-markdown-parser/sdk/go"

runtime, err := markdown.NewRuntime(ctx, wasmBytes)
if err != nil { return err }
defer runtime.Close(ctx)

parser, err := runtime.NewParser(ctx, markdown.PresetTraQV1)
if err != nil { return err }
defer parser.Close(ctx)
document, err := parser.ParseInline(ctx, "[資料](https://example.com)")
if err != nil { return err }
for _, node := range document.Children {
    if link, ok := node.Data.(*markdown.Link); ok {
        fmt.Println(link.Destination)
    }
}
```

`Document` と全 payload は生成した型です。JSON が必要なら `json.Marshal(document)` を使います。結果は Wasm memory を参照しないため、次の解析や解放後も保持できます。

Runtime は wazero の実行環境と一度だけコンパイルした Wasm module を所有します。`runtime.NewParser(ctx, preset)` は独立した instance を作成します。Parser 内の解析呼び出しは直列化します。並列に解析する用途では Parser を複数作ります。待機中または呼び出し前の context キャンセルは instance を閉じません。実行中の Wasm を中断した場合は wazero が instance を閉じるため、同じ Runtime から新しい Parser を作成してください。自動的な再生成や pool はありません。`Parser.Close(ctx)` はその instance だけを解放します。`Runtime.Close(ctx)` は全 Parser とコンパイル済み module を解放し、以降の Parser 作成も拒否します。どちらも処理中の呼び出しを中断できます。

## 契約と生成

通信形式は `{source, children}` とノードの `{kind, span, data, children?}` です。`kind` は Rust 型から生成する通信キー、`span` は原文の UTF-8 バイト範囲です。利用側は生成した `names` / Go の定数と型を使います。

AST の構造・意味・資源制限は Rust の parser と codec が検証してから出力します。ホストはこれを再実装せず、対応する Rust ビルド ID を初期化時に照合して解析結果を読みます。Go の union decoder と TypeScript の任意利用の payload guard は Rust の契約から生成します。

生成器は object、string、boolean、文字列 enum、nullable、u8 / u32 を扱います。未対応の形・制約や payload 型名の衝突は生成エラーです。新しい形を追加する場合は生成器を拡張します。通常のノードやプリセット追加では手書きのホスト実装を変更しません。

Wasm ABI 3 は input buffer、`configure`、`parse(mode)`、output buffer の小さな通信面です。一つの instance が一つの文法を所有します。入力上限 64 KiB、出力上限 1 MiB、memory 上限 32 MiB を Rust で定義し、ホストに必要な上限も生成します。

ビルド ID は Rust ソース、Cargo.lock、manifest と固定 toolchain 設定の内容から生成する不一致検出用の値です。改行とパス表記を正規化し、ビルド環境が違っても同じ値になります。配布物の真正性を証明する署名ではありません。`dist/contract.json` は診断用に実際の Wasm の SHA-256 も記録します。

## 描画と保存

[traq-markdown-it](https://github.com/traPtitech/traq-markdown-it) は受け取った Document から HTML を作ります。レンダラーの Plugin 宣言はそのパッケージが所有します。文法を構成する Rust の Plugin とは別の API です。

通知・参照抽出のネイティブ Rust API は [traq-processing](https://github.com/traq-markdown-parser/trap/tree/main/crates/traq-processing) にあります。保存済みメッセージの文法版は利用側で管理し、原文を対応するプリセットで再解析します。永続 AST の互換層は設けません。
