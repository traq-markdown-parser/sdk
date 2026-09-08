package ast_test

import (
	"encoding/json"
	"fmt"
	"github.com/traq-markdown-parser/sdk/go/ast"
	"github.com/traq-markdown-parser/sdk/go/binding"
	"github.com/traq-markdown-parser/sdk/go/extensions/commonmark"
	"github.com/traq-markdown-parser/sdk/go/extensions/trap"
	"reflect"
	"strings"
	"testing"
)

func TestStrictCommonAndExtensionContracts(t *testing.T) {
	valid := fmt.Sprintf(`{"source":"日本","children":[{"kind":"%s","span":{"start":0,"end":6},"data":{"type":"user","id":"u","label":"@u"}}]}`, trap.ReferenceName)
	if _, err := ast.Decode([]byte(valid), binding.Registry()); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{
		strings.Replace(valid, `"id":"u"`, `"id":null`, 1),
		strings.Replace(valid, `"id":"u"`, `"id":3`, 1),
		strings.Replace(valid, `"label":"@u"`, `"unexpected":"@u"`, 1),
		strings.Replace(valid, `"start":0`, `"start":1`, 1),
		strings.Replace(valid, `"end":6`, `"end":7`, 1),
		strings.Replace(valid, `"kind":"`+trap.ReferenceName+`"`, `"kind":"`+trap.ReferenceName+`","value":"extra"`, 1),
		strings.Replace(valid, `"kind":"`+trap.ReferenceName+`"`, `"kind":"`+trap.ReferenceName+`","children":null`, 1),
		strings.Replace(valid, trap.ReferenceName, `unknown@1`, 1),
		valid + `{}`,
	} {
		if _, err := ast.Decode([]byte(raw), binding.Registry()); err == nil {
			t.Fatal("invalid contract was accepted")
		}
	}
}

func TestCommonPayloadAndSemanticValidationBoundary(t *testing.T) {
	raw := fmt.Sprintf(`{"source":"link","children":[{"kind":"%s","span":{"start":0,"end":4},"data":{"destination":"https://example.com","form":"explicit","title":null}}]}`, commonmark.LinkName)
	document, err := ast.Decode([]byte(raw), binding.Registry())
	if err != nil {
		t.Fatal(err)
	}
	link, ok := document.Children[0].Payload.(commonmark.Link)
	if !ok || link.Destination != "https://example.com" || link.Title != nil {
		t.Fatal("CommonMark payload was not restored to its generated type")
	}
	encoded, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := ast.Decode(encoded, binding.Registry())
	if err != nil || !reflect.DeepEqual(decoded, document) {
		t.Fatalf("typed payload roundtrip failed: %v", err)
	}
	// Hosts enforce u8 representation. Heading's semantic range belongs to Rust.
	for _, level := range []int{7, 256} {
		raw := fmt.Sprintf(`{"source":"x","children":[{"kind":"%s","span":{"start":0,"end":1},"data":{"level":%d}}]}`, commonmark.HeadingName, level)
		_, err := ast.Decode([]byte(raw), binding.Registry())
		if (err == nil) != (level == 7) {
			t.Fatalf("unexpected host validation for level %d: %v", level, err)
		}
	}
}
