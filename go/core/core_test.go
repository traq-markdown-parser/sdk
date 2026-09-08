package core

import (
	"context"
	"errors"
	"github.com/traq-markdown-parser/sdk/go/ast"
	"github.com/traq-markdown-parser/sdk/go/extensions/trap"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestWasmIsolationAndFailure(t *testing.T) {
	wasm, err := os.ReadFile("../../dist/parser.wasm")
	if err != nil {
		t.Fatal("build Wasm first: ", err)
	}
	ctx := context.Background()
	parser, err := New(ctx, wasm)
	if err != nil {
		t.Fatal(err)
	}
	defer parser.Close(ctx)
	for _, test := range []struct {
		source string
		code   string
	}{
		{strings.Repeat("あ", 22000), "resource_limit"}, {string([]byte{255}), "invalid_utf8"},
	} {
		_, err := parseDefault(parser, ctx, test.source)
		var parseError *Error
		if !errors.As(err, &parseError) || parseError.Code != test.code {
			t.Fatalf("unexpected parse error: %v", err)
		}

	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := parseDefault(parser, cancelled, "text"); err == nil {
		t.Fatal("cancelled parse succeeded")
	}
	var workers sync.WaitGroup
	for i := 0; i < 12; i++ {
		workers.Go(func() {
			source := "日本語🦀\r\n`!{\"type\":\"user\",\"id\":\"hidden\",\"raw\":\"@hidden\"}`\n\n!{\"type\":\"user\",\"id\":\"visible\",\"raw\":\"@visible\"}"
			raw, err := parseDefault(parser, ctx, source)
			if err != nil {
				t.Error(err)
				return
			}
			document := raw.Document
			if document.Source != source {
				t.Error("document contract broken")
			}
			var ids []string
			var visit func([]ast.Node)
			visit = func(nodes []ast.Node) {
				for _, node := range nodes {
					if reference, ok := node.Payload.(trap.Reference); ok && reference.Type == "user" {
						ids = append(ids, reference.ID)
					}
					visit(node.Children)
				}
			}
			visit(document.Children)
			if strings.Join(ids, ",") != "visible" {
				t.Errorf("opaque content leaked mentions: %v", ids)
			}
		})
	}
	workers.Wait()

	columns := 7500
	source := strings.Repeat("|", columns+1) + "\n" + strings.Repeat("|-", columns) + "|\n" + strings.Repeat("|", columns+1)
	_, err = parseDefault(parser, ctx, source)
	var limit *Error
	if !errors.As(err, &limit) || limit.Resource != "output_bytes" {
		t.Fatalf("output limit: %v", err)
	}
	// A cancelled request must not poison the next borrower of its instance.
	deadline, stop := context.WithTimeout(ctx, time.Microsecond)
	defer stop()
	if _, err := parseDefault(parser, deadline, strings.Repeat("[", 5000)); err == nil {
		t.Fatal("deadline did not terminate parse")
	}
	for i := 0; i < cap(parser.pool)+1; i++ {
		if _, err := parseDefault(parser, ctx, "after"); err != nil {
			t.Fatal("reuse after failure: ", err)
		}
	}
}

func TestCloseAndPoolWait(t *testing.T) {
	wasm, err := os.ReadFile("../../dist/parser.wasm")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	p, err := New(ctx, wasm)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close(ctx)
	// Exhaust the pool to exercise cancellation while waiting, deterministically.
	for range cap(p.pool) {
		<-p.pool
	}
	deadline, cancel := context.WithTimeout(ctx, time.Millisecond)
	defer cancel()
	if _, err := parseDefault(p, deadline, "waiting"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("pool wait: %v", err)
	}
	waiter := make(chan error)
	go func() { _, err := parseDefault(p, ctx, "waiting"); waiter <- err }()
	if err := p.Close(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-waiter:
		if err == nil {
			t.Fatal("closed parser succeeded")
		}
	case <-time.After(time.Second):
		t.Fatal("close did not unblock pool waiter")
	}
	if _, err := parseDefault(p, ctx, "after close"); err == nil {
		t.Fatal("closed parser succeeded")
	}
	if err := p.Close(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestInvalidArtifact(t *testing.T) {
	for _, wasm := range [][]byte{{1, 2, 3}, {0, 97, 115, 109, 1, 0, 0, 0}} {
		if p, err := New(context.Background(), wasm); err == nil {
			p.Close(context.Background())
			t.Fatal("incompatible artifact accepted")
		}
	}
}

func BenchmarkParse(b *testing.B) {
	wasm, err := os.ReadFile("../../dist/parser.wasm")
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	parser, err := New(ctx, wasm)
	if err != nil {
		b.Fatal(err)
	}
	defer parser.Close(ctx)
	source := "日本語の本文 **strong** :stamp: https://example.com\n\n- 項目\n- !{\"type\":\"user\",\"id\":\"u\",\"raw\":\"@user\"}"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := parseDefault(parser, ctx, source); err != nil {
			b.Fatal(err)
		}
	}
}

// User-stamp runs previously paid for fuzzy email recognition before the inline
// pass made them opaque. Keep this representative workload visible in benchmarks.
func BenchmarkUserStampTable(b *testing.B) {
	wasm, err := os.ReadFile("../../dist/parser.wasm")
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	parser, err := New(ctx, wasm)
	if err != nil {
		b.Fatal(err)
	}
	defer parser.Close(ctx)
	source := "| rank | user | value |\n| - | - | - |\n" + strings.Repeat("| 1 | [:@user_name: user_name](https://example.jp/users/user_name) | 1234 |\n", 100)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := parseDefault(parser, ctx, source); err != nil {
			b.Fatal(err)
		}
	}
}

func parseDefault(runtime *Runtime, ctx context.Context, source string) (*Result, error) {
	parser, err := runtime.Parser(ctx, runtime.Presets.TraQ.V1)
	if err != nil {
		return nil, err
	}
	defer parser.Close()
	return parser.Parse(ctx, source)
}
