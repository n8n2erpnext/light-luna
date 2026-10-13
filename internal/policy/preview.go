package policy

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

type Preview struct {
	Schema         string   `json:"schema"`
	PolicyName     string   `json:"policyName"`
	Revision       int64    `json:"revision"`
	HashSHA256     string   `json:"hashSha256"`
	Mode           string   `json:"mode"`
	DefaultInbound string   `json:"defaultInbound"`
	Effects        string   `json:"effects"`
	Enforceable    bool     `json:"enforceable"`
	Rules          []Rule   `json:"rules"`
	Warnings       []string `json:"warnings"`
}

// MakePreview is an inert representation; NEVER interpret it as an nft script.
// No kernel operation, write to privileged IPC, or actual flow simulation exists.
func MakePreview(b Bundle) (Preview, error) {
	if err := b.Validate(); err != nil {
		return Preview{}, err
	}
	data, err := json.Marshal(b)
	if err != nil {
		return Preview{}, err
	}
	hash := sha256.Sum256(data)
	rules := append([]Rule{}, b.Spec.Rules...)
	sort.Slice(rules, func(i, j int) bool { return rules[i].Priority < rules[j].Priority })
	return Preview{
		Schema:         "lightluna.preview/v1alpha1",
		PolicyName:     b.Metadata.Name,
		Revision:       b.Metadata.Revision,
		HashSHA256:     hex.EncodeToString(hash[:]),
		Mode:           b.Spec.Mode,
		DefaultInbound: b.Spec.DefaultInbound,
		Effects:        "none: validation/preview only",
		Enforceable:    false,
		Rules:          rules,
		Warnings: []string{
			"No nftables syntax or kernel coexistence validation has been performed.",
			"Emergency networks are declarations, NOT independently verified working rescue paths.",
			"No staged/commit/rollback, simulation, NetBird connector or live ban actions are implemented.",
		},
	}, nil
}
