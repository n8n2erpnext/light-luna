# 15 — P0 Standalone Foundation (experimental, non-enforcing)

**Status:** Engineering scaffold on branch `feat/p0-standalone-foundation`. This does NOT complete P0, does NOT satisfy G-STANDALONE, and must not be installed as a firewall.

## Implemented in the first slice

- Standard library Go policy reader with 1 MiB input cap, unknown JSON field rejection, and one-object-only parsing.
- Typed minimal PolicyBundle contract for standalone `host` / `edge` only.
- CIDR and port alias validation, uniqueness, canonical CIDR checks, rule UUID/priority checks, P0 supported TCP/UDP/ICMP directions and actions.
- Require declaration of rescue networks before accepting an inbound drop/reject policy even for preview. **Declaring rescue networks does NOT prove reachability.**
- Deny real auto-bans; `off`, `observe` and `shadow` are the only recognized autonomous modes.
- Deterministic, non-enforcing JSON policy preview, with explicit kernel/coexistence/recovery unknowns.
- Synthetic documentation-only sample policy. Documentation IP ranges `192.0.2.0/24` MUST NOT be used for real administration.

## Commands (development only)

```sh
go test ./...
go vet ./...
go run ./cmd/lunactl version
go run ./cmd/lunactl firewall policies validate testdata/host-baseline.json
go run ./cmd/lunactl firewall policies preview testdata/host-baseline.json
```

## P0 VM automatic monotonic TTL watchdog — independent service lab PASS x3 (2026-10-14)

[Verified implementation](architecture/p0-vm-autonomous-ttl-watchdog.md) adds `guest-auto-watchdog.py` to the existing **synthetic-only** KVM VM stage flow, with strict no-NIC/KVM checks, stage ID and journal verification, exact PID process birth+netns inode, atomic READY before commit and autonomous monotonic expiry→nft table removal→journal `ROLLED_BACK`→DONE. A PID1-controlled systemd namespace keeper persists independently after original lxc-exec shell anchor termination. One-shot previous timer approaches were flaky; the new explicit autonomously armed service has **three consecutive combined VM regression PASS** after cleanup-PID scoping, with restored IPv4/IPv6 TCP/UDP. No Light Luna-generated rule was applied; this is **not** production privileged executor or reboot-safe rescue.

## P0 typed synthetic VM transaction contract and bounded recovery test (2026-10-14)

[Stage contract](architecture/p0-vm-synthetic-stage-revert.md) now implemented in `internal/vmstage`, exposed as `lunactl vm stage-intent <policy> <TTL>`: **non-enforcing** manifest pinned to policy digest/revision/table, TTL bounds and four phase labels. The VM test uses an fsynced **single-process mock journal** and a separate systemd rollback service explicitly triggered by the test harness; negative cases reject wrong TTL/stage ID, duplicate and out-of-order progression. TCP/UDP and IPv6 synthetic hooks show packet DROP with nft counters then restored connectivity. Combined VM suite **PASS x2**. Prior independent `systemd-run --on-active` TTL timer was unreliable: **automatic TTL rollback NOT proved**; this is not an authorization gate, product executor or host policy apply. Interrupted journal states were manually reconciled after verified cleanup, with pre-commit abort evidence archived.

## P0 independent VM nft kernel parser/packet/watchdog evidence (2026-10-14 — PASS)

[Kernel-isolated LXD VM](architecture/p0-lxd-kernel-nft-rollback.md) running Ubuntu 24.04 kernel 6.8.0-142 vs HOMESV 6.8.0-146 now supports repeatable no-NIC netns nft checks. [Runner](../scripts/lab/vm/run-tests.sh) asserts separate KVM kernel, no NIC and exact profile before transferring the synthetic baseline's **unhooked Go compiler script**, and uses `nft -c` only. [Guest test](../scripts/lab/vm/guest-nft-test.sh) verifies baseline ICMP success, negative syntax, independent **synthetic-only** nft hook packet drop/counter, controller-placeholder termination and separately triggered systemd service rollback restoring traffic (autonomous TTL timer NOT verified). **PASS on two consecutive runs**. Never interpret this as Light Luna enforcing a live rule, a shipping privileged executor or power-loss-safe rollback. No changes to host firewall or Cloud/NetBird.

## Networkless LXD Agent/Gate UID lab (2026-10-14; EXECUTED)

**Latest evidence:** After the earlier denial was resolved with explicit Light Remote `lxd`/`package-manager`/`sudo-on-demand` capabilities, LXD 5.21.8 was installed. A fresh no-network unprivileged Ubuntu 24.04 container `light-luna-p0-lab` passed actual multi-UID positive/negative IPC. The expanded LXD config has no NIC; the only automatically created `lxdbr0` was removed. `LXD_DISTINCT_UID_POSITIVE_NEGATIVE_RESTART_PASS` and `LXD_CROSS_UID_TEST_RUN_COMPLETED` were observed. Initial test runner failures (Go absent from command PATH and snap confinement hiding host `/tmp`) were corrected with explicit toolchain PATH and a repo-local transient artifacts dir. Follow the latest [test evidence](architecture/p0-lxd-cross-uid-lab.md); earlier BLOCKED paragraph below is a historical pre-install snapshot.



[Lab design/evidence](architecture/p0-lxd-cross-uid-lab.md) and [lab runner scripts](../scripts/lab/lxd/README.md) implement preparation-only orchestration to test **distinct Linux UIDs** and Unix group ACLs inside a disposable *NIC-less* LXD container. **Historical state, before approved installation:** the LXD snap was initially absent and installation denied. The later approved install, no-NIC container and live guest test results recorded above supersede this checkpoint. Only shell static checks are presently available; do **not** mark real multi-UID isolation or kernel nftables acceptance PASS. Container shares host kernel; LXD *VM*, not container, is needed later for netfilter/BPF tests.

## Ed25519 private/public request signing and optional group-owned IPC (2026-10-14)

[Current source design/test](architecture/p0-ed25519-agent-gate-identity.md) adds Ed25519 signing with Agent-only private key; Gate only reads the public verifier and fsynced replay journal. Both old HMAC tests and new Ed25519 tests remain separately configured, with strict **no downgrade** rule. `ListenGroup` takes pre-provisioned Gate-owned 0710 IPC directory and 0660 socket, with pinned Agent/Gate UID/GID checks; it does **not** add users or alter service permissions on HOMESV. End-to-end tests ran as separate PIDs under the **same development UID**. No production distinct-UID service, owner authorization, Cloud link, privileged executor or live nftables.

## Separate P0 Linux Agent/Gate processes with restart replay (2026-10-14)

[Current developer implementation and evidence](architecture/p0-two-process-gate-and-journal.md): `luna-gate` (explicit foreground unprivileged 1–30s serve lease), `luna-agent` (one-shot signed client), private throwaway `internal/devkeys`, and locked append-only `fsync` replay journal `internal/controlgate/journal_linux.go`. End-to-end test uses **two actual PIDs**, stops/restarts the Gate and verifies previously used requests still reject; invalid journal prevents Gate startup. The same-UID key is an **insecure production pattern** deliberately restricted to temporary tests. No production owner auth, separate UID sandbox, durable power-loss proof, privileged executor, Cloud, NetBird or kernel actions. This updates but does not delete the earlier in-process and UDS transport snapshots below.

## Actual Linux Unix socket IPC prototype (2026-10-14)

[Source and tests](architecture/p0-linux-unix-ipc.md) prove a **real, bounded Linux Unix socket**, kernel-level peer UID check using `SO_PEERCRED`, owner-private temporary socket, strict JSON frame input, timeout/concurrency limits and a **distinct test subprocess** connecting to the P0 unprivileged gate. The `lunactl control ipc-selftest` output is labeled `SAME_PROCESS_GOROUTINES_REAL_LINUX_UNIX_SOCKET_NO_EXECUTOR`; **production separated Agent/Gate daemons, persistent replay, OIDC and isolated owner/session keys remain absent**. All resulting receipts are advisory or rejected; no nftables kernel modifications. Do not assume same UID proves real owner identity.

## Authenticated control-gate prototype (2026-10-14)

[Source-level control gate](architecture/p0-local-control-gate.md) and `lunactl control selftest` now model authenticated signed local proposals using an ephemeral injected key, exact policy content/revision, expiring nonce and request IDs. Replays, unsigned/tampered changes, stale policy and **every enforcement attempt** are rejected. All receipts show `kernelWrites=false`; no IPC, service, persistent replay storage, local user authentication, restricted privileged executor or rollback exists. HOMESV's existing GitHub deploy key is usable directly after adding a **verified GitHub SSH host key** to `known_hosts`, avoiding the earlier VPS-ARM relay. This Git transport configuration does not change Light Luna runtime.

## Current flow simulator development slice (2026-10-14)

[Source-level P0 flow simulator](architecture/p0-flow-simulator.md) adds `internal/flowsim`, a strict bounded synthetic flow fixture and `lunactl firewall policies simulate <policy> <flows>`. It matches intent in rule priority order, flags **potential SSH/management lockout**, reports an undefined output default as `UNSPECIFIED`, and labels every effective kernel verdict `UNKNOWN` because it does **not** model Docker/UFW coexistence, routing or conntrack. This remains non-enforcing and has no actual netns/nft kernel test or owner recovery mechanism.

## Non-claims and strict prohibitions

- There is now only inert unhooked text generation (see architecture/p0-nft-unhooked-compiler.md), but NO `nft -c` compatibility test, kernel transaction, privileged executor, watchdog, rollback, rescue path, WebUI, service, network simulator, log collector, firewall installer, ban or NetBird connector in this first slice.
- Policy JSON Schema covers declarative syntax; the CLI additionally checks CIDR network masking, duplicate identities and cross-field constraints. They must remain aligned and will need schema conformance testing.
- No real firewall changes, no production test, no permissions escalation and no networking side effects are authorized by these files.
- Never run `nft add`/`nft flush`/`iptables -F` on a host as part of this scaffold.

## Next safe P0 stages

1. Add negative tests for duplicate JSON keys and stricter alias schemas; add fuzz/regression corpus.
2. Introduce an nft compiler interface that emits *inert text* and is checked in **disposable netns only**, with no application on the parent host.
3. Build an independent local policy gate, least-privilege IPC and out-of-process rollback design; keep them disabled by default.
4. Add standalone CLI/API status and simulation with explicit limits and no real apply.
5. Define CI matrix for Ubuntu/Debian amd64/arm64, compatibility fixture snapshots, SBOM and package dry-runs.
6. Obtain explicit owner approval before ANY live firewall enforcement testing. NetBird work stays parked until G-STANDALONE and separate approval.
