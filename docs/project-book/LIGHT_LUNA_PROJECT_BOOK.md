# Light Luna Project Book

Status: current project synthesis (not a substitute for code)
Date: 2026-10-13
Scope: product direction, architecture boundaries, documented P0 reality, development order, evidence and open gates
Source baseline: public `n8n2erpnext/light-luna` `main@66920c251603b0b0c4bcd0d0cfc94ca8f7220d41`; development `feat/p0-standalone-foundation@9ae27b267e72e465ea7808c3936fabcf6b0cac05`

## Reading contract

Read [governance](LIBRARY_RULES.md), [library map](DOCUMENT_LIBRARY_MAP.md), [Code Map](LIGHT_LUNA_CODE_MAP.md), [truth status](PROJECT_TRUTH_STATUS.json), and exact source before changing implementation. Product documentation may describe intended capability, not shipped behavior.

## Identity and owner direction

Light Luna is an independent, installable Linux **software firewall first**, with optional Luna Brain, WAF/IDS, Cloud and later hardware modes. It should function without Cloud or NetBird. Owner confirmed **public GitHub repository is intentional** on 2026-10-13, superseding the earlier private-repository preference in historical handoffs.

## Non-negotiable stage order

[Roadmap](../11-ROADMAP.md) and [ADR register](../12-ADRS.md): P0–P4 independent Host/Edge, explicit gate **G-STANDALONE**, then P5 optional NetBird development **only after separately authorized by owner**. No early connector, read-only API probing, dependency, credentials or compatibility integration tests. Host/Edge packet visibility depends on topology; one-NIC passive Edge is not an automatic LAN packet sensor.

## Technology roles (implemented vs proposed)

- **Go — actual P0 implementation**: `lunactl` CLI, typed policy validation, inert preview. Target main backend, API and separately privileged executor. The executor **does not exist**.
- **React + TypeScript — WebUI proposal only**: browser dashboard, policy editor, observability; no frontend source or service exists in this branch.
- **Go-native Luna Brain — owner-approved architecture direction, NOT IMPLEMENTED**: deterministic analytics, evidence correlation and typed recommendations should use Go; even with Go shared across components, Brain cannot access the privileged executor. See [ADR-020](../adr/ADR-020-go-core-python-optional.md).
- **Python — optional worker only, NOT IMPLEMENTED**: selective, justified ML/stats/inference work using separate sandboxed unprivileged process, bounded CPU/memory/time and versioned inputs/outputs; no mandatory Python runtime for normal Host/Edge operation. `.mjs` is no longer the default Brain choice.
- **SQLite — local store proposal only**: no database schema or service yet.
- **nftables — future enforcement backend**: no compiler, transaction, watchdog, kernel change, service or real rules deployed.

Rust, a **mandatory Python runtime**, and a separate Node backend are not required by the P0 firewall. Python 3 currently runs only the docs catalog tool; a future optional ML worker is permitted but does not yet exist. React/TypeScript WebUI and SQLite remain proposals rather than delivered code.

## Verified engineering snapshot

[Code Map](LIGHT_LUNA_CODE_MAP.md) traces the exact Go CLI path. HOMESV is primary development workspace `/home/ubuntu/n8n2erpnext/light-luna`; Go 1.27.2 user-space toolchain used for tests. Development commit `9ae27b2` adds policy validator + preview + synthetic fixtures; recorded `go test ./...`, `go vet ./...`, CLI checks and amd64/arm64 cross-builds passed on HOMESV. **Those tests do not establish a working firewall or P0 acceptance.**

Previously the docs-only `main` commit `66920c2` was published using VPS-ARM deploy key. All production infrastructure/firewall remained untouched in those operations.

## Optional Linux eBPF/XDP direction (decision, not deployed)

[ADR-021](../adr/ADR-021-ebpf-observability-xdp-optional.md) adds **non-mandatory eBPF tracing** as a potential source of bounded host/network/socket/process telemetry and **optional XDP ingress DDoS filtering** after isolated tests and explicit enablement. nftables remains the primary firewall. No source object, attach, XDP_DROP, loader or benchmark exists in the current project. Do not confuse an existing kernel BPF attachment on HOMESV with Light Luna functionality. Kernel/driver support and coexistence must be checked before any isolated lab attachment, with no interruption of third-party XDP. Optional XDP is not a G-STANDALONE requirement; NetBird remains deferred.

## Brain V2 and collective security knowledge (agreed architecture, not built)

After read-only inspection of the independent VPS-ARM Sentinel services, the source-audited legacy “Luna” is **Python rules, hand-authored heuristic scores, SQLite observations, JSONL cognition templates and bounded learning curriculum**—not a demonstrated model with trained neural weights. [Evidence audit](../history/audits/SENTINEL_LUNA_COGNITION_AUDIT_2026-10-13.md) records the exact logic, confidence blend/decay and generated knowledge record counts. [ADR-022](../adr/ADR-022-rebuild-luna-brain-governed-knowledge-migration.md) chooses a new Go-native Luna Brain V2 and selective, reviewed knowledge migration rather than cloning the live system. [Brain V2](../architecture/luna-brain-v2.md) defines typed evidence/weights/scenario and action limits, while [Threat Knowledge Pipeline](../architecture/threat-intelligence-knowledge-pipeline.md) defines trusted CVE/KEV/EPSS/vendor+ATT&CK inputs, exact affected-asset match, signed knowledge snapshots and a future opt-in, privacy-preserving Cloud learning contract. Static schemas are provided for design but **are not Go runtime API**. No Brain V2 code, local threat intel matcher, neural training/inference, Cloud knowledge service or live response has been deployed.

## Pending review: dual-role Brain, bounded memory and Cloud intelligence

[ADR-023](../adr/ADR-023-dual-brain-memory-cloud.md) **PROPOSES** two logical Brain roles (Go deterministic analytical path and optional hypothesis/ML reasoning) using one evidence/knowledge workspace, with independent policy authorization. It proposes bounded data storage at experimental Starter 256 MiB / Standard 512 MiB / Extended 2 GiB; these are NOT measured deployment figures and exclude application binaries, other apps' logs and kernel map RAM. [Local memory design](../architecture/luna-memory-lifecycle.md) defines sampling/TTL/compact/pressure, while [Cloud intelligence](../architecture/luna-cloud-intelligence.md) defines a future optional feed curator, signed scoped deltas, consent-based aggregate upload and offline local authority. No Cloud Brain is required for per-packet decisions. These concepts must not block P0–P4 standalone operation or bypass the G-STANDALONE and NetBird restrictions.

## Owner-decided offline installation and separate component versions (NOT IMPLEMENTED)

[ADR-024](../adr/ADR-024-offline-base-independent-component-updates.md) **decides** the first shipping installer must contain **Basic Go Luna Brain plus verified baseline knowledge**, a usable standalone App/UI/CLI and separately restricted firewall/recovery, even with no Cloud connection. The Brain is **required inside the install bundle but not allowed to block firewall/TTL operation when it crashes**; heavy optional Python inference remains extra. [Lifecycle specification](../architecture/offline-base-component-update-lifecycle.md) defines separate App, Brain, Executor, per-module, KnowledgePack-revision and optional ML model versions, explicit compatibility graph, signed Cloud/offline distribution and isolated rollback. This supersedes earlier packaging text calling the Brain optional. These are release acceptance criteria only: no installer/Brain/update system exists in this P0 Go CLI branch. Store/budget proposals in ADR-023 remain unaccepted sizing experiments.

## Cloud portability and scale-out direction (DECIDED design, no deployment)

[ADR-025](../adr/ADR-025-portable-scalable-cloud.md) formalizes the owner's requirement that Cloud be movable among VPS-ARM, VPS-AMD, other providers and eventual k3s **without reinstalling devices, changing public API/trust identity or losing KnowledgePack revisions**. Start small with modular Go API + independent workers, **single PostgreSQL writer**, signed immutable object storage via a provider-neutral adapter, container images/config templates and offsite restorable backups. [Architecture](../architecture/cloud-portability-and-scale.md) defines later stateless replica scaling, while [operator runbook](../architecture/cloud-migration-and-recovery.md) protects old-writer fencing, Postgres WAL/PITR, object checksums, key custody, client sync/idempotency and safe failure. **No cloud service, backup, hosted storage or migration is implemented or operated.** The standalone-first gate and Cloud-optional device model remain unchanged.

## One-account local/Cloud identity and owner-managed remote control (DECIDED design)

[ADR-026](../adr/ADR-026-unified-account-remote-fleet-control.md) selects one human OIDC identity for Cloud and enrolled local WebUI accounts, plus **offline local bootstrap owner and protected break-glass**. Cloud Console can later show actual reported firewall/Brain/module status and enqueue versioned typed configuration intents via outbound device connection, not a direct inbound root/shell. [Architecture](../architecture/unified-account-cloud-fleet-control.md) keeps reported local config authoritative, desired Cloud revisions pending until acknowledged and local Policy Gate enforcing scope/revision/MFA/risk checks. Cloud enrollment, remote control and learning telemetry opt-in are different permissions. No Account/Cloud/device enrollment or remote-action runtime is yet implemented; P0 standalone remains sole reachable implementation.

## Safety boundary

Mandatory scope and TTL, owner emergency paths, audit, independent privileged policy gate, transactional rollback lease and separation of host/vendor nftables tables are [design requirements](../07-SECURITY-SAFETY.md), **not implemented test results**. No root host firewall modifications; no development code is deployed.

## Open gates

1. P0 strict input/schema negative/fuzz checks, inert compiler, isolated netns tests, independent watchdog contract, scoped CLI/API.
2. Dedicated HOMESV Git SSH authorization and CI workflow (pending separate verification).
3. Full package/install/upgrade/uninstall, isolated enforcement and recovery proof; benchmarked amd64/arm64.
4. P1–P4 Host/Edge release-quality implementation and tests.
5. G-STANDALONE owner signoff, then *separate* NetBird go-ahead.
6. Cloud/brain/WAF/hardware decisions require explicit verified evidence.

## Source bookmarks

- [Architecture](../01-ARCHITECTURE.md), [Firewall](../02-FIREWALL-ENGINE.md), [Security](../07-SECURITY-SAFETY.md), [Roadmap](../11-ROADMAP.md), [ADRs](../12-ADRS.md)
- [CLI source](../../cmd/lunactl/main.go), [policy](../../internal/policy/validate.go), [tests](../../internal/policy/policy_test.go), [fixture](../../testdata/host-baseline.json)
- [Initial handoff](../14-HANDOFF-2026-10-13.md), [P0 implementation record](../15-P0-IMPLEMENTATION-NOTES.md), [HOMESV handoff](../16-P0-HOMESV-HANDOFF-2026-10-13.md) — historical context, not runtime truth.
