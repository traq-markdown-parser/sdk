package core

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
)

type definition struct {
	plugins []Plugin
	order   []int
}

func (d definition) copy() definition {
	return definition{slices.Clone(d.plugins), slices.Clone(d.order)}
}

// Builders are mutable, confined to one goroutine. Built Grammars are immutable.
type GrammarBuilder struct {
	runtime    *Runtime
	definition definition
}

func (r *Runtime) Builder() *GrammarBuilder { return &GrammarBuilder{runtime: r} }
func (b *GrammarBuilder) Add(plugin *Plugin) error {
	if plugin == nil {
		return fmt.Errorf("nil plugin")
	}
	for _, p := range b.definition.plugins {
		if p.identity == plugin.identity {
			return &BuildError{Code: "duplicate", Element: plugin.name}
		}
	}
	if plugin.providerRuntime != nil && plugin.providerRuntime != b.runtime {
		return fmt.Errorf("text provider belongs to another runtime")
	}
	order := slices.Clone(b.definition.order)
	for _, rule := range plugin.rules {
		if rule.runtime != b.runtime {
			return fmt.Errorf("rule belongs to another runtime")
		}
		if slices.Contains(order, rule.index) {
			return &BuildError{Code: "duplicate", Element: rule.name}
		}
		order = append(order, rule.index)
	}
	b.definition.plugins = append(b.definition.plugins, *plugin)
	b.definition.order = order
	return nil
}
func (b *GrammarBuilder) Remove(plugin *Plugin) error {
	if plugin == nil {
		return fmt.Errorf("nil plugin")
	}
	index := slices.IndexFunc(b.definition.plugins, func(p Plugin) bool { return p.identity == plugin.identity })
	if index < 0 {
		return &BuildError{Code: "missing", Element: plugin.name}
	}
	for _, rule := range b.definition.plugins[index].rules {
		b.definition.order = slices.DeleteFunc(b.definition.order, func(i int) bool { return i == rule.index })
	}
	b.definition.plugins = slices.Delete(b.definition.plugins, index, index+1)
	return nil
}
func (b *GrammarBuilder) Before(rule, anchor *Rule) error {
	if rule == nil || anchor == nil || rule.runtime != b.runtime || anchor.runtime != b.runtime {
		return fmt.Errorf("rules must belong to this runtime")
	}
	if rule.phase != anchor.phase {
		return fmt.Errorf("rules must belong to the same phase")
	}
	from, to := slices.Index(b.definition.order, rule.index), slices.Index(b.definition.order, anchor.index)
	if from < 0 {
		return &BuildError{Code: "missing", Element: rule.name}
	}
	if to < 0 {
		return &BuildError{Code: "missing", Element: anchor.name}
	}
	if from != to {
		b.definition.order = slices.Delete(b.definition.order, from, from+1)
		if from < to {
			to--
		}
		b.definition.order = slices.Insert(b.definition.order, to, rule.index)
	}
	return nil
}
func (b *GrammarBuilder) Build(ctx context.Context) (*Grammar, error) {
	snapshot := b.definition.copy()
	recipe, err := json.Marshal(snapshot.composition())
	if err != nil {
		return nil, err
	}
	var grammar *Grammar
	// Reserve across the whole pool: every accepted grammar must fit on every
	// worker when first parsed, regardless of which worker built it.
	if err := b.runtime.reserveGrammar(); err != nil {
		return nil, err
	}
	defer func() {
		if grammar == nil {
			b.runtime.grammarCount.Add(-1)
		}
	}()
	err = b.runtime.withWorker(ctx, func(w *worker) error {
		result, err := w.build(ctx, recipe)
		if err != nil {
			return err
		}
		grammar = b.runtime.newGrammar(snapshot, recipe, result.Description)
		grammar.state.initialized.Store(true)
		w.grammars[grammar.state] = result.Handle
		return nil
	})
	return grammar, err
}
