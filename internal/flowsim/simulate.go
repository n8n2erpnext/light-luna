// Package flowsim models the *declared* PolicyBundle rule order for synthetic
// traffic. It intentionally cannot predict the effective kernel firewall verdict.
package flowsim

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"sort"

	"github.com/n8n2erpnext/light-luna/internal/policy"
)

const MaxCases = 128
const MaxInputBytes int64 = 256 << 10
const InputSchema = "lightluna.flow-fixtures/v1alpha1"

type Fixture struct {
	Schema string `json:"schema"`
	Cases  []Flow `json:"cases"`
}
type Flow struct {
	CaseID             string `json:"caseId"`
	Direction          string `json:"direction"`
	SourceIP           string `json:"sourceIp"`
	Protocol           string `json:"protocol"`
	DestinationPort    int    `json:"destinationPort"`
	ConnectionState    string `json:"connectionState"`
	ManagementCritical bool   `json:"managementCritical"`
}
type Result struct {
	CaseID           string   `json:"caseId"`
	PolicyIntent     string   `json:"policyIntent"`
	MatchedRuleID    string   `json:"matchedRuleId,omitempty"`
	LogRuleIDs       []string `json:"logRuleIds"`
	EffectiveVerdict string   `json:"effectiveVerdict"`
	ExternalFirewall string   `json:"externalFirewall"`
	ManagementRisk   string   `json:"managementRisk"`
	Explanation      []string `json:"explanation"`
}
type Report struct {
	Schema         string   `json:"schema"`
	Mode           string   `json:"mode"`
	PolicyName     string   `json:"policyName"`
	PolicyRevision int64    `json:"policyRevision"`
	KernelWrites   bool     `json:"kernelWrites"`
	Cases          []Result `json:"cases"`
	Warnings       []string `json:"warnings"`
}

func (f Fixture) Validate() error {
	if f.Schema != InputSchema {
		return errors.New("unsupported flow fixture schema")
	}
	if len(f.Cases) == 0 || len(f.Cases) > MaxCases {
		return fmt.Errorf("flow count must be 1..%d", MaxCases)
	}
	ids := make(map[string]struct{}, len(f.Cases))
	for i, v := range f.Cases {
		if len(v.CaseID) < 3 || len(v.CaseID) > 100 {
			return fmt.Errorf("case %d invalid caseId", i)
		}
		for _, r := range v.CaseID {
			if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.') {
				return fmt.Errorf("case %d invalid caseId characters", i)
			}
		}
		if _, ok := ids[v.CaseID]; ok {
			return fmt.Errorf("duplicate caseId %q", v.CaseID)
		}
		ids[v.CaseID] = struct{}{}
		if v.Direction != "input" && v.Direction != "output" {
			return fmt.Errorf("case %q unsupported direction", v.CaseID)
		}
		ip, err := netip.ParseAddr(v.SourceIP)
		if err != nil || !ip.IsGlobalUnicast() && !ip.IsPrivate() || ip.Is4In6() || ip.Zone() != "" {
			// Allow global unicast and private/documentation unicast; reject link-local,
			// unspecified, multicast, IPv4-mapped and scoped addresses.
			return fmt.Errorf("case %q invalid/unusable sourceIp", v.CaseID)
		}
		if v.Protocol != "tcp" && v.Protocol != "udp" && v.Protocol != "icmp" && v.Protocol != "icmpv6" {
			return fmt.Errorf("case %q unsupported protocol", v.CaseID)
		}
		if v.Protocol == "icmp" && !ip.Is4() || v.Protocol == "icmpv6" && !ip.Is6() {
			return fmt.Errorf("case %q ICMP family mismatch", v.CaseID)
		}
		if v.Protocol == "tcp" || v.Protocol == "udp" {
			if v.DestinationPort < 1 || v.DestinationPort > 65535 {
				return fmt.Errorf("case %q invalid L4 port", v.CaseID)
			}
		} else if v.DestinationPort != 0 {
			return fmt.Errorf("case %q ICMP cannot specify port", v.CaseID)
		}
		switch v.ConnectionState {
		case "new", "established", "unknown":
		default:
			return fmt.Errorf("case %q invalid connection state", v.CaseID)
		}
	}
	return nil
}

// Simulate matches the documented minimal rule subset. It is neither an
// nftables interpreter nor an actual conntrack/external ruleset simulator.
func Simulate(p policy.Bundle, f Fixture) (Report, error) {
	if err := p.Validate(); err != nil {
		return Report{}, err
	}
	if err := f.Validate(); err != nil {
		return Report{}, err
	}
	rules := append([]policy.Rule(nil), p.Spec.Rules...)
	sort.Slice(rules, func(i, j int) bool { return rules[i].Priority < rules[j].Priority })
	aliases := map[string][]netip.Prefix{}
	for _, a := range p.Spec.Aliases {
		if a.Type != "cidr-set" {
			continue
		}
		arr := make([]netip.Prefix, 0, len(a.Values))
		for _, raw := range a.Values {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return Report{}, err
			}
			pr, err := netip.ParsePrefix(s)
			if err != nil {
				return Report{}, err
			}
			arr = append(arr, pr)
		}
		aliases[a.ID] = arr
	}
	out := Report{
		Schema: "lightluna.flow-report/v1alpha1", Mode: "SYNTHETIC_POLICY_INTENT_ONLY",
		PolicyName: p.Metadata.Name, PolicyRevision: p.Metadata.Revision, KernelWrites: false,
		Cases: make([]Result, 0, len(f.Cases)),
		Warnings: []string{
			"All effective kernel verdicts are UNKNOWN. Docker/UFW, base chain priority and conntrack are not modeled.",
			"An emergency CIDR declaration never proves a working rescue or management route.",
			"Output default, real interface/zone routing, established-flow state, NAT and external firewall coexistence are not modeled.",
			"This is a pure-Go fixture simulator and DOES NOT call nftables or modify the host.",
		},
	}
	for _, flow := range f.Cases {
		ip, _ := netip.ParseAddr(flow.SourceIP)
		r := Result{
			CaseID: flow.CaseID, PolicyIntent: "UNSPECIFIED", EffectiveVerdict: "UNKNOWN",
			ExternalFirewall: "UNKNOWN", ManagementRisk: "NOT_ASSESSED",
			LogRuleIDs: []string{}, Explanation: []string{},
		}
		for _, rule := range rules {
			if rule.Direction != flow.Direction || rule.Protocol != flow.Protocol {
				continue
			}
			if rule.DestinationPort != 0 && rule.DestinationPort != flow.DestinationPort {
				continue
			}
			if !matchesSource(rule.Source, ip, aliases) {
				continue
			}
			if rule.Action == "log" {
				r.LogRuleIDs = append(r.LogRuleIDs, rule.ID)
				continue // log is nonterminal; keep searching.
			}
			r.PolicyIntent = rule.Action
			r.MatchedRuleID = rule.ID
			r.Explanation = append(r.Explanation, "First matching terminal rule under declared priority order only.")
			break
		}
		if r.MatchedRuleID == "" && flow.Direction == "input" {
			r.PolicyIntent = p.Spec.DefaultInbound
			r.Explanation = append(r.Explanation, "No terminal rule matched; declared defaultInbound applies in policy model.")
		}
		if r.MatchedRuleID == "" && flow.Direction == "output" {
			r.Explanation = append(r.Explanation, "No terminal rule: output default is undefined in current PolicyBundle schema.")
		}
		if flow.ConnectionState != "new" {
			r.Explanation = append(r.Explanation, "Established/unknown connection state is not modeled; conntrack can affect actual verdict.")
		}
		if flow.ManagementCritical {
			switch r.PolicyIntent {
			case "drop", "reject":
				r.ManagementRisk = "POTENTIAL_LOCKOUT"
				r.Explanation = append(r.Explanation, "Synthetic management-critical flow would be denied under declared intent; preserve independent rescue access.")
			default:
				r.ManagementRisk = "UNVERIFIED_REACHABILITY"
				r.Explanation = append(r.Explanation, "No proof that management access works: an accept or rescue-CIDR declaration does not override other kernel chains.")
			}
		} else {
			r.ManagementRisk = "NOT_APPLICABLE"
		}
		out.Cases = append(out.Cases, r)
	}
	return out, nil
}
func matchesSource(expr string, ip netip.Addr, aliases map[string][]netip.Prefix) bool {
	if expr == "any" {
		return true
	}
	if list, ok := aliases[expr]; ok {
		for _, pr := range list {
			if pr.Contains(ip) {
				return true
			}
		}
		return false
	}
	pr, err := netip.ParsePrefix(expr)
	return err == nil && pr.Contains(ip)
}
