# traQ processing presets

通知の描画と参照の抽出を組み合わせる配布 crate です。通常の依存に parser・codec を含まず、
各実装は `commonmark-text` / `generic-text` / `trap-text` / `trap-extraction` が所有します。

```rust
use markdown_extractor::Extractor;
use markdown_renderer::Renderer;
use markdown_traq_processing::presets::traq::v1;

let renderer = Renderer::new(&v1::notification::preset("https://q.example.test")?);
let extractor = Extractor::new(&v1::references::preset()?);
// parser が生成した同じ Document を借用する。
let text = renderer.render(&document)?;
let references = extractor.extract(&document)?;
```

実行可能な例は [examples/notification.rs](examples/notification.rs) にあります。
リポジトリ root で `cargo run -p markdown-traq-processing --example notification` を実行してください。
AST の JSON 変換を挟まず、最終結果だけを JSON として出力します。

`notification::builder(origin)` / `references::builder()` は編集可能な builder を返します。
Plugin の `add` / `remove` 後に `build()` し、汎用 Renderer / Extractor に渡せます。
作成済みの実行インスタンスは、別の構成の編集や構築によって変化しません。

通知はブロック間の改行を残します。一行への整形は利用側で行います。
spoiler は描画した内容の Unicode scalar value 数だけ `█` を並べ、改行は保持します。
空の origin はリンクの特別表示を無効にします。origin は最大 2,048 UTF-8 バイトです。
同じ origin の `/files/{uuid}` / `/messages/{uuid}` は添付・引用の表示になりますが、
明示的なリンクラベルとコード内の文字列は維持します。URL 判定は文字列による比較です。

参照抽出は user / group / channel の ID を正規化して返します。順序・重複を保持し、
spoiler 内も対象です。コード内は parser が参照ノードを生成しないため対象になりません。
添付・引用 ID の抽出、ユーザー情報の解決、通知送信は含みません。

JSON 設定や Wasm のリソース管理は SDK 接続層の責務です。
