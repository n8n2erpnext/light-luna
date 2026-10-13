package policy

import (
	"bytes"
	"strings"
	"testing"
)

func TestPolicyRejectsDuplicateKeysRecursively(t *testing.T) {
	good := string(fixture(t))
	for name, body := range map[string]string{
		"metadata duplicate": strings.Replace(good, `"revision": 1`, `"revision": 1, "revision": 10`, 1),
		"rule duplicate":     strings.Replace(good, `"priority": 200`, `"priority": 200, "priority": 201`, 1),
		"nested duplicate":   strings.Replace(good, `"enabled": true`, `"enabled": true, "enabled": false`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Decode(bytes.NewBufferString(body)); err == nil {
				t.Fatal("duplicate key accepted")
			}
		})
	}
}
