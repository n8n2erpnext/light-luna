// Package foundation declares inert, bounded P0 simulation contracts.
// It does not collect traffic, inspect a host, or authorize enforcement.
package foundation

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

const ObservationSchema = "lightluna.observation/v1alpha1"
const MaxReplayEvents = 128
const MaxReplayBytes = 256 << 10

var idPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._:-]{2,127}$`)

type Observation struct {
	Schema      string `json:"schema"`
	EventID     string `json:"eventId"`
	SiteID      string `json:"siteId"`
	Category    string `json:"category"`
	OccurredUTC string `json:"occurredUtc"`
	Coverage    string `json:"coverage"`
	SourceTrust string `json:"sourceTrust"`
	Count       int    `json:"count"`
	Windows     int    `json:"windows"`
}
type ReplayInput struct {
	Schema       string        `json:"schema"`
	Observations []Observation `json:"observations"`
}

func (in ReplayInput) Validate() error {
	if in.Schema != "lightluna.replay/v1alpha1" {
		return errors.New("unsupported replay schema")
	}
	if len(in.Observations) < 1 || len(in.Observations) > MaxReplayEvents {
		return errors.New("invalid replay event count")
	}
	ids := map[string]bool{}
	for i, e := range in.Observations {
		if err := e.Validate(); err != nil {
			return fmt.Errorf("observations[%d]: %w", i, err)
		}
		if ids[e.EventID] {
			return fmt.Errorf("duplicate eventId %q", e.EventID)
		}
		ids[e.EventID] = true
	}
	return nil
}
func (o Observation) Validate() error {
	if o.Schema != ObservationSchema {
		return errors.New("unsupported observation schema")
	}
	if !idPattern.MatchString(o.EventID) || !idPattern.MatchString(o.SiteID) {
		return errors.New("invalid eventId/siteId")
	}
	if o.Category != "AUTH_FAILURE_BURST" && o.Category != "EGRESS_FANOUT" {
		return errors.New("unsupported observation category")
	}
	if o.Coverage != "COMPLETE" && o.Coverage != "PARTIAL" && o.Coverage != "MISSING" {
		return errors.New("invalid telemetry coverage")
	}
	if o.SourceTrust != "VERIFIED_INGRESS" && o.SourceTrust != "UNVERIFIED" {
		return errors.New("invalid source provenance")
	}
	t, err := time.Parse(time.RFC3339, o.OccurredUTC)
	if err != nil || t.IsZero() {
		return errors.New("occurredUtc must be RFC3339")
	}
	if o.Count < 0 || o.Count > 100000 || o.Windows < 0 || o.Windows > 144 || o.Windows > o.Count {
		return errors.New("invalid bounded event counts")
	}
	return nil
}

type Assessment struct {
	Schema         string   `json:"schema"`
	SiteID         string   `json:"siteId"`
	EventID        string   `json:"eventId"`
	Scenario       string   `json:"scenario"`
	Status         string   `json:"status"`
	Reasons        []string `json:"reasons"`
	ProposedAction string   `json:"proposedAction"`
	CanEnforce     bool     `json:"canEnforce"`
}

type GateResult struct {
	Decision     string `json:"decision"`
	KernelWrites bool   `json:"kernelWrites"`
	Reason       string `json:"reason"`
}

type ObservationResult struct {
	Assessment Assessment `json:"assessment"`
	Gate       GateResult `json:"gate"`
}
type ReplayReport struct {
	Schema     string              `json:"schema"`
	Mode       string              `json:"mode"`
	CloudState string              `json:"cloudState"`
	Results    []ObservationResult `json:"results"`
}

type ModuleStatus struct {
	Module  string `json:"module"`
	Version string `json:"version"`
	State   string `json:"state"`
}
type LocalSnapshot struct {
	Schema              string         `json:"schema"`
	Stage               string         `json:"stage"`
	AppVersion          string         `json:"appVersion"`
	BrainVersion        string         `json:"brainVersion"`
	ExecutorVersion     string         `json:"executorVersion"`
	KnowledgeRevision   string         `json:"knowledgeRevision"`
	CloudState          string         `json:"cloudState"`
	FirewallEnforcement string         `json:"firewallEnforcement"`
	HostIntegrity       string         `json:"hostIntegrity"`
	Modules             []ModuleStatus `json:"modules"`
}
