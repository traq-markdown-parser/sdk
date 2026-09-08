package core

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"testing"
	"time"
)

func compiledHandles(t *testing.T, r *Runtime) uint32 {
	t.Helper()
	var total uint32
	for range cap(r.pool) {
		requireOK(t, r.withWorker(context.Background(), func(w *worker) error {
			count, err := call(context.Background(), w.module, "grammar_count")
			if int(count) != len(w.grammars) {
				t.Fatal("untracked Wasm grammar")
			}
			total += count
			return err
		}))
	}
	return total
}
func TestPresetConstructionIsLazyAndShared(t *testing.T) {
	r := runtimeForTest(t)
	ctx := context.Background()
	g := r.Presets.TraQ.V1
	description := g.Describe()
	plugins := g.Plugins()
	b := g.ToBuilder()
	if r.grammarCount.Load() != 0 || compiledHandles(t, r) != 0 {
		t.Fatal("preset metadata compiled grammars")
	}
	independent, err := b.Build(ctx)
	requireOK(t, err)
	if independent.Describe() != description || len(independent.Plugins()) != len(plugins) {
		t.Fatal("metadata changed after compilation")
	}
	requireOK(t, independent.Close())
	const callers = 24
	parsers := make(chan *Parser, callers)
	var group sync.WaitGroup
	for range callers {
		group.Go(func() {
			p, err := r.Parser(ctx, g)
			if err != nil {
				t.Error(err)
				return
			}
			parsers <- p
		})
	}
	group.Wait()
	close(parsers)
	if len(parsers) != callers || r.grammarCount.Load() != 1 || compiledHandles(t, r) != 1 {
		t.Fatal("constructors must share one initial compilation")
	}
	requireOK(t, g.Close())
	for p := range parsers {
		_, err := p.Parse(ctx, "**text** $x$")
		requireOK(t, err)
		requireOK(t, p.Close())
	}
	if r.grammarCount.Load() != 0 || compiledHandles(t, r) != 0 {
		t.Fatal("last owner failed to release")
	}
	requireOK(t, r.Presets.CommonMark.Close())
	if _, err := r.Parser(ctx, r.Presets.CommonMark); err == nil {
		t.Fatal("closed unused preset accepted")
	}
}

// Occupying all workers makes a first constructor wait without timing a Wasm call.
func borrowAll(r *Runtime) func() {
	held := make([]*worker, cap(r.pool))
	for i := range held {
		held[i] = <-r.pool
	}
	return func() {
		for _, w := range held {
			r.pool <- w
		}
	}
}
func awaitReservation(t *testing.T, r *Runtime) {
	t.Helper()
	deadline := time.After(time.Second)
	for r.grammarCount.Load() != 1 {
		select {
		case <-deadline:
			t.Fatal("constructor did not reserve capacity")
		default:
			runtime.Gosched()
		}
	}
}
func TestPresetInitializationCancellationAndRetry(t *testing.T) {
	r := runtimeForTest(t)
	g := r.Presets.TraQ.V1
	returnWorkers := borrowAll(r)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	first := make(chan error, 1)
	go func() {
		p, err := r.Parser(ctx, g)
		if p != nil {
			p.Close()
		}
		first <- err
	}()
	awaitReservation(t, r)
	// This constructor waits for the first constructor, rather than a worker.
	timeout, stop := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer stop()
	if _, err := r.Parser(timeout, g); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("initialization gate: %v", err)
	}
	cancel()
	select {
	case err := <-first:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("worker wait: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancellation did not unblock constructor")
	}
	if r.grammarCount.Load() != 0 || g.state.references.Load() != 1 {
		t.Fatal("canceled constructor leaked capacity or ownership")
	}
	returnWorkers()
	if compiledHandles(t, r) != 0 {
		t.Fatal("canceled initialization left a handle")
	}
	p, err := r.Parser(context.Background(), g)
	requireOK(t, err)
	defer p.Close()
	if r.grammarCount.Load() != 1 || compiledHandles(t, r) != 1 {
		t.Fatal("retry did not initialize")
	}
	canceled, done := context.WithCancel(context.Background())
	done()
	if _, err := r.Parser(canceled, g); !errors.Is(err, context.Canceled) {
		t.Fatal("initialized constructor ignored cancellation")
	}
}
func TestCloseDuringPresetInitialization(t *testing.T) {
	r := runtimeForTest(t)
	g := r.Presets.TraQ.V1
	returnWorkers := borrowAll(r)
	result := make(chan *Parser, 1)
	failure := make(chan error, 1)
	go func() { p, err := r.Parser(context.Background(), g); result <- p; failure <- err }()
	awaitReservation(t, r)
	// The in-flight constructor already retained the grammar.
	requireOK(t, g.Close())
	returnWorkers()
	p := <-result
	requireOK(t, <-failure)
	_, err := p.Parse(context.Background(), "survives grammar close")
	requireOK(t, err)
	requireOK(t, p.Close())
	if r.grammarCount.Load() != 0 || compiledHandles(t, r) != 0 {
		t.Fatal("constructor lease leaked")
	}
}
