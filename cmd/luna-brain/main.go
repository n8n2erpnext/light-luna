// luna-brain is a separate, offline-only P0 Brain executable.
// It cannot inspect a host, access the network, update a model, or enforce rules.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/n8n2erpnext/light-luna/internal/brain"
	"github.com/n8n2erpnext/light-luna/internal/foundation"
	"github.com/n8n2erpnext/light-luna/internal/offline"
	"github.com/n8n2erpnext/light-luna/internal/strictjson"
)

func run(args []string) error {
	if len(args) == 1 && args[0] == "version" {
		fmt.Println("luna-brain", brain.Version, "(offline synthetic replay only)")
		return nil
	}
	if len(args) != 2 || args[0] != "replay" {
		return fmt.Errorf("usage: luna-brain version | replay <synthetic-file>")
	}
	f, err := os.Open(args[1])
	if err != nil {
		return err
	}
	defer f.Close()
	var input foundation.ReplayInput
	if err := strictjson.Decode(f, foundation.MaxReplayBytes, &input); err != nil {
		return err
	}
	report, err := offline.Replay(input)
	if err != nil {
		return err
	}
	output, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}
func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
