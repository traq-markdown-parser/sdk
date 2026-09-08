package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/traq-markdown-parser/sdk/go/core"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	wasmPath := flag.String("wasm", "../../dist/parser.wasm", "path to the shared Wasm artifact")
	flag.Parse()
	ctx := context.Background()
	wasm, err := os.ReadFile(*wasmPath)
	if err != nil {
		return err
	}
	runtime, err := core.New(ctx, wasm)
	if err != nil {
		return err
	}
	defer runtime.Close(ctx)

	parser, err := runtime.Parser(ctx, runtime.Presets.TraQ.V1)
	if err != nil {
		return err
	}
	defer parser.Close()
	result, err := parser.Parse(ctx, "**hello** :stamp: $x$")
	if err != nil {
		return err
	}
	fmt.Println(string(result.JSON))

	// Independent compositions can reuse the same Runtime and worker pool.
	builder := runtime.Presets.TraQ.V1.ToBuilder()
	if err := builder.Remove(runtime.Plugins.Generic.Math); err != nil {
		return err
	}
	grammar, err := builder.Build(ctx)
	if err != nil {
		return err
	}
	defer grammar.Close()
	withoutMath, err := runtime.Parser(ctx, grammar)
	if err != nil {
		return err
	}
	defer withoutMath.Close()
	grammar.Close() // Existing Parsers retain their own ownership.
	result, err = withoutMath.ParseInline(ctx, "$x$")
	if err != nil {
		return err
	}
	fmt.Println(string(result.JSON))
	return nil
}
