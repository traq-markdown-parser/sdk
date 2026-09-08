// Generated from Rust node contracts. Do not edit.
import type { BlankLineData } from './BlankLineData.js';
import type { ReferenceData } from './ReferenceData.js';
import type { SpoilerData } from './SpoilerData.js';
import type { StampData } from './StampData.js';
export type { BlankLineData } from './BlankLineData.js';
export type { ReferenceData } from './ReferenceData.js';
export type { ReferenceKind } from './ReferenceKind.js';
export type { SpoilerData } from './SpoilerData.js';
export type { StampData } from './StampData.js';
export type NodeKind =
  | { kind: "markdown_trap_contracts::compat::BlankLineData"; data: BlankLineData }
  | { kind: "markdown_trap_contracts::reference::ReferenceData"; data: ReferenceData }
  | { kind: "markdown_trap_contracts::spoiler::SpoilerData"; data: SpoilerData }
  | { kind: "markdown_trap_contracts::stamp::StampData"; data: StampData };
// Generated from Rust node payload types. Do not edit.
import {fields,string,boolean,nullable,oneOf} from "../../core/validation/fields.js"
export const names = Object.freeze({"BlankLine":"markdown_trap_contracts::compat::BlankLineData","Reference":"markdown_trap_contracts::reference::ReferenceData","Spoiler":"markdown_trap_contracts::spoiler::SpoilerData","Stamp":"markdown_trap_contracts::stamp::StampData"} as const)
const validators = new Map<string, (data: unknown) => boolean>([
  ["markdown_trap_contracts::compat::BlankLineData",value => fields(value,{},{})],
  ["markdown_trap_contracts::reference::ReferenceData",value => fields(value,{"id":string,"label":string,"type":oneOf("user","group","channel")},{})],
  ["markdown_trap_contracts::spoiler::SpoilerData",value => fields(value,{},{})],
  ["markdown_trap_contracts::stamp::StampData",value => fields(value,{"literal":string},{})],
])
export const nodes: ReadonlyMap<string, (data: unknown) => boolean> = new Map(validators)
// Check this payload only; children can still contain unknown nodes.
export function isKnownNode<T extends {kind: string; data: unknown}>(node: T): node is T & NodeKind {
  return validators.get(node.kind)?.(node.data) ?? false
}
