package strictjson

import (
	"strings"
	"testing"
)

func TestRejectDuplicateKeysAndTrailingData(t *testing.T) {
	inputs := []string{
		`{"x":1,"x":2}`,
		`{"x":{"y":1,"y":2}}`,
		`{"x":[{"y":1,"y":2}]}`,
		`{"x":1}{"x":2}`,
		`{"x":1,"extra":2}`,
		`{"x":1`,
	}
	for _, payload := range inputs {
		var dst struct {
			X int `json:"x"`
		}
		if err := Decode(strings.NewReader(payload), 1024, &dst); err == nil {
			t.Errorf("accepted invalid JSON: %s", payload)
		}
	}
}

// Existing callers keep permissive JSON null behavior, but typed policy
// inputs can opt into an unambiguous no-null interpretation at the parser.
func TestDecodeRejectNullIsOptInAndRecursive(t *testing.T) {
	for _, raw := range []string{
		`null`, `{"x":null}`, `{"x":[false,{"y":null}]}`,
		`{"x":[null]}`, `[{"x":null}]`,
	} {
		t.Run(raw, func(t *testing.T) {
			var strict any
			if err := DecodeRejectNull(strings.NewReader(raw), 1024, &strict); err == nil {
				t.Fatal("strict parser accepted a nested or root null")
			}
			var compatible any
			if err := Decode(strings.NewReader(raw), 1024, &compatible); err != nil {
				t.Fatalf("legacy Decode changed null behavior: %v", err)
			}
		})
	}
	allowed := `{"x":[0,false,"",[],{}],"y":true}`
	var out any
	if err := DecodeRejectNull(strings.NewReader(allowed), 1024, &out); err != nil {
		t.Fatalf("non-null falsy values rejected: %v", err)
	}
	if err := DecodeRejectNull(strings.NewReader(`{"x":1,"x":2}`), 1024, &out); err == nil {
		t.Fatal("strict parser skipped existing duplicate-key protection")
	}
}

func TestValidAndBounded(t *testing.T) {
	var dst struct {
		X int `json:"x"`
	}
	if err := Decode(strings.NewReader(`{"x":5}`), 100, &dst); err != nil || dst.X != 5 {
		t.Fatalf("valid JSON failed: %v %+v", err, dst)
	}
	if err := Decode(strings.NewReader(`{"x":5}`), 4, &dst); err == nil {
		t.Fatal("input cap ignored")
	}
}
