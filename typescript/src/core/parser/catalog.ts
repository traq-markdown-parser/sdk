import type { Owner, Catalog, CatalogTree, Grammar } from "./types.js";
import { Plugin as Declaration, PluginGroup } from "../definitions/index.js";
import { Plugin, makePlugin, makeRule } from "./plugin.js";
import { data } from "./identity.js";
export function loadCatalog(runtime: Owner, catalog: Catalog) {
  const groups: PluginGroup[] = [];
  for (const definition of catalog.groups) {
    let group =
      definition.parent === null
        ? new PluginGroup(definition.name)
        : groups[definition.parent].group(definition.name);
    groups.push(group);
  }
  const rules = catalog.rules.map((definition, index) =>
    makeRule(runtime, index, definition),
  );
  const plugins = catalog.plugins.map((definition) => {
    const plugin = new Plugin(
      definition.group === null
        ? new Declaration(definition.name)
        : groups[definition.group].new(definition.name),
    );
    const state = data(plugin, "plugin");
    return makePlugin({
      ...state,
      name: definition.name,
      frozen: true,
      rules: definition.rules.map((index) => rules[index]),
      text: definition.text.map((index) => ({ runtime, index })),
    });
  });
  function tree<T>(
    value: number | CatalogTree,
    leaf: (index: number) => T,
  ): T | { readonly [key: string]: T | object } {
    if (typeof value === "number") return leaf(value);
    return Object.freeze(
      Object.fromEntries(
        Object.entries(value).map(([name, child]) => [name, tree(child, leaf)]),
      ),
    );
  }
  const presets = new Map<number, Grammar>();
  return {
    plugins: tree(catalog.exports.plugins, (index) => plugins[index]),
    presets: tree(catalog.exports.presets, (index) => {
      if (!presets.has(index)) {
        const definition = catalog.presets[index];
        presets.set(
          index,
          runtime.preset(
            {
              plugins: definition.plugins.map((index) =>
                data(plugins[index], "plugin"),
              ),
              order: definition.order,
            },
            definition,
          ),
        );
      }
      return presets.get(index)!;
    }),
  };
}
