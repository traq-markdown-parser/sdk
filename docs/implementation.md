# 実装と利用

## 文法から Parser を作る

```rust
use markdown_traq::{Parser, presets, syntax::extensions::math};

let parser = presets::traq::v1::parser();
let document = parser.parse("**本文** :stamp:")?;
let inline = parser.parse_inline("**本文**")?;

let mut builder = presets::traq::v1::grammar().to_builder();
builder.remove(math::plugin())?;
let grammar = builder.build()?;
let without_math = Parser::new(&grammar);
```

Grammar は確定した文法、Parser はその共有データと制限値を保持します。
本文や参照定義などの状態は各 parse に属します。Grammar を解放しても、作成済みの Parser は有効です。
CommonMark 0.31.2 は presets::commonmark、traQ は presets::traq::v1 にあります。

文法は CommonMark・汎用拡張・traP 拡張に分離しています。
型付きルール、text provider、SourceView、Budget の説明は [parser core](https://github.com/traq-markdown-parser/core/blob/main/crates/parser/README.md)、
組み合わせは [文法 preset](https://github.com/traq-markdown-parser/trap/blob/main/crates/traq/README.md) を参照してください。
空の builder は text provider を持たないため build で拒否します。

## 名前・階層・同一性

Rust の共有宣言:

```rust
use markdown_definitions::Plugin;
let generic = Plugin::group("generic");
let github = generic.group("github");
let issue = github.new("issue");
```

名前は必須の表示名で、ID ではありません。同じ group インスタンスを使い回して階層を作ります。
parser / renderer の実装 Plugin は、この共有宣言を参照して生成します。
同じ plugin / rule の二重登録と、採用された同じ親の表示名の衝突を拒否します。
未採用の定義は衝突検査の対象になりません。

登録・削除・順序変更はインスタンスを渡します。変更は登録済みの snapshot に影響しません。
TS / Go の組み込み文法 Plugin は読み取り専用です。独自 Plugin を作り、組み込み Rule を追加して構成できます。
表示名の変更で新しい Plugin を作る API はありません。

## TypeScript / JavaScript

```ts
import { loadRuntime, Plugin } from '@traq-markdown-parser/ts'
import * as definitions from '@traq-markdown-parser/ts/definitions'
import { names } from '@traq-markdown-parser/ts/nodes'

const runtime = await loadRuntime(wasmBytes)
const grammar = runtime.presets.traq.v1.toBuilder()
  .remove(runtime.plugins.generic.math).build()
const parser = runtime.parser(grammar)
grammar.dispose()

const document = parser.parseInline('[資料](https://example.com)')
for (const node of document.children) {
  if (node.kind === names.Link) {
    const destination: string = node.data.destination
  }
}
parser.dispose()
runtime.dispose()

const generic = definitions.Plugin.group('custom')
const declaration = generic.new('math')
const math = new Plugin(declaration)
```

通常の parse は生成済みのノード検証を使い、codec の指定を要求しません。
独自の検証登録は loadRuntime の nodes オプションに渡します。
未登録の型は実際に parse で返ったときにエラーになります。
allowUnknownNodes は既定で false。true は未知型を受け取り、原文表示などへ回す用途です。
既知の型で payload が不正な場合は、未知型の許可にかかわらず拒否します。

省略 / false の戻り値は Document、true は Document<true> です。
両者は同じノードを返しますが、後者の型は既知型と未知型の両方を含みます。
isKnownNode は生成済みの payload 検証を使い、対象ノードを既知型へ絞り込みます。
子ノードが既知型であることまでは保証しません。子も個別に確認します。

```ts
import { isKnownNode } from '@traq-markdown-parser/ts'

const parser = runtime.parser(runtime.presets.traq.v1, { allowUnknownNodes: true })
for (const node of parser.parseInline(source).children) {
  if (isKnownNode(node) && node.kind === names.Link) {
    console.log(node.data.destination) // string
  }
}
parser.dispose()
```

isKnownNode は受信済み AST の payload を確認するための関数です。
任意の JSON の構造・span・意味を検証する API ではありません。

Grammar の toBuilder / describe だけでは Wasm の文法をコンパイルしません。
preset から初めて parser を作る際に初期化し、同じ Grammar から作る Parser で共有します。
独自構成の build はその場で検証・構築します。dispose は繰り返し呼べます。

## Go

```go
runtime, err := core.New(ctx, wasmBytes)
if err != nil { return err }
defer runtime.Close(ctx)
parser, err := runtime.Parser(ctx, runtime.Presets.TraQ.V1)
if err != nil { return err }
defer parser.Close()
result, err := parser.ParseInline(ctx, "[資料](https://example.com)")
if err != nil { return err }
for _, node := range result.Document.Children {
    if link, ok := node.Payload.(commonmark.Link); ok {
        fmt.Println(link.Destination)
    }
}
```

core は go/core、commonmark は go/extensions/commonmark です。
Node は Kind / Span / Children / Payload を持ち、Payload は全種類が生成済みの具体的な型です。
生 JSON をノードごとに保持しません。Result.JSON は必要な呼び出し向けに文書全体の通信表現を残します。

階層は core.NewPluginGroup("generic")、子 group は generic.Group("github")、
plugin は generic.New("math") で作ります。Builder.Add / Remove / Before は error を返します。
Grammar.ToBuilder().Build(ctx) で独立した構成を作れます。

Runtime の4つの Wasm worker を全 Parser が共有します。builder は一つの goroutine 内で編集し、
Grammar / Parser は並行利用できます。Parser(ctx, grammar) は初期化とキャンセルを扱い、失敗後に再試行できます。
最後の所有者が Close すると文法を解放します。処理中の呼び出しも文法を保持するため、並行する Close から保護されます。
停止した Wasm instance は次回利用時に再生成します。Go は未知ノードを許可しません。

## AST と検証

Rust の各ノードは契約 package が所有する struct です。AST core は文法の型一覧を持ちません。
JSON は codec の境界でのみ生成します。Document 自体には Serialize / Deserialize を要求しません。

通信形式は全種類で {kind, span, data, children}。kind は Rust 型から生成する通信キーで、
利用側は生成された names や型を使います。キーを手書きしたり永続化したりする契約ではありません。
source は原文、span は原文の UTF-8 バイト範囲です。

TS / Go は共通構造、payload の型、span、サイズ・深さ・ノード数の上限を検査します。
Heading の段数や葉ノードの子の禁止などの意味検証は、型の NodeData::validate() を
Rust の parser と codec から呼び出します。Go の単独 ast.Decode は意味検証を複製しません。

Wasm の配布層にある一つの型一覧から、codec 登録、schema、TS / Go の型を生成します。
parser core と renderer core はこの一覧を参照しません。生成器は閉じた object、string、boolean、
文字列 enum、nullable、u8 / u32 を扱い、未対応の型・制約や型名の衝突は生成時に拒否します。

## 描画と保存

HTML / CSS は別リポジトリの [traq-markdown-it](https://github.com/traPtitech/traq-markdown-it) が担当します。parser が返した Document を renderer に渡し、HTML 文字列を得ます。bindings は描画ライブラリへ依存しません。

```ts
import { renderer } from '@traptitech/traq-markdown-it'
import { v1 } from '@traptitech/traq-markdown-it/traq'

const view = renderer(v1.html({ store }))
const html: string = view.render(parser.parse(source))
```

Rust の通知・参照抽出は [traq-processing](https://github.com/traq-markdown-parser/trap/tree/main/crates/traq-processing) を参照してください。同じ native Document を借用でき、JSON 変換を挟みません。

保存済みメッセージの文法版と Parser の対応は利用側が管理します。SDK / AST に syntaxVersion はありません。原文を適切な preset で再解析する運用で、永続 AST の互換層や自動キャッシュは設けません。

## 配布と制限

ABI 2 / AST 4。旧 AST 3 の Wasm と新 SDK の混在は初期化時に拒否します。
入力64KiB、出力1MiB、Wasm memory32MiB、コンパイル済み文法256件です。未使用 preset は文法枠を消費しません。
WebAssembly.Memory は解放しても縮小せず、allocator が領域を再利用します。

npm run build が Wasm・型・検証登録・catalog を生成します。
検証と独立 consumer の手順は [開発](../CONTRIBUTING.md)、実行例は [examples](../examples/README.md) を参照してください。
