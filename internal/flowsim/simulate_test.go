package flowsim

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/n8n2erpnext/light-luna/internal/policy"
)

const fixturePolicy = `{
"apiVersion":"lightluna.io/v1alpha1","kind":"PolicyBundle",
"metadata":{"name":"test-firewall","revision":8},
"spec":{"mode":"host","defaultInbound":"drop",
"emergencyAccess":{"enabled":true,"networks":["192.0.2.0/24"]},
"aliases":[{"id":"admin-subnets","type":"cidr-set","values":["192.0.2.0/24","2001:db8::/32"]}],
"rules":[
{"id":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","priority":100,"direction":"input","source":"admin-subnets","protocol":"tcp","destinationPort":22,"action":"log"},
{"id":"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb","priority":110,"direction":"input","source":"admin-subnets","protocol":"tcp","destinationPort":22,"action":"accept"},
{"id":"cccccccc-cccc-cccc-cccc-cccccccccccc","priority":200,"direction":"input","source":"any","protocol":"tcp","destinationPort":443,"action":"accept"},
{"id":"dddddddd-dddd-dddd-dddd-dddddddddddd","priority":210,"direction":"input","source":"any","protocol":"udp","destinationPort":53,"action":"reject"},
{"id":"eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee","priority":120,"direction":"output","source":"any","protocol":"tcp","destinationPort":8443,"action":"drop"}
],
"autonomous":{"mode":"shadow","allowRealBans":false,"maxBanTtlSeconds":300}
}}`

func policyFixture(t *testing.T) policy.Bundle {
	t.Helper()
	p, err := policy.Decode(strings.NewReader(fixturePolicy))
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func flow(id, dir, source, proto string, port int, critical bool) Flow {
	return Flow{CaseID: id, Direction: dir, SourceIP: source, Protocol: proto, DestinationPort: port, ConnectionState: "new", ManagementCritical: critical}
}
func TestDeterministicIntentAndManagementWarnings(t *testing.T) {
	p := policyFixture(t)
	f := Fixture{Schema: InputSchema, Cases: []Flow{
		flow("admin-v4", "input", "192.0.2.14", "tcp", 22, true),
		flow("admin-v6", "input", "2001:db8::5", "tcp", 22, true),
		flow("admin-denied", "input", "198.51.100.7", "tcp", 22, true),
		flow("web", "input", "203.0.113.8", "tcp", 443, false),
		flow("dns-reject", "input", "198.51.100.8", "udp", 53, false),
		flow("outbound-deny", "output", "192.0.2.14", "tcp", 8443, false),
		flow("outbound-unknown", "output", "192.0.2.14", "tcp", 443, false),
	}}
	report, err := Simulate(p, f)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Simulate(p, f)
	if err != nil {
		t.Fatal(err)
	}
	aJSON, _ := json.Marshal(report)
	bJSON, _ := json.Marshal(b)
	if string(aJSON) != string(bJSON) {
		t.Fatal("nondeterministic output")
	}
	if report.KernelWrites || len(report.Cases) != 7 || report.Mode != "SYNTHETIC_POLICY_INTENT_ONLY" {
		t.Fatal("unsafe report")
	}
	want := []string{"accept", "accept", "drop", "accept", "reject", "drop", "UNSPECIFIED"}
	for i, r := range report.Cases {
		if r.PolicyIntent != want[i] || r.EffectiveVerdict != "UNKNOWN" || r.ExternalFirewall != "UNKNOWN" {
			t.Errorf("bad intent %d: %+v", i, r)
		}
	}
	if len(report.Cases[0].LogRuleIDs) != 1 || report.Cases[0].MatchedRuleID != "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb" {
		t.Fatal("log action improperly terminal or alias mismatch")
	}
	if report.Cases[0].ManagementRisk != "UNVERIFIED_REACHABILITY" || report.Cases[2].ManagementRisk != "POTENTIAL_LOCKOUT" {
		t.Fatal("management risk warning missing")
	}
	if report.Cases[2].MatchedRuleID != "" {
		t.Fatal("default denied must not forge a matched rule ID")
	}
}
func TestUntrustedInputFailsClosed(t *testing.T) {
	p := policyFixture(t)
	good := flow("test-a", "input", "192.0.2.14", "tcp", 22, true)
	cases := []struct {
		name   string
		change func(*Flow)
	}{
		{"bad-direction", func(x *Flow) { x.Direction = "forward" }},
		{"mapped-IP", func(x *Flow) { x.SourceIP = "::ffff:192.0.2.14" }},
		{"missing-IP", func(x *Flow) { x.SourceIP = "" }},
		{"multicast", func(x *Flow) { x.SourceIP = "224.0.0.1" }},
		{"port-overflow", func(x *Flow) { x.DestinationPort = 65536 }},
		{"icmp-port", func(x *Flow) { x.Protocol = "icmp"; x.DestinationPort = 22 }},
		{"wrong-icmp-family", func(x *Flow) { x.Protocol = "icmpv6"; x.DestinationPort = 0 }},
		{"unknown-state", func(x *Flow) { x.ConnectionState = "pending" }},
		{"invalid-case-id", func(x *Flow) { x.CaseID = "  strange" }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bad := good
			c.change(&bad)
			if _, err := Simulate(p, Fixture{Schema: InputSchema, Cases: []Flow{bad}}); err == nil {
				t.Fatalf("accepted %s", c.name)
			}
		})
	}
	duplicated := Fixture{Schema: InputSchema, Cases: []Flow{good, good}}
	if _, err := Simulate(p, duplicated); err == nil {
		t.Fatal("accepted duplicate case IDs")
	}
	many := make([]Flow, MaxCases+1)
	for i := range many {
		many[i] = good
	}
	if _, err := Simulate(p, Fixture{Schema: InputSchema, Cases: many}); err == nil {
		t.Fatal("accepted oversized input")
	}
	if _, err := Simulate(p, Fixture{Schema: "wrong", Cases: []Flow{good}}); err == nil {
		t.Fatal("accepted wrong schema")
	}
}
func TestLogWithoutTerminalDoesNotBecomeDenial(t *testing.T) {
	p := policyFixture(t)
	p.Spec.Rules = p.Spec.Rules[:1] // log-only rule
	f := Fixture{Schema: InputSchema, Cases: []Flow{flow("logged", "input", "192.0.2.14", "tcp", 22, true)}}
	r, err := Simulate(p, f)
	if err != nil {
		t.Fatal(err)
	}
	if r.Cases[0].PolicyIntent != "drop" || len(r.Cases[0].LogRuleIDs) != 1 || r.Cases[0].ManagementRisk != "POTENTIAL_LOCKOUT" {
		t.Fatalf("log should not terminate: %+v", r.Cases[0])
	}
}
