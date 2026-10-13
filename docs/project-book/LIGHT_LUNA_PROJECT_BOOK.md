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
