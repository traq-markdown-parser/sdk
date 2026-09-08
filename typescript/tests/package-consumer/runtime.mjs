import { names } from '@traq-markdown-parser/ts/nodes';
import { readFile } from 'node:fs/promises';
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { loadRuntime } from '@traq-markdown-parser/ts';
const bytes = await readFile(new URL(import.meta.resolve('@traq-markdown-parser/ts/parser.wasm')));
assert.equal(createHash('sha256').update(bytes).digest('hex'), process.argv[2]);
const runtime = await loadRuntime(bytes);
try {
  const parser = runtime.parser(runtime.presets.traq.v1);
  try {
    assert.equal(parser.parse('**package**').children[0].children[0].kind, names.Strong);
    assert.equal(parser.parseInline(':stamp:').children[0].kind, names.Stamp);
    assert.equal(parser.parse('- parent\n\t- child').children[0].children[0].children[1].kind, names.List);
  } finally { parser.dispose(); }
} finally { runtime.dispose(); }
console.log('Packed bindings: ESM, Wasm digest, AST parsing, and NodeNext declarations passed');
