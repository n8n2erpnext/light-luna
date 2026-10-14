package observability

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func demoEvents() SyntheticEventBatch {
	return SyntheticEventBatch{
		Schema: EventSchema, Origin: Origin, Authority: Authority,
		Events: []SyntheticEvent{
			{EventID: "evt_aaaabbbbcccc", Kind: "BAN_PROPOSED", Severity: "WARN",
				Reason: "AUTH_FAILURE_BURST", OccurredUTC: "2026-10-14T10:00:00Z",
				Count: 3, EnforcementState: "NOT_ENFORCED_SYNTHETIC"},
			{EventID: "evt_ddddeeeeffff", Kind: "GUARDIAN_UNKNOWN", Severity: "HIGH",
				Reason: "HEALTH_UNAVAILABLE", OccurredUTC: "2026-10-14T11:00:00Z",
				Count: 1, EnforcementState: "NOT_ENFORCED_SYNTHETIC"},
		},
	}
}

func TestOfflineSyntheticNotificationPlan(t *testing.T) {
	input := demoEvents()
	routes := []SyntheticRoute{{Channel: "slack", Enabled: true}, {Channel: "telegram", Enabled: true}}
	out, err := PlanSyntheticNotifications(input, routes)
	if err != nil {
		t.Fatal(err)
	}
	if out.DeliveryAllowed || out.Authority != Authority || len(out.Notices) != 4 {
		t.Fatalf("invalid plan: %+v", out)
	}
	if out.Notices[0].EventID != "evt_aaaabbbbcccc" ||
		out.Notices[0].Channel != "telegram" ||
		out.Notices[1].Channel != "slack" ||
		out.Notices[0].State != "SIMULATED_NOT_SENT" {
		t.Fatalf("nondeterministic plan %+v", out.Notices)
	}
	input.Events[0], input.Events[1] = input.Events[1], input.Events[0]
	routes[0], routes[1] = routes[1], routes[0]
	second, err := PlanSyntheticNotifications(input, routes)
	if err != nil || fmt.Sprint(second.Notices) != fmt.Sprint(out.Notices) {
		t.Fatalf("not deterministic: %v", err)
	}
	s, err := SyntheticQueueSnapshot(out)
	if err != nil {
		t.Fatal(err)
	}
	text, err := RenderSynthetic(s)
	if err != nil {
		t.Fatal(err)
	}
	for _, part := range []string{
		`lightluna_p0_synthetic_notification_queue_depth{channel="telegram"} 2`,
		`lightluna_p0_synthetic_notification_queue_depth{channel="slack"} 2`,
		`lightluna_p0_synthetic_notification_queue_depth{channel="email"} 0`,
	} {
		if !strings.Contains(text, part) {
			t.Errorf("missing %q", part)
		}
	}
	if strings.Contains(text, "lightluna_bans_active") || strings.Contains(text, "evt_aaaabbbbcccc") {
		t.Fatal("exposed event identity or forged applied ban")
	}
}

func TestRepositoryEventFixtureAndQueuePlan(t *testing.T) {
	file, err := os.Open("../../testdata/local-security-events-synthetic.sample.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	batch, err := DecodeEventBatch(file)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := PlanSyntheticNotifications(batch, []SyntheticRoute{{Channel: "telegram", Enabled: true}, {Channel: "email", Enabled: true}})
	if err != nil || len(plan.Notices) != 4 || plan.DeliveryAllowed {
		t.Fatalf("unexpected fixture plan: %+v %v", plan, err)
	}
	aggregate, err := SyntheticQueueSnapshot(plan)
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := RenderSynthetic(aggregate)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, `lightluna_p0_synthetic_notification_queue_depth{channel="email"} 2`) {
		t.Fatal("fixture missing synthetic queue metric")
	}
	if strings.Contains(rendered, "evt_aaaabbbbcccc") || strings.Contains(rendered, "lightluna_bans_active") {
		t.Fatal("export leaked event identity or claimed real ban")
	}
}

func TestPlanBudgetHardLimit(t *testing.T) {
	batch := demoEvents()
	batch.Events = nil
	for i := 0; i < MaxSyntheticEvents; i++ {
		batch.Events = append(batch.Events, SyntheticEvent{
			EventID: fmt.Sprintf("evt_%012d", i),
			Kind:    "DETECTION", Severity: "INFO", Reason: "AUTH_FAILURE_BURST",
			OccurredUTC: "2026-10-14T10:00:00Z", Count: 1,
			EnforcementState: "NOT_ENFORCED_SYNTHETIC",
		})
	}
	routes := []SyntheticRoute{{Channel: "telegram", Enabled: true}, {Channel: "discord", Enabled: true}, {Channel: "slack", Enabled: true}, {Channel: "email", Enabled: true}}
	if _, err := PlanSyntheticNotifications(batch, routes); err == nil {
		t.Fatal("unbounded aggregate delivery plan accepted")
	}
}

func TestNoRouteNoNoticeAndNoSend(t *testing.T) {
	plan, err := PlanSyntheticNotifications(demoEvents(), nil)
	if err != nil || len(plan.Notices) != 0 || plan.DeliveryAllowed {
		t.Fatal("no route must not send", plan, err)
	}
	empty, err := SyntheticQueueSnapshot(plan)
	if err != nil || len(empty.BanStateCounts) != 0 || len(empty.NotificationQueues) != 4 {
		t.Fatalf("invalid empty queue aggregate: %v %+v", err, empty)
	}
}

func TestRejectUntrustedSyntheticEvents(t *testing.T) {
	cases := []struct {
		name   string
		change func(*SyntheticEventBatch)
	}{
		{"forged schema", func(b *SyntheticEventBatch) { b.Schema = "lightluna.live" }},
		{"claim real origin", func(b *SyntheticEventBatch) { b.Origin = "ENFORCED_KERNEL" }},
		{"claim owner auth", func(b *SyntheticEventBatch) { b.Authority = "GRANTED" }},
		{"empty events", func(b *SyntheticEventBatch) { b.Events = nil }},
		{"too many", func(b *SyntheticEventBatch) {
			for len(b.Events) <= MaxSyntheticEvents {
				e := b.Events[0]
				e.EventID = fmt.Sprintf("evt_%012d", len(b.Events))
				b.Events = append(b.Events, e)
			}
		}},
		{"duplicate id", func(b *SyntheticEventBatch) { b.Events[1].EventID = b.Events[0].EventID }},
		{"raw ip as id", func(b *SyntheticEventBatch) { b.Events[0].EventID = "203.0.113.17" }},
		{"raw URL id", func(b *SyntheticEventBatch) { b.Events[0].EventID = "https://attacker.invalid" }},
		{"ban actually applied", func(b *SyntheticEventBatch) { b.Events[0].Kind = "BAN_APPLIED" }},
		{"owner approved", func(b *SyntheticEventBatch) { b.Events[0].OwnerApproved = true }},
		{"delivery sent", func(b *SyntheticEventBatch) { b.Events[0].NotificationSent = true }},
		{"fake enforcement", func(b *SyntheticEventBatch) { b.Events[0].EnforcementState = "APPLIED_VERIFIED" }},
		{"arbitrary severity", func(b *SyntheticEventBatch) { b.Events[0].Severity = "event-1234" }},
		{"arbitrary reason", func(b *SyntheticEventBatch) { b.Events[0].Reason = "https://invalid" }},
		{"count zero", func(b *SyntheticEventBatch) { b.Events[0].Count = 0 }},
		{"count unbounded", func(b *SyntheticEventBatch) { b.Events[0].Count = MaxCount + 1 }},
		{"missing date", func(b *SyntheticEventBatch) { b.Events[0].OccurredUTC = "" }},
		{"not UTC", func(b *SyntheticEventBatch) { b.Events[0].OccurredUTC = "2026-10-14T17:00:00+07:00" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := demoEvents()
			tc.change(&b)
			if _, err := PlanSyntheticNotifications(b, nil); err == nil {
				t.Fatal("accepted event without offline trust")
			}
		})
	}
}

func TestStrictEventJSONRejectsCredentialsAndDuplicateFields(t *testing.T) {
	valid := fmt.Sprintf(`{"schema":%q,"origin":%q,"authority":%q,"events":[{"eventId":"evt_aaaabbbbcccc","kind":"DETECTION","severity":"INFO","reason":"AUTH_FAILURE_BURST","occurredUtc":"2026-10-14T10:00:00Z","count":1,"enforcementState":"NOT_ENFORCED_SYNTHETIC","ownerApproved":false,"notificationSent":false}]}`, EventSchema, Origin, Authority)
	b, err := DecodeEventBatch(strings.NewReader(valid))
	if err != nil || len(b.Events) != 1 {
		t.Fatalf("valid event refused: %v", err)
	}
	for _, fragment := range []string{
		`,"ip":"203.0.113.1"`, `,"webhookUrl":"https://hooks.slack.com/abc"`,
		`,"smtpPassword":"secret"`, `,"recipient":"alice@example.test"`,
		`,"payload":"private request body"`, `,"subject":"198.51.100.1"`,
	} {
		altered := strings.Replace(valid, `"notificationSent":false`, `"notificationSent":false`+fragment, 1)
		if _, err := DecodeEventBatch(strings.NewReader(altered)); err == nil {
			t.Fatalf("accepted secret/raw user field %s", fragment)
		}
	}
	for _, v := range []string{
		strings.Replace(valid, `"count":1`, `"count":1,"count":2`, 1),
		strings.Repeat(valid, 2),
		strings.Repeat("X", MaxEventBytes+1),
		strings.Replace(valid, `"eventId":"evt_aaaabbbbcccc"`, `"eventId":"203.0.113.5"`, 1),
	} {
		if _, err := DecodeEventBatch(strings.NewReader(v)); err == nil {
			t.Fatal("accepted ambiguous/malformed input")
		}
	}
}

func TestRoutesAndPlansCannotEnableExternalDelivery(t *testing.T) {
	invalidRoutes := [][]SyntheticRoute{
		{{Channel: "telegram", Enabled: false}},
		{{Channel: "https://discord.com/api/webhooks/X", Enabled: true}},
		{{Channel: "alice@example.test", Enabled: true}},
		{{Channel: "telegram", Enabled: true}, {Channel: "telegram", Enabled: true}},
		{{Channel: "telegram", Enabled: true}, {Channel: "discord", Enabled: true}, {Channel: "slack", Enabled: true}, {Channel: "email", Enabled: true}, {Channel: "other", Enabled: true}},
	}
	for _, routes := range invalidRoutes {
		if _, err := PlanSyntheticNotifications(demoEvents(), routes); err == nil {
			t.Fatalf("accepted unapproved route: %+v", routes)
		}
	}
	valid, err := PlanSyntheticNotifications(demoEvents(), []SyntheticRoute{{Channel: "email", Enabled: true}})
	if err != nil {
		t.Fatal(err)
	}
	invalid := valid
	invalid.DeliveryAllowed = true
	if _, err := SyntheticQueueSnapshot(invalid); err == nil {
		t.Fatal("accepted claimed delivery permission")
	}
	invalid = valid
	invalid.Notices = append([]SyntheticNotice(nil), valid.Notices...)
	invalid.Notices[0].State = "SENT"
	if _, err := SyntheticQueueSnapshot(invalid); err == nil {
		t.Fatal("accepted forged sent status")
	}
	invalid = valid
	invalid.Notices = append(append([]SyntheticNotice(nil), valid.Notices...), valid.Notices[0])
	if _, err := SyntheticQueueSnapshot(invalid); err == nil {
		t.Fatal("accepted duplicated notice")
	}
}
