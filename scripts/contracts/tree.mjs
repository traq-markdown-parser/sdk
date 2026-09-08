import { quoted as q } from './schema.mjs';
import { contractsDirectory, relativeImport } from './layout.mjs';

// The distribution owns the union. Tree traversal never contains this type list.
export function treeFiles(entries, metadata) {
  const header = '// Generated from the Rust node contracts. Do not edit.\n';
  return new Map([
    ['typescript/src/parser/generated/Span.ts', header + 'export type Span = { start: number; end: number };\n'],
    ['typescript/src/parser/generated/NodeKind.ts', header +
      entries.map(([key, s]) => `import type { ${s.title} } from ${q(relativeImport('typescript/src/parser/generated', contractsDirectory(metadata[key].group) + '/' + s.title + '.js'))};`).join('\n') +
      '\nexport type NodeKind =\n' + entries.map(([key, s]) =>
        `  | { kind: ${q(key)}; data: ${s.title} }`).join('\n') + ';\n'],
    ['typescript/src/parser/generated/Node.ts', header +
      "import type { Span } from './Span.js';\nimport type { NodeKind } from './NodeKind.js';\n" +
      'export type Node<AllowUnknown extends boolean = false> =\n' +
      '  (NodeKind | (AllowUnknown extends true ? { kind: string; data: unknown } : never)) &\n' +
      '  { span: Span; children?: Node<AllowUnknown>[] };\n'],
    ['typescript/src/parser/generated/Document.ts', header +
      "import type { Node } from './Node.js';\n" +
      'export type Document<AllowUnknown extends boolean = false> = {\n' +
      '  source: string; children: Node<AllowUnknown>[];\n};\n'],
  ]);
}
