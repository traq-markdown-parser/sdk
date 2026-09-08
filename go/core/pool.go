// Package core runs the same freestanding Wasm used by the browser.
package core

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/tetratelabs/wazero"
	"github.com/traq-markdown-parser/sdk/go/ast"
	"github.com/traq-markdown-parser/sdk/go/binding"
)

type Runtime struct {
	Presets      Presets
	Plugins      Plugins
	runtime      wazero.Runtime
	module       wazero.CompiledModule
	pool         chan *worker
	closed       chan struct{}
	closeOnce    sync.Once
	lifecycle    sync.Mutex
	nodes        ast.Registry
	grammarCount atomic.Int64
}

func New(ctx context.Context, wasm []byte) (*Runtime, error) {
	rt := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().WithMemoryLimitPages(512).WithCloseOnContextDone(true))
	module, err := rt.CompileModule(ctx, wasm)
	if err != nil {
		rt.Close(ctx)
		return nil, err
	}
	r := &Runtime{runtime: rt, module: module, pool: make(chan *worker, 4), closed: make(chan struct{}), nodes: binding.Registry()}
	for range cap(r.pool) {
		w, err := r.instantiate(ctx)
		if err != nil {
			r.Close(ctx)
			return nil, err
		}
		r.pool <- w
	}
	if err := r.loadCatalog(ctx); err != nil {
		r.Close(ctx)
		return nil, err
	}
	return r, nil
}
func (r *Runtime) isClosed() bool {
	select {
	case <-r.closed:
		return true
	default:
		return false
	}
}
func (r *Runtime) Close(ctx context.Context) error {
	var err error
	r.closeOnce.Do(func() {
		r.lifecycle.Lock()
		close(r.closed)
		// Runtime.Close below also closes borrowed modules.
		for len(r.pool) > 0 {
			<-r.pool
		}
		r.lifecycle.Unlock()
		err = r.runtime.Close(ctx)
	})
	return err
}
func (r *Runtime) instantiate(ctx context.Context) (*worker, error) {
	module, err := r.runtime.InstantiateModule(ctx, r.module, wazero.NewModuleConfig().WithName(""))
	if err != nil {
		return nil, err
	}
	if err := validateArtifact(ctx, module); err != nil {
		module.Close(ctx)
		return nil, err
	}
	return &worker{module: module, grammars: map[*grammarState]uint32{}}, nil
}
func (r *Runtime) withWorker(ctx context.Context, run func(*worker) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.isClosed() {
		return fmt.Errorf("runtime is closed")
	}
	var w *worker
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-r.closed:
		return fmt.Errorf("runtime is closed")
	case w = <-r.pool:
	}
	defer func() {
		r.lifecycle.Lock()
		defer r.lifecycle.Unlock()
		if w != nil && w.module.IsClosed() {
			w = nil
		}
		if r.isClosed() {
			if w != nil {
				w.module.Close(context.Background())
			}
			return
		}
		if w != nil {
			w.sweep()
		}
		r.pool <- w
	}()
	if w != nil {
		w.sweep()
	}
	if w == nil || w.module.IsClosed() {
		var err error
		w, err = r.instantiate(ctx)
		if err != nil {
			return err
		}
	}
	err := run(w)
	if err != nil {
		var parseError *Error
		var buildError *BuildError
		if !errors.As(err, &parseError) && !errors.As(err, &buildError) {
			w.module.Close(context.Background())
		}
	}
	return err
}
