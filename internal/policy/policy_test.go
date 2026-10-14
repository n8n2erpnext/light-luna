package policy

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func fixture(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("../../testdata/host-baseline.json")
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func decode(t *testing.T, data []byte) (Bundle, error) {
	t.Helper()
	return Decode(bytes.NewReader(data))
}
func TestValidFixture(t *testing.T) {
	b, err := decode(t, fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	if b.Spec.Mode != "host" || len(b.Spec.Rules) != 2 {
		t.Fatalf("unexpected bundle: %+v", b)
	}
}
func TestPreviewIsDeterministicAndInert(t *testing.T) {
	b, err := decode(t, fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	a, err := MakePreview(b)
	if err != nil {
		t.Fatal(err)
	}
	c, err := MakePreview(b)
	if err != nil {
		t.Fatal(err)
	}
	if a.HashSHA256 != c.HashSHA256 || a.HashSHA256 == "" {
		t.Fatal("unstable hash")
	}
	if a.Effects != "none: validation/preview only" || a.Enforceable {
		t.Fatal("unsafe preview")
	}
	if len(a.Rules) != 2 || a.Rules[0].Priority != 100 || a.Rules[1].Priority != 200 {
		t.Fatal("rule sorting wrong")
	}
}
func TestRejectMalformedPolicies(t *testing.T) {
	good := string(fixture(t))
	tests := map[string]string{
		"unknown field":       strings.Replace(good, `"revision": 1`, `"revision": 1, "unsafeExtra": true`, 1),
		"unknown nested rule": strings.Replace(good, `"priority": 200`, `"priority": 200, "command": "flush ruleset"`, 1),
		"real auto ban":       strings.Replace(good, `"allowRealBans": false`, `"allowRealBans": true`, 1),
		"gateway forbidden":   strings.Replace(good, `"mode": "host"`, `"mode": "gateway"`, 1),
		"forward forbidden":   strings.Replace(good, `"direction": "input"`, `"direction": "forward"`, 1),
		"bad cidr":            strings.Replace(good, `192.0.2.14/32`, `192.0.2.14/999`, 1),
		"noncanonical cidr":   strings.Replace(good, `192.0.2.0/24`, `192.0.2.14/24`, 1),
		"port out of range":   strings.Replace(good, `"destinationPort": 22`, `"destinationPort": 65536`, 1),
		"duplicate priority":  strings.Replace(good, `"priority": 200`, `"priority": 100`, 1),
		"port source alias":   strings.Replace(good, `"source": "trusted-admin"`, `"source": "web-ports"`, 1),
		"no rescue":           strings.Replace(good, `"enabled": true`, `"enabled": false`, 1),
		"missing mode":        strings.Replace(good, `"mode": "shadow"`, `"mode": ""`, 1),
		"trailing json":       good + "{}",
		"malformed json":      good[:len(good)/2],
	}
	for name, payload := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := decode(t, []byte(payload)); err == nil {
				t.Fatal("expected rejection")
			}
		})
	}
}

// Family agreement is enforced at the policy trust boundary, not deferred to
// nft-preview. A source alias is rejected if ANY member is incompatible.
func TestICMPSourceFamiliesValidateConsistently(t *testing.T) {
	base, err := decode(t, fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name        string
		protocol    string
		source      string
		aliasValues []string
		reject      bool
	}{
		{"IPv4 ICMP literal", "icmp", "192.0.2.0/24", nil, false},
		{"IPv6 ICMPv6 literal", "icmpv6", "2001:db8::/32", nil, false},
		{"IPv6 for IPv4 ICMP", "icmp", "2001:db8::/32", nil, true},
		{"IPv4 for ICMPv6", "icmpv6", "192.0.2.0/24", nil, true},
		{"any allowed for ICMP", "icmp", "any", nil, false},
		{"any allowed for ICMPv6", "icmpv6", "any", nil, false},
		{"IPv4 alias for IPv4 ICMP", "icmp", "trusted-admin", nil, false},
		{"IPv4 alias rejected for ICMPv6", "icmpv6", "trusted-admin", nil, true},
		{"mixed alias rejected for IPv4 ICMP", "icmp", "trusted-admin", []string{"192.0.2.14/32", "2001:db8::/32"}, true},
		{"mixed alias rejected for ICMPv6", "icmpv6", "trusted-admin", []string{"192.0.2.14/32", "2001:db8::/32"}, true},
		{"IPv6 alias for ICMPv6", "icmpv6", "trusted-admin", []string{"2001:db8::/32"}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := base
			b.Spec.Rules = append([]Rule(nil), base.Spec.Rules...)
			b.Spec.Aliases = append([]Alias(nil), base.Spec.Aliases...)
			if tc.aliasValues != nil {
				b.Spec.Aliases[0].Values = nil
				for _, cidr := range tc.aliasValues {
					raw, _ := json.Marshal(cidr)
					b.Spec.Aliases[0].Values = append(b.Spec.Aliases[0].Values, raw)
				}
			}
			b.Spec.Rules[1].Protocol = tc.protocol
			b.Spec.Rules[1].Source = tc.source
			b.Spec.Rules[1].DestinationPort = 0
			err := b.Validate()
			if (err != nil) != tc.reject {
				t.Fatalf("reject=%v got err=%v", tc.reject, err)
			}
		})
	}
}

// "any" is a policy-language source sentinel, not a user-defined alias.
// Its precedence differs across consumers (FlowSimulator/ConflictAnalyzer vs
// nft preview alias expansion), so validation must reject the collision.
func TestReservedAnyAliasCannotShadowSourceSentinel(t *testing.T) {
	b, err := decode(t, fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	b.Spec.Aliases[0].ID = "any"
	b.Spec.Rules[1].Source = "any"
	if err := b.Validate(); err == nil {
		t.Fatal("reserved any alias was accepted as a user-defined source")
	}
	if _, err := MakePreview(b); err == nil {
		t.Fatal("preview accepted ambiguous any alias")
	}
	if _, err := AnalyzeConflicts(b); err == nil {
		t.Fatal("static analyzer accepted ambiguous any alias")
	}
	// "any" remains legal as a rule source in bundles with no alias collision.
	b.Spec.Aliases[0].ID = "trusted-admin"
	if err := b.Validate(); err != nil {
		t.Fatalf("legitimate any wildcard source rejected: %v", err)
	}
	// A port-set alias is also forbidden from masquerading as the wildcard,
	// even if that alias is not referenced by the current rules.
	b.Spec.Aliases[1].ID = "any"
	if err := b.Validate(); err == nil || !strings.Contains(err.Error(), "reserved source id") {
		t.Fatalf("reserved port-set alias was not rejected: %v", err)
	}
}

// Source-of-truth Go semantic gate: fail closed when defaultInbound would
// deny new inbound traffic but no emergency CIDR is explicitly declared.
// This is declaration validation, not proof that a rescue route is reachable.
func TestDefaultInboundRequiresDeclaredEmergencyAccess(t *testing.T) {
	base, err := decode(t, fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name     string
		mode     string
		enabled  bool
		networks []string
		valid    bool
	}{
		{"accept can omit rescue", "accept", false, nil, true},
		{"accept may declare rescue", "accept", true, []string{"192.0.2.0/24"}, true},
		{"drop with approved rescue declaration", "drop", true, []string{"192.0.2.0/24"}, true},
		{"reject with approved rescue declaration", "reject", true, []string{"192.0.2.0/24"}, true},
		{"drop rescue disabled", "drop", false, []string{"192.0.2.0/24"}, false},
		{"reject rescue disabled", "reject", false, []string{"192.0.2.0/24"}, false},
		{"drop enabled but no CIDRs", "drop", true, nil, false},
		{"reject enabled but no CIDRs", "reject", true, []string{}, false},
		{"drop missing both rescue safeguards", "drop", false, nil, false},
		{"reject missing both rescue safeguards", "reject", false, nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := base
			b.Spec.DefaultInbound = tc.mode
			b.Spec.EmergencyAccess.Enabled = tc.enabled
			b.Spec.EmergencyAccess.Networks = tc.networks
			err := b.Validate()
			if (err == nil) != tc.valid {
				t.Fatalf("go policy valid=%t expected %t; err=%v", err == nil, tc.valid, err)
			}
		})
	}
}

// Ports use a zero wildcard in the P0 PolicyBundle model. ICMP families
// have no L4 destination port, so a nonzero value is a semantic error.
func TestProtocolDestinationPortContract(t *testing.T) {
	base, err := decode(t, fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name     string
		protocol string
		port     int
		valid    bool
	}{
		{"TCP wildcard zero", "tcp", 0, true},
		{"UDP wildcard zero", "udp", 0, true},
		{"TCP max port", "tcp", 65535, true},
		{"UDP real port", "udp", 53, true},
		{"TCP negative port", "tcp", -1, false},
		{"UDP over maximum", "udp", 65536, false},
		{"ICMP zero no port", "icmp", 0, true},
		{"ICMPv6 zero no port", "icmpv6", 0, true},
		{"ICMP with TCP port", "icmp", 443, false},
		{"ICMPv6 with UDP port", "icmpv6", 53, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := base
			b.Spec.Rules = append([]Rule(nil), base.Spec.Rules...)
			b.Spec.Rules[0].Protocol = tc.protocol
			b.Spec.Rules[0].DestinationPort = tc.port
			got := b.Validate()
			if (got == nil) != tc.valid {
				t.Fatalf("protocol=%s port=%d valid=%t expected=%t: %v",
					tc.protocol, tc.port, got == nil, tc.valid, got)
			}
		})
	}
}

// JSON Schema should advertise the exact representable metadata revision
// domain of the authoritative Go int64 decoder, not unbounded JSON integers.
func TestMetadataRevisionInt64Bounds(t *testing.T) {
	b, err := decode(t, fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	b.Metadata.Revision = int64(^uint64(0) >> 1)
	if err := b.Validate(); err != nil {
		t.Fatalf("maximum int64 revision rejected: %v", err)
	}
	original := string(fixture(t))
	for name, raw := range map[string]string{
		"int64-overflow": "9223372036854775808",
		"huge-integer":   "100000000000000000000000000000000000000000000000000000000000000000000000000000000",
		"zero":           "0",
		"negative":       "-1",
	} {
		t.Run(name, func(t *testing.T) {
			input := strings.Replace(original, `"revision": 1`, `"revision": `+raw, 1)
			if input == original {
				t.Fatal("fixture revision did not match")
			}
			if _, err := decode(t, []byte(input)); err == nil {
				t.Fatalf("accepted invalid metadata revision %q", raw)
			}
		})
	}
}

// Explicit JSON null is never a valid typed PolicyBundle field. For optional
// properties, omission/zero can remain valid, but null must not silently map
// to Go zero values and change declared policy intent.
func TestDecodeRejectExplicitNullFields(t *testing.T) {
	cases := []struct {
		name   string
		change func(map[string]any)
	}{
		{"null rules", func(spec map[string]any) { spec["rules"] = nil }},
		{"null aliases with no reference", func(spec map[string]any) {
			spec["aliases"] = nil
			spec["rules"].([]any)[1].(map[string]any)["source"] = "any"
		}},
		{"null optional note", func(spec map[string]any) {
			spec["rules"].([]any)[0].(map[string]any)["note"] = nil
		}},
		{"null optional destination port", func(spec map[string]any) {
			spec["rules"].([]any)[0].(map[string]any)["destinationPort"] = nil
		}},
		{"null autonomous TTL", func(spec map[string]any) {
			spec["autonomous"].(map[string]any)["maxBanTtlSeconds"] = nil
		}},
		{"null rescue networks with accept", func(spec map[string]any) {
			spec["defaultInbound"] = "accept"
			spec["emergencyAccess"].(map[string]any)["networks"] = nil
		}},
		{"null rescue enabled with accept", func(spec map[string]any) {
			spec["defaultInbound"] = "accept"
			spec["emergencyAccess"].(map[string]any)["enabled"] = nil
		}},
		{"null alias member", func(spec map[string]any) {
			spec["aliases"].([]any)[0].(map[string]any)["values"].([]any)[0] = nil
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var payload map[string]any
			if err := json.Unmarshal(fixture(t), &payload); err != nil {
				t.Fatal(err)
			}
			tc.change(payload["spec"].(map[string]any))
			raw, err := json.Marshal(payload)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := decode(t, raw); err == nil {
				t.Fatalf("accepted explicitly null %s", tc.name)
			}
		})
	}
}

// The public schema declares several mandatory keys. Unlike a JSON
// Schema object, Go structs silently default omitted keys, and encoding/json
// also accepts case-insensitive field spellings. The policy file boundary
// must require the actual case-sensitive contract without requiring optional
// arrays or ports.
func TestDecodeRejectMissingOrCaseFoldedRequiredKeys(t *testing.T) {
	tests := []struct {
		name   string
		change func(map[string]any)
	}{
		{"missing emergency object", func(b map[string]any) {
			delete(b["spec"].(map[string]any), "emergencyAccess")
		}},
		{"missing emergency enabled", func(b map[string]any) {
			delete(b["spec"].(map[string]any)["emergencyAccess"].(map[string]any), "enabled")
		}},
		{"missing emergency networks", func(b map[string]any) {
			delete(b["spec"].(map[string]any)["emergencyAccess"].(map[string]any), "networks")
		}},
		{"missing autonomous allowRealBans", func(b map[string]any) {
			delete(b["spec"].(map[string]any)["autonomous"].(map[string]any), "allowRealBans")
		}},
		{"missing autonomous maxBanTtlSeconds", func(b map[string]any) {
			delete(b["spec"].(map[string]any)["autonomous"].(map[string]any), "maxBanTtlSeconds")
		}},
		{"folded root Spec", func(b map[string]any) {
			b["Spec"] = b["spec"]
			delete(b, "spec")
		}},
		{"folded emergency Enabled", func(b map[string]any) {
			e := b["spec"].(map[string]any)["emergencyAccess"].(map[string]any)
			e["Enabled"] = e["enabled"]
			delete(e, "enabled")
		}},
		{"folded metadata Revision", func(b map[string]any) {
			m := b["metadata"].(map[string]any)
			m["Revision"] = m["revision"]
			delete(m, "revision")
		}},
		{"folded rule Action", func(b map[string]any) {
			r := b["spec"].(map[string]any)["rules"].([]any)[0].(map[string]any)
			r["Action"] = r["action"]
			delete(r, "action")
		}},
		{"folded alias Type", func(b map[string]any) {
			a := b["spec"].(map[string]any)["aliases"].([]any)[0].(map[string]any)
			a["Type"] = a["type"]
			delete(a, "type")
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var b map[string]any
			if err := json.Unmarshal(fixture(t), &b); err != nil {
				t.Fatal(err)
			}
			b["spec"].(map[string]any)["defaultInbound"] = "accept"
			tc.change(b)
			data, err := json.Marshal(b)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := decode(t, data); err == nil {
				t.Fatalf("accepted missing or case-folded required field: %s", tc.name)
			}
		})
	}
	// Actual optional properties stay optional. Absence is not a null token.
	var b map[string]any
	if err := json.Unmarshal(fixture(t), &b); err != nil {
		t.Fatal(err)
	}
	b["spec"].(map[string]any)["defaultInbound"] = "accept"
	delete(b["spec"].(map[string]any), "rules")
	delete(b["spec"].(map[string]any), "aliases")
	raw, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decode(t, raw); err != nil {
		t.Fatalf("rejecting optional omitted rules and aliases: %v", err)
	}
}

func TestRejectLargeInput(t *testing.T) {
	if _, err := decode(t, bytes.Repeat([]byte("a"), MaxPolicyBytes+1)); err == nil {
		t.Fatal("oversized input accepted")
	}
}
func TestNoMutationOfInput(t *testing.T) {
	b, err := decode(t, fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	before := b.Spec.Rules[0].Priority
	_, err = MakePreview(b)
	if err != nil {
		t.Fatal(err)
	}
	if b.Spec.Rules[0].Priority != before {
		t.Fatal("preview mutated rule order")
	}
}
