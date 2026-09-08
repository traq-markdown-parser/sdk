package core

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/tetratelabs/wazero/api"
	"github.com/traq-markdown-parser/sdk/go/ast"
	"github.com/traq-markdown-parser/sdk/go/binding"
)

func validateArtifact(ctx context.Context, instance api.Module) error {
	if instance.Memory() == nil {
		return fmt.Errorf("missing Wasm memory")
	}
	for name, arity := range map[string]int{"abi_version": 0, "ast_version": 0, "input_ptr": 1, "parse": 2, "output_ptr": 0, "contract_ptr": 0, "contract_len": 0, "grammar_build": 0, "grammar_drop": 1, "grammar_count": 0} {
		fn := instance.ExportedFunction(name)
		if fn == nil {
			return fmt.Errorf("missing Wasm export %s", name)
		}
		definition := fn.Definition()
		if len(definition.ParamTypes()) != arity || len(definition.ResultTypes()) != 1 || definition.ResultTypes()[0] != api.ValueTypeI32 {
			return fmt.Errorf("invalid export signature %s", name)
		}
		for _, param := range definition.ParamTypes() {
			if param != api.ValueTypeI32 {
				return fmt.Errorf("invalid parameter type %s", name)
			}
		}
	}
	for name, expected := range map[string]uint32{"abi_version": 2, "ast_version": 4} {
		value, err := call(ctx, instance, name)
		if err != nil {
			return err
		}
		if value != expected {
			return fmt.Errorf("unsupported %s: %d", name, value)
		}
	}
	ptr, err := call(ctx, instance, "contract_ptr")
	if err != nil {
		return err
	}
	size, err := call(ctx, instance, "contract_len")
	if err != nil {
		return err
	}
	if size == 0 || size > 65536 {
		return fmt.Errorf("invalid contract size")
	}
	raw, ok := instance.Memory().Read(ptr, size)
	if !ok {
		return fmt.Errorf("invalid contract buffer")
	}
	var contract struct {
		ABI     uint32 `json:"abiVersion"`
		AST     uint32 `json:"astVersion"`
		Catalog any    `json:"catalog"`
		Limits  struct {
			Input    uint32 `json:"inputBytes"`
			Output   uint32 `json:"outputBytes"`
			Memory   uint32 `json:"memoryBytes"`
			Grammars uint32 `json:"grammars"`
		} `json:"limits"`
	}
	if err := ast.DecodeFields(raw, &contract, []string{"abiVersion", "astVersion", "catalog", "limits"}, nil, nil); err != nil {
		return err
	}
	if contract.ABI != 2 || contract.AST != 4 {
		return fmt.Errorf("invalid artifact contract")
	}
	if contract.Limits.Input != 65536 || contract.Limits.Output != 1048576 || contract.Limits.Memory != 33554432 || contract.Limits.Grammars != 256 {
		return fmt.Errorf("incompatible resource limits")
	}
	var expected any
	if err := json.Unmarshal(binding.CatalogJSON, &expected); err != nil {
		return err
	}
	if !reflect.DeepEqual(contract.Catalog, expected) {
		return fmt.Errorf("artifact catalog does not match SDK")
	}
	return nil
}
