package observability

// This file provides a pure, offline synthetic security-event and outbound
// notification-plan contract. Plans are data: they never send, write to disk,
// connect to a queue or grant an owner/firewall capability.

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"time"

	"github.com/n8n2erpnext/light-luna/internal/strictjson"
)

const EventSchema = "lightluna.synthetic-security-events/v1alpha1"
const MaxEventBytes = 32 << 10
const MaxSyntheticEvents = 64
const MaxSyntheticPlans = 128

type SyntheticEvent struct {
	EventID          string `json:"eventId"`
	Kind             string `json:"kind"`
	Severity         string `json:"severity"`
	Reason           string `json:"reason"`
	OccurredUTC      string `json:"occurredUtc"`
	Count            uint64 `json:"count"`
	EnforcementState string `json:"enforcementState"`
	OwnerApproved    bool   `json:"ownerApproved"`
	NotificationSent bool   `json:"notificationSent"`
}

type SyntheticEventBatch struct {
	Schema    string           `json:"schema"`
	Origin    string           `json:"origin"`
	Authority string           `json:"authority"`
	Events    []SyntheticEvent `json:"events"`
}

type SyntheticRoute struct {
	Channel string `json:"channel"`
	Enabled bool   `json:"enabled"`
}

type SyntheticNotice struct {
	EventID string `json:"eventId"`
	Channel string `json:"channel"`
	State   string `json:"state"`
}

// SyntheticPlan deliberately contains NO webhook, recipient, token, delivery
// attempt, HTTP/SMTP transport, live ban receipt, or network address.
type SyntheticPlan struct {
	Schema          string            `json:"schema"`
	Origin          string            `json:"origin"`
	Authority       string            `json:"authority"`
	DeliveryAllowed bool              `json:"deliveryAllowed"`
	Notices         []SyntheticNotice `json:"notices"`
}

var syntheticEventID = regexp.MustCompile(`^evt_[a-z0-9]{12,32}$`)
var kinds = []string{"DETECTION", "BAN_PROPOSED", "BAN_REJECTED", "GUARDIAN_UNKNOWN", "SYSTEM_DEGRADED"}
var severities = []string{"INFO", "WARN", "HIGH", "CRITICAL"}
var reasons = []string{"AUTH_FAILURE_BURST", "EGRESS_FANOUT", "COVERAGE_GAP", "POLICY_DENIED", "HEALTH_UNAVAILABLE"}

// DecodeEventBatch rejects unknown fields, duplicated nested JSON keys,
// oversized payloads and any real enforcement/notification authority.
func DecodeEventBatch(r io.Reader) (SyntheticEventBatch, error) {
	var batch SyntheticEventBatch
	if err := strictjson.Decode(r, MaxEventBytes, &batch); err != nil {
		return SyntheticEventBatch{}, err
	}
	if err := batch.Validate(); err != nil {
		return SyntheticEventBatch{}, err
	}
	return batch, nil
}

func (b SyntheticEventBatch) Validate() error {
	if b.Schema != EventSchema || b.Origin != Origin || b.Authority != Authority {
		return errors.New("synthetic security-event authority mismatch")
	}
	if len(b.Events) < 1 || len(b.Events) > MaxSyntheticEvents {
		return errors.New("invalid bounded event batch size")
	}
	seen := make(map[string]bool, len(b.Events))
	for _, event := range b.Events {
		if !syntheticEventID.MatchString(event.EventID) {
			return errors.New("event id not an opaque synthetic identifier")
		}
		if seen[event.EventID] {
			return errors.New("duplicated event id")
		}
		seen[event.EventID] = true
		if !allowed(event.Kind, kinds) || !allowed(event.Severity, severities) || !allowed(event.Reason, reasons) {
			return fmt.Errorf("event kind, reason or severity outside bounded allowlist")
		}
		occurred, err := time.Parse(time.RFC3339, event.OccurredUTC)
		if err != nil || occurred.IsZero() || !occurred.Equal(occurred.UTC()) || event.OccurredUTC[len(event.OccurredUTC)-1] != 'Z' {
			return errors.New("event time must use RFC3339 UTC Z")
		}
		if event.Count == 0 || event.Count > MaxCount {
			return errors.New("event count outside bounded range")
		}
		if event.EnforcementState != "NOT_ENFORCED_SYNTHETIC" ||
			event.OwnerApproved || event.NotificationSent {
			return errors.New("event attempts to claim real ban or outbound delivery")
		}
	}
	return nil
}

// PlanSyntheticNotifications only computes a deterministic, bounded list of
// hypothetical notices. Caller-selected channels are never credentials or
// evidence of owner permission to send. No transport implementation exists.
func PlanSyntheticNotifications(batch SyntheticEventBatch, routes []SyntheticRoute) (SyntheticPlan, error) {
	if err := batch.Validate(); err != nil {
		return SyntheticPlan{}, err
	}
	if len(routes) > len(channels) {
		return SyntheticPlan{}, errors.New("too many synthetic routes")
	}
	enabled := map[string]bool{}
	for _, r := range routes {
		if !allowed(r.Channel, channels) || !r.Enabled {
			return SyntheticPlan{}, errors.New("route must be known and explicitly enabled for simulation")
		}
		if enabled[r.Channel] {
			return SyntheticPlan{}, errors.New("duplicate synthetic route")
		}
		enabled[r.Channel] = true
	}
	events := append([]SyntheticEvent(nil), batch.Events...)
	sort.Slice(events, func(i, j int) bool { return events[i].EventID < events[j].EventID })
	plan := SyntheticPlan{Schema: EventSchema, Origin: Origin,
		Authority: Authority, DeliveryAllowed: false,
		Notices: make([]SyntheticNotice, 0, len(events)*len(enabled))}
	for _, event := range events {
		for _, channel := range channels {
			if enabled[channel] {
				if len(plan.Notices) >= MaxSyntheticPlans {
					return SyntheticPlan{}, errors.New("too many simulated notification plans")
				}
				plan.Notices = append(plan.Notices, SyntheticNotice{
					EventID: event.EventID, Channel: channel, State: "SIMULATED_NOT_SENT"})
			}
		}
	}
	return plan, nil
}

// SyntheticQueueSnapshot makes an intentionally untrustworthy metric snapshot
// of a plan, never an applied-ban or actual delivery-status gauge.
func SyntheticQueueSnapshot(plan SyntheticPlan) (Snapshot, error) {
	if plan.Schema != EventSchema || plan.Origin != Origin ||
		plan.Authority != Authority || plan.DeliveryAllowed ||
		len(plan.Notices) > MaxSyntheticPlans {
		return Snapshot{}, errors.New("invalid synthetic notification plan")
	}
	seen := make(map[string]bool, len(plan.Notices))
	counts := make(map[string]uint64, len(channels))
	for _, notice := range plan.Notices {
		if !syntheticEventID.MatchString(notice.EventID) ||
			!allowed(notice.Channel, channels) || notice.State != "SIMULATED_NOT_SENT" {
			return Snapshot{}, errors.New("not a synthetic undelivered notice")
		}
		key := notice.EventID + "/" + notice.Channel
		if seen[key] {
			return Snapshot{}, errors.New("duplicate planned notice")
		}
		seen[key] = true
		counts[notice.Channel]++
	}
	snapshot := Snapshot{Schema: Schema, Origin: Origin, Authority: Authority}
	for _, channel := range channels {
		snapshot.NotificationQueues = append(snapshot.NotificationQueues,
			NotificationQueueCount{Channel: channel, Count: counts[channel]})
	}
	if err := snapshot.Validate(); err != nil {
		return Snapshot{}, err
	}
	return snapshot, nil
}
