package policy

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"testing"
)

func policyWithCIDRUnion(t *testing.T, earlier, later []string) Bundle {
	t.Helper()
	b := twoCoreRules(t)
	if len(later) != 1 {
		t.Fatal("test helper requires single later prefix")
	}
	b.Spec.Rules[1].Source = "trusted-admin"
	b.Spec.Rules[0].Source = later[0]
	b.Spec.Aliases[0].Values = nil
	for _, cidr := range earlier {
		raw, err := json.Marshal(cidr)
		if err != nil {
			t.Fatal(err)
		}
		b.Spec.Aliases[0].Values = append(b.Spec.Aliases[0].Values, json.RawMessage(raw))
	}
	return b
}

func TestCIDRUnionShadowAndMissingGap(t *testing.T) {
	tests := []struct {
		name    string
		earlier []string
		later   string
		want    string
	}{
		{"two ipv4 halves cover whole /24",
			[]string{"192.0.2.0/25", "192.0.2.128/25"},
			"192.0.2.0/24", "LATER_RULE_SHADOWED"},
		{"three ipv4 quarters with gap",
			[]string{"192.0.2.0/26", "192.0.2.64/26", "192.0.2.192/26"},
			"192.0.2.0/24", "DIFFERENT_ACTIONS_OVERLAP"},
		{"overlapping ipv4 members do not fill gap",
			[]string{"192.0.2.0/25", "192.0.2.64/26"},
			"192.0.2.0/24", "DIFFERENT_ACTIONS_OVERLAP"},
		{"ipv6 halves cover whole /32",
			[]string{"2001:db8::/33", "2001:db8:8000::/33"},
			"2001:db8::/32", "LATER_RULE_SHADOWED"},
		{"ipv6 half leaves gap",
			[]string{"2001:db8::/33"},
			"2001:db8::/32", "DIFFERENT_ACTIONS_OVERLAP"},
		{"mixed ipv4 ipv6 does not fill ipv4 gap",
			[]string{"192.0.2.0/25", "2001:db8::/32"},
			"192.0.2.0/24", "DIFFERENT_ACTIONS_OVERLAP"},
		{"disjoint family",
			[]string{"2001:db8::/32"},
			"192.0.2.0/24", ""},
		{"single superset is still covered",
			[]string{"192.0.2.0/24"},
			"192.0.2.128/25", "LATER_RULE_SHADOWED"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := policyWithCIDRUnion(t, tt.earlier, []string{tt.later})
			got, err := AnalyzeConflicts(b)
			if err != nil {
				t.Fatal(err)
			}
			if !got.Complete || got.KernelWrites || got.Enforceable {
				t.Fatalf("must remain complete non-enforcing: %+v", got)
			}
			if tt.want == "" {
				if len(got.Findings) != 0 {
					t.Fatalf("expected no matching source: %+v", got)
				}
				return
			}
			if len(got.Findings) != 1 || got.Findings[0].Code != tt.want {
				t.Fatalf("got %+v, expected %s", got, tt.want)
			}
		})
	}
}

func TestExactUnionCoveragePropertyIPv4(t *testing.T) {
	target := netip.MustParsePrefix("192.0.2.0/28")
	parts := []string{"192.0.2.0/30", "192.0.2.4/30", "192.0.2.8/30", "192.0.2.12/30"}
	for mask := 1; mask < 16; mask++ {
		var selected []netip.Prefix
		for i, s := range parts {
			if mask&(1<<i) != 0 {
				selected = append(selected, netip.MustParsePrefix(s))
			}
		}
		expected := mask == 15
		got := unionContainsPrefix(target, selected)
		if got != expected {
			t.Fatalf("mask %04b coverage %v expected %v", mask, got, expected)
		}
	}
}

func TestExactUnionCoveragePropertyIPv6(t *testing.T) {
	target := netip.MustParsePrefix("2001:db8::/126")
	for mask := 1; mask < 16; mask++ {
		var members []netip.Prefix
		for i := 0; i < 4; i++ {
			if mask&(1<<i) != 0 {
				members = append(members, netip.MustParsePrefix(fmt.Sprintf("2001:db8::%x/128", i)))
			}
		}
		if got := unionContainsPrefix(target, members); got != (mask == 15) {
			t.Fatalf("IPv6 mask %04b wrongly considered covered: %v", mask, got)
		}
	}
}

func TestCIDRComparisonBudgetForAliasUnionFailClosed(t *testing.T) {
	a := matchSources{prefixes: []netip.Prefix{
		netip.MustParsePrefix("192.0.2.0/25"), netip.MustParsePrefix("192.0.2.128/25"),
	}}
	b := matchSources{prefixes: []netip.Prefix{netip.MustParsePrefix("192.0.2.0/24")}}
	counter := MaxCIDRComparisons - 1
	_, _, limit := compareSources(a, b, &counter)
	if !limit || counter != MaxCIDRComparisons {
		t.Fatalf("budget overrun not fail-closed: limit=%v comparisons=%d", limit, counter)
	}
}

func TestPrefixEndsAtMaximumIPv4IPv6Address(t *testing.T) {
	for _, src := range []string{"0.0.0.0/0", "255.255.255.255/32", "::/0", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128"} {
		p := netip.MustParsePrefix(src)
		end := prefixLast(p)
		if !p.Contains(end) {
			t.Fatalf("%s last addr escaped: %s", p, end)
		}
		if end.Next().IsValid() {
			t.Fatalf("%s should finish address family", p)
		}
	}
}
