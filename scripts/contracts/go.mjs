import { shape, typeName, quoted as q } from "./schema.mjs";
const fieldName = (name) =>
  name === "id"
    ? "ID"
    : name.replace(/(^|_)([a-z])/g, (_, p, c) => c.toUpperCase());
function goType(s) {
  if (s.kind === "string" || s.kind === "enum") return "string";
  if (s.kind === "boolean") return "bool";
  if (s.kind === "integer") return s.format;
  if (s.kind === "nullable") return "*" + goType(s.inner);
  throw new Error("Unsupported Go field: " + s.kind);
}
const strings = (names) =>
  names.length ? "[]string{" + names.map(q).join(",") + "}" : "nil";
export function goPayload(group, wireName, schema) {
  const name = typeName(schema),
    s = shape(schema);
  if (s.kind !== "object") throw new Error("Payload must be object");
  const required = s.fields.filter((f) => f.required).map((f) => f.name);
  const optional = s.fields.filter((f) => !f.required).map((f) => f.name);
  const nullable = s.fields
    .filter((f) => f.shape.kind === "nullable")
    .map((f) => f.name);
  const enums = s.fields.filter(
    (f) =>
      (f.shape.kind === "nullable" ? f.shape.inner : f.shape).kind === "enum",
  );
  const checks = enums
    .map((f) => {
      const nullable = f.shape.kind === "nullable",
        inner = nullable ? f.shape.inner : f.shape;
      const field = `value.${fieldName(f.name)}`;
      const access = nullable ? `*${field}` : field;
      const invalid = inner.values.map((v) => `${access} != ${q(v)}`).join(" && ");
      const condition = nullable ? `${field} != nil && (${invalid})` : invalid;
      return `if ${condition} { return nil, fmt.Errorf("invalid ${f.name}") }`;
    })
    .join("\n");
  const fields = s.fields.map((f) =>
    `${fieldName(f.name)} ${goType(f.shape)} \`json:${q(f.name)}\``).join("\n");
  return `// Code generated from Rust node payload types. DO NOT EDIT.
package ${group}
import ("encoding/json"; ${enums.length ? '"fmt"; ' : ""}"github.com/traq-markdown-parser/sdk/go/ast")
const ${name}Name = ${q(wireName)}
type ${name} struct {
${fields}
}
func decode${name}(raw json.RawMessage) (any,error) {
var value ${name}
if err := ast.DecodeFields(raw,&value,${strings(required)},${strings(optional)},${strings(nullable)}); err != nil {return nil,err}
${checks}
return value,nil
}
`;
}
export function goRegistry(group, entries) {
  const decoders = entries.map(([, s]) => `${typeName(s)}Name: decode${typeName(s)},`).join("\n");
  return `// Code generated from Rust node payload types. DO NOT EDIT.
package ${group}
import "github.com/traq-markdown-parser/sdk/go/ast"
func Registry() ast.Registry {return ast.Registry{
${decoders}
}}
`;
}
