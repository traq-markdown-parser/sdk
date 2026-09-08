package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tetratelabs/wazero/api"
	"github.com/traq-markdown-parser/sdk/go/ast"
)

// JSON preserves the wire representation for hashing; Document has typed payloads.
type Result struct {
	Document *ast.Document
	JSON     json.RawMessage
}

func call(ctx context.Context, instance api.Module, name string, args ...uint64) (uint32, error) {
	f := instance.ExportedFunction(name)
	if f == nil {
		return 0, fmt.Errorf("missing Wasm export %s", name)
	}
	results, err := f.Call(ctx, args...)
	if err != nil {
		return 0, err
	}
	if len(results) != 1 {
		return 0, fmt.Errorf("invalid Wasm result: %s", name)
	}
	return api.DecodeU32(results[0]), nil
}
func invoke(ctx context.Context, instance api.Module, input []byte, name string, args ...uint64) ([]byte, error) {
	if len(input) > 65536 {
		return nil, &Error{Code: "resource_limit", Resource: "input_bytes"}
	}
	ptr, err := call(ctx, instance, "input_ptr", uint64(len(input)))
	if err != nil {
		return nil, err
	}
	if ptr == 0 || !instance.Memory().Write(ptr, input) {
		return nil, fmt.Errorf("invalid input buffer")
	}
	size, err := call(ctx, instance, name, args...)
	if err != nil {
		return nil, err
	}
	if size == 0 || size > 1048576 {
		return nil, fmt.Errorf("invalid output length")
	}
	ptr, err = call(ctx, instance, "output_ptr")
	if err != nil {
		return nil, err
	}
	output, ok := instance.Memory().Read(ptr, size)
	if !ok {
		return nil, fmt.Errorf("invalid output buffer")
	}
	// Decoding is complete before this worker returns to the pool.
	return output, nil
}
func parseInstance(ctx context.Context, instance api.Module, source string, handle, mode uint32, registry ast.Registry) (*Result, error) {
	output, err := invoke(ctx, instance, []byte(source), "parse", uint64(handle), uint64(mode))
	if err != nil {
		return nil, err
	}
	var result struct {
		Document json.RawMessage `json:"document"`
		Error    json.RawMessage `json:"error"`
	}
	if err := ast.DecodeFields(output, &result, nil, []string{"document", "error"}, nil); err != nil {
		return nil, err
	}
	if (len(result.Document) == 0) == (len(result.Error) == 0) {
		return nil, fmt.Errorf("ambiguous Wasm result")
	}
	if len(result.Error) > 0 {
		return nil, decodeParseError(result.Error)
	}
	document, err := ast.Decode(result.Document, registry)
	if err != nil {
		return nil, err
	}
	if document.Source != source {
		return nil, fmt.Errorf("document source mismatch")
	}
	return &Result{Document: document, JSON: result.Document}, nil
}
