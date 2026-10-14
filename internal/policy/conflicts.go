package policy

// Non-enforcing, bounded static relationship analysis for validated Host/Edge
// rules. No kernel, socket, filesystem or privileged-executor operations.
import (
	"encoding/json"
	"fmt"
	"net/netip"
	"sort"
)

const (
	MaxRulePairs       = 8192
	MaxCIDRComparisons = 100000
	MaxRuleFindings    = 128
)

type ConflictFinding struct {
	Code            string `json:"code"`
	EarlierID       string `json:"earlierId"`
	LaterID         string `json:"laterId"`
	EarlierPriority int    `json:"earlierPriority"`
	LaterPriority   int    `json:"laterPriority"`
	EarlierAction   string `json:"earlierAction"`
	LaterAction     string `json:"laterAction"`
}

type ConflictReport struct {
	Schema          string            `json:"schema"`
	Status          string            `json:"status"`
	Complete        bool              `json:"complete"`
	LimitedBy       string            `json:"limitedBy,omitempty"`
	RulePairs       int               `json:"rulePairs"`
	CIDRComparisons int               `json:"cidrComparisons"`
	Findings        []ConflictFinding `json:"findings"`
	Warnings        []string          `json:"warnings"`
	Enforceable     bool              `json:"enforceable"`
	KernelWrites    bool              `json:"kernelWrites"`
}

type matchSources struct {
	any      bool
	prefixes []netip.Prefix
}

func ruleSources(b Bundle, r Rule) (matchSources, error) {
	if r.Source == "any" {
		return matchSources{any: true}, nil
	}
	for _, a := range b.Spec.Aliases {
		if a.ID != r.Source {
			continue
		}
		if a.Type != "cidr-set" {
			return matchSources{}, fmt.Errorf("unsupported source alias")
		}
		s := matchSources{prefixes: make([]netip.Prefix, 0, len(a.Values))}
		for _, raw := range a.Values {
			var value string
			if err := json.Unmarshal(raw, &value); err != nil {
				return matchSources{}, err
			}
			pr, err := netip.ParsePrefix(value)
			if err != nil {
				return matchSources{}, err
			}
			s.prefixes = append(s.prefixes, pr)
		}
		return s, nil
	}
	pr, err := netip.ParsePrefix(r.Source)
	if err != nil {
		return matchSources{}, err
	}
	return matchSources{prefixes: []netip.Prefix{pr}}, nil
}

func prefixContains(outer, inner netip.Prefix) bool {
	return outer.Addr().Is4() == inner.Addr().Is4() &&
		outer.Bits() <= inner.Bits() && outer.Contains(inner.Addr())
}

func compareSources(a, b matchSources, counter *int) (overlaps, covers, limit bool) {
	if a.any {
		return true, true, false
	}
	if b.any {
		return len(a.prefixes) > 0, false, false
	}
	covers = true
	for _, bp := range b.prefixes {
		contained := false
		fragments := make([]netip.Prefix, 0, len(a.prefixes))
		for _, ap := range a.prefixes {
			if *counter >= MaxCIDRComparisons {
				return false, false, true
			}
			*counter = *counter + 1
			if prefixContains(ap, bp) {
				overlaps = true
				contained = true
			} else if prefixContains(bp, ap) {
				overlaps = true
				fragments = append(fragments, ap)
			}
		}
		// Each earlier alias can be a *union* of disjoint networks.
		// Prove full coverage only if all address gaps are closed,
		// never merely because the sets overlap.
		if !contained && unionContainsPrefix(bp, fragments) {
			contained = true
		}
		if !contained {
			covers = false
		}
	}
	return overlaps, covers, false
}

// AnalyzeConflicts reports definite single-rule shadows and overlapping
// different actions. These are REVIEW observations, not valid/invalid
// verdicts and never grants to modify firewall state.
func AnalyzeConflicts(b Bundle) (ConflictReport, error) {
	if err := b.Validate(); err != nil {
		return ConflictReport{}, err
	}
	report := ConflictReport{
		Schema: "lightluna.rule-conflicts/v1alpha1", Status: "COMPLETE_NO_FINDINGS",
		Complete: true, Findings: []ConflictFinding{},
		Warnings: []string{
			"Static rule conflicts do not establish effective kernel verdicts or third-party firewall coexistence.",
			"emergencyAccess networks are declarations only; independent management and rescue reachability remain UNKNOWN.",
			"This report does not authenticate an owner, grant executor authority, or authorize APPLY.",
		},
	}
	rules := append([]Rule(nil), b.Spec.Rules...)
	sort.Slice(rules, func(i, j int) bool { return rules[i].Priority < rules[j].Priority })
	sources := make([]matchSources, len(rules))
	for i, r := range rules {
		s, err := ruleSources(b, r)
		if err != nil {
			return ConflictReport{}, err
		}
		sources[i] = s
	}
	for i := 0; i < len(rules); i++ {
		for j := 0; j < i; j++ {
			if report.RulePairs >= MaxRulePairs {
				report.Complete = false
				report.LimitedBy = "RULE_PAIR_BUDGET"
				report.Status = "INCOMPLETE_REVIEW_REQUIRED"
				return report, nil
			}
			report.RulePairs++
			early, late := rules[j], rules[i]
			if early.Direction != late.Direction || early.Protocol != late.Protocol || early.Action == "log" {
				continue
			}
			portOverlap := early.DestinationPort == 0 || late.DestinationPort == 0 || early.DestinationPort == late.DestinationPort
			if !portOverlap {
				continue
			}
			overlap, sourceCovers, limit := compareSources(sources[j], sources[i], &report.CIDRComparisons)
			if limit {
				report.Complete = false
				report.LimitedBy = "CIDR_COMPARISON_BUDGET"
				report.Status = "INCOMPLETE_REVIEW_REQUIRED"
				return report, nil
			}
			if !overlap {
				continue
			}
			portCovers := early.DestinationPort == 0 || early.DestinationPort == late.DestinationPort
			code := ""
			if sourceCovers && portCovers {
				code = "LATER_RULE_SHADOWED"
			} else if early.Action != late.Action {
				code = "DIFFERENT_ACTIONS_OVERLAP"
			}
			if code == "" {
				continue
			}
			report.Findings = append(report.Findings, ConflictFinding{
				Code: code, EarlierID: early.ID, LaterID: late.ID,
				EarlierPriority: early.Priority, LaterPriority: late.Priority,
				EarlierAction: early.Action, LaterAction: late.Action,
			})
			if len(report.Findings) >= MaxRuleFindings {
				report.Complete = false
				report.LimitedBy = "FINDING_BUDGET"
				report.Status = "INCOMPLETE_REVIEW_REQUIRED"
				return report, nil
			}
		}
	}
	if len(report.Findings) > 0 {
		report.Status = "FINDINGS_REQUIRE_REVIEW"
	}
	return report, nil
}
