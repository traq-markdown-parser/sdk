const q = JSON.stringify;
function leaves(tree, path = []) {
  return Object.entries(tree).flatMap(([key, value]) =>
    typeof value === "number"
      ? [[...path, key]]
      : leaves(value, [...path, key]),
  );
}
function named(tree, path = []) {
  return Object.fromEntries(
    Object.entries(tree).map(([key, value]) => [
      key,
      typeof value === "number"
        ? [...path, key].join(".")
        : named(value, [...path, key]),
    ]),
  );
}
export function presetFiles(
  tree,
  type = "Preset",
  value = "presets",
  file = "presets",
) {
  const paths = leaves(tree);
  const goName = (part) =>
    ({ traq: "TraQ", commonmark: "CommonMark" })[part] ??
    part[0].toUpperCase() + part.slice(1);
  return new Map([
    [
      `typescript/generated/${file}.ts`,
      "// Generated from Rust preset exports. Do not edit.\n" +
        `export type ${type} = ${paths.map((p) => q(p.join("."))).join(" | ")};\nexport const ${value} = ${q(named(tree))} as const;\n`,
    ],
    [
      `go/${file}_generated.go`,
      `// Code generated from Rust preset exports. DO NOT EDIT.\npackage markdown\ntype ${type} string\nconst (\n` +
        paths
          .map(
            (p) =>
              `${type}${p.map(goName).join("")} ${type} = ${q(p.join("."))}`,
          )
          .join("\n") +
        "\n)\n",
    ],
  ]);
}
