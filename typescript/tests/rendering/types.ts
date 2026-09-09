import { Plugin as Declaration } from '@traq-markdown-parser/core/definitions'
import {
  createRuntime,
  presets,
  type Document
} from '@traq-markdown-parser/traq'
import { names, isKnownNode } from '@traq-markdown-parser/commonmark/nodes'
import {
  renderer,
  Plugin,
  PresetBuilder,
  type RenderContext
} from '@traq-markdown-parser/core/renderer'
import { plugin } from '@traq-markdown-parser/commonmark/renderer'
import { v1 } from '@traq-markdown-parser/traq/renderer'
import type { Store } from '@traq-markdown-parser/trap-extension/renderer'

const runtime = await createRuntime(new Uint8Array())
const parser = runtime.createParser(presets.traq.v1)
const view = renderer(v1.html())
const result: string = view.render(parser.parse('text'))
const messageView = v1.messageRenderer({ origin: 'https://q.example.test' })
const messageHtml: string = messageView.render(
  parser.parse('text')
).renderedText
const previewHtml: string = messageView.renderInline(
  parser.parse('text')
).renderedText

const openDocument = {} as Document<true>
const openResult: string = view.render(openDocument)
const custom = plugin().replace(names.Link, (node, context) => {
  if (isKnownNode(node) && node.kind === names.Link) {
    const destination: string = node.data.destination
    return context.escape(destination) + context.inline(node.children)
  }
  return context.fallback(node)
})
const builder = new PresetBuilder().add(custom)
const declaration = Declaration.group('custom').new('annotation')
builder.add(
  new Plugin(declaration).on('custom::annotation', (node, context) =>
    context.blocks(node.children)
  )
)
const store: Store = {
  getMe: () => ({ id: 'me' }),
  generateUserHref: id => '#' + id
}
renderer(v1.html({ store, math: tex => tex, highlight: code => code }))

// @ts-expect-error handlers return HTML strings
new Plugin(declaration).on('invalid', () => [])
// @ts-expect-error presets have runtime identity
renderer({})
// @ts-expect-error shared declarations are explicit objects
new Plugin('name')
declare const context: RenderContext
const renderedChildren: string = context.inline([])
void [result, openResult, renderedChildren]
