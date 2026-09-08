// Package markdown calls the Rust parser. Grammar behavior and AST contracts live in Rust.
package markdown

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// Runtime compiles Wasm once and owns the parsers instantiated from it.
type Runtime struct {
	runtime  wazero.Runtime
	compiled wazero.CompiledModule
}

// Parser owns one Wasm instance. Calls are serialized; use separate parsers for parallel execution.
// A context canceled during a Wasm call closes the instance; create a new Parser to resume.
type Parser struct {
	module api.Module
	gate   chan struct{}
}

func NewRuntime(ctx context.Context, wasm []byte) (*Runtime, error) {
	rt := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().WithMemoryLimitPages(memoryPages).WithCloseOnContextDone(true))
	compiled, err := rt.CompileModule(ctx, wasm)
	if err != nil {
		rt.Close(context.Background())
		return nil, err
	}
	for _, name := range []string{"input_ptr", "output_ptr", "configure", "parse"} {
		if compiled.ExportedMemories()["memory"] == nil || compiled.ExportedFunctions()[name] == nil {
			rt.Close(context.Background())
			return nil, fmt.Errorf("invalid parser Wasm exports")
		}
	}
	return &Runtime{runtime: rt, compiled: compiled}, nil
}

// Close releases the compiled module and all parsers created by this Runtime.
func (r *Runtime) Close(ctx context.Context) error { return r.runtime.Close(ctx) }

// NewParser creates an independent instance using the shared compiled module.
func (r *Runtime) NewParser(ctx context.Context, preset Preset) (*Parser, error) {
	module, err := r.runtime.InstantiateModule(ctx, r.compiled, wazero.NewModuleConfig().WithName(""))
	if err != nil {
		return nil, err
	}
	p := &Parser{module: module, gate: make(chan struct{}, 1)}
	if _, err := p.call(ctx, "configure", string(preset)); err != nil {
		p.Close(context.Background())
		return nil, err
	}
	return p, nil
}

// Close releases only this parser's instance, leaving its Runtime usable.
func (p *Parser) Close(ctx context.Context) error { return p.module.Close(ctx) }
func (p *Parser) Parse(ctx context.Context, source string) (*Document, error) {
	return p.parse(ctx, source, 0)
}
func (p *Parser) ParseInline(ctx context.Context, source string) (*Document, error) {
	return p.parse(ctx, source, 1)
}
func (p *Parser) parse(ctx context.Context, source string, mode uint64) (*Document, error) {
	select {
	case p.gate <- struct{}{}:
		defer func() { <-p.gate }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	raw, err := p.call(ctx, "parse", source, mode)
	if err != nil {
		return nil, err
	}
	var document Document
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	return &document, nil
}

func (p *Parser) call(ctx context.Context, operation, input string, args ...uint64) (json.RawMessage, error) {
	if p.module.IsClosed() {
		return nil, fmt.Errorf("parser is closed")
	}
	if len(input) > inputBytes {
		return nil, fmt.Errorf("Wasm input limit exceeded")
	}
	pointer, err := p.module.ExportedFunction("input_ptr").Call(ctx, uint64(len(input)))
	if err != nil {
		return nil, err
	}
	if pointer[0] == 0 {
		return nil, fmt.Errorf("Wasm input limit exceeded")
	}
	if !p.module.Memory().Write(uint32(pointer[0]), []byte(input)) {
		return nil, fmt.Errorf("invalid Wasm input range")
	}
	length, err := p.module.ExportedFunction(operation).Call(ctx, args...)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}
	pointer, err = p.module.ExportedFunction("output_ptr").Call(ctx)
	if err != nil {
		return nil, err
	}
	output, ok := p.module.Memory().Read(uint32(pointer[0]), uint32(length[0]))
	if !ok {
		return nil, fmt.Errorf("invalid Wasm output range")
	}
	var reply struct {
		Document   json.RawMessage `json:"document"`
		Error      json.RawMessage `json:"error"`
		Configured string          `json:"configured"`
	}
	if err := json.Unmarshal(output, &reply); err != nil {
		return nil, err
	}
	if reply.Error != nil {
		return nil, fmt.Errorf("markdown: %s", reply.Error)
	}
	if operation == "configure" && reply.Configured != buildID {
		return nil, fmt.Errorf("Wasm does not match this SDK build")
	}
	return reply.Document, nil
}
