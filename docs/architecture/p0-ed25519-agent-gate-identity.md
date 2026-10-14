# P0 Agent/Gate Ed25519 Identity Boundary and Group-Owned Unix Socket

Status: **IMPLEMENTED in standalone Go development source; same-UID HOMESV test only; NOT a certified multi-user production service**
Date: 2026-10-14
Scope: asymmetric per-Agent request signature, pinned Gate public verifier, kernel Unix peer UID checks, group-owned socket ACL contract, no privileged executor
Earlier baselines: [two separate development processes and replay journal](p0-two-process-gate-and-journal.md), [Linux UDS](p0-linux-unix-ipc.md), [control gate](p0-local-control-gate.md)
Sources: [Ed25519 signer/verifier](../../internal/controlgate/control.go), [Ed journal constructor](../../internal/controlgate/journal_linux.go), [identity key file boundary](../../internal/identitykeys/ed25519_linux.go), [Unix group socket](../../internal/localipc/group_linux.go), [client server UID check](../../internal/localipc/unix_linux.go), [Gate entrypoint](../../cmd/luna-gate/main.go), [Agent entrypoints](../../cmd/luna-agent/main.go), [Agent Ed client](../../cmd/luna-agent/ed_linux.go), [integration test](../../scripts/test_p0_ed25519_identities.sh)
Security rules: [Safety](../07-SECURITY-SAFETY.md), [self-defense](../adr/ADR-028-self-defense-host-compromise.md), [project truth](../project-book/PROJECT_TRUTH_STATUS.json)
Commit: verify actual `feat/p0-standalone-foundation` HEAD; no release tag or production deployment asserted.

## Canonical Agent signature spelling — development source hardening (2026-10-14)

[`Boundary.Evaluate`](../../internal/controlgate/control.go) now refuses uppercase hexadecimal spellings of a valid Agent Ed25519 signature. It verifies canonical lowercase hex round-trip **before** the existing pinned-key signature check or replay recording. The [negative unit test](../../internal/controlgate/ed25519_test.go) was RED against previous code and GREEN after the change, checking that invalid aliases leave replay untouched and that a canonical signed preview works once. Existing Ed25519 two-process Linux restart/corruption tests PASS. This change does **not** authenticate a human owner, authorize APPLY, integrate offline reviewer proof, or add privileged executor/kernel operations.

## Purpose and flow

The previous HMAC prototype required **both** demo processes to read the **same secret**. This incremental contract eliminates that sharing for the new Ed25519 mode: the Agent signs with a private key, while Gate holds **only a pinned public key** and cannot produce valid Agent-signed requests using its stored verification material.

```text
Agent UID [VM target, not yet tested]    Gate UID [VM target, not yet tested]
  agent dir 0700 private key 0600         gate dir 0700 public key 0600
  SignEd25519(Intent)                     NewEd25519BoundaryWithJournal(public)
             |                             -> pinned site/actor, rev/hash, TTL
    group-accessible gate.sock 0660 ------> SO_PEERCRED expected Agent UID
             |                             replay journal append + fsync
             <------- receipt ------------- advisory or rejected ONLY
  Check SO_PEERCRED expected Gate UID     no private Agent key, no executor
```

Domain-separated Ed25519 message prefix differs from HMAC; `SignedIntent.signatureEd25519` is mutually exclusive with legacy `macSha256`. A Gate configured for Ed25519 **rejects HMAC-only and mixed envelopes**, verifies signatures against its pinned public key and never accepts the public key from the incoming message. Original HMAC tests and tooling still run in their separately constructed HMAC Gate mode; there is no silent downgrade. Replay journal remains per Gate (single writer, locked, fsynced, bounded, corruption fail-closed).

## Runnable commands (DEVELOPMENT ONLY)

```sh
# Test creates separate key directories and two OS processes in a temporary
# private scratch space. It NEVER creates Linux users or modifies host firewall.
bash scripts/test_p0_ed25519_identities.sh
```

Individual code paths:
- `luna-gate init-ed25519 <new-agent-key-dir> <new-gate-key-dir>`: creates two fresh private `0700` directories; an Ed25519 private key exists **only** in the agent directory, and public key exists **only** in Gate directory, each file `0600`. This is a development bootstrap run as one UID. **Never use this bootstrap as a production cross-UID provisioner.**
- `luna-gate serve-ed25519 <policy-file> <gate-key-dir> <group-socket-dir> <allowed-agent-UID> <expected-IPC-GID> <1..30s>`: loads the public key only, opens a separate `replay-ed25519.jsonl` journal, validates a group-owned `0710` directory, opens socket `0660`, verifies its actual Gate UID/GID, and only evaluates HMAC-free Ed25519 requests from the pinned Agent UID. Does not run as root or change group membership.
- `luna-agent preview-ed25519|apply-ed25519 <policy-file> <agent-private-dir> <socket-path> <expected-gate-UID> [requestID nonce]`: loads private key only, signs exact policy digest + revision + actor + TTL, checks the Gate server's **kernel** peer UID and sends bounded framed JSON. A legitimate signed `apply` still returns `REJECTED`.
- `ListenGroup` refuses a socket directory with wrong owner/GID/mode, a pre-existing socket, and unexpected created socket GID. The directory must already have owner Gate UID + IPC GID, mode exactly `0710`. Socket is mode `0660`; only group membership established by an **independent administrator** can make cross-UID connects work in a future isolated test. No chmod/chown/group edits on host are performed.

**Security boundary:** `SO_PEERCRED` UID verifies a Linux user at connect time, **not** the code identity, human owner authentication or integrity of that user's process. Group membership permits transport access, **not authorization**; Ed25519/nonce, policy/scope and future owner consent still matter. Agent private-key compromise allows impersonation until revoked/rotated. The old HMAC `SignedIntent` is a separate P0 legacy test path, not a fallback to the Ed Gate.

## Verified tests and important non-claims

- Unit: [Ed verifier negatives](../../internal/controlgate/ed25519_test.go) reject wrong public key, tampered intent, HMAC-only downgrade, mixed signature, signed APPLY and replay; existing HMAC tests still PASS.
- Files: [identity key tests](../../internal/identitykeys/ed25519_linux_test.go) reject wrong mode, symlink, reinitialization; public/private correspondence validated.
- Linux socket ACL: [group socket tests](../../internal/localipc/group_linux_test.go) reject incorrect GID, overly permissive directory, socket overwrite; output socket must be mode `0660`.
- E2E: `scripts/test_p0_ed25519_identities.sh` uses **two different OS processes but the same HOMESV UID/GID**, detects mismatch of expected Gate UID, replay before/after normal Gate restart and corrupt journal startup refusal; ensures the Gate private-key file does not exist.
- Existing HMAC two-process test still runs, proving backward test compatibility, not production downgrade negotiation.
- **No actual separate Linux UID test or user/group provisioning occurred.** Current validation proves support for **configuration** of different expected UIDs and IPC group, not real isolation. A disposable VM must create distinct Linux users, explicit limited group membership and verify unauthorized peers fail.
- **No actual owner login, signed server response, key rotation or remote attestation** exists. Reply is validated structurally and gated on socket peer UID; a compromised Gate UID can still forge replies. No trust on local kernel integrity under root compromise.
- No persistent service, sudo, nftables, eBPF/XDP, Cloud/NetBird, binary signing, real firewall application or rollback; G-STANDALONE remains blocked.

## HOMESV verification (2026-10-14, source only)

- `go test -count=3 ./...` **PASS**, including Ed25519 mixed signature/downgrade/tamper tests, key permissions, group socket ownership/permission negative tests and all earlier HMAC/UDS modules.
- `go vet ./...` **PASS**.
- `bash scripts/test_p0_ed25519_identities.sh` returned **`P0_ED25519_TWO_BINARIES_RESTART_FAILCLOSED_PASS`**, verifying distinct PIDs under same UID, correct preview, denied APPLY, wrong expected Gate UID, refused replay after restart, corrupt journal failure.
- The original `bash scripts/test_p0_gate_processes.sh` **PASS** for the HMAC regression path.
- Four binaries `lunactl`, `luna-brain`, `luna-gate`, `luna-agent` independently cross-built for **linux/amd64 and linux/arm64**, `CGO_ENABLED=0` **PASS**.
- Source catalog generated and checked with valid repo links: **107 indexed sources PASS**. Cross-UID positive/negative VM acceptance, race detector and kernel nftables acceptance **not tested**.

## Next acceptance gate

1. Run a **multi-UID negative and positive matrix in a disposable Linux VM** (Gate UID, Agent UID, unauthorized UID; allowed socket GID). Do not add users, change system groups or permissions on HOMESV.
2. Protect key material with independent service UID, secure provisioning/rotation/revocation, and trusted owner approval. Avoid agent private-key copying into Gate.
3. Add authenticated server response/session-binding and replay journal tamper/rollback detection under intentional fs corruption/reboot tests.
4. Run `go test -race` in compiler-equipped CI; HOMESV currently lacks gcc.
5. Finish isolated nftables `nft -c` + functional packet/rollback canary before considering a privileged executor. No Cloud/NetBird integration until G-STANDALONE and explicit owner signoff.
