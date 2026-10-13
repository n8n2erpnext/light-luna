# Initial Standalone Runtime Framework (P0 executable slice)

Status: **source-level implementation, NON-ENFORCING and NOT DEPLOYED**
Date: 2026-10-14
Scope: Go source module boundaries, synthetic offline Brain replay, independent safety preview and component version visibility
Sources: [CLI entry](../../cmd/lunactl/main.go), [separate Brain executable](../../cmd/luna-brain/main.go), [coordinator](../../internal/offline/replay.go), [foundation contracts](../../internal/foundation/observation.go), [policy input](../../internal/policy/model.go), [strict JSON boundary](../../internal/strictjson/decode.go), [tests](../../internal/offline/replay_test.go)
Safety authority: [Security rules](../07-SECURITY-SAFETY.md), [ADR-024](../adr/ADR-024-offline-base-independent-component-updates.md), [ADR-028](../adr/ADR-028-self-defense-host-compromise.md)
Commit: source SHA to be read with git; this document must not be interpreted as deployment evidence.

## Executable skeleton implemented

```text
cmd/lunactl
   |-- version
   |-- system snapshot ------> internal/offline.Snapshot()
   |                             |-- internal/guardian.Snapshot() [UNKNOWN]
   |                             |-- independent App/Brain version strings
   |                             |-- explicit absence of executor/Cloud/BPF
   |-- brain replay <file> --> internal/strictjson.Decode()
   |                             --> internal/foundation.ReplayInput.Validate()
   |                             --> internal/offline.Replay()
   |                                  --> internal/brain.Evaluate()
   |                                  --> internal/gate.Preview() [NO WRITES]
   |                                  --> structured JSON result
   |-- firewall policies validate/preview --> internal/policy.Decode()
   |                                            --> internal/strictjson.Decode()
   |                                            --> Policy.Validate/MakePreview()
   |                                            --> NO kernel writes
   |
cmd/luna-brain (separate Go binary/packaging target)
   |-- version
   |-- replay <file> --> same offline strict JSON/Brain+Gate replay library
   |
future: local auth/API + telemetry collectors + signed KnowledgePack reader
   -> separate unprivileged Basic Brain service
   -> independent owner Policy Gate
   -> privileged nftables executor + watchdog
   -> optional Host Guardian eBPF, optional Cloud and WebUI
```

This **does not yet create independent running processes** or IPC boundaries at runtime, even though two binaries can be compiled separately. The in-process CLI replay is development harness only; future Brain/Agent IPC must be authenticated, scoped and verified before any firewall implementation. The `internal/gate` preview is a deliberate reject/observe stub, **not a security-certified Policy Gate**.

## Currently meaningful behaviors

- `internal/strictjson`: bounded size, recursive duplicate-key rejection, strict unknown-field rejection and exactly one JSON value. Used by policy CLI and synthetic Brain replay.
- `internal/foundation`: bounded versioned observation/replay contracts (128 input observations, 256 KiB file budget, IDs/time/source/coverage constraints).
- `internal/brain`: Go deterministic *demo*, independent `brain.Version=0.0.0-brain.p0`; evaluates synthetic multi-window auth failures as an advisory. No calibrated ML, no local logs, no actual packets or user activity, no real attack attribution.
- `internal/guardian`: honest `UNKNOWN` and `NOT_INSTALLED` statuses; **does not scan host/files/OS**.
- `internal/gate`: bounded `WATCH_ONLY` advisory preview. Other or enforceable proposals are explicitly rejected. It has no capability to reach the kernel or cloud.
- `internal/offline`: local snapshot reports per-component state without conflating App/Brain/Knowledge revisions; simulated chain `observation → brain → gate` emits `kernelWrites=false` always.
- `cmd/luna-brain`: separate buildable executable, currently one-shot offline replay rather than background daemon; normal runtime service isolation still future work.

## Developer acceptance (synthetic, offline)

```sh
go test -count=1 ./...
go vet ./...
go run ./cmd/lunactl system snapshot
go run ./cmd/lunactl brain replay testdata/offline-brain-replay.synthetic.json
go run ./cmd/luna-brain version
go run ./cmd/luna-brain replay testdata/offline-brain-replay.synthetic.json
go run ./cmd/lunactl firewall policies validate testdata/host-baseline.json
go run ./cmd/lunactl firewall policies preview testdata/host-baseline.json
```

Cross build `lunactl` and `luna-brain` for `linux/amd64` and `linux/arm64` with `CGO_ENABLED=0`, `GOPROXY=off` and `GOTOOLCHAIN=local`; only local temp outputs, no installer and **no server software started**.

Security test examples: reject duplicate JSON keys recursively, unknown schema/field, oversize replay, duplicate event IDs, event time/count contradictions, unverified ingress/partial telemetry escalations and any proposed actual ban from the demonstration gate. Replay `SUSPICIOUS_PATTERN` is a **heuristic candidate**, not an attack verdict.

## Concrete next engineering increments (without scope creep)

1. Finish P0 policy/compiler+standalone safety gate in isolated network namespaces, and add test fixtures for admin recovery/coexistence.
2. Establish a stable, bounded, authenticated local Agent ↔ Brain IPC and separately privileged restricted executor (still no live enforcement until authorization).
3. Deliver signed immutable **base knowledge** for actual offline installer later; currently `BASE_PACK_NOT_IMPLEMENTED` is accurate.
4. Add local telemetry adapters and safe observer in shadow with explicit provenance, including sensor coverage/loss reporting.
5. Implement Host Guardian integrity checks and optional malware scanner in isolated VM before enabling on real systems.
6. Keep Cloud sync, fleet, NetBird and XDP control inactive until respective gates/approval. No Cloud dependency for current binary.

## Non-claims / environment

No nftables policy applied, no eBPF program attached, no host scanned, no root privilege granted, no network socket opened by the framework, no customer data used, no model trained, no KnowledgePack verified, no status from a live operating host asserted. CI, E2E production or package acceptance remains to be done.
