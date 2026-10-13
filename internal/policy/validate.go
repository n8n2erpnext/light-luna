package policy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/netip"
	"regexp"
	"strings"
)

var nameRE = regexp.MustCompile("^[a-z][a-z0-9-]{0,62}[a-z0-9]$")
var uuidRE = regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")

func bytesReader(data []byte) *bytes.Reader { return bytes.NewReader(data) }

func prefix(s string) error {
	p, err := netip.ParsePrefix(s)
	if err != nil {
		return fmt.Errorf("invalid CIDR %q: %w", s, err)
	}
	if p.Masked() != p {
		return fmt.Errorf("CIDR %q must be network-masked", s)
	}
	if p.Addr().Is4In6() {
		return fmt.Errorf("IPv4-mapped IPv6 is unsupported")
	}
	return nil
}

func (b Bundle) Validate() error {
	if b.APIVersion != "lightluna.io/v1alpha1" || b.Kind != "PolicyBundle" {
		return fmt.Errorf("unsupported apiVersion or kind")
	}
	if len(b.Metadata.Name) < 2 || len(b.Metadata.Name) > 64 || !nameRE.MatchString(b.Metadata.Name) {
		return fmt.Errorf("metadata.name must be 2-64 lowercase slug characters")
	}
	if b.Metadata.Revision < 1 {
		return fmt.Errorf("metadata.revision must be >= 1")
	}
	if b.Spec.Mode != "host" && b.Spec.Mode != "edge" {
		return fmt.Errorf("only standalone host/edge are supported")
	}
	switch b.Spec.DefaultInbound {
	case "accept", "drop", "reject":
	default:
		return fmt.Errorf("invalid defaultInbound")
	}
	if b.Spec.DefaultInbound != "accept" && (!b.Spec.EmergencyAccess.Enabled || len(b.Spec.EmergencyAccess.Networks) == 0) {
		return fmt.Errorf("drop/reject default requires explicit emergencyAccess networks, even for preview")
	}
	if len(b.Spec.EmergencyAccess.Networks) > 32 {
		return fmt.Errorf("too many emergency networks")
	}
	for i, n := range b.Spec.EmergencyAccess.Networks {
		if err := prefix(n); err != nil {
			return fmt.Errorf("emergencyAccess.networks[%d]: %w", i, err)
		}
	}
	if len(b.Spec.Aliases) > 256 || len(b.Spec.Rules) > 2048 {
		return fmt.Errorf("policy size limits exceeded")
	}
	aliases := make(map[string]string, len(b.Spec.Aliases))
	for i, a := range b.Spec.Aliases {
		if len(a.ID) < 2 || !nameRE.MatchString(a.ID) {
			return fmt.Errorf("aliases[%d]: invalid id", i)
		}
		if _, exists := aliases[a.ID]; exists {
			return fmt.Errorf("aliases[%d]: duplicate id %q", i, a.ID)
		}
		if a.Type != "cidr-set" && a.Type != "port-set" {
			return fmt.Errorf("aliases[%d]: unsupported type", i)
		}
		if len(a.Values) == 0 || len(a.Values) > 256 {
			return fmt.Errorf("aliases[%d]: values length invalid", i)
		}
		seen := map[string]bool{}
		for j, raw := range a.Values {
			var canonical string
			if a.Type == "cidr-set" {
				var value string
				if err := json.Unmarshal(raw, &value); err != nil || value == "" {
					return fmt.Errorf("aliases[%d].values[%d]: expected CIDR string", i, j)
				}
				if err := prefix(value); err != nil {
					return fmt.Errorf("aliases[%d].values[%d]: %w", i, j, err)
				}
				canonical = value
			} else {
				var value int
				if err := json.Unmarshal(raw, &value); err != nil || value < 1 || value > 65535 {
					return fmt.Errorf("aliases[%d].values[%d]: expected TCP/UDP port 1..65535", i, j)
				}
				canonical = fmt.Sprintf("%d", value)
			}
			if seen[canonical] {
				return fmt.Errorf("aliases[%d]: duplicate value %s", i, canonical)
			}
			seen[canonical] = true
		}
		aliases[a.ID] = a.Type
	}
	ids := map[string]bool{}
	priorities := map[int]bool{}
	for i, r := range b.Spec.Rules {
		if !uuidRE.MatchString(r.ID) || ids[r.ID] {
			return fmt.Errorf("rules[%d]: invalid or duplicate UUID", i)
		}
		ids[r.ID] = true
		if r.Priority < 1 || r.Priority > 65535 || priorities[r.Priority] {
			return fmt.Errorf("rules[%d]: invalid or duplicate priority", i)
		}
		priorities[r.Priority] = true
		if r.Direction != "input" && r.Direction != "output" {
			return fmt.Errorf("rules[%d]: only host/edge input/output supported in P0; forward is later Gateway", i)
		}
		switch r.Protocol {
		case "tcp", "udp", "icmp", "icmpv6":
		default:
			return fmt.Errorf("rules[%d]: unsupported protocol", i)
		}
		switch r.Action {
		case "accept", "drop", "reject", "log":
		default:
			return fmt.Errorf("rules[%d]: unsupported action", i)
		}
		if r.Source == "" {
			return fmt.Errorf("rules[%d]: source required (CIDR or cidr-set)", i)
		}
		if kind, exists := aliases[r.Source]; exists {
			if kind != "cidr-set" {
				return fmt.Errorf("rules[%d]: source alias must be cidr-set", i)
			}
		} else if r.Source != "any" {
			if err := prefix(r.Source); err != nil {
				return fmt.Errorf("rules[%d]: source %q is neither CIDR nor cidr-set: %w", i, r.Source, err)
			}
		}
		if r.DestinationPort != 0 {
			if r.DestinationPort < 1 || r.DestinationPort > 65535 || (r.Protocol != "tcp" && r.Protocol != "udp") {
				return fmt.Errorf("rules[%d]: destinationPort requires tcp/udp and port 1..65535", i)
			}
		}
		if len(r.Note) > 512 || strings.ContainsRune(r.Note, '\x00') {
			return fmt.Errorf("rules[%d]: note invalid", i)
		}
	}
	switch b.Spec.Autonomous.Mode {
	case "off", "observe", "shadow":
	default:
		return fmt.Errorf("P0 Luna mode must be off, observe or shadow")
	}
	if b.Spec.Autonomous.AllowRealBans {
		return fmt.Errorf("P0 forbids autonomous real bans")
	}
	if b.Spec.Autonomous.MaxBanTTLSeconds < 0 || b.Spec.Autonomous.MaxBanTTLSeconds > 3600 {
		return fmt.Errorf("invalid maximum ban TTL")
	}
	return nil
}
