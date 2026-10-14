// Package vmstage defines a typed, INERT contract for a synthetic nft VM lab.
// It never shells out, opens a socket, installs nft rules or grants privileges.
// Synthetic VM hooks are authored and exercised by separate VM-only test code.
package vmstage

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

	"github.com/n8n2erpnext/light-luna/internal/nftpreview"
	"github.com/n8n2erpnext/light-luna/internal/policy"
)

const Schema = "lightluna.vm-stage-intent/v1alpha1"
const MaxTTLSeconds = 10
const IsolatedTable = "luna_p0_stage"

type Intent struct {
	Schema            string   `json:"schema"`
	Mode              string   `json:"mode"`
	PolicyName        string   `json:"policyName"`
	PolicyRevision    int64    `json:"policyRevision"`
	SourcePreviewHash string   `json:"sourcePreviewHash"`
	StageID           string   `json:"stageId"`
	Table             string   `json:"table"`
	TTLSeconds        int      `json:"ttlSeconds"`
	Phases            []string `json:"phases"`
	Effect            string   `json:"effect"`
	KernelWrites      bool     `json:"kernelWrites"`
	OwnerApproved     bool     `json:"ownerApproved"`
	Scope             string   `json:"scope"`
}

func Build(p policy.Bundle, ttlSeconds int) (Intent, error) {
	if ttlSeconds < 1 || ttlSeconds > MaxTTLSeconds {
		return Intent{}, errors.New("VM stage TTL outside 1..10s")
	}
	preview, err := nftpreview.Compile(p)
	if err != nil {
		return Intent{}, fmt.Errorf("invalid source preview: %w", err)
	}
	if preview.KernelWrites || preview.Status != "UNHOOKED_TEXT_PREVIEW_ONLY" || preview.SyntaxChecked || preview.KernelCompatibilityChecked {
		return Intent{}, errors.New("unsafe preview status")
	}
	// The stage identity must bind the intended TTL as well as policy/table.
	// It is deterministic for idempotent synthetic lab replay; it is NOT an
	// owner signature, durable lease, unique deployment ID or APPLY permit.
	digest := sha256.Sum256([]byte(Schema + "\x00" + preview.PolicyHashSHA256 + "\x00" + IsolatedTable + "\x00" + strconv.Itoa(ttlSeconds)))
	return Intent{
		Schema: Schema, Mode: "SYNTHETIC_VM_ONLY_NO_EXECUTOR",
		PolicyName: p.Metadata.Name, PolicyRevision: p.Metadata.Revision,
		SourcePreviewHash: preview.PolicyHashSHA256,
		StageID:           "vm-" + hex.EncodeToString(digest[:12]), Table: IsolatedTable,
		TTLSeconds:   ttlSeconds,
		Phases:       []string{"STAGED", "VALIDATED", "COMMITTED_SYNTHETIC", "ROLLED_BACK"},
		Effect:       "none: manifest only; synthetic VM-only hook tested separately",
		KernelWrites: false, OwnerApproved: false, Scope: "KVM_GUEST_PRIVATE_NETNS_ONLY",
	}, nil
}
