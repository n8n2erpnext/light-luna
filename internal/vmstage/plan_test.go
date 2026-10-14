package vmstage

import (
	"os"
	"strings"
	"testing"

	"github.com/n8n2erpnext/light-luna/internal/policy"
)

func fixture(t *testing.T) policy.Bundle {
	t.Helper()
	f, e := os.Open("../../testdata/host-baseline.json")
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	p, e := policy.Decode(f)
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func TestVMStageManifestInertAndDeterministic(t *testing.T) {
	p := fixture(t)
	a, e := Build(p, 5)
	if e != nil {
		t.Fatal(e)
	}
	b, e := Build(p, 5)
	if e != nil {
		t.Fatal(e)
	}
	if a.StageID != b.StageID || a.SourcePreviewHash != b.SourcePreviewHash {
		t.Fatal("not deterministic")
	}
	if a.KernelWrites || a.OwnerApproved || a.Mode != "SYNTHETIC_VM_ONLY_NO_EXECUTOR" || a.Table != IsolatedTable || a.TTLSeconds != 5 {
		t.Fatal(a)
	}
	if len(a.Phases) != 4 || a.Phases[0] != "STAGED" || a.Phases[3] != "ROLLED_BACK" {
		t.Fatal(a.Phases)
	}
	if !strings.HasPrefix(a.StageID, "vm-") {
		t.Fatal(a.StageID)
	}
}

// A synthetic stage ID must bind the requested monotonic rollback deadline.
// Reusing a policy with a different TTL must NOT alias a previous VM journal.
func TestVMStageIdentityBindsTTL(t *testing.T) {
	p := fixture(t)
	short, err := Build(p, 1)
	if err != nil {
		t.Fatal(err)
	}
	long, err := Build(p, MaxTTLSeconds)
	if err != nil {
		t.Fatal(err)
	}
	if short.StageID == long.StageID {
		t.Fatalf("different TTL intents collide: ttl=1 and ttl=%d share stage ID %q", MaxTTLSeconds, short.StageID)
	}
	if short.SourcePreviewHash != long.SourcePreviewHash {
		t.Fatal("policy content hash unexpectedly depends on TTL")
	}
	same, err := Build(p, 1)
	if err != nil || same.StageID != short.StageID {
		t.Fatalf("same policy and TTL did not reproduce stage ID: %+v, %v", same, err)
	}
	mutated := p
	mutated.Metadata.Revision++
	newRevision, err := Build(mutated, 1)
	if err != nil || newRevision.StageID == short.StageID {
		t.Fatalf("stage identity is not tied to policy revision: %+v, %v", newRevision, err)
	}
}

func TestVMStageTTLBoundsAndInvalidPolicy(t *testing.T) {
	p := fixture(t)
	for _, ttl := range []int{-2, 0, 11, 3600} {
		if _, e := Build(p, ttl); e == nil {
			t.Fatal("unsafe TTL accepted", ttl)
		}
	}
	if _, e := Build(p, 1); e != nil {
		t.Fatal(e)
	}
	invalid := p
	invalid.Metadata.Name = ""
	if _, e := Build(invalid, 5); e == nil {
		t.Fatal("invalid policy accepted")
	}
}
