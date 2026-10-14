// lunactl P0 is intentionally unprivileged and cannot change kernel rules.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/n8n2erpnext/light-luna/internal/controlgate"
	"github.com/n8n2erpnext/light-luna/internal/flowsim"
	"github.com/n8n2erpnext/light-luna/internal/foundation"
	"github.com/n8n2erpnext/light-luna/internal/localipc"
	"github.com/n8n2erpnext/light-luna/internal/nftpreview"
	"github.com/n8n2erpnext/light-luna/internal/offline"
	"github.com/n8n2erpnext/light-luna/internal/policy"
	"github.com/n8n2erpnext/light-luna/internal/strictjson"
	"github.com/n8n2erpnext/light-luna/internal/vmstage"
)

func usage() {
	fmt.Fprintln(os.Stderr, "usage: lunactl version | system snapshot | brain replay <synthetic-file> | firewall policies validate <file> | firewall policies preview <file> | firewall policies nft-preview <file> | firewall policies simulate <policy-file> <synthetic-flow-file> | control selftest <policy-file> | control ipc-selftest <policy-file> | vm stage-intent <policy-file> <ttl-seconds>")
}
func pretty(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
func run(args []string) error {
	if len(args) == 1 && args[0] == "version" {
		fmt.Println("lunactl", offline.AppVersion, "(standalone, non-enforcing)")
		return nil
	}
	if len(args) == 2 && args[0] == "system" && args[1] == "snapshot" {
		return pretty(offline.Snapshot())
	}
	if len(args) == 3 && args[0] == "brain" && args[1] == "replay" {
		f, err := os.Open(args[2])
		if err != nil {
			return fmt.Errorf("open replay: %w", err)
		}
		defer f.Close()
		var in foundation.ReplayInput
		if err := strictjson.Decode(f, foundation.MaxReplayBytes, &in); err != nil {
			return fmt.Errorf("invalid replay: %w", err)
		}
		report, err := offline.Replay(in)
		if err != nil {
			return fmt.Errorf("invalid replay: %w", err)
		}
		return pretty(report)
	}
	if len(args) == 3 && args[0] == "control" && args[1] == "ipc-selftest" {
		f, err := os.Open(args[2])
		if err != nil {
			return fmt.Errorf("open policy: %w", err)
		}
		defer f.Close()
		p, err := policy.Decode(f)
		if err != nil {
			return fmt.Errorf("invalid policy: %w", err)
		}
		report, err := localipc.SelfTest(p)
		if err != nil {
			return fmt.Errorf("Linux IPC selftest: %w", err)
		}
		return pretty(report)
	}
	if len(args) == 3 && args[0] == "control" && args[1] == "selftest" {
		f, err := os.Open(args[2])
		if err != nil {
			return fmt.Errorf("open policy: %w", err)
		}
		defer f.Close()
		p, err := policy.Decode(f)
		if err != nil {
			return fmt.Errorf("invalid policy: %w", err)
		}
		result, err := controlgate.SelfTest(p, time.Now().UTC())
		if err != nil {
			return fmt.Errorf("local control selftest: %w", err)
		}
		return pretty(result)
	}
	if len(args) == 4 && args[0] == "vm" && args[1] == "stage-intent" {
		f, err := os.Open(args[2])
		if err != nil {
			return fmt.Errorf("open policy: %w", err)
		}
		defer f.Close()
		p, err := policy.Decode(f)
		if err != nil {
			return fmt.Errorf("invalid policy: %w", err)
		}
		ttl, err := strconv.Atoi(args[3])
		if err != nil {
			return fmt.Errorf("invalid TTL: %w", err)
		}
		intent, err := vmstage.Build(p, ttl)
		if err != nil {
			return fmt.Errorf("VM stage intent: %w", err)
		}
		return pretty(intent)
	}
	if len(args) == 5 && args[0] == "firewall" && args[1] == "policies" && args[2] == "simulate" {
		f, err := os.Open(args[3])
		if err != nil {
			return fmt.Errorf("open policy: %w", err)
		}
		defer f.Close()
		p, err := policy.Decode(f)
		if err != nil {
			return fmt.Errorf("invalid policy: %w", err)
		}
		flowsFile, err := os.Open(args[4])
		if err != nil {
			return fmt.Errorf("open synthetic flow file: %w", err)
		}
		defer flowsFile.Close()
		var data flowsim.Fixture
		if err := strictjson.Decode(flowsFile, flowsim.MaxInputBytes, &data); err != nil {
			return fmt.Errorf("invalid flow fixture: %w", err)
		}
		report, err := flowsim.Simulate(p, data)
		if err != nil {
			return fmt.Errorf("flow simulation: %w", err)
		}
		return pretty(report)
	}
	if len(args) != 4 || args[0] != "firewall" || args[1] != "policies" || (args[2] != "validate" && args[2] != "preview" && args[2] != "nft-preview") {
		usage()
		return fmt.Errorf("unsupported command")
	}
	f, err := os.Open(args[3])
	if err != nil {
		return fmt.Errorf("open policy: %w", err)
	}
	defer f.Close()
	p, err := policy.Decode(f)
	if err != nil {
		return fmt.Errorf("invalid policy: %w", err)
	}
	if args[2] == "validate" {
		fmt.Printf("VALID: %s revision %d (no rules applied)\n", p.Metadata.Name, p.Metadata.Revision)
		return nil
	}
	if args[2] == "nft-preview" {
		compiled, err := nftpreview.Compile(p)
		if err != nil {
			return fmt.Errorf("nft preview: %w", err)
		}
		return pretty(compiled)
	}
	plan, err := policy.MakePreview(p)
	if err != nil {
		return err
	}
	return pretty(plan)
}
func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
