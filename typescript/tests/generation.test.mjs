import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, writeFile, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { goNodes } from '../../scripts/contracts/go.mjs';
import { typescriptFiles } from '../../scripts/contracts/typescript.mjs';
import { presetFiles } from '../../scripts/contracts/presets.mjs';

test('new Rust-exported payloads and presets generate both host APIs', async () => {
  const directory = await mkdtemp(path.join(tmpdir(), 'markdown-contract-'));
  try {
    // The same inputs produced by ts-rs and schemars for a new Rust node.
    await writeFile(path.join(directory, 'BadgeData.ts'), 'export type BadgeData = { label:string; active:boolean };\n');
    await writeFile(path.join(directory, 'ParseError.ts'), 'export type ParseError = {code:"internal_error"};\n');
    const schema = {title:'BadgeData',type:'object',additionalProperties:false,
      properties:{label:{type:'string'},active:{type:'boolean'}},required:['label','active']};
    const key = 'custom::BadgeData';
    const generated = await typescriptFiles({nodes:{[key]:{schema,group:'custom'}}},directory);
    assert.match(generated.get('typescript/generated/custom.ts'), /kind: "custom::BadgeData"; data: BadgeData/);
    assert.match(generated.get('typescript/generated/nodes.ts'), /custom.NodeKind/);
    const go = goNodes([[key,schema]]);
    assert.match(go, /Label string/);
    assert.match(go, /Active bool/);
    assert.match(go, /case BadgeName: payload = &Badge\{\}/);
    assert.throws(()=>goNodes([[key,schema],['another::BadgeData',schema]]),/Duplicate generated payload type/);
    const presets = presetFiles({commonmark:0,custom:{compact:1}});
    assert.match(presets.get('typescript/generated/presets.ts'), /"custom.compact"/);
    assert.match(presets.get('go/presets_generated.go'), /PresetCustomCompact Preset = "custom.compact"/);
  } finally {
    if (path.dirname(directory) !== path.resolve(tmpdir())) throw new Error('Unexpected temporary path');
    await rm(directory,{recursive:true,force:true});
  }
});
