// Package policy provides the non-privileged, non-enforcing policy foundation.
// No function in this package executes nft, iptables, subprocesses, or firewall changes.
package policy

import (
	"encoding/json"
	"io"

	"github.com/n8n2erpnext/light-luna/internal/strictjson"
)

const MaxPolicyBytes = 1 << 20

type Bundle struct {
	APIVersion string   `json:"apiVersion"`
	Kind       string   `json:"kind"`
	Metadata   Metadata `json:"metadata"`
	Spec       Spec     `json:"spec"`
}
type Metadata struct {
	Name     string `json:"name"`
	Revision int64  `json:"revision"`
}
type Spec struct {
	Mode            string          `json:"mode"`
	DefaultInbound  string          `json:"defaultInbound"`
	EmergencyAccess EmergencyAccess `json:"emergencyAccess"`
	Aliases         []Alias         `json:"aliases"`
	Rules           []Rule          `json:"rules"`
	Autonomous      Autonomous      `json:"autonomous"`
}
type EmergencyAccess struct {
	Enabled  bool     `json:"enabled"`
	Networks []string `json:"networks"`
}
type Alias struct {
	ID     string            `json:"id"`
	Type   string            `json:"type"`
	Values []json.RawMessage `json:"values"`
}
type Rule struct {
	ID              string `json:"id"`
	Priority        int    `json:"priority"`
	Direction       string `json:"direction"`
	Source          string `json:"source"`
	Protocol        string `json:"protocol"`
	DestinationPort int    `json:"destinationPort,omitempty"`
	Action          string `json:"action"`
	Note            string `json:"note,omitempty"`
}
type Autonomous struct {
	Mode             string `json:"mode"`
	AllowRealBans    bool   `json:"allowRealBans"`
	MaxBanTTLSeconds int    `json:"maxBanTtlSeconds"`
}

// Decode enforces bounded input, unique keys at every JSON depth, known fields
// and policy semantics. No kernel/netfilter operation is performed.
func Decode(r io.Reader) (Bundle, error) {
	var b Bundle
	if err := strictjson.DecodeRejectNull(r, MaxPolicyBytes, &b); err != nil {
		return b, err
	}
	if err := b.Validate(); err != nil {
		return b, err
	}
	return b, nil
}
