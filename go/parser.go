// Package markdown selects traQ presets and pairs them with the distributed Wasm.
package markdown

import (
	"context"
	"encoding/json"

	"github.com/traq-markdown-parser/core/go/binding"
)

type Runtime struct{ runtime *binding.Runtime }
type Parser struct{ instance *binding.Instance }
type Processor struct{ instance *binding.Instance }

func NewRuntime(ctx context.Context, wasm []byte) (*Runtime, error) {
	runtime, err := binding.NewRuntime(
		ctx,
		wasm,
		binding.Artifact{
			BuildID:     buildID,
			InputBytes:  inputBytes,
			MemoryPages: memoryPages,
		},
	)
	if err != nil {
		return nil, err
	}
	return &Runtime{runtime: runtime}, nil
}

func (r *Runtime) Close(ctx context.Context) error   { return r.runtime.Close(ctx) }
func (p *Parser) Close(ctx context.Context) error    { return p.instance.Close(ctx) }
func (p *Processor) Close(ctx context.Context) error { return p.instance.Close(ctx) }

func (r *Runtime) NewParser(ctx context.Context, preset Preset) (*Parser, error) {
	instance, err := r.runtime.NewInstance(ctx, "configure", string(preset))
	if err != nil {
		return nil, err
	}
	return &Parser{instance: instance}, nil
}

func (r *Runtime) NewProcessor(
	ctx context.Context,
	preset Preset,
	options ProcessorOptions,
) (*Processor, error) {
	config, err := json.Marshal(map[string]any{"preset": preset, "options": options})
	if err != nil {
		return nil, err
	}
	instance, err := r.runtime.NewInstance(ctx, "configure_processor", string(config))
	if err != nil {
		return nil, err
	}
	return &Processor{instance: instance}, nil
}

func (p *Parser) Parse(ctx context.Context, source string) (*Document, error) {
	return p.parse(ctx, source, 0)
}
func (p *Parser) ParseInline(ctx context.Context, source string) (*Document, error) {
	return p.parse(ctx, source, 1)
}

func (p *Parser) parse(ctx context.Context, source string, mode uint64) (*Document, error) {
	raw, err := p.instance.Call(ctx, "parse", source, mode)
	if err != nil {
		return nil, err
	}
	return DecodeDocument(raw)
}

func (p *Processor) Process(ctx context.Context, source string) (*ProcessOutput, error) {
	raw, err := p.instance.Call(ctx, "process", source)
	if err != nil {
		return nil, err
	}
	var result ProcessOutput
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
