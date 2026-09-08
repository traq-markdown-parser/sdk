package core

import (
	"context"
	"fmt"
	"sync/atomic"
)

type grammarState struct {
	definition     definition
	recipe         []byte
	description    string
	references     atomic.Int64
	initialized    atomic.Bool
	initialization chan struct{}
}

// Grammar is an immutable definition. Presets compile on first Parser construction;
// compiled data is shared with every Parser created from the same Grammar.
type Grammar struct {
	runtime *Runtime
	state   *grammarState
	closed  atomic.Bool
}

func (r *Runtime) newGrammar(snapshot definition, recipe []byte, description string) *Grammar {
	state := &grammarState{
		definition:     snapshot,
		recipe:         recipe,
		description:    description,
		initialization: make(chan struct{}, 1),
	}
	state.references.Store(1)
	return &Grammar{runtime: r, state: state}
}
func (g *Grammar) ToBuilder() *GrammarBuilder {
	return &GrammarBuilder{runtime: g.runtime, definition: g.state.definition.copy()}
}
func (g *Grammar) Plugins() []*Plugin {
	result := make([]*Plugin, len(g.state.definition.plugins))
	for index, plugin := range g.state.definition.plugins {
		plugin.frozen = true
		result[index] = &plugin
	}
	return result
}
func (g *Grammar) Describe() string { return g.state.description }
func (g *Grammar) Close() error {
	if g.closed.CompareAndSwap(false, true) {
		g.runtime.release(g.state)
	}
	return nil
}

// retain races safely with the last owner's Close: zero is never resurrected.
func (s *grammarState) retain() bool {
	for refs := s.references.Load(); refs > 0; refs = s.references.Load() {
		if s.references.CompareAndSwap(refs, refs+1) {
			return true
		}
	}
	return false
}
func (r *Runtime) release(state *grammarState) {
	if state.references.Add(-1) != 0 || !state.initialized.Load() {
		return
	}
	r.grammarCount.Add(-1)
	// Idle workers release immediately. Busy workers sweep on return; an
	// in-flight request owns a reference until its Wasm call has completed.
	r.lifecycle.Lock()
	defer r.lifecycle.Unlock()
	count := len(r.pool)
	for range count {
		var w *worker
		select {
		case w = <-r.pool:
		default:
			return
		}
		if w != nil {
			w.sweep()
		}
		r.pool <- w
	}
}

// Parser initializes an unused preset before returning. Cancellation and failed
// initialization leave an open Grammar available for retry.
func (r *Runtime) Parser(ctx context.Context, grammar *Grammar) (*Parser, error) {
	if grammar == nil || grammar.runtime != r {
		return nil, fmt.Errorf("grammar belongs to another runtime")
	}
	if r.isClosed() || grammar.closed.Load() || !grammar.state.retain() {
		return nil, fmt.Errorf("grammar is closed")
	}
	if err := grammar.state.initialize(ctx, r); err != nil {
		r.release(grammar.state)
		return nil, err
	}
	return &Parser{runtime: r, state: grammar.state}, nil
}
