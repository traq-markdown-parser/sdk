import type { Rule, PluginState, Owner, Phase } from "./types.js";
import { Plugin as Declaration } from "../definitions/index.js";
import { metadata, data } from "./identity.js";

export class Plugin {
  constructor(declaration: Declaration) {
    if (!(declaration instanceof Declaration))
      throw new TypeError("Expected Plugin declaration");
    metadata.set(this, {
      kind: "plugin",
      symbol: {},
      declaration,
      name: declaration.name,
      group: declaration.namespace,
      rules: [],
      text: [],
      frozen: false,
    });
  }

  get name() {
    return data(this, "plugin").name;
  }
  get namespace() {
    return data(this, "plugin").group;
  }
  get inlineRules() {
    return Object.freeze(
      data(this, "plugin").rules.filter(
        (r): r is Rule<"inline"> => r.phase === "inline",
      ),
    );
  }
  get blockRules() {
    return Object.freeze(
      data(this, "plugin").rules.filter(
        (r): r is Rule<"block"> => r.phase === "block",
      ),
    );
  }
  get textRules() {
    return Object.freeze(
      data(this, "plugin").rules.filter(
        (r): r is Rule<"text"> => r.phase === "text",
      ),
    );
  }
  add(rule: Rule) {
    data(rule, "rule");
    const previous = data(this, "plugin");
    if (previous.frozen)
      throw new TypeError("Bundled plugin definitions are immutable");
    metadata.set(this, {
      ...previous,
      symbol: {},
      rules: [...previous.rules, rule],
    });
    return this;
  }
}
export function makePlugin(state: PluginState) {
  const plugin = new Plugin(state.declaration);
  metadata.set(plugin, state);
  return plugin;
}
export function makeRule(
  runtime: Owner,
  index: number,
  definition: { name: string | null; phase: Phase },
): Rule {
  const rule = Object.freeze({
    name: definition.name,
    phase: definition.phase,
  });
  metadata.set(rule, { kind: "rule", runtime, index, ...definition });
  return rule as Rule;
}
