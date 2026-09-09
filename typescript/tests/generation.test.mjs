import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, writeFile, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { goNodes } from '@traq-markdown-parser/core/codegen/go';
import { nodeFiles } from '@traq-markdown-parser/core/codegen/nodes';
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
    assert.match((await nodeFiles({nodes:{[key]:{schema,group:'custom'}}},directory)).get('custom.ts'), /kind: "custom::BadgeData"; data: BadgeData/);
    assert.match(generated.get('typescript/generated/nodes.ts'), /custom.NodeKind/);
    const go = goNodes([[key,schema]]);
    assert.match(go, /Label string/);
    assert.match(go, /Active bool/);
    assert.match(go, /case BadgeName: return &Badge\{\}/);
    assert.throws(()=>goNodes([[key,schema],['another::BadgeData',schema]]),/Duplicate generated payload type/);
    const presets = presetFiles({commonmark:0,custom:{compact:1}});
    assert.match(presets.get('typescript/generated/presets.ts'), /"custom.compact"/);
    assert.match(presets.get('go/presets_generated.go'), /PresetCustomCompact Preset = "custom.compact"/);
  } finally {
    if (path.dirname(directory) !== path.resolve(tmpdir())) throw new Error('Unexpected temporary path');
    await rm(directory,{recursive:true,force:true});
  }
});

test('Rust processing options, nested results and presets generate without host changes', async () => {
  const { processingFiles } = await import('../../scripts/contracts/processing.mjs');
  const directory = await mkdtemp(path.join(tmpdir(), 'processing-contract-'));
  try {
    const object = (title, properties) => ({title, type:'object', additionalProperties:false, properties, required:Object.keys(properties)});
    const details = object('Details', {labels:{type:'array',items:{type:'string'}}});
    const options = object('ProcessorOptions', {compact:{type:'boolean'},details:{$ref:'#/$defs/Details'}});
    options.$defs = {Details:details};
    const output = object('ProcessOutput', {details:{$ref:'#/$defs/Details'}, batches:{type:'array',items:{$ref:'#/$defs/Details'}}});
    output.$defs = {Details:details};
    await writeFile(path.join(directory,'ProcessorOptions.ts'), 'import type { Details } from "./Details.js";\nexport type ProcessorOptions = { compact:boolean; details:Details };');
    await writeFile(path.join(directory,'ProcessOutput.ts'), 'import type { Details } from "./Details.js";\nexport type ProcessOutput = { details:Details; batches:Array<Details> };');
    await writeFile(path.join(directory,'Details.ts'), 'export type Details = { labels:Array<string> };');
    const schemas = {ProcessorPreset:{type:'string',enum:['traq.v1','custom.compact']},ProcessorOptions:options,ProcessOutput:output};
    const files = await processingFiles(schemas,directory);
    const go = files.get('go/processing_generated.go');
    assert.match(go,/ProcessorPresetCustomCompact ProcessorPreset = "custom.compact"/);
    assert.match(go,/Compact bool/);
    assert.match(go,/Batches \[\]Details/);
    assert.match(go,/Labels \[\]string/);
    assert.equal(go.match(/type Details struct/g).length,1);
    const ts = files.get('typescript/generated/processing.ts');
    assert.match(ts,/"custom.compact"/);
    assert.match(ts,/compact:boolean/);
    assert.match(ts,/batches:Array<Details>/);
    assert.equal(ts.match(/export type Details/g).length,1);
    output.$defs = {Details:object('Details',{other:{type:'boolean'}})};
    await assert.rejects(processingFiles(schemas,directory),/Conflicting processing type/);
  } finally {
    if (path.dirname(directory) !== path.resolve(tmpdir())) throw new Error('Unexpected temporary path');
    await rm(directory,{recursive:true,force:true});
  }
});
