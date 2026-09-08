// Generated from Rust contracts. Do not edit.
import * as commonmark from './commonmark.js';
import * as generic from './generic.js';
import * as trap from './trap.js';
export type NodeKind = commonmark.NodeKind | generic.NodeKind | trap.NodeKind;
export type Node<AllowUnknown extends boolean = false> =
 (NodeKind | (AllowUnknown extends true ? {kind:string;data:unknown} : never)) &
 {span:{start:number;end:number};children?:Node<AllowUnknown>[]};
export type Document<AllowUnknown extends boolean = false> = {source:string;children:Node<AllowUnknown>[]};
export const names = Object.freeze({...commonmark.names,...generic.names,...trap.names});
export const nodes: ReadonlyMap<string,(data:unknown)=>boolean> = new Map([...commonmark.nodes,...generic.nodes,...trap.nodes]);
export function isKnownNode(node:Node<true>):node is Node<true> & NodeKind {return nodes.get(node.kind)?.(node.data) ?? false;}

export type ParseError = { "code": "invalid_utf8" } | { "code": "resource_limit", resource: string, } | { "code": "internal_error" };
