package markdown

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func parserFor(t *testing.T, preset Preset) *Parser {
	t.Helper()
	bytes, err := os.ReadFile("../dist/parser.wasm")
	if err != nil {
		t.Fatal(err)
	}
	p, err := New(context.Background(), bytes, preset)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { p.Close(context.Background()) })
	return p
}
func TestFixtureAST(t *testing.T) {
	p := parserFor(t, PresetTraQV1)
	raw, err := os.ReadFile("../tests/fixtures/commonmark-0.31.2.json")
	if err != nil {
		t.Fatal(err)
	}
	var common []struct {
		Example  int
		Markdown string
	}
	if err = json.Unmarshal(raw, &common); err != nil {
		t.Fatal(err)
	}
	sources := map[int]string{}
	for _, c := range common {
		sources[c.Example] = c.Markdown
	}
	for _, file := range []string{"traq-v1-commonmark", "traq-v1-extensions"} {
		raw, err := os.ReadFile("../tests/fixtures/" + file + ".json")
		if err != nil {
			t.Fatal(err)
		}
		var cases []struct {
			Name     string
			Source   string
			Example  int
			Expected map[string]struct {
				OK  json.RawMessage `json:"Ok"`
				Err json.RawMessage
			}
		}
		if err = json.Unmarshal(raw, &cases); err != nil {
			t.Fatal(err)
		}
		for i, fixture := range cases {
			for mode, parse := range map[string]func(context.Context, string) (*Document, error){"block": p.Parse, "inline": p.ParseInline} {
				source := fixture.Source
				if fixture.Example != 0 {
					source = sources[fixture.Example]
				}
				got, err := parse(context.Background(), source)
				expected := fixture.Expected[mode]
				if expected.OK == nil {
					if err == nil {
						t.Fatalf("%s %d %s: expected error", file, i, mode)
					}
					continue
				}
				if err != nil {
					t.Fatalf("%s %d %s: %v", file, i, mode, err)
				}
				actual, err := json.Marshal(got)
				if err != nil {
					t.Fatal(err)
				}
				var a, b any
				json.Unmarshal(actual, &a)
				json.Unmarshal(expected.OK, &b)
				if !reflect.DeepEqual(a, b) {
					t.Fatalf("%s %d %s: AST mismatch\ngot %s\nwant %s", file, i, mode, actual, expected.OK)
				}
			}
		}
	}
}
func TestLifecycleAndBounds(t *testing.T) {
	ctx := context.Background()
	p := parserFor(t, PresetTraQV1)
	common := parserFor(t, PresetCommonMark)
	first, err := p.ParseInline(ctx, ":stamp:")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := first.Children[0].Data.(*Stamp); !ok {
		t.Fatalf("wrong payload %T", first.Children[0].Data)
	}
	plain, err := common.ParseInline(ctx, ":stamp:")
	if err != nil {
		t.Fatal(err)
	}
	if plain.Children[0].Kind != TextName {
		t.Fatal("commonmark used traq grammar")
	}
	for _, source := range []string{strings.Repeat("x", inputBytes+1), string([]byte{255}), strings.Repeat("!!", 100) + "deep" + strings.Repeat("!!", 100)} {
		if _, err := p.Parse(ctx, source); err == nil {
			t.Fatal("expected error")
		}
		if _, err := p.Parse(ctx, "after"); err != nil {
			t.Fatal(err)
		}
	}
	if first.Source != ":stamp:" {
		t.Fatal("result aliases Wasm memory")
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := p.Parse(canceled, "x"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	// Canceling while waiting must not close the Wasm instance.
	p.gate <- struct{}{}
	waiting, stop := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { _, err := p.Parse(waiting, "waiting"); done <- err }()
	stop()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	<-p.gate
	if _, err := p.Parse(ctx, "still open"); err != nil {
		t.Fatal(err)
	}
	if err := p.Close(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.Close(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Parse(ctx, "closed"); err == nil {
		t.Fatal("closed parser accepted input")
	}
	if _, err := common.Parse(ctx, "independent"); err != nil {
		t.Fatal(err)
	}
}
func TestConcurrentCalls(t *testing.T) {
	p := parserFor(t, PresetTraQV1)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				got, err := p.Parse(context.Background(), "**hello** 日本語")
				if err != nil {
					t.Error(err)
					return
				}
				if got.Source != "**hello** 日本語" {
					t.Error("wrong source")
				}
			}
		}()
	}
	wg.Wait()
}
func TestInitialization(t *testing.T) {
	if _, err := New(context.Background(), []byte{0}, PresetTraQV1); err == nil {
		t.Fatal("accepted mismatched artifact")
	}
	bytes, err := os.ReadFile("../dist/parser.wasm")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := New(context.Background(), bytes, Preset("invalid")); err == nil {
		t.Fatal("accepted invalid preset")
	}
}
