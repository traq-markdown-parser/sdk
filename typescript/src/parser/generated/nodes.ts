// Generated from Rust node contracts. Do not edit.
import type { Node } from './Node.js';
import type { NodeKind } from './NodeKind.js';
import * as commonmark from "../../commonmark/contracts/nodes.js";
import * as generic from "../../commonmark/extensions/contracts/nodes.js";
import * as trap from "../../trap/contracts/nodes.js";
export const names = Object.freeze({...commonmark.names,...generic.names,...trap.names});
const validators = new Map([...commonmark.nodes,...generic.nodes,...trap.nodes]);
export const nodes: ReadonlyMap<string, (data: unknown) => boolean> = new Map(validators);
export function isKnownNode(node: Node<true>): node is Node<true> & NodeKind { return validators.get(node.kind)?.(node.data) ?? false; }
