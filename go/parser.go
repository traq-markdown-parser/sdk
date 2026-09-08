// Package markdown calls Rust parsing and processing implementations.
package markdown

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// Runtime compiles Wasm once and owns its parser and processor instances.
type Runtime struct {
	runtime  wazero.Runtime
	compiled wazero.CompiledModule
}

// Parser owns one Wasm instance. Calls are serialized; use separate parsers for parallel execution.
// A context canceled during a Wasm call closes the instance; create a new Parser to resume.
type Parser struct {
	*instance
}

// Processor parses once in Rust, then renders notifications and extracts references.
type Processor struct {
	*instance
}

type instance struct {
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
	for _, name := range []string{"input_ptr", "output_ptr", "configure", "parse", "configure_processor", "process"} {
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
	i, err := r.instantiate(ctx, "configure", string(preset))
	if err != nil {
		return nil, err
	}
	return &Parser{i}, nil
}

func (r *Runtime) NewProcessor(ctx context.Context, preset ProcessorPreset, options ProcessorOptions) (*Processor, error) {
	config, err := json.Marshal(map[string]any{"preset": preset, "options": options})
	if err != nil {
		return nil, err
	}
	i, err := r.instantiate(ctx, "configure_processor", string(config))
	if err != nil {
		return nil, err
	}
	return &Processor{i}, nil
}

func (r *Runtime) instantiate(ctx context.Context, operation, config string) (*instance, error) {
	module, err := r.runtime.InstantiateModule(ctx, r.compiled, wazero.NewModuleConfig().WithName(""))
	if err != nil {
		return nil, err
	}
	p := &instance{module: module, gate: make(chan struct{}, 1)}
	if _, err := p.call(ctx, operation, config); err != nil {
		p.Close(context.Background())
		return nil, err
	}
	return p, nil
}

// Close releases only this instance, leaving its Runtime usable.
func (p *instance) Close(ctx context.Context) error { return p.module.Close(ctx) }

func (p *Parser) Parse(ctx context.Context, source string) (*Document, error) {
	return p.parse(ctx, source, 0)
}

func (p *Parser) ParseInline(ctx context.Context, source string) (*Document, error) {
	return p.parse(ctx, source, 1)
}

func (p *Parser) parse(ctx context.Context, source string, mode uint64) (*Document, error) {
	raw, err := p.invoke(ctx, "parse", source, mode)
	if err != nil {
		return nil, err
	}
	var document Document
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	return &document, nil
}

func (p *Processor) Process(ctx context.Context, source string) (*ProcessOutput, error) {
	raw, err := p.invoke(ctx, "process", source)
	if err != nil {
		return nil, err
	}
	var result ProcessOutput
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (p *instance) invoke(ctx context.Context, operation, source string, args ...uint64) (json.RawMessage, error) {
	select {
	case p.gate <- struct{}{}:
		defer func() { <-p.gate }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return p.call(ctx, operation, source, args...)
}

func (p *instance) call(ctx context.Context, operation, input string, args ...uint64) (json.RawMessage, error) {
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
		Result     json.RawMessage `json:"result"`
		Error      json.RawMessage `json:"error"`
		Configured string          `json:"configured"`
	}
	if err := json.Unmarshal(output, &reply); err != nil {
		return nil, err
	}

	if reply.Error != nil {
		return nil, fmt.Errorf("markdown: %s", reply.Error)
	}
	if (operation == "configure" || operation == "configure_processor") && reply.Configured != buildID {
		return nil, fmt.Errorf("Wasm does not match this SDK build")
	}

	if reply.Document != nil {
		return reply.Document, nil
	}
	return reply.Result, nil
}
