package core

import (
	"context"
	"encoding/json"

	"github.com/traq-markdown-parser/sdk/go/binding"
)

type catalog struct {
	Groups  []groupDefinition  `json:"groups"`
	Plugins []pluginDefinition `json:"plugins"`
	Rules   []struct {
		Name  *string `json:"name"`
		Phase string  `json:"phase"`
	} `json:"rules"`
	Presets []struct {
		Description string `json:"description"`
		Plugins     []int  `json:"plugins"`
		Order       []int  `json:"order"`
	} `json:"presets"`
}

func displayName(name *string) string {
	if name == nil {
		return ""
	}
	return *name
}
func (r *Runtime) loadCatalog(ctx context.Context) error {
	// The artifact was checked against this generated catalog on instantiation.
	var spec catalog
	if err := json.Unmarshal(binding.CatalogJSON, &spec); err != nil {
		return err
	}
	groups := make([]*PluginGroup, len(spec.Groups))
	for index, item := range spec.Groups {
		group := NewPluginGroup(displayName(item.Name))
		if item.Parent != nil {
			group.parent = groups[*item.Parent]
		}
		groups[index] = group
	}
	rules := make([]*Rule, len(spec.Rules))
	for index, item := range spec.Rules {
		rules[index] = &Rule{runtime: r, index: index, name: displayName(item.Name), phase: item.Phase}
	}
	plugins := make([]*Plugin, len(spec.Plugins))
	for index, item := range spec.Plugins {
		plugin := NewPlugin(displayName(item.Name))
		plugin.text = item.Text
		plugin.providerRuntime = r
		if item.Group != nil {
			plugin.group = groups[*item.Group]
		}
		for _, rule := range item.Rules {
			plugin.rules = append(plugin.rules, rules[rule])
		}
		plugin.frozen = true
		plugins[index] = plugin
	}
	presets := make([]*Grammar, len(spec.Presets))
	for index, item := range spec.Presets {
		if err := ctx.Err(); err != nil {
			return err
		}
		b := r.Builder()
		for _, plugin := range item.Plugins {
			if err := b.Add(plugins[plugin]); err != nil {
				return err
			}
		}
		b.definition.order = item.Order
		snapshot := b.definition.copy()
		recipe, err := json.Marshal(snapshot.composition())
		if err != nil {
			return err
		}
		presets[index] = r.newGrammar(snapshot, recipe, item.Description)
	}
	r.exportCatalog(plugins, presets)
	return nil
}
