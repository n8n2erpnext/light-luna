# P0 Two-Process Agent/Gate and Restart-Persistent Replay Journal

Status: **IMPLEMENTED in Go development source; tested on HOMESV; not installed, not authorized for firewall enforcement**
Date: 2026-10-14
Scope: nonroot **separate Linux process** demo using actual Unix domain socket transport, ephemeral dev-secret provisioning, restart-persistent bounded replay records, fail-closed journal locking and corruption detection
Supersedes: none; extends the prior [Linux Unix IPC prototype](p0-linux-unix-ipc.md) and [local control gate contract](p0-local-control-gate.md)
Primary sources: [Gate process entrypoint](../../cmd/luna-gate/main.go), [Agent process entrypoint](../../cmd/luna-agent/main.go), [demo-key boundary](../../internal/devkeys/keys_linux.go), [journal implementation](../../internal/controlgate/journal_linux.go), [HMAC Control Gate](../../internal/controlgate/control.go), [end-to-end test script](../../scripts/test_p0_gate_processes.sh), [journal unit tests](../../internal/controlgate/journal_linux_test.go), [key unit tests](../../internal/devkeys/keys_linux_test.go)
Security authority: [Light Luna safety](../07-SECURITY-SAFETY.md), [ADR-028 self-defense](../adr/ADR-028-self-defense-host-compromise.md), [project truth](../project-book/PROJECT_TRUTH_STATUS.json)
Git evidence: see current branch `feat/p0-standalone-foundation` HEAD and Worklog. This document is **not a release, kernel safety certification or packaging guide**.

## First-create journal entry durability ordering (2026-10-14 — development source, nonroot verified)

**Scope:** Core P0, parent feature branch commit `1611a6a6461c0f6ce61ec86eecbc516a5aaa02dd` at HOMESV prior to this increment. [Source](../../internal/controlgate/journal_linux.go) and [tests](../../internal/controlgate/journal_linux_test.go). This is a **development journal startup-safety improvement**, not production crash recovery, signed human-owner approval or privileged executor authorization.

The Gate journal previously opened with `O_CREAT|O_APPEND|O_NOFOLLOW` and synced each later replay record, but did **not** explicitly sync the newly created **parent directory entry**. The code now first attempts `O_CREAT|O_EXCL|O_NOFOLLOW` to distinguish a new inode from an existing one; on `EEXIST` it opens the existing path **without `O_CREAT`** and preserves prior content. As before, it validates a private 0700 owner directory, a single-link 0600 regular file, and an exclusive nonblocking `flock`. **Only when newly created**, while locked and before accepting any signed intent, it runs `fsync(file)` then opens its parent with `O_DIRECTORY|O_NOFOLLOW` and runs `fsync(parent-directory)`. Sync/open/close error aborts startup; it does not fall back to in-memory replay protection. A second Gate still cannot take ownership of the journal.

**Evidence:** three successive Go test runs across all packages, `go vet`, two nonroot, two-process HMAC and Ed25519 gate scripts (restart replay and corrupted journal deny), plus 91 offline VM Python regression tests all PASS. Added [journal unit cases](../../internal/controlgate/journal_linux_test.go) confirm new 0600 journal creation and safe reopen; reject existing directory, hardlinked replay file, attempted sync on closed fd, and symlinked parent path; confirm normal private-parent sync. No actual power interruption, disk-full crash, filesystem fault injection, or physical storage durability was measured.

**Nonclaims:** fsync directory is a necessary ordering step on supported filesystems, not a guarantee against defective storage, kernel crash or power-cut. The test Agent Ed25519 key authenticates a **local test signer**, not the real owner; string role `OWNER` does not establish independent human approval. `APPLY_POLICY` remains **REJECTED** with `KernelWrites=false` and `ExecutorStatus=NOT_AVAILABLE`. This increment starts no services and modifies no live nft, firewall, host users, NetBird or Cloud. P0 OPEN, G-STANDALONE BLOCKED.

## Concrete tested execution (safe local scratch only)

```sh
# Runs on HOMESV as nonroot; creates and destroys tempdir, random key and replay journal.
bash scripts/test_p0_gate_processes.sh
```

The test builds **two independent Linux executables** `luna-gate` and `luna-agent` with `CGO_ENABLED=0`, creates a private temporary working directory and initializes a *new*, random, 32-byte test-only shared key.

- `luna-gate init <new-private-demo-dir>` creates the directory `0700` and `p0-shared.key` `0600` with cryptographic randomness. Refuses root and reinitializing an existing directory, preventing accidental destructive overwrite.
- `luna-gate serve <policy.json> <private-dir> <max-seconds>` starts an **explicitly invoked foreground process** for 1–30 seconds, reads locally validated policy, verifies key ownership/mode/size, acquires a nonblocking exclusive lock on a `replay.jsonl` journal, reconstructs the active replay cache, opens the `gate.sock` private Unix socket, checks `SO_PEERCRED` and routes each signed request to `controlgate.Boundary.Evaluate`. It exits after its short lease/SIGTERM and removes its socket. There is **no installer, systemd unit or privileged executor**.
- `luna-agent <preview|apply> <policy.json> <private-dir> [requestId nonce]` starts a **different one-shot Linux process**, reads the *same synthetic demo key*, checks Gate server UID, signs bounded intent pinned to local site/actor/revision/content hash and sends one framed request. It prints a redacted typed receipt; **even a signed `apply` request returns REJECTED**.
- Optional explicit IDs are **dev/test fixture controls only** to replay exact request ID/nonce on a later process. No production session, authenticated human OWNER identity, Cloud access token or legitimate approval is implied by the `ActorRole: OWNER` demo field.

```text
luna-agent (separate PID, same dev UID)   luna-gate (separate PID, same dev UID)
  private test key 0600                     same private test key 0600
  Sign HMAC (policy SHA/revision)                ListenPrivate UDS 0600
         |                                  SO_PEERCRED UID check
         |==== signed length-framed JSON ======>|
         |                                  validate MAC+policy+expiry+nonce/ID
         |                                  journal append -> fsync BEFORE reply
         |                                  [advisory preview / denied apply]
         |<=== bounded JSON receipt ===========|
     no kernel writes                      no privileged executor
```

## What is actually protected, and what is not

- **Restart replay continuity:** `NewBoundaryWithJournal` hydrates prior unexpired request IDs/nonces and holds `flock LOCK_EX|LOCK_NB` for the Gate process lifetime. A second Gate on the same journal fails rather than operating without replay protection. Already-used IDs and nonces continue to be rejected after the Gate process is stopped/restarted with the same key.
- **Fail-closed journal:** append-only JSON lines (source-validated, strictly bounded), `O_NOFOLLOW`, single-link owner-private `0600` file in `0700` dir, max 2MiB, `fsync` on each recorded accepted/denied-signed request. Truncated, malformed, oversized, insecure, symlinked or concurrently locked journal prevents Gate startup or new signed actions, **never falls back to RAM-only mode**. Exhaustion causes denial. No silent file repair or truncation.
- **No unattended daemon:** lease max 30 seconds; test script starts/stops both programs in a disposable scratch directory and traps cleanup on exit. Tests do not grant root, capabilities, firewall network namespace access, VPN or Cloud permissions.
- **OS peer identity is limited:** Gate requires caller Linux UID to match its dev user, but the key is deliberately shared between same-UID processes for this P0 test. Another malicious process running under **that UID may read the key**; this is NOT an acceptable production key-distribution design. We have not demonstrated different-UID isolation. `OWNER` role is synthetic, not an authenticated person.
- **Disk persistence ≠ crash-consistent auth transaction:** `fsync` protects the single append file under normal restart tests, but there is **no power-loss/FS replay durability proof**, no journal compaction/backup, no protected monotonic clock, no durable policy-executor transaction, no hardware key attestation. A 2MiB full journal fails closed and requires manual controlled migration. A signed `APPLY_POLICY` is stored as spent but still categorically **rejected**.
- **Sender and receiver no privilege:** the Gate's socket communicates only with `localipc` and `controlgate`; no `nft`, `nft -c`, XDP, eBPF, system service, runtime malware scanner or Cloud port exists. Even a valid preview receipt has `kernelWrites=false` and `executorStatus=NOT_AVAILABLE`.

## Verified test coverage and evidence

`go test -count=1 ./...`, `go vet ./...`, `bash scripts/test_p0_gate_processes.sh` exercise:
- Two real different PIDs performing signed preview and deny-only apply.
- Repeat request ID/nonce denied before and **after restart** (same key), while a new unique preview works.
- Corrupt JSON journal forces server startup rejection; second journal lock fails, insecure file mode and symlinks rejected.
- Failed journal write is sticky fail-closed in unit tests; unexpected contents or expiry do not silently allow a privileged operation.
- Permission checks on `0700` directory and `0600` key; refusal to reinitialize existing key and key symlinks.

Go unit tests run without CGO. The `go test -race` detector still requires a CI runner with GCC and is not marked PASS on HOMESV. Separate Linux/amd64 and Linux/arm64 cross-build results are recorded in Worklog when run. Existing `lunactl` IPC/selftest, synthetic flow simulator and Brain demo continue independently.

## Verified HOMESV results for this source increment

- `GOTOOLCHAIN=local GOPROXY=off CGO_ENABLED=0 go test -count=3 ./...`: PASS (three consecutive full-suite runs, including journal, key-file and local IPC tests).
- `go vet ./...`: PASS.
- `bash scripts/test_p0_gate_processes.sh`: **P0_TWO_PROCESSES_JOURNAL_RESTART_CORRUPTION_PASS**; two separate executables, restart, deny-old request, fail-corrupt-store behavior.
- `linux/amd64` and `linux/arm64` static (`CGO_ENABLED=0`) cross-build: **all four** `lunactl`, `luna-brain`, `luna-gate`, `luna-agent` PASS.
- `scripts/update_doc_catalog.py --check --check-links`: **99 indexed sources PASS**; `git diff --check`: PASS. Compiler/job results are from HOMESV source checks, not deployed acceptance.
- Race detector `go test -race` still **NOT verified** on HOMESV because `gcc` is unavailable; this is a CI/isolated-runner follow-up. No nft/kernel capability tests were attempted.

## Subsequent Ed25519 key separation mode (not part of the original HMAC proof)

[Ed25519 Agent/Gate verification](p0-ed25519-agent-gate-identity.md) adds a **separate** test mode in the same binaries: the Agent alone reads an Ed25519 private key and Gate only reads a pinned public key, with group Unix socket UID/GID checks. The previous HMAC mode documented above is still used for regression. New same-UID Ed mode does **not** prove distinct Linux service user deployment.

## What to implement before production

1. **Different service UIDs, OS peer credential policy and process sandbox**, with separate rights for broker, Agent, Gate and never-shared owner secrets. There must be no private key file jointly readable by potentially hostile same-UID processes in production.
2. **Owner login and explicit policy authorization** (local recovery owner, MFA/Cloud roles and revocation), not synthetic `OWNER` label. Mutual response authentication, boot/session key rotation, key theft containment, and a real trust-store manager.
3. **Durable replay and action transaction with crash/reboot semantics**, protected independent watchdog and journaling/compaction/restore against disk full and rollback attacks; current fsynced JSONL is bounded test evidence, not certified durable authorization.
4. **Isolated VM Linux nft syntax+functional traffic tests, Docker/UFW coexistence and rescue canary** before any privileged executor or automatic blocking.
5. CI race detector and multi-user negative tests in disposable Linux VM. P0 and G-STANDALONE remain OPEN/BLOCKED; NetBird integration is not approved until separate gate and owner consent.

**Non-claims:** Neither process is a shipping Luna agent/service. No firewall rules or Linux kernel hooks were changed, no network was reconfigured and no customer telemetry entered Cloud.
