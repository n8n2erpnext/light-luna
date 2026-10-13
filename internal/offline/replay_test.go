package offline

import (
	"github.com/n8n2erpnext/light-luna/internal/foundation"
	"github.com/n8n2erpnext/light-luna/internal/gate"
	"github.com/n8n2erpnext/light-luna/internal/guardian"
	"testing"
)

func sample(id, coverage, trust string, count, windows int) foundation.Observation {
	return foundation.Observation{
		Schema:  foundation.ObservationSchema,
		EventID: id, SiteID: "example-site", Category: "AUTH_FAILURE_BURST",
		OccurredUTC: "2026-10-14T00:00:00Z", Coverage: coverage, SourceTrust: trust,
		Count: count, Windows: windows,
	}
}
func TestSyntheticReplayOnly(t *testing.T) {
	in := foundation.ReplayInput{Schema: "lightluna.replay/v1alpha1", Observations: []foundation.Observation{
		sample("verified-example", "COMPLETE", "VERIFIED_INGRESS", 12, 3),
		sample("incomplete-example", "PARTIAL", "VERIFIED_INGRESS", 40, 5),
		sample("unverified-example", "COMPLETE", "UNVERIFIED", 50, 5),
		sample("benign-example", "COMPLETE", "VERIFIED_INGRESS", 2, 1),
	}}
	out, err := Replay(in)
	if err != nil {
		t.Fatal(err)
	}
	if out.Mode != "SYNTHETIC_OFFLINE_NON_ENFORCING" || len(out.Results) != 4 {
		t.Fatal("unsafe or missing replay")
	}
	wants := []string{"SUSPICIOUS_PATTERN", "INSUFFICIENT_TELEMETRY", "INSUFFICIENT_TELEMETRY", "OBSERVE"}
	for i, r := range out.Results {
		if r.Assessment.Status != wants[i] || r.Assessment.CanEnforce || r.Assessment.ProposedAction != "WATCH_ONLY" {
			t.Fatalf("wrong assessment %d: %+v", i, r.Assessment)
		}
		if r.Gate.Decision != "ADVISORY_ONLY" || r.Gate.KernelWrites {
			t.Fatalf("unsafe gate %d: %+v", i, r.Gate)
		}
	}
}
func TestRejectUnscopedAndExecutableProposals(t *testing.T) {
	a := foundation.Assessment{Schema: "lightluna.assessment/v1alpha1", SiteID: "example-site", EventID: "evt-123", ProposedAction: "BAN_IP", CanEnforce: true}
	if got := gate.Preview(a); got.Decision != "REJECTED" || got.KernelWrites {
		t.Fatalf("allowed ban: %+v", got)
	}
	a.ProposedAction = "WATCH_ONLY"
	a.CanEnforce = false
	a.SiteID = ""
	if got := gate.Preview(a); got.Decision != "REJECTED" {
		t.Fatalf("allowed unscoped: %+v", got)
	}
}
func TestRejectMalformedAndDuplicate(t *testing.T) {
	cases := []foundation.ReplayInput{
		{Schema: "wrong", Observations: []foundation.Observation{sample("event-111", "COMPLETE", "VERIFIED_INGRESS", 9, 2)}},
		{Schema: "lightluna.replay/v1alpha1", Observations: []foundation.Observation{sample("event-111", "COMPLETE", "VERIFIED_INGRESS", 9, 2), sample("event-111", "COMPLETE", "VERIFIED_INGRESS", 9, 2)}},
		{Schema: "lightluna.replay/v1alpha1", Observations: []foundation.Observation{sample("event-111", "COMPLETE", "VERIFIED_INGRESS", 2, 50)}},
		{Schema: "lightluna.replay/v1alpha1", Observations: []foundation.Observation{sample("event-111", "COMPLETE", "FAKED", 10, 2)}},
	}
	for _, in := range cases {
		if _, err := Replay(in); err == nil {
			t.Fatalf("accepted invalid replay %+v", in)
		}
	}
}
func TestMissingComponentsAreNotClaimedHealthy(t *testing.T) {
	s := Snapshot()
	if s.FirewallEnforcement != "NOT_IMPLEMENTED" || s.ExecutorVersion != "NOT_IMPLEMENTED" || s.HostIntegrity != "UNKNOWN_NO_SENSORS" || s.CloudState != "NEVER_ENROLLED" {
		t.Fatalf("misleading snapshot: %+v", s)
	}
	g := guardian.Snapshot()
	if g.Status != "UNKNOWN" || g.FileScanner != "NOT_INSTALLED" || g.EBPF != "NOT_INSTALLED" {
		t.Fatalf("misleading guardian: %+v", g)
	}
}
