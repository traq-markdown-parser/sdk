package core

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/traq-markdown-parser/sdk/go/extensions/generic"
)

func runtimeForTest(t *testing.T) *Runtime {
	t.Helper()
	wasm, err := os.ReadFile("../../dist/parser.wasm")
	if err != nil {
		t.Fatal(err)
	}
	r, err := New(context.Background(), wasm)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { r.Close(context.Background()) })
	return r
}
func requireOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func TestNamespaceAndComposition(t *testing.T) {
	ctx := context.Background()
	r := runtimeForTest(t)
	generic := NewPluginGroup("generic")
	github := generic.Group("github")
	math := generic.New("math")
	b := r.Builder()
	requireOK(t, b.Add(r.Plugins.CommonMark.Core))
	requireOK(t, b.Add(math))
	requireOK(t, b.Add(github.New("math")))
	if b.Add(math) == nil {
		t.Fatal("duplicate symbol accepted")
	}
	g, err := b.Build(context.Background())
	requireOK(t, err)
	defer g.Close()
	if !strings.Contains(g.Describe(), "generic/github/math") {
		t.Fatal("namespace missing")
	}
	requireOK(t, b.Add(generic.New("math")))
	_, err = b.Build(context.Background())
	var build *BuildError
	if !errors.As(err, &build) || build.Code != "duplicate_name" {
		t.Fatalf("duplicate name: %v", err)
	}
	foreign := runtimeForTest(t)
	if r.Builder().Add(foreign.Plugins.Generic.Math) == nil {
		t.Fatal("foreign rule accepted")
	}
	if _, err := r.Parser(ctx, foreign.Presets.TraQ.V1); err == nil {
		t.Fatal("foreign grammar accepted")
	}
}
func TestIndependentLeasesAndForks(t *testing.T) {
	r := runtimeForTest(t)
	ctx := context.Background()
	original, err := r.Parser(ctx, r.Presets.TraQ.V1)
	requireOK(t, err)
	defer original.Close()
	b := r.Presets.TraQ.V1.ToBuilder()
	requireOK(t, b.Remove(r.Plugins.Generic.Math))
	g, err := b.Build(ctx)
	requireOK(t, err)
	p, err := r.Parser(ctx, g)
	requireOK(t, err)
	if len(g.Plugins()) != len(r.Presets.TraQ.V1.Plugins())-1 {
		t.Fatal("composition metadata lost")
	}
	requireOK(t, g.Close())
	requireOK(t, g.Close())
	if _, err := r.Parser(ctx, g); err == nil {
		t.Fatal("closed grammar accepted")
	}
	result, err := p.Parse(ctx, "$x$")
	requireOK(t, err)
	if strings.Contains(string(result.JSON), generic.InlineMathName) {
		t.Fatal("removed math still enabled")
	}
	result, err = original.Parse(ctx, "$x$")
	requireOK(t, err)
	if !strings.Contains(string(result.JSON), generic.InlineMathName) {
		t.Fatal("original preset changed")
	}
	rebuilt, err := g.ToBuilder().Build(ctx)
	requireOK(t, err)
	requireOK(t, rebuilt.Close())
	requireOK(t, p.Close())
	requireOK(t, p.Close())
	if _, err := p.Parse(ctx, "x"); err == nil {
		t.Fatal("closed parser accepted")
	}
}
func TestReleaseAcrossWorkersAndConcurrentClose(t *testing.T) {
	r := runtimeForTest(t)
	ctx := context.Background()
	for range 30 {
		g, err := r.Presets.TraQ.V1.ToBuilder().Build(ctx)
		requireOK(t, err)
		p, err := r.Parser(ctx, g)
		requireOK(t, err)
		// Sequential borrowers rotate through all workers.
		for range cap(r.pool) {
			_, err = p.Parse(ctx, "$x$")
			requireOK(t, err)
		}
		requireOK(t, g.Close())
		requireOK(t, p.Close())
		if g.state.references.Load() != 0 {
			t.Fatal("lease leaked")
		}
		for range cap(r.pool) {
			requireOK(t, r.withWorker(ctx, func(w *worker) error {
				if _, found := w.grammars[g.state]; found {
					t.Fatal("released grammar retained")
				}
				count, err := call(ctx, w.module, "grammar_count")
				if int(count) != len(w.grammars) {
					t.Fatal("Wasm handle leaked")
				}
				return err
			}))
		}
	}
	for range 30 {
		g, err := r.Presets.TraQ.V1.ToBuilder().Build(ctx)
		requireOK(t, err)
		p, err := r.Parser(ctx, g)
		requireOK(t, err)
		var workers sync.WaitGroup
		for range 8 {
			workers.Go(func() {
				_, err := p.Parse(ctx, "**text**")
				if err != nil && !strings.Contains(err.Error(), "closed") {
					t.Error(err)
				}
			})
		}
		workers.Go(func() { g.Close(); p.Close() })
		workers.Wait()
		if g.state.references.Load() != 0 {
			t.Fatal("concurrent lease leaked")
		}
	}
}
