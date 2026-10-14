package observability

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func sample() Snapshot {
	return Snapshot{
		Schema: Schema, Origin: Origin, Authority: Authority,
		BanStateCounts: []BanStateCount{
			{Family: "ipv4", State: "proposed", Count: 7},
			{Family: "ipv6", State: "unknown", Count: 2},
		},
		NotificationQueues: []NotificationQueueCount{
			{Channel: "telegram", Count: 5},
			{Channel: "discord", Count: 0},
		},
	}
}

func TestSyntheticRendersDeterministicBoundedPrometheus(t *testing.T) {
	s := sample()
	first, err := RenderSynthetic(s)
	if err != nil {
		t.Fatal(err)
	}
	s.BanStateCounts[0], s.BanStateCounts[1] = s.BanStateCounts[1], s.BanStateCounts[0]
	s.NotificationQueues[0], s.NotificationQueues[1] = s.NotificationQueues[1], s.NotificationQueues[0]
	second, err := RenderSynthetic(s)
	if err != nil || first != second {
		t.Fatal("non-deterministic rendered snapshot", err)
	}
	for _, expected := range []string{
		"# TYPE lightluna_p0_synthetic_ban_state_count gauge",
		"lightluna_p0_synthetic_ban_state_count{family=\"ipv4\",state=\"proposed\"} 7",
		"lightluna_p0_synthetic_ban_state_count{family=\"ipv6\",state=\"unknown\"} 2",
		"lightluna_p0_synthetic_notification_queue_depth{channel=\"telegram\"} 5",
		"lightluna_p0_synthetic_contract_info{authority=\"none\"} 1",
	} {
		if !strings.Contains(first, expected) {
			t.Errorf("missing %q", expected)
		}
	}
	for _, dangerous := range []string{"lightluna_bans_active", "applied_verified", "192.0.2.8", "token", "webhook"} {
		if strings.Contains(first, dangerous) {
			t.Errorf("exported forbidden name %q", dangerous)
		}
	}
	if len(first) > 4096 {
		t.Fatal("snapshot unexpectedly large")
	}
}

func TestSyntheticRejectsUnauthorizedOrHighCardinalityValues(t *testing.T) {
	cases := []struct {
		name   string
		modify func(*Snapshot)
	}{
		{"wrong schema", func(s *Snapshot) { s.Schema = "other" }},
		{"invalid origin", func(s *Snapshot) { s.Origin = "LIVE_KERNEL" }},
		{"forged authority", func(s *Snapshot) { s.Authority = "OWNER_APPROVED" }},
		{"source IP instead family", func(s *Snapshot) { s.BanStateCounts[0].Family = "192.0.2.8" }},
		{"CIDR instead family", func(s *Snapshot) { s.BanStateCounts[0].Family = "198.51.100.0/24" }},
		{"proposed auto promoted", func(s *Snapshot) { s.BanStateCounts[0].State = "applied_verified" }},
		{"event id instead state", func(s *Snapshot) { s.BanStateCounts[0].State = "event-1234" }},
		{"freeform channel", func(s *Snapshot) { s.NotificationQueues[0].Channel = "mailto:alice@example.test" }},
		{"webhook URL", func(s *Snapshot) { s.NotificationQueues[0].Channel = "https://hooks.slack.com/test" }},
		{"unbounded count", func(s *Snapshot) { s.BanStateCounts[0].Count = MaxCount + 1 }},
		{"unbounded queue", func(s *Snapshot) { s.NotificationQueues[0].Count = MaxCount + 1 }},
		{"duplicate ban series", func(s *Snapshot) { s.BanStateCounts = append(s.BanStateCounts, s.BanStateCounts[0]) }},
		{"duplicate provider", func(s *Snapshot) { s.NotificationQueues = append(s.NotificationQueues, s.NotificationQueues[0]) }},
		{"too many rows", func(s *Snapshot) {
			for i := 0; i < MaxRows+1; i++ {
				s.BanStateCounts = append(s.BanStateCounts, BanStateCount{})
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := sample()
			tc.modify(&s)
			if _, err := RenderSynthetic(s); err == nil {
				t.Fatal("accepted forbidden synthetic data")
			}
		})
	}
}

func TestStrictJSONBoundaryRejectsSensitiveUnknownOrDuplicateInputs(t *testing.T) {
	good := fmt.Sprintf(`{"schema":%q,"origin":%q,"authority":%q,"banStateCounts":[{"family":"ipv4","state":"proposed","count":3}],"notificationQueues":[]}`, Schema, Origin, Authority)
	s, err := Decode(strings.NewReader(good))
	if err != nil || len(s.BanStateCounts) != 1 {
		t.Fatalf("failed valid bounded contract: %v", err)
	}
	for _, extra := range []string{
		`"sourceIp":"192.0.2.8"`, `"eventId":"abc"`, `"webhook":"https://secrets.invalid"`,
		`"smtpPassword":"secret"`, `"url":"https://metadata.google.internal"`, `"ownerApproved":true`,
	} {
		payload := strings.TrimSuffix(good, "}") + "," + extra + "}"
		if _, err := Decode(strings.NewReader(payload)); err == nil {
			t.Errorf("accepted extra field %q", extra)
		}
	}
	for _, payload := range []string{
		strings.Replace(good, `"count":3`, `"count":3,"count":4`, 1),
		good + good,
		good + " trailing",
		strings.Repeat("A", MaxInputBytes+1),
		strings.Replace(good, `"count":3`, `"count":-1`, 1),
		strings.Replace(good, `"state":"proposed"`, `"state":"applied_verified"`, 1),
	} {
		if _, err := Decode(strings.NewReader(payload)); err == nil {
			t.Fatal("accepted malformed/unsafe synthetic input")
		}
	}
}

func TestRepositorySyntheticFixtureDoesNotClaimAppliedBan(t *testing.T) {
	file, err := os.Open("../../testdata/local-observability-synthetic.sample.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	s, err := Decode(file)
	if err != nil {
		t.Fatal(err)
	}
	out, err := RenderSynthetic(s)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "family=\"ipv4\",state=\"proposed\"} 7") {
		t.Fatal("fixture metrics mismatch")
	}
	for _, name := range []string{"lightluna_bans_active", "lightluna_ban_transitions_total", "applied_verified"} {
		if strings.Contains(out, name) {
			t.Fatalf("fixture claimed real policy metric %q", name)
		}
	}
}

func TestNoSideEffects(t *testing.T) {
	s := sample()
	if _, err := RenderSynthetic(s); err != nil {
		t.Fatal(err)
	}
	// Merely importing/rendering snapshots must not require a network
	// interface, http.Server, token, queue worker, nft binary, root or sudo.
	// The package depends only on Go stdlib plus existing strictjson.
}
