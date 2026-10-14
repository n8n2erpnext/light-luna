// Package offline ties the initial local simulation components together.
package offline

import (
	"github.com/n8n2erpnext/light-luna/internal/brain"
	"github.com/n8n2erpnext/light-luna/internal/foundation"
	"github.com/n8n2erpnext/light-luna/internal/gate"
	"github.com/n8n2erpnext/light-luna/internal/guardian"
)

const AppVersion = "0.0.0-p0.dev"

func Snapshot() foundation.LocalSnapshot {
	health := guardian.Snapshot()
	return foundation.LocalSnapshot{
		Schema: "lightluna.local-snapshot/v1alpha1", Stage: "P0_NON_ENFORCING_SCAFFOLD",
		AppVersion: AppVersion, BrainVersion: brain.Version,
		ExecutorVersion: "NOT_IMPLEMENTED", KnowledgeRevision: "BASE_PACK_NOT_IMPLEMENTED",
		CloudState: "NEVER_ENROLLED", FirewallEnforcement: "NOT_IMPLEMENTED",
		HostIntegrity: health.Status + "_NO_SENSORS",
		Modules: []foundation.ModuleStatus{
			{Module: "policy-validator", Version: AppVersion, State: "PREVIEW_ONLY"},
			{Module: "nft-preview-compiler", Version: AppVersion, State: "UNHOOKED_TEXT_ONLY"},
			{Module: "flow-simulator", Version: AppVersion, State: "SYNTHETIC_INTENT_ONLY"},
			{Module: "local-control-gate", Version: AppVersion, State: "IN_PROCESS_EPHEMERAL_AUTH_TEST_ONLY"},
			{Module: "brain-basic", Version: brain.Version, State: "SYNTHETIC_REPLAY_ONLY"},
			{Module: "host-guardian", Version: "not-installed", State: "UNKNOWN_NO_SENSORS"},
			{Module: "ebpf-observer", Version: "not-installed", State: "NOT_INSTALLED"},
			{Module: "xdp-guard", Version: "not-installed", State: "NOT_INSTALLED"},
			{Module: "cloud-sync", Version: "not-installed", State: "NOT_INSTALLED"},
		},
	}
}

func Replay(in foundation.ReplayInput) (foundation.ReplayReport, error) {
	if err := in.Validate(); err != nil {
		return foundation.ReplayReport{}, err
	}
	out := foundation.ReplayReport{
		Schema: "lightluna.replay-report/v1alpha1", Mode: "SYNTHETIC_OFFLINE_NON_ENFORCING",
		CloudState: "NEVER_ENROLLED", Results: make([]foundation.ObservationResult, 0, len(in.Observations)),
	}
	for _, ob := range in.Observations {
		assessment, err := brain.Evaluate(ob)
		if err != nil {
			return foundation.ReplayReport{}, err
		}
		out.Results = append(out.Results, foundation.ObservationResult{Assessment: assessment, Gate: gate.Preview(assessment)})
	}
	return out, nil
}
