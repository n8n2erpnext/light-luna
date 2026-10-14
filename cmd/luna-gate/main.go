//go:build linux

// luna-gate is an explicitly launched nonroot P0 development transport.
// It cannot call nftables, open privileged sockets or execute firewall changes.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/n8n2erpnext/light-luna/internal/controlgate"
	"github.com/n8n2erpnext/light-luna/internal/devkeys"
	"github.com/n8n2erpnext/light-luna/internal/localipc"
	"github.com/n8n2erpnext/light-luna/internal/policy"
)

func main() {
	if err := execute(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
func execute(args []string) error {
	if len(args) == 2 && args[0] == "init" {
		return devkeys.Init(args[1])
	}
	if len(args) != 4 || args[0] != "serve" {
		return errors.New("usage: luna-gate init <new-private-dir> | serve <policy.json> <private-dir> <max-seconds>")
	}
	seconds, err := strconv.Atoi(args[3])
	if err != nil || seconds < 1 || seconds > 30 {
		return errors.New("P0 serve lease must be integer 1..30 seconds")
	}
	dir := args[2]
	key, err := devkeys.Load(dir)
	if err != nil {
		return fmt.Errorf("read demo key: %w", err)
	}
	defer func() {
		for i := range key {
			key[i] = 0
		}
	}()
	f, err := os.Open(args[1])
	if err != nil {
		return err
	}
	p, err := policy.Decode(f)
	f.Close()
	if err != nil {
		return err
	}
	gate, journal, err := controlgate.NewBoundaryWithJournal(key, "synthetic-site", "synthetic-agent", filepath.Join(dir, "replay.jsonl"), time.Now().UTC())
	if err != nil {
		return fmt.Errorf("open replay journal: %w", err)
	}
	defer journal.Close()
	listener, _, err := localipc.ListenPrivate(dir)
	if err != nil {
		return fmt.Errorf("open local socket: %w", err)
	}
	defer listener.Close()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	ctx, expire := context.WithTimeout(ctx, time.Duration(seconds)*time.Second)
	defer expire()
	return localipc.Serve(ctx, listener, uint32(os.Geteuid()), gate, p, time.Now)
}
