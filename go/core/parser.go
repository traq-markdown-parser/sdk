package core

import (
	"context"
	"fmt"
	"sync/atomic"
	"unicode/utf8"
)

// Parser is safe for concurrent calls. It owns a lease independent of Grammar.
type Parser struct {
	runtime *Runtime
	state   *grammarState
	closed  atomic.Bool
}

func (p *Parser) Close() error {
	if p.closed.CompareAndSwap(false, true) {
		p.runtime.release(p.state)
	}
	return nil
}
func (p *Parser) Parse(ctx context.Context, source string) (*Result, error) {
	return p.parse(ctx, source, 0)
}
func (p *Parser) ParseInline(ctx context.Context, source string) (*Result, error) {
	return p.parse(ctx, source, 1)
}
func (p *Parser) parse(ctx context.Context, source string, mode uint32) (*Result, error) {
	if p.closed.Load() || !p.state.retain() {
		return nil, fmt.Errorf("parser is closed")
	}
	defer p.runtime.release(p.state)
	if len(source) > 65536 {
		return nil, &Error{Code: "resource_limit", Resource: "input_bytes"}
	}
	if !utf8.ValidString(source) {
		return nil, &Error{Code: "invalid_utf8"}
	}
	var result *Result
	err := p.runtime.withWorker(ctx, func(w *worker) error {
		handle, err := w.grammar(ctx, p.state)
		if err != nil {
			return err
		}
		result, err = parseInstance(ctx, w.module, source, handle, mode, p.runtime.nodes)
		return err
	})
	return result, err
}
