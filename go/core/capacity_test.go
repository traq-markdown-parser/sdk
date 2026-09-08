package core

import (
	"context"
	"testing"
)

func TestGrammarCapacityIsSharedAcrossWorkers(t *testing.T) {
	r := runtimeForTest(t)
	ctx := context.Background()
	var held []*Grammar
	for range 256 {
		g, err := r.Presets.CommonMark.ToBuilder().Build(ctx)
		requireOK(t, err)
		held = append(held, g)
	}
	if _, err := r.Presets.CommonMark.ToBuilder().Build(ctx); err == nil {
		t.Fatal("runtime exceeded worker capacity")
	}
	if _, err := r.Parser(ctx, r.Presets.TraQ.V1); err == nil {
		t.Fatal("lazy preset exceeded capacity")
	}
	if r.grammarCount.Load() != 256 {
		t.Fatal("failed preset initialization leaked capacity")
	}
	// Force the most recently built grammar to be realized on every worker.
	p, err := r.Parser(ctx, held[len(held)-1])
	requireOK(t, err)
	for range cap(r.pool) {
		_, err = p.ParseInline(ctx, "text")
		requireOK(t, err)
	}
	p.Close()
	held[0].Close()
	retry, err := r.Parser(ctx, r.Presets.TraQ.V1)
	requireOK(t, err)
	_, err = retry.Parse(ctx, "after retry")
	requireOK(t, err)
	retry.Close()
	r.Presets.TraQ.V1.Close()
	for _, g := range held {
		g.Close()
	}
	if r.grammarCount.Load() != 0 {
		t.Fatal("capacity reservation leaked")
	}
	// A build rejected by core validation must return its reserved slot.
	b := r.Builder()
	requireOK(t, b.Add(NewPlugin("same")))
	requireOK(t, b.Add(NewPlugin("same")))
	if _, err := b.Build(ctx); err == nil {
		t.Fatal("duplicate accepted")
	}
	if r.grammarCount.Load() != 0 {
		t.Fatal("failed build leaked capacity")
	}
	g, err := r.Presets.CommonMark.ToBuilder().Build(ctx)
	requireOK(t, err)
	g.Close()
}
