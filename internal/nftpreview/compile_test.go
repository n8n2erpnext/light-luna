package nftpreview

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/n8n2erpnext/light-luna/internal/policy"
)

func example(t *testing.T) policy.Bundle {
	t.Helper()
	f, err := os.Open("../../testdata/host-baseline.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	p, err := policy.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func TestCompileInertAndDeterministic(t *testing.T) {
	p := example(t)
	a, err := Compile(p)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Compile(p)
	if err != nil {
		t.Fatal(err)
	}
	if a.UnhookedScript != b.UnhookedScript || a.PolicyHashSHA256 != b.PolicyHashSHA256 {
		t.Fatal("output is not deterministic")
	}
	if a.KernelWrites || a.Status != "UNHOOKED_TEXT_PREVIEW_ONLY" || a.ExpandedRules != 2 {
		t.Fatalf("unsafe plan: %+v", a)
	}
	for _, token := range []string{" hook ", " priority ", " flush ruleset", " policy drop", " policy reject", " add rule ", "insert rule"} {
		if strings.Contains(a.UnhookedScript, token) {
			t.Fatalf("forbidden hook/apply statement %q", token)
		}
	}
	if !strings.Contains(a.UnhookedScript, "table inet luna_p0_preview") ||
		!strings.Contains(a.UnhookedScript, "chain preview_inbound") ||
		!strings.Contains(a.UnhookedScript, "ip saddr 192.0.2.14/32 tcp dport 22 accept") {
		t.Fatal("missing basic nftables syntax")
	}
	first := strings.Index(a.UnhookedScript, "tcp dport 22")
	second := strings.Index(a.UnhookedScript, "tcp dport 8080")
	if first < 0 || second < 0 || first >= second {
		t.Fatal("rule priority order not preserved")
	}
	if !strings.Contains(a.Warnings[0], "NOT compiled") {
		t.Fatal("missing limitations")
	}
}

func TestProtocolAndFamilies(t *testing.T) {
	tests := []struct {
		protocol, source, action string
		port                     int
		want                     string
		reject                   bool
	}{
		{"tcp", "any", "accept", 0, "meta l4proto tcp accept", false},
		{"udp", "2001:db8::/32", "log", 5353, "ip6 saddr 2001:db8::/32 udp dport 5353 log", false},
		{"icmp", "any", "drop", 0, "meta nfproto ipv4 meta l4proto icmp drop", false},
		{"icmpv6", "2001:db8::/32", "reject", 0, "ip6 saddr 2001:db8::/32 meta nfproto ipv6 meta l4proto ipv6-icmp reject", false},
		{"icmpv6", "192.0.2.0/24", "accept", 0, "", true},
		{"icmp", "2001:db8::/32", "accept", 0, "", true},
	}
	for _, tc := range tests {
		r := policy.Rule{ID: "11111111-1111-1111-1111-111111111111", Priority: 100, Direction: "input", Source: tc.source, Protocol: tc.protocol, Action: tc.action, DestinationPort: tc.port}
		got, err := render(r, tc.source)
		if tc.reject {
			if err == nil {
				t.Errorf("%v expected rejection", tc)
			}
			continue
		}
		if err != nil || got != tc.want {
			t.Errorf("%v got %q err %v", tc, got, err)
		}
	}
}

func TestRejectUnsafeCrossFamilyPolicyAndExpansion(t *testing.T) {
	p := example(t)
	// A syntactically valid policy with mismatched address family still MUST NOT compile.
	p.Spec.Rules[0].Protocol = "icmpv6"
	p.Spec.Rules[0].Source = "192.0.2.0/24"
	p.Spec.Rules[0].DestinationPort = 0
	if _, err := Compile(p); err == nil {
		t.Fatal("accepted IPv4 source with ICMPv6")
	}
	p = example(t)
	// Aliases may expand to many items; generated text must remain bounded.
	values := make([]json.RawMessage, 256)
	for i := range values {
		b, _ := json.Marshal(fmt.Sprintf("192.0.2.%d/32", i))
		values[i] = b
	}
	// For many values, the policy's canonical CIDR check remains active.
	p.Spec.Aliases[0].Values = values
	for i := 0; i < 40; i++ {
		r := p.Spec.Rules[1]
		r.ID = makeUUID(i)
		r.Priority = 1000 + i
		p.Spec.Rules = append(p.Spec.Rules, r)
	}
	if _, err := Compile(p); err == nil {
		t.Fatal("accepted unbounded expansion")
	}
}

func TestNoNoteInjection(t *testing.T) {
	p := example(t)
	p.Spec.Rules[0].Note = "bad\"; hook input priority -500; policy accept;"
	a, err := Compile(p)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(a.UnhookedScript, "priority -500") || strings.Contains(a.UnhookedScript, "bad") {
		t.Fatal("untrusted note interpolated into nft syntax")
	}
}
func makeUUID(i int) string {
	// only test fixtures, deterministic UUIDs
	digits := "0123456789abcdef"
	return "11111111-1111-1111-1111-1111111111" + string([]byte{digits[i/16], digits[i%16]})
}
