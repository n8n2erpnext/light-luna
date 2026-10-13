// lunactl P0 is intentionally unprivileged and cannot change kernel rules.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/n8n2erpnext/light-luna/internal/policy"
)

const version = "0.0.0-p0.dev"

func usage() {
	fmt.Fprintln(os.Stderr, "usage: lunactl version | firewall policies validate <file> | firewall policies preview <file>")
}
func run(args []string) error {
	if len(args) == 1 && args[0] == "version" {
		fmt.Println("lunactl", version, "(standalone, non-enforcing)")
		return nil
	}
	if len(args) != 4 || args[0] != "firewall" || args[1] != "policies" || (args[2] != "validate" && args[2] != "preview") {
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
	plan, err := policy.MakePreview(p)
	if err != nil {
		return err
	}
	result, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(result))
	return nil
}
func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
