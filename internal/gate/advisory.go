// Package gate contains a non-enforcing preview of an independent approval
// boundary. It never invokes a privileged executor or a kernel interface.
package gate

import "github.com/n8n2erpnext/light-luna/internal/foundation"

func Preview(a foundation.Assessment) foundation.GateResult {
	if a.Schema != "lightluna.assessment/v1alpha1" || a.SiteID == "" || a.EventID == "" {
		return foundation.GateResult{Decision: "REJECTED", Reason: "Invalid or unscoped proposal.", KernelWrites: false}
	}
	if a.CanEnforce || a.ProposedAction != "WATCH_ONLY" {
		return foundation.GateResult{Decision: "REJECTED", Reason: "P0 rejects any enforcement request.", KernelWrites: false}
	}
	return foundation.GateResult{Decision: "ADVISORY_ONLY", Reason: "No executor exists. Explicit local approval and independent rollback would be required for enforcement.", KernelWrites: false}
}
