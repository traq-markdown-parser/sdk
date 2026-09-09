import {
  shape,
  typeName,
  quoted as q,
} from "@traq-markdown-parser/core/codegen/schema";

const fieldName = (name) =>
  name === "id"
    ? "ID"
    : name.replace(/(^|_)([a-z])/g, (_, p, c) => c.toUpperCase());

function goType(s) {
  if (s.kind === "string" || s.kind === "enum") return "string";
  if (s.kind === "boolean") return "bool";
  if (s.kind === "integer") return s.format;
  if (s.kind === "nullable") return "*" + goType(s.inner);
  if (s.kind === "array") return "[]" + goType(s.items);
  if (s.kind === "object") return s.name ?? `struct {${fields(s)}}`;
  throw new Error("Unsupported Go field: " + s.kind);
}

const fields = (s) =>
  s.fields
    .map((f) => `${fieldName(f.name)} ${goType(f.shape)} \`json:${q(f.name)}\``)
    .join("\n");

export function goContract(schema, root = schema) {
  return `type ${typeName(schema)} struct {\n${fields(shape(schema, root))}\n}\n`;
}

export function goPayload(wireName, schema) {
  const name = typeName(schema),
    s = shape(schema);
  if (s.kind !== "object") throw new Error("Payload must be object");
  return (
    `const ${name}Name = ${q(wireName)}\ntype ${name} struct {\n` +
    s.fields
      .map(
        (f) => `${fieldName(f.name)} ${goType(f.shape)} \`json:${q(f.name)}\``,
      )
      .join("\n") +
    `\n}\nfunc (*${name}) nodePayload() {}\n`
  );
}

export function goNodes(entries) {
  const names = entries.map(([, schema]) => typeName(schema));
  if (new Set(names).size !== names.length)
    throw new Error("Duplicate generated payload type");
  return (
    '// Code generated from Rust contracts. DO NOT EDIT.\npackage markdown\nimport("encoding/json";"fmt")\n' +
    'type Span struct {Start uint32 `json:"start"`; End uint32 `json:"end"`}\n' +
    'type Document struct {Source string `json:"source"`; Children []Node `json:"children"`}\n' +
    "type Payload interface {nodePayload()}\n" +
    'type Node struct {Kind string `json:"kind"`; Span Span `json:"span"`; Data Payload `json:"data"`; Children []Node `json:"children,omitempty"`}\n' +
    entries.map(([key, schema]) => goPayload(key, schema)).join("\n") +
    "func (n *Node) UnmarshalJSON(raw []byte) error {\n" +
    'var wire struct {Kind string `json:"kind"`; Span Span `json:"span"`; Data json.RawMessage `json:"data"`; Children []Node `json:"children"`}\n' +
    "if err:=json.Unmarshal(raw,&wire);err!=nil{return err}\nvar payload Payload\nswitch wire.Kind {\n" +
    entries
      .map(([, s]) => `case ${typeName(s)}Name: payload = &${typeName(s)}{}`)
      .join("\n") +
    '\ndefault:return fmt.Errorf("unsupported Rust node: %s",wire.Kind)\n}\n' +
    "if err:=json.Unmarshal(wire.Data,payload);err!=nil{return err}\n*n=Node{wire.Kind,wire.Span,payload,wire.Children};return nil\n}\n"
  );
}
