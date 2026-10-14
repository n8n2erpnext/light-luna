# P0 Linux Local IPC — Unix Socket and SO_PEERCRED Prototype

Status: **IMPLEMENTED in development source, tested on HOMESV; not a packaged daemon or privileged policy executor**
Date: 2026-10-14
Scope: bounded Linux UDS transport for authenticated `LOCAL_AGENT` requests to unprivileged, deny-only Policy Gate
Supersedes: none. Extends [in-process control gate](p0-local-control-gate.md), which remains source-valid.
Primary sources: [Linux transport](../../internal/localipc/unix_linux.go), [temporary socket selftest](../../internal/localipc/selftest_linux.go), [socket/cross-process tests](../../internal/localipc/unix_linux_test.go), [HMAC gate](../../internal/controlgate/control.go), [CLI](../../cmd/lunactl/main.go), [security](../07-SECURITY-SAFETY.md).
Canonical evidence: current branch `feat/p0-standalone-foundation`, verify latest commit via `git rev-parse HEAD`. The HEAD shown in older architecture snapshots is historical.

## Entrypoint and tested flow

```sh
go run ./cmd/lunactl control ipc-selftest testdata/host-baseline.json
```

`lunactl control ipc-selftest` reads a strict local policy, creates a 32-byte random **per-invocation test-only HMAC key**, a short-lived `0700` owner-private temporary directory and `0600` Unix socket. It starts a **goroutine** serving `localipc.Serve` in the same unprivileged CLI process, and uses `localipc.Call` across the **actual Linux socket kernel transport**. It runs four requests: valid owner-scoped preview → `ADVISORY_ONLY`, replay → `REJECTED`, modified signed content → `REJECTED`, correctly signed `APPLY_POLICY` → `REJECTED`. Receipts always have `executorStatus=NOT_AVAILABLE` and `kernelWrites=false`.

A separate **automated subprocess integration test** `TestLinuxSocketCrossProcess` launches a second instance of the Go test binary, connects to the test listener from a different PID with the same UID and proves a valid signed preview can traverse the socket. The test fixture uses a **public fixed synthetic key** only in test code; production key provisioning and secret-at-rest protections are NOT implemented. Neither the CLI nor the test installs any daemon.

```text
Synthetic local test client (same UID)                  Mock Policy Gate server
  signed typed HMAC Intent                             owned 0700 tempdir/socket 0600
        |      Linux Unix domain socket / 4-byte bounded frame       |
        +--------------------------> SO_PEERCRED peer UID verification
                                    strict JSON duplicate/unknown rejection
                                    controlgate.Boundary.Evaluate
                                    (MAC, actor+site, policy SHA+revision,
                                     TTL, request ID+nonce replay cache)
        <-------------------------  advisory-only or rejection receipt
     no nft command, no privileges, no Cloud, no kernel/host change
```

## Implemented transport security boundaries

- Linux-only source under `//go:build linux`. Refuse to **listen as root** in this P0 demo; directory must be owned by current UID, true directory (no symlink), mode exactly `0700`; socket path must be short/unused. Socket mode is set to `0600` inside inaccessible parent directory. Listener unlinks its own socket when closed.
- Linux `SO_PEERCRED` is read using `syscall.GetsockoptUcred` for each accepted connection. The server refuses a UID different from its explicitly configured expected UID. Client also rejects a server socket whose kernel peer UID differs from its own UID. **This proves matching local Unix user identity only, not the authorized Light Luna owner or trusted executable identity.**
- Each connection processes **one** four-byte big-endian length-prefixed JSON request (1 to 16KiB), one response (1 to 4KiB), then closes. Strict JSON decoding rejects unknown fields and duplicate nested keys. Per-connection deadline is 3 seconds; at most eight requests are handled concurrently and excess accepted clients are immediately closed.
- HMAC and replay/expiry checks remain independently owned by [controlgate](../../internal/controlgate/control.go). Even a signed request from an accepted UID cannot reach an executor; `APPLY_POLICY` always returns `REJECTED`.
- Every response is checked by the local client for schema, matching request ID, denial/advisory-only decision, `kernelWrites=false` and `executorStatus=NOT_AVAILABLE`. Malformed or oversized frames receive no authenticated response.

## Exact limitations and threat model

1. **Same UID is NOT strong service isolation.** A malicious process running under the allowed UID may connect; code/credential signing remains required. No UID-to-owner/OIDC or peer binary/namespace attestation is implemented. Production Agent/Gate must use dedicated service users and isolated key custody, in addition to OS peer and local-owner authorization.
2. **One real socket is not a production IPC service.** The CLI selftest starts client/server goroutines in one unprivileged process. A subprocess unit test demonstrates cross-process interoperability, but there is no systemd unit, persistent process, user database, session key broker, WebUI/API listener, Cloud control channel, Linux capability enforcement or update lifecycle.
3. The `controlgate.Boundary` replay cache is **memory-only**, reset at restart. Future privileged processing requires durable authorization replay/idempotency or boot/nonce scoped tickets, independently protected by a watchdog. No file-backed replay state or crash-recovery protocol exists.
4. No UNIX file-descriptor passing, bulk streamed telemetry, root shell, executor socket, nftables apply, BPF attach, file malware scan, firewall/sudo changes or NetBird integration. **A test result does not satisfy G-STANDALONE.**
5. The demo cannot substitute for a production MAC-protected reply channel with service keys and a protocol downgrade strategy. Kernel UDS peer credentials do not protect against all threats once root/kernel is compromised.
6. `CGO_ENABLED=0` Go tests and Linux amd64/arm64 cross-builds work. HOMESV still lacks `gcc`, so the race detector should run in CI/build runner rather than be reported as PASS.

## HOMESV verification (2026-10-14 source build)

- `GOTOOLCHAIN=local GOPROXY=off CGO_ENABLED=0 go test -count=1 ./...`: **PASS**, including Linux Unix socket peer checks and subprocess roundtrip.
- `go vet ./...`: **PASS**.
- `go run ./cmd/lunactl control ipc-selftest testdata/host-baseline.json`: **PASS**, preview/advisory and negative decisions as declared, `kernelWrites=false`.
- `go run ./cmd/lunactl system snapshot`: **PASS**, new `linux-local-ipc=UNIX_SOCKET_SELFTEST_ONLY` and firewall `NOT_IMPLEMENTED`.
- `GOOS=linux GOARCH=amd64/arm64 CGO_ENABLED=0 go build` for `lunactl` and `luna-brain`: **PASS**.
- Documentation catalog/link checks **91 indexed sources PASS**. No live agent, host firewall or privileged daemon was started. Race detector **NOT verified** because compiler `gcc` is absent on HOMESV.

## Verified scenarios

- Valid local authenticated preview, duplicate HMAC nonce/request, MAC tampering, signed `APPLY_POLICY` denial, same-UID cross-process socket roundtrip.
- Deny wrong accepted UID, reject directory mode `0755`, symlinked directory and attempts to overwrite an existing socket path; verify socket mode `0600`.
- Reject oversize frame length and invalid/duplicate-key JSON without a successful receipt. Verify hard maximum frame output size.
- Shutdown by caller-owned context and listener close; socket temporary directory removed and no permanent daemon started.

## Later implementation — actual separate test binaries and journal

[The follow-up development slice](p0-two-process-gate-and-journal.md) now provides manually launched `luna-gate`/`luna-agent` in **two separate OS processes**, plus a locked fsynced replay journal that persists nonce/request ID across normal Gate restarts. Both still use a shared key under **the same developer UID**; that is **not distinct service identity/production auth**. The original `lunactl control ipc-selftest` remains an in-process socket demonstration, and neither mode can apply rules to the kernel.

## Next safe increment

1. Implement **dedicated unprivileged process** with a **protected authenticated local key broker** and independently verifiable UID/group model, not demo same-UID.
2. Persist replay/idempotency and roll back pending transactions safely across restart before any privilege-bearing `apply` instruction.
3. Test actual UDS peer access matrix under multiple Linux users in a disposable VM, and run race detector in GCC-enabled CI.
4. Add independently guarded, default-deny privileged executor **only** after isolated nft grammar/packet tests and tested recovery watchdog; no production host nftables changes or automatic bans.
