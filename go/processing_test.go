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

func TestProcessing(t *testing.T) {
	ctx := context.Background()
	wasm, err := os.ReadFile("../dist/parser.wasm")
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := NewRuntime(ctx, wasm)
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Close(ctx)
	processor, err := runtime.NewProcessor(ctx, ProcessorPresetTraQV1, ProcessorOptions{Origin: "https://q.example.test"})
	if err != nil {
		t.Fatal(err)
	}
	plain, err := runtime.NewProcessor(ctx, ProcessorPresetTraQV1, ProcessorOptions{})
	if err != nil {
		t.Fatal(err)
	}
	parser, err := runtime.NewParser(ctx, PresetTraQV1)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../tests/fixtures/processing-notifications.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []struct{ Name, Source, Notification string }
	if err := json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatal(err)
	}
	if len(fixtures) != 787 {
		t.Fatal(len(fixtures))
	}
	for _, fixture := range fixtures {
		output, err := processor.Process(ctx, fixture.Source)
		if err != nil {
			t.Fatalf("%s: %v", fixture.Name, err)
		}
		if output.NotificationText != fixture.Notification {
			t.Fatalf("%s: got %q, want %q", fixture.Name, output.NotificationText, fixture.Notification)
		}
	}
	const id = "00000000-0000-0000-0000-000000000001"
	const user = `!{"type":"user","id":"` + id + `","raw":"@alice"}`
	result, err := processor.Process(ctx, user+" !!"+user+"!! `"+user+"`")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.References.Mentions, []string{id, id}) || result.NotificationText != "@alice ██████ "+user {
		t.Fatalf("%+v", result)
	}
	url := "https://q.example.test/files/" + id
	output, err := plain.Process(ctx, url)
	if err != nil || output.NotificationText != url {
		t.Fatalf("%+v %v", output, err)
	}
	if _, err := runtime.NewProcessor(ctx, ProcessorPreset("missing"), ProcessorOptions{}); err == nil {
		t.Fatal("accepted missing preset")
	}
	if _, err := runtime.NewProcessor(ctx, ProcessorPresetTraQV1, ProcessorOptions{Origin: strings.Repeat("x", 2049)}); err == nil {
		t.Fatal("accepted oversized origin")
	}
	for _, source := range []string{strings.Repeat("x", inputBytes+1), string([]byte{255}), strings.Repeat("!!", 100) + "deep" + strings.Repeat("!!", 100)} {
		if _, err := processor.Process(ctx, source); err == nil {
			t.Fatal("accepted invalid input")
		}
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := processor.Process(canceled, "x"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 8; j++ {
				output, err := processor.Process(ctx, "**next**")
				if err != nil {
					t.Error(err)
					return
				}
				if output.NotificationText != "next" {
					t.Error(output)
				}
			}
		}()
	}
	wg.Wait()

	if err := processor.Close(ctx); err != nil {
		t.Fatal(err)
	}
	if err := processor.Close(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := processor.Process(ctx, "closed"); err == nil {
		t.Fatal("closed processor accepted input")
	}
	if _, err := plain.Process(ctx, "still alive"); err != nil {
		t.Fatal(err)
	}
	if _, err := parser.Parse(ctx, "still alive"); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Close(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := plain.Process(ctx, "closed"); err == nil {
		t.Fatal("Runtime.Close left a processor open")
	}
	if _, err := runtime.NewProcessor(ctx, ProcessorPresetTraQV1, ProcessorOptions{}); err == nil {
		t.Fatal("closed runtime accepted creation")
	}
	if !reflect.DeepEqual(result.References.Mentions, []string{id, id}) {
		t.Fatal("results alias Wasm memory")
	}
}
