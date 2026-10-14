// Package nftpreview converts validated P0 policies into syntactically
// nftables-like, deliberately UNHOOKED preview chains. It never invokes nft,
// acquires privileges, or accesses the kernel. Its output is NOT an apply plan.
package nftpreview

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"sort"
	"strings"

	"github.com/n8n2erpnext/light-luna/internal/policy"
)

const MaxExpandedRules = 8192
const MaxOutputBytes = 1 << 20

// Plan contains code-generation evidence, NOT an executable firewall policy.
// The script has NO base-chain hooks, priorities or default drop policy.
type Plan struct {
	Schema                     string   `json:"schema"`
	PolicyName                 string   `json:"policyName"`
	Revision                   int64    `json:"revision"`
	PolicyHashSHA256           string   `json:"policyHashSha256"`
	Status                     string   `json:"status"`
	Effect                     string   `json:"effect"`
	KernelWrites               bool     `json:"kernelWrites"`
	SyntaxChecked              bool     `json:"syntaxChecked"`
	KernelCompatibilityChecked bool     `json:"kernelCompatibilityChecked"`
	ExpandedRules              int      `json:"expandedRules"`
	UnhookedScript             string   `json:"unhookedScript"`
	Warnings                   []string `json:"warnings"`
}

// Compile creates isolated regular chains only: without a base-chain hook, a
// chain will never receive packets on its own. No live policy is constructed.
func Compile(p policy.Bundle) (Plan, error) {
	if err := p.Validate(); err != nil {
		return Plan{}, err
	}
	preview, err := policy.MakePreview(p)
	if err != nil {
		return Plan{}, err
	}
	aliases := make(map[string][]string, len(p.Spec.Aliases))
	for _, a := range p.Spec.Aliases {
		if a.Type != "cidr-set" {
			continue
		} // port-set use is not supported by this P0 schema.
		values := make([]string, 0, len(a.Values))
		for _, raw := range a.Values {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return Plan{}, err
			}
			values = append(values, s)
		}
		sort.Strings(values)
		aliases[a.ID] = values
	}
	rules := append([]policy.Rule(nil), p.Spec.Rules...)
	sort.Slice(rules, func(i, j int) bool { return rules[i].Priority < rules[j].Priority })
	count := 0
	var b bytes.Buffer
	b.WriteString("# Light Luna P0 preview ONLY - this is NOT a firewall apply transaction.\n")
	b.WriteString("# Chains have NO netfilter hooks and NO default chain policies.\n")
	b.WriteString("# Rescue CIDRs are declarations only; reachability is not verified.\n")
	b.WriteString("table inet luna_p0_preview {\n")
	for _, dir := range []string{"input", "output"} {
		chain := "preview_inbound"
		if dir == "output" {
			chain = "preview_outbound"
		}
		fmt.Fprintf(&b, "  chain %s {\n", chain)
		for _, r := range rules {
			if r.Direction != dir {
				continue
			}
			sources := []string{r.Source}
			if expanded, ok := aliases[r.Source]; ok {
				sources = expanded
			}
			for _, source := range sources {
				if count >= MaxExpandedRules {
					return Plan{}, fmt.Errorf("preview expansion exceeds %d rules", MaxExpandedRules)
				}
				clause, err := render(r, source)
				if err != nil {
					return Plan{}, fmt.Errorf("rule %s: %w", r.ID, err)
				}
				// Safe fixed comment; never interpolate user-supplied notes in nft text.
				fmt.Fprintf(&b, "    %s comment \"luna-p0-%s\"\n", clause, r.ID)
				count++
				if b.Len() > MaxOutputBytes {
					return Plan{}, errors.New("nft preview exceeds byte budget")
				}
			}
		}
		b.WriteString("  }\n")
	}
	b.WriteString("}\n")
	if b.Len() > MaxOutputBytes {
		return Plan{}, errors.New("nft preview exceeds byte budget")
	}
	return Plan{
		Schema:     "lightluna.nft-preview/v1alpha1",
		PolicyName: p.Metadata.Name, Revision: p.Metadata.Revision, PolicyHashSHA256: preview.HashSHA256,
		Status: "UNHOOKED_TEXT_PREVIEW_ONLY", Effect: "none: regular chains without hooks or base policies",
		KernelWrites: false, SyntaxChecked: false, KernelCompatibilityChecked: false, ExpandedRules: count, UnhookedScript: b.String(),
		Warnings: []string{
			"Do not apply this text as a firewall policy: defaultInbound and emergencyAccess are NOT compiled.",
			"No nft kernel capability, existing ruleset coexistence, Docker/UFW, or admin reachability checks are implied.",
			"No executor, stage, commit, rollback or production enforcement exists in P0.",
		},
	}, nil
}

func render(r policy.Rule, source string) (string, error) {
	var parts []string
	if source != "any" {
		p, err := netip.ParsePrefix(source)
		if err != nil {
			return "", err
		}
		if p.Masked() != p {
			return "", errors.New("noncanonical source CIDR")
		}
		if r.Protocol == "icmp" && !p.Addr().Is4() {
			return "", errors.New("IPv6 source with IPv4 ICMP")
		}
		if r.Protocol == "icmpv6" && !p.Addr().Is6() {
			return "", errors.New("IPv4 source with ICMPv6")
		}
		if p.Addr().Is4() {
			parts = append(parts, "ip saddr "+p.String())
		} else {
			parts = append(parts, "ip6 saddr "+p.String())
		}
	}
	switch r.Protocol {
	case "tcp", "udp":
		if r.DestinationPort != 0 {
			parts = append(parts, fmt.Sprintf("%s dport %d", r.Protocol, r.DestinationPort))
		} else {
			parts = append(parts, "meta l4proto "+r.Protocol)
		}
	case "icmp":
		parts = append(parts, "meta nfproto ipv4", "meta l4proto icmp")
	case "icmpv6":
		parts = append(parts, "meta nfproto ipv6", "meta l4proto ipv6-icmp")
	default:
		return "", errors.New("unsupported protocol")
	}
	switch r.Action {
	case "accept", "drop", "reject", "log":
		parts = append(parts, r.Action)
	default:
		return "", errors.New("unsupported action")
	}
	return strings.Join(parts, " "), nil
}
