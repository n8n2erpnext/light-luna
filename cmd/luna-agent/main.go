//go:build linux

// luna-agent is a one-shot P0 client. The OWNER role string in demo envelopes
// is NOT a real authenticated human identity. No live node registration exists.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"encoding/json"
	"github.com/n8n2erpnext/light-luna/internal/controlgate"
	"github.com/n8n2erpnext/light-luna/internal/devkeys"
	"github.com/n8n2erpnext/light-luna/internal/localipc"
	"github.com/n8n2erpnext/light-luna/internal/policy"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
func randomID(prefix string) (string, error) {
	data := make([]byte, 16)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(data), nil
}
func run(args []string) error {
	if len(args) > 0 && (args[0] == "preview-ed25519" || args[0] == "apply-ed25519") {
		return runEd25519(args)
	}
	if len(args) != 3 && len(args) != 5 {
		return errors.New("usage: luna-agent <preview|apply> <policy-file> <private-demo-dir> [request-id nonce]")
	}
	if args[0] != "preview" && args[0] != "apply" {
		return errors.New("only synthetic preview/apply-denial operations supported")
	}
	key, err := devkeys.Load(args[2])
	if err != nil {
		return fmt.Errorf("open shared demo key: %w", err)
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
	pv, err := policy.MakePreview(p)
	if err != nil {
		return err
	}
	id, nonce := "", ""
	if len(args) == 5 {
		id = args[3]
		nonce = args[4]
	} else {
		id, err = randomID("synthetic-request-")
		if err != nil {
			return err
		}
		nonce, err = randomID("synthetic-nonce-")
		if err != nil {
			return err
		}
	}
	action := "PREVIEW_POLICY"
	if args[0] == "apply" {
		action = "APPLY_POLICY"
	}
	now := time.Now().UTC()
	signed, err := controlgate.Sign(key, controlgate.Intent{
		Protocol: controlgate.Protocol, RequestID: id, Nonce: nonce, SiteID: "synthetic-site", ActorID: "synthetic-agent",
		ActorRole: "OWNER", Origin: "LOCAL_AGENT", PolicyName: p.Metadata.Name, ExpectedRevision: p.Metadata.Revision,
		PolicyHashSHA256: pv.HashSHA256, IssuedUTC: now.Add(-time.Second), ExpiresUTC: now.Add(time.Minute), Operation: action,
	})
	if err != nil {
		return err
	}
	receipt, err := localipc.Call(context.Background(), filepath.Join(args[2], "gate.sock"), signed)
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	// A denied explicit operation is an expected test outcome. JSON decision is
	// the source of truth, rather than a pretend applied-operation exit status.
	return nil
}
