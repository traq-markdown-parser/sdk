import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { createRuntime, presets } from '@traq-markdown-parser/traq'
import { renderer, PresetBuilder } from '@traq-markdown-parser/core/renderer'
import { plugin } from '@traq-markdown-parser/commonmark/renderer'
import { v1 } from '@traq-markdown-parser/traq/renderer'

assert.throws(() => import.meta.resolve('markdown-it'), {
  code: 'ERR_MODULE_NOT_FOUND'
})

const css = await readFile(
  new URL(import.meta.resolve('@traq-markdown-parser/traq/index.css')),
  'utf8'
)
assert(css.includes('.markdown-body') && css.includes('.emoji'))
const runtime = await createRuntime(
  await readFile(
    new URL(import.meta.resolve('@traq-markdown-parser/traq/parser.wasm'))
  )
)
try {
  const parser = runtime.createParser(presets.traq.v1)
  const document = parser.parse('**package** $x$ !!hidden!! :0xff0000:')
  const output = renderer(v1.html()).render(document)
  for (const text of [
    '<strong>package</strong>',
    'katex',
    'spoiler',
    'background-color: #ff0000'
  ])
    assert(output.includes(text), text)
  const messages = v1.messageRenderer({ origin: 'https://q.example.test' })
  const message = parser.parse(
    'hello\nhttps://q.example.test/files/00000000-0000-0000-0000-000000000001'
  )
  assert.equal(messages.renderInline(message).renderedText, 'hello')
  assert.equal(messages.render(message).embeddings[0].type, 'file')
  const custom = renderer(new PresetBuilder().add(plugin()).build())
  assert.equal(
    custom.renderInline(parser.parseInline('**shared declaration**')),
    '<strong>shared declaration</strong>'
  )
} finally {
  runtime.dispose()
}
console.log(
  'Packed renderer consumer: declarations, HTML, CSS, and Wasm integration passed'
)
