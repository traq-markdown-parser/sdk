package core

import (
	"context"
	"fmt"
)

func (r *Runtime) reserveGrammar() error {
	if r.grammarCount.Add(1) > 256 {
		r.grammarCount.Add(-1)
		return &BuildError{Code: "invalid_definition", Reason: "live grammar limit exceeded"}
	}
	return nil
}
func (s *grammarState) initialize(ctx context.Context, r *Runtime) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.initialized.Load() {
		return nil
	}
	// Only the first constructor compiles the definition. Waiting constructors
	// can cancel, and failed attempts leave the preset available for retry.
	select {
	case s.initialization <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	case <-r.closed:
		return fmt.Errorf("runtime is closed")
	}
	defer func() { <-s.initialization }()
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.initialized.Load() {
		return nil
	}
	if err := r.reserveGrammar(); err != nil {
		return err
	}
	err := r.withWorker(ctx, func(w *worker) error {
		_, err := w.grammar(ctx, s)
		return err
	})
	if err != nil {
		r.grammarCount.Add(-1)
		return err
	}
	s.initialized.Store(true)
	return nil
}
