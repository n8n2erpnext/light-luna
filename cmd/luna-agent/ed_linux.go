//go:build linux

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/n8n2erpnext/light-luna/internal/controlgate"
	"github.com/n8n2erpnext/light-luna/internal/identitykeys"
	"github.com/n8n2erpnext/light-luna/internal/localipc"
	"github.com/n8n2erpnext/light-luna/internal/policy"
)

func runEd25519(args []string) error {
	if len(args) != 5 && len(args) != 7 {
		return errors.New("usage: luna-agent preview-ed25519|apply-ed25519 <policy> <agent-private-dir> <group-socket-path> <trusted-Gate-UID> [request-id nonce]")
	}
	expected, err := strconv.ParseUint(args[4], 10, 32)
	if err != nil {
		return errors.New("invalid expected Gate UID")
	}
	key, err := identitykeys.LoadPrivate(args[2])
	if err != nil {
		return err
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
	var id, nonce string
	if len(args) == 7 {
		id, nonce = args[5], args[6]
	} else {
		id, err = randomID("synthetic-ed-request-")
		if err != nil {
			return err
		}
		nonce, err = randomID("synthetic-ed-nonce-")
		if err != nil {
			return err
		}
	}
	op := "PREVIEW_POLICY"
	if args[0] == "apply-ed25519" {
		op = "APPLY_POLICY"
	}
	now := time.Now().UTC()
	msg, err := controlgate.SignEd25519(key, controlgate.Intent{
		Protocol: controlgate.Protocol, RequestID: id, Nonce: nonce,
		SiteID: "synthetic-site", ActorID: "synthetic-agent",
		ActorRole: "OWNER", Origin: "LOCAL_AGENT",
		PolicyName: p.Metadata.Name, ExpectedRevision: p.Metadata.Revision,
		PolicyHashSHA256: pv.HashSHA256, IssuedUTC: now.Add(-time.Second), ExpiresUTC: now.Add(time.Minute), Operation: op,
	})
	if err != nil {
		return err
	}
	receipt, err := localipc.CallExpectedGateUID(context.Background(), args[3], msg, uint32(expected))
	if err != nil {
		return err
	}
	result, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(result))
	return nil
}
