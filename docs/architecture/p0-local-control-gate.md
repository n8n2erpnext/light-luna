# P0 Local Control Gate — Authenticated Intent, Replay and Safe Denial

Status: **IMPLEMENTED SOURCE-LEVEL DEMO — in-process only; NOT IPC, service auth, executor, rollback or firewall enforcement**
Date: 2026-10-14
Scope: local Agent → Policy Gate → future restricted Executor trust boundary, authenticated request envelope and negative tests
Primary source: [controlgate Go implementation](../../internal/controlgate/control.go), [ephemeral selftest](../../internal/controlgate/selftest.go), [negative/concurrency tests](../../internal/controlgate/control_test.go), [lunactl entrypoint](../../cmd/lunactl/main.go), [policy](../../internal/policy/model.go), [security invariants](../07-SECURITY-SAFETY.md)
Prior implementation: [Synthetic Flow Simulator](p0-flow-simulator.md), [unhooked nft preview](p0-nft-unhooked-compiler.md)
Git evidence: use current commit SHA on `feat/p0-standalone-foundation` after source verification. **No deployed services or production controls are implied.**

## Actual runnable code and entrypoints

```sh
go run ./cmd/lunactl control selftest testdata/host-baseline.json
```

`lunactl.run` → `policy.Decode` → `controlgate.SelfTest` → random 32-byte **ephemeral in-memory test-only HMAC key** → signed `PREVIEW_POLICY` and `APPLY_POLICY` requests → `controlgate.Boundary.Evaluate` → JSON receipts. The demo verifies one advisory-only preview plus rejection of duplicate nonce, tampered MAC-covered content and a signed apply attempt. **No private key is checked into Git, printed, stored or sent over a network; no listener or Unix socket is created.**

## Trust-boundary prototype

- `Intent` binds protocol version, request ID, nonce, site ID, actor ID, `OWNER` role and `LOCAL_AGENT` origin to exact policy name/revision/**SHA256 content digest**, issuance/expiry (max 5 minutes) and typed `PREVIEW_POLICY` or `APPLY_POLICY` operation.
- `Sign` uses HMAC-SHA256 over deterministic, typed JSON plus a domain-separated protocol prefix. A 32+ byte **externally supplied** secret is required. `Boundary` stores an independent copy of the secret and pins the expected site and actor. Unknown roles and Cloud relay identities fail P0 validation.
- `Evaluate` validates structure, MAC using constant-time HMAC equality, pinned identity, clock window, exact local policy revision/hash and an **in-memory** replay cache of both nonce and request ID. It denies new requests when replay cache is full; it never silently disables replay protection.
- A fresh authenticated `PREVIEW_POLICY` returns `ADVISORY_ONLY`; **even a correctly signed `APPLY_POLICY` returns `REJECTED`** because there is no privileged executor, watchdog, independent safety/rollback, authenticated cross-process session or real owner confirmation. All receipts always say `executorStatus=NOT_AVAILABLE`, `kernelWrites=false` and `replayProtection=IN_MEMORY_ONLY`.
- Separate `internal/gate.Preview` for Brain proposals remains an inert independent observation boundary; this source demo does **not** merge Cloud governance, Brain decisions, local owner account login or kernel authorization into the same role.

## Tests and explicitly absent guarantees

**HOMESV verification:** `go test -count=1 ./...`, `go vet ./...`, CLI control selftest/snapshot, Linux amd64+arm64 cross-build and catalog/link checks PASS. The 32-way concurrent replay unit test passed under normal Go tests. **Go race instrumentation was attempted but NOT completed** because HOMESV has no `gcc` and the Go race detector requires CGO; do not claim a race-detector PASS. Run `CGO_ENABLED=1 go test -race ./internal/controlgate` in a build environment with a C compiler before using this as a real IPC authorization module.


`go test ./...` covers valid signed preview, repeated request and nonce rejection, altered request data/MAC, invalid origin/role/site, future/expired TTL, incompatible policy revision or changed content at **the same revision**, signed APPLY denial, weak key, excess token lifetime and 32-way concurrent replay attempts (only one accepted); `go vet ./...` and cross-build should be recorded separately when run.

**Important limitations:**
1. In-memory replay protection resets when the process restarts. A new Boundary initialized with the **same key** could accept a previously valid unexpired signed request. In production **persist replay and idempotency state** (or bind requests to one-use boot/session identities with durable authorization semantics) before granting any executor access. The demo generates a **new key per invocation** and does not issue reusable production credentials.
2. HMAC possession is not proof of **OS peer identity**, local-owner presence, TTY confirmation, Cloud session or a trusted agent binary. Production IPC requires Unix socket peer credentials/ACLs or authenticated mutual channel, key rotation, credential storage, sandboxing, subject capabilities and explicit owner approval.
3. No privileged executor is connected, and no `nft` command, eBPF/XDP attach, network/Cloud request or live firewall mutation can be made through this package.
4. Timestamps depend on a trustworthy device clock; no clock sync, anti-rollback, durable transaction IDs, rate limiting across restarts, disaster recovery or real lease/watchdog has been implemented.
5. The gate only uses policy digest from `policy.MakePreview` and **does not establish rescue-route reachability**, multi-tenant policy or actual kernel packet behavior. NetBird integration remains prohibited until standalone acceptance and owner signoff.

## Next implementation gate

Introduce an independently testable **local Agent ↔ Brain/Policy Gate IPC** with peer credential verification and full input framing/limits, but keep all privileges off. Build the future executor as an isolated stub with a **denial-only default** before considering a temporary TTL-based staging, preflight and independent watchdog. Test in a disposable VM/network namespace that cannot touch HOMESV's production nftables, test rollback after forced process termination and loss of network; until then do not expose any apply path.
