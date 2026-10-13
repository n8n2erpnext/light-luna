// Package brain contains the P0 offline deterministic *synthetic observation*
// evaluator. It makes no firewall, process or Cloud changes.
package brain

import (
	"fmt"
	"github.com/n8n2erpnext/light-luna/internal/foundation"
)

const Version = "0.0.0-brain.p0"

// Evaluate accepts normalized, externally provided replay observations only.
// The thresholds are demonstrative, NOT trained risk scores or production detection rules.
func Evaluate(o foundation.Observation) (foundation.Assessment, error) {
	if err := o.Validate(); err != nil {
		return foundation.Assessment{}, err
	}
	a := foundation.Assessment{
		Schema: "lightluna.assessment/v1alpha1",
		SiteID: o.SiteID, EventID: o.EventID,
		Scenario:       "basic." + o.Category + ".demo.v1",
		Status:         "OBSERVE",
		Reasons:        []string{"Synthetic replay only; no live telemetry or calibrated model."},
		ProposedAction: "WATCH_ONLY", CanEnforce: false,
	}
	if o.Coverage != "COMPLETE" || o.SourceTrust != "VERIFIED_INGRESS" {
		a.Status = "INSUFFICIENT_TELEMETRY"
		a.Reasons = append(a.Reasons, "Unverified provenance or incomplete telemetry; no benign or malicious conclusion.")
		return a, nil
	}
	switch o.Category {
	case "AUTH_FAILURE_BURST":
		if o.Count >= 8 && o.Windows >= 2 {
			a.Status = "SUSPICIOUS_PATTERN"
			a.Reasons = append(a.Reasons, fmt.Sprintf("Replay counts (%d) span %d windows; candidate only.", o.Count, o.Windows))
		} else {
			a.Reasons = append(a.Reasons, "Insufficient independent windows or events for demo pattern.")
		}
	case "EGRESS_FANOUT":
		a.Reasons = append(a.Reasons, "Egress pattern needs process provenance/benign alternatives; observe only.")
	}
	return a, nil
}
