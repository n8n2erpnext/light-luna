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
