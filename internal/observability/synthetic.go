// Package observability contains pure synthetic, non-enforcing P0 observation
// contracts. This package is NOT a Prometheus server or actual ban exporter.
// Nothing here reads network state, opens sockets, or controls firewall policy.
package observability

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/n8n2erpnext/light-luna/internal/strictjson"
)

const (
	Schema        = "lightluna.synthetic-observability/v1alpha1"
	Origin        = "P0_SYNTHETIC_TEST_ONLY"
	Authority     = "NO_KERNEL_OR_NOTIFICATION_AUTHORITY"
	MaxInputBytes = 16 << 10
	MaxRows       = 10
	MaxCount      = 1000000
)

type BanStateCount struct {
	Family string `json:"family"`
	State  string `json:"state"`
	Count  uint64 `json:"count"`
}

type NotificationQueueCount struct {
	Channel string `json:"channel"`
	Count   uint64 `json:"count"`
}

// Snapshot deliberately has no ban IP, CIDR, URL, recipient, event ID, owner
// approval, real executor status, credential or arbitrary label field.
type Snapshot struct {
	Schema             string                   `json:"schema"`
	Origin             string                   `json:"origin"`
	Authority          string                   `json:"authority"`
	BanStateCounts     []BanStateCount          `json:"banStateCounts"`
	NotificationQueues []NotificationQueueCount `json:"notificationQueues"`
}

var families = []string{"ipv4", "ipv6"}
var states = []string{"proposed", "unknown", "rejected"}
var channels = []string{"telegram", "discord", "slack", "email"}

func allowed(v string, domain []string) bool {
	for _, item := range domain {
		if v == item {
			return true
		}
	}
	return false
}

// Decode uses existing bounded strict JSON and refuses arbitrary properties,
// repeated JSON object keys, extra JSON values and overlong messages.
func Decode(r io.Reader) (Snapshot, error) {
	var s Snapshot
	if err := strictjson.Decode(r, MaxInputBytes, &s); err != nil {
		return Snapshot{}, err
	}
	if err := s.Validate(); err != nil {
		return Snapshot{}, err
	}
	return s, nil
}

func (s Snapshot) Validate() error {
	if s.Schema != Schema || s.Origin != Origin || s.Authority != Authority {
		return errors.New("synthetic contract identity/authority mismatch")
	}
	if len(s.BanStateCounts) > len(families)*len(states) ||
		len(s.NotificationQueues) > len(channels) ||
		len(s.BanStateCounts)+len(s.NotificationQueues) > MaxRows {
		return errors.New("unbounded synthetic cardinality")
	}
	seen := make(map[string]bool, MaxRows)
	for _, row := range s.BanStateCounts {
		if !allowed(row.Family, families) || !allowed(row.State, states) {
			return errors.New("ban state/family not in synthetic allowlist")
		}
		if row.Count > MaxCount {
			return errors.New("ban aggregate count exceeds P0 limit")
		}
		key := row.Family + "/" + row.State
		if seen[key] {
			return fmt.Errorf("duplicate synthetic ban aggregate %s", key)
		}
		seen[key] = true
	}
	seen = make(map[string]bool, len(channels))
	for _, row := range s.NotificationQueues {
		if !allowed(row.Channel, channels) {
			return errors.New("notification channel not in allowlist")
		}
		if row.Count > MaxCount {
			return errors.New("notification queue count exceeds P0 limit")
		}
		if seen[row.Channel] {
			return errors.New("duplicate notification channel")
		}
		seen[row.Channel] = true
	}
	return nil
}

// RenderSynthetic returns Prometheus text only for explicitly synthetic
// aggregate fixtures. It NEVER emits the future real lightluna_bans_active
// metric and never accepts arbitrary label keys or label values.
func RenderSynthetic(s Snapshot) (string, error) {
	if err := s.Validate(); err != nil {
		return "", err
	}
	banCounts := make(map[string]uint64, len(s.BanStateCounts))
	for _, item := range s.BanStateCounts {
		banCounts[item.Family+"/"+item.State] = item.Count
	}
	queueCounts := make(map[string]uint64, len(s.NotificationQueues))
	for _, item := range s.NotificationQueues {
		queueCounts[item.Channel] = item.Count
	}
	var b strings.Builder
	b.Grow(1800)
	b.WriteString("# HELP lightluna_p0_synthetic_contract_info Synthetic P0 contract, NOT live firewall or notification state.\n")
	b.WriteString("# TYPE lightluna_p0_synthetic_contract_info gauge\n")
	b.WriteString("lightluna_p0_synthetic_contract_info{authority=\"none\"} 1\n")
	b.WriteString("# HELP lightluna_p0_synthetic_ban_state_count Inert synthetic samples; NEVER verified real bans.\n")
	b.WriteString("# TYPE lightluna_p0_synthetic_ban_state_count gauge\n")
	for _, family := range families {
		for _, state := range states {
			b.WriteString("lightluna_p0_synthetic_ban_state_count{family=\"")
			b.WriteString(family)
			b.WriteString("\",state=\"")
			b.WriteString(state)
			b.WriteString("\"} ")
			b.WriteString(strconv.FormatUint(banCounts[family+"/"+state], 10))
			b.WriteByte('\n')
		}
	}
	b.WriteString("# HELP lightluna_p0_synthetic_notification_queue_depth Synthetic queue fixtures; NO notification transport exists.\n")
	b.WriteString("# TYPE lightluna_p0_synthetic_notification_queue_depth gauge\n")
	for _, channel := range channels {
		b.WriteString("lightluna_p0_synthetic_notification_queue_depth{channel=\"")
		b.WriteString(channel)
		b.WriteString("\"} ")
		b.WriteString(strconv.FormatUint(queueCounts[channel], 10))
		b.WriteByte('\n')
	}
	return b.String(), nil
}
