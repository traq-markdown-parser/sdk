package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tetratelabs/wazero/api"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

type worker struct {
	module   api.Module
	grammars map[*grammarState]uint32
}
type builtGrammar struct {
	Handle      uint32 `json:"handle"`
	Description string `json:"description"`
}

func (w *worker) build(ctx context.Context, recipe []byte) (*builtGrammar, error) {
	raw, err := invoke(ctx, w.module, recipe, "grammar_build")
	if err != nil {
		return nil, err
	}
	var result struct {
		Grammar json.RawMessage `json:"grammar"`
		Error   json.RawMessage `json:"error"`
	}
	if err := ast.DecodeFields(raw, &result, nil, []string{"grammar", "error"}, nil); err != nil {
		return nil, err
	}
	if (len(result.Grammar) == 0) == (len(result.Error) == 0) {
		return nil, fmt.Errorf("ambiguous build result")
	}
	if len(result.Error) > 0 {
		var failure struct{ Code string }
		if err := json.Unmarshal(result.Error, &failure); err != nil {
			return nil, err
		}
		if failure.Code == "resource_limit" {
			return nil, decodeParseError(result.Error)
		}
		return nil, decodeBuildError(result.Error)
	}
	var grammar builtGrammar
	if err := ast.DecodeFields(result.Grammar, &grammar, []string{"handle", "description"}, nil, nil); err != nil {
		return nil, err
	}
	if grammar.Handle == 0 {
		return nil, fmt.Errorf("invalid grammar handle")
	}
	return &grammar, nil
}
func (w *worker) grammar(ctx context.Context, state *grammarState) (uint32, error) {
	if handle, exists := w.grammars[state]; exists {
		return handle, nil
	}
	compiled, err := w.build(ctx, state.recipe)
	if err != nil {
		return 0, err
	}
	if compiled.Description != state.description {
		return 0, fmt.Errorf("inconsistent grammar compilation")
	}
	w.grammars[state] = compiled.Handle
	return compiled.Handle, nil
}
func (w *worker) sweep() {
	if w.module.IsClosed() {
		return
	}
	for state, handle := range w.grammars {
		if state.references.Load() != 0 {
			continue
		}
		deleted, err := call(context.Background(), w.module, "grammar_drop", uint64(handle))
		delete(w.grammars, state)
		if err != nil || deleted != 1 {
			w.module.Close(context.Background())
			return
		}
	}
}
