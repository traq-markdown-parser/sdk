import path from "node:path";

// Distribution-owned layout. Neither runtime nor renderer knows these groups.
const directories = {
  commonmark: "typescript/src/commonmark/contracts",
  generic: "typescript/src/commonmark/extensions/contracts",
  trap: "typescript/src/trap/contracts",
};
export function contractsDirectory(group) {
  const directory = directories[group];
  if (!directory) throw new Error("Unknown contract owner: " + group);
  return directory;
}
export function relativeImport(directory, file) {
  const relative = path.relative(directory, file).replaceAll("\\", "/");
  return relative.startsWith(".") ? relative : "./" + relative;
}
