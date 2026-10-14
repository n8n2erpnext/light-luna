package policy

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func twoCoreRules(t *testing.T) Bundle {
	t.Helper()
	b, err := decode(t, fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	early := b.Spec.Rules[1]
	early.Priority = 100
	early.Source = "any"
	early.DestinationPort = 22
	early.Action = "drop"
	late := b.Spec.Rules[0]
	late.Priority = 200
	late.Source = "192.0.2.14/32"
	late.DestinationPort = 22
	late.Action = "accept"
	b.Spec.Rules = []Rule{late, early} // inverse file order to prove priority sorting
	return b
}

// A complete clean *static* audit is not a rescue-route, kernel or
// externally coexisting firewall safety certificate. Every report path,
// including budgets that force INCOMPLETE, must carry explicit limitations.
func TestConflictReportsAlwaysDiscloseUnverifiedSafety(t *testing.T) {
	tests := []struct {
		name         string
		mutate       func(*Bundle)
		wantComplete bool
	}{
		{"complete no findings", func(b *Bundle) {
			b.Spec.Rules[0].DestinationPort = 443
		}, true},
		{"complete finding", func(b *Bundle) {}, true},
		{"incomplete pair budget", func(b *Bundle) {
			b.Spec.Rules = nil
			for i := 0; i < 200; i++ {
				b.Spec.Rules = append(b.Spec.Rules, Rule{
					ID:              fmt.Sprintf("%08x-1111-4111-8111-111111111111", i+1),
					Priority:        i + 1,
					Direction:       "input",
					Source:          "any",
					Protocol:        "tcp",
					DestinationPort: 443,
					Action:          "log",
				})
			}
		}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := twoCoreRules(t)
			tc.mutate(&b)
			r, err := AnalyzeConflicts(b)
			if err != nil {
				t.Fatal(err)
			}
			if r.Complete != tc.wantComplete || r.Enforceable || r.KernelWrites {
				t.Fatalf("incorrect audit authority/completeness: %+v", r)
			}
			serialized, err := json.Marshal(r)
			if err != nil {
				t.Fatal(err)
			}
			var decoded struct {
				Warnings []string `json:"warnings"`
			}
			if err := json.Unmarshal(serialized, &decoded); err != nil {
				t.Fatal(err)
			}
			joined := strings.ToLower(strings.Join(decoded.Warnings, " "))
			for _, keyword := range []string{"kernel", "emergencyaccess", "reachability", "owner", "executor"} {
				if !strings.Contains(joined, keyword) {
					t.Fatalf("report lost non-authoritative %q warning: %s", keyword, string(serialized))
				}
			}
		})
	}
}

func TestDefiniteShadowAndDeterministicOrder(t *testing.T) {
	b := twoCoreRules(t)
	r, err := AnalyzeConflicts(b)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Complete || r.Status != "FINDINGS_REQUIRE_REVIEW" || len(r.Findings) != 1 ||
		r.Findings[0].Code != "LATER_RULE_SHADOWED" || r.Findings[0].EarlierPriority != 100 ||
		r.KernelWrites || r.Enforceable {
		t.Fatalf("wrong inert shadow classification: %+v", r)
	}
	b.Spec.Rules[0], b.Spec.Rules[1] = b.Spec.Rules[1], b.Spec.Rules[0]
	r2, err := AnalyzeConflicts(b)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := json.Marshal(r)
	z, _ := json.Marshal(r2)
	if string(a) != string(z) {
		t.Fatal("source rule order changed static result")
	}
}

func TestIntentionalExceptionOverlapNotInvalidation(t *testing.T) {
	b := twoCoreRules(t)
	b.Spec.Rules[1].Action = "accept"
	b.Spec.Rules[1].Source = "192.0.2.14/32"
	b.Spec.Rules[0].Action = "drop"
	b.Spec.Rules[0].Source = "192.0.2.0/24"
	r, err := AnalyzeConflicts(b)
	if err != nil || len(r.Findings) != 1 || r.Findings[0].Code != "DIFFERENT_ACTIONS_OVERLAP" || !r.Complete {
		t.Fatalf("intentional exception must be reviewed, not rejected: %+v %v", r, err)
	}
	if err := b.Validate(); err != nil {
		t.Fatal("policy should remain valid", err)
	}
}

func TestPortsDirectionsProtocolsAndNonterminalLog(t *testing.T) {
	testCases := []struct {
		name   string
		mutate func(*Bundle)
		want   int
	}{
		{"different ports", func(b *Bundle) { b.Spec.Rules[0].DestinationPort = 443 }, 0},
		{"different directions", func(b *Bundle) { b.Spec.Rules[0].Direction = "output" }, 0},
		{"different L4 protocols", func(b *Bundle) { b.Spec.Rules[0].Protocol = "udp" }, 0},
		{"nonterminal earlier log", func(b *Bundle) { b.Spec.Rules[1].Action = "log" }, 0},
		{"wildcard TCP port shadows", func(b *Bundle) { b.Spec.Rules[1].DestinationPort = 0 }, 1},
		{"later log shadowed", func(b *Bundle) { b.Spec.Rules[0].Action = "log" }, 1},
		{"duplicate-action shadow", func(b *Bundle) { b.Spec.Rules[0].Action = "drop" }, 1},
		{"different IP families", func(b *Bundle) { b.Spec.Rules[1].Source = "2001:db8::/32" }, 0},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := twoCoreRules(t)
			tc.mutate(&b)
			r, err := AnalyzeConflicts(b)
			if err != nil || !r.Complete || len(r.Findings) != tc.want {
				t.Fatalf("wrong precedence case: %+v %v", r, err)
			}
		})
	}
}

func TestCIDRSetAndValidation(t *testing.T) {
	b := twoCoreRules(t)
	b.Spec.Rules[1].Source = "trusted-admin"
	b.Spec.Aliases[0].Values = []json.RawMessage{json.RawMessage(`"192.0.2.0/24"`)}
	r, err := AnalyzeConflicts(b)
	if err != nil || len(r.Findings) != 1 || r.Findings[0].Code != "LATER_RULE_SHADOWED" {
		t.Fatalf("CIDR alias must shadow narrow source: %+v %v", r, err)
	}
	b.Spec.Rules[0].Source = "web-ports"
	if _, err := AnalyzeConflicts(b); err == nil {
		t.Fatal("accepted port-set as CIDR source")
	}
}

func TestBudgetNeverReturnsFalseClean(t *testing.T) {
	b := twoCoreRules(t)
	b.Spec.Rules = nil
	for i := 0; i < 200; i++ {
		b.Spec.Rules = append(b.Spec.Rules, Rule{
			ID:       fmt.Sprintf("%08x-1111-4111-8111-111111111111", i+1),
			Priority: i + 1, Direction: "input", Source: "any", Protocol: "tcp",
			DestinationPort: 443, Action: "log",
		})
	}
	r, err := AnalyzeConflicts(b)
	if err != nil || r.Complete || r.Status != "INCOMPLETE_REVIEW_REQUIRED" ||
		r.LimitedBy != "RULE_PAIR_BUDGET" || r.RulePairs != MaxRulePairs ||
		r.KernelWrites || r.Enforceable {
		t.Fatalf("over-budget policy falsely called complete: %+v %v", r, err)
	}
}

func TestFindingBudgetReportsIncomplete(t *testing.T) {
	b := twoCoreRules(t)
	b.Spec.Rules = nil
	for i := 0; i < 20; i++ {
		b.Spec.Rules = append(b.Spec.Rules, Rule{
			ID:       fmt.Sprintf("%08x-1111-4111-8111-111111111111", i+1),
			Priority: i + 1, Direction: "input", Source: "any", Protocol: "tcp",
			DestinationPort: 443, Action: "accept",
		})
	}
	r, err := AnalyzeConflicts(b)
	if err != nil || r.Complete || r.LimitedBy != "FINDING_BUDGET" ||
		len(r.Findings) != MaxRuleFindings || r.Enforceable {
		t.Fatalf("finding limit must force explicit incomplete: %+v %v", r, err)
	}
}
