# Light Luna Worklog

Status: chronological worklog; not canonical runtime evidence
Date: 2026-10-13
Scope: durable decisions, verified engineering milestones and blockers

## 2026-10-13 — Documentation baseline, standalone-first lock

- Architecture, security, stages P0–P7 and explicit G-STANDALONE were written in [legacy docs](DOCUMENT_LIBRARY_MAP.md). NetBird connector remains deferred until G-STANDALONE and owner approval.
- Git `main` published to public `n8n2erpnext/light-luna` at `66920c2`; owner later confirmed repository may remain **public**.

## 2026-10-13 — HOMESV P0 source slice

- HOMESV became primary development environment. Feature branch `feat/p0-standalone-foundation@9ae27b267e72e465ea7808c3936fabcf6b0cac05`.
- Go CLI PolicyBundle validator and inert preview added. HOMESV `go test ./...`, `go vet ./...`, validate/preview, Linux amd64/arm64 builds passed. No firewall runtime or release.
- Source branch propagated to GitHub with SHA-256 verified Git bundle via VPS-ARM. GitHub PR creation via connector returned 403; remote branch exists. See [historical handoff](../16-P0-HOMESV-HANDOFF-2026-10-13.md).

## 2026-10-13 — Documentation governance adopted from LightBI

- Inspected read-only VPS-ARM LightBI document library in `LightBI-docs-cli-lightbi/docs/project-book/`: its `LIBRARY_RULES`, `DOCUMENT_LIBRARY_MAP`, `DOCUMENT_TEMPLATE`, `LIGHTBI_CODE_MAP`, Project Book and source catalog mechanism.
- Adapted a **new Light Luna-owned** library with [governance](LIBRARY_RULES.md), [navigation](DOCUMENT_LIBRARY_MAP.md), [Project Book](LIGHT_LUNA_PROJECT_BOOK.md), [Code Map](LIGHT_LUNA_CODE_MAP.md), machine truth status and generated SHA source catalog.
- Existing `docs/00`–`docs/16` paths kept unchanged. Future docs use functional shelves and authority labels; no LightBI code/doc copied wholesale.

## 2026-10-13 — Go/Python component decision

- Owner accepted the Go-native security/backend/Brain direction and explicitly allowed Python in components that benefit from its ML/data-science ecosystem.
- Recorded [ADR-020](../adr/ADR-020-go-core-python-optional.md): Go deterministic Brain remains independent of an optional sandboxed non-root Python worker. `.mjs` is no longer the default Luna Brain runtime.
- This is an **architecture decision only**; no Brain, worker, Python runtime dependency, firewall enforcement or production changes were made.

## 2026-10-13 — eBPF/XDP extension decision

- Reviewed [Cloudflare L4Drop 2018](https://blog.cloudflare.com/l4drop-xdp-ebpf-based-ddos-mitigations/), Linux kernel eBPF/XDP documentation, cilium/ebpf and libxdp multiprogram coexistence guidelines.
- Recorded [ADR-021](../adr/ADR-021-ebpf-observability-xdp-optional.md): nftables authoritative; optional read-only eBPF tracing; XDP opt-in ingress mitigation only after independent gate/recovery and owner approval; existing XDP programs must never be replaced.
- HOMESV read-only link/kernel inventory confirms the machine has Linux 6.8 kernel/BTF but **does not prove safe/native XDP support**. No Light Luna BPF program loaded, attached or built; do not publish user infrastructure-specific device details in the repository.

## 2026-10-13 — Legacy Luna source audit and Brain V2 redesign

- Performed **read-only** VPS-ARM inspection of the original Python Luna/Selene service sources, live systemd unit metadata, source-only local Git checkout, cognitive JSONL schemas/counts, SQLite schema/size and intelligence adapters. No live user data rows exported and no configuration changed.
- Verified three generated cognition zones: 11,980 reasoning, 17,115 knowledge, 5,863 context records; generator creates repeated template combinations to fill memory budget, not neural model weights. Source audit lists site-scoped decision thresholds, pattern scoring/decay and missing evidence gates.
- Chose [ADR-022](../adr/ADR-022-rebuild-luna-brain-governed-knowledge-migration.md): rebuild Go-native Brain V2, migrate reviewed/deduplicated facts and tests **without copying live production DB or generated bulk context**. Original Sentinel stays independent.
- Created [Brain V2](../architecture/luna-brain-v2.md), [Threat Knowledge Pipeline](../architecture/threat-intelligence-knowledge-pipeline.md) and two versioned advisory knowledge/proposal JSON Schema drafts (not wired to runtime). Future Cloud learns from opt-in reviewed and privacy-minimized evidence, not raw traffic or automatic global ban policies.
- Site-specific legacy production status and all security actions left unchanged. Brain implementation and G-STANDALONE remain open.

## 2026-10-13 — Design review: hemispheres, training, memory and Cloud

- Reviewed existing 3-region Sentinel corpus semantics against the requested two-brain metaphor; recommended **logical analytical vs hypothesis reasoning** roles rather than actual neural hemispheres, preserving independent critic and Policy Gate.
- [ADR-023](../adr/ADR-023-dual-brain-memory-cloud.md) remains **PROPOSED** pending owner approval. Training a core neural model is not a baseline prerequisite; small optional ML worker requires labelled benchmarking.
- Proposed bounded client storage profiles, detailed quota/compaction and stale-evidence rules in [Local Memory Lifecycle](../architecture/luna-memory-lifecycle.md).
- Proposed offline-first, signed update/delta and privacy-preserving future multi-tenant service in [Cloud Intelligence](../architecture/luna-cloud-intelligence.md). No Cloud service, model, database size caps or client sync actually implemented.

## 2026-10-13 — Owner approves offline Basic Brain and split release versions

- Clarified product contract: Light Luna installs as a **self-contained, Cloud-optional firewall with mandatory compact Go Basic Brain and baseline knowledge**, not a Cloud-dependent empty shell. Brain failure still cannot halt or loosen independent firewall policy/TTL.
- [ADR-024](../adr/ADR-024-offline-base-independent-component-updates.md) DECIDED separate `appVersion`, `brainVersion`, `executorVersion`, module versions, `knowledgeRevision`, optional `modelVersion`, compatibility matrix/ABI, signed Cloud/offline update bundles and per-component rollback.
- [Lifecycle](../architecture/offline-base-component-update-lifecycle.md) and two static JSON schema contracts plus a fictional offline-state fixture were created; these are design only and not a running updater, signed artifacts or executable firewall release.
- Preserve P0–P4 standalone-first, defer optional Cloud service and NetBird integration, and test first-install with no network before any release claim.

## 2026-10-14 — Cloud portability and migration-first architecture

- Owner required Light Luna Cloud to **scale and move VPS/provider/stack** with minimal client disruption. Recorded [ADR-025](../adr/ADR-025-portable-scalable-cloud.md) DECIDED portability acceptance, leaving specific providers/thresholds and recovery SLOs subject to testing.
- Designed [modular scalable Cloud](../architecture/cloud-portability-and-scale.md) (Go stateless API, workers, single Postgres primary, S3-compatible immutable objects, Compose-first → k3s) and [migration runbook](../architecture/cloud-migration-and-recovery.md) (restore/PITR, old-primary fencing, signing keys/device identities, no split brain and verified handover).
- Introduced a **synthetic schema + non-approved inventory fixture** for validating future migration preparation without recording customer credentials/logs. No Cloud instantiated, backups initiated or infrastructure changed.

## 2026-10-14 — Owner decides unified account and CrowdSec-style Cloud fleet management

- Reviewed official CrowdSec Console enrollment and polling-based decision-management docs as **product precedent**, not an exact implementation dependency or promise of identical features.
- [ADR-026](../adr/ADR-026-unified-account-remote-fleet-control.md) sets one Cloud-account human identity for enrolled local and Cloud WebUIs; preserve offline bootstrap owner/emergency recovery and avoid Cloud-password replication to client.
- [Architecture](../architecture/unified-account-cloud-fleet-control.md) defines explicit device enrollment and remote-control authorization, RBAC, device outbound polling, desired vs actual config, compare-and-swap revisions, local Policy Gate, approval, expiry, audit and conflict receipts; Cloud cannot root-shell local nodes.
- Added design-only typed fleet intent JSON Schema and synthetic intent fixture; no accounts, Cloud API, local WebUI or firewall changed.

## 2026-10-14 — CrowdSec behavior/scenario comparison and federated immune-style direction

- Reviewed [CrowdSec scenarios](https://docs.crowdsec.net/docs/log_processor/scenarios/intro/), [Central API contribution](https://docs.crowdsec.net/docs/central_api/intro/), [community blocklists](https://docs.crowdsec.net/docs/central_api/community_blocklist/) and [Remediation Sync beta](https://docs.crowdsec.net/u/console/remediation_sync/).
- [ADR-027](../adr/ADR-027-collective-threat-immunity.md) DECIDED owner direction: opted-in client sends bounded **subject + behavior + scenario + evidence and local action/outcome**, Cloud vets/deduplicates and issues **signed scoped watch/curated bulletins** so other nodes recognize attack-like behavior early. No automatic global ban from one client.
- [Federated design](../architecture/federated-immune-threat-intelligence.md) distinguishes local mitigation receipts vs proven maliciousness, official scenario hashes vs custom, public indicator provenance/TTL, corrections/retractions and anti-Sybil. Added two static candidate JSON schemas and synthetic fixtures. Nothing executed on host firewall or Cloud.

## 2026-10-14 — Owner requires Light Luna self-defense and host-compromise containment

- Investigated authoritative eBPF/Tetragon/Falco runtime security telemetry, Linux IMA/fs-verity integrity enforcement and ClamAV on-demand/on-access fanotify scanning. **XDP alone is a packet hook, not a host malware scanner**; severe root/kernel compromise can falsify all in-host sensors.
- [ADR-028](../adr/ADR-028-self-defense-host-compromise.md) DECIDED new guardian/threat model: tamper/signed update protection, limited eBPF process/file/network observation, optional bounded AV, safe local egress/lateral containment after Policy Gate, independent recovery; Cloud rejects/quarantines compromised contributions and never accepts raw malware/executable samples from clients by default.
- [Architecture](../architecture/host-self-defense-and-compromise-containment.md) and static HostIntegrityReport schema + synthetic compromised-device fixture added. No malware scanning, system installation, firewall attach, Cloud upload or code enforcement performed.

## 2026-10-14 — First executable P0 Go component skeleton

- Implemented a **genuine runnable but non-enforcing** foundation slice in [local runtime design/verification](../architecture/local-runtime-foundation.md): `internal/foundation` bounded event contracts; `internal/brain` deterministic, synthetic-only auth-burst observation; `internal/guardian` UNKNOWN coverage; `internal/gate` advisory-only/no-write; `internal/offline` first observable call chain and component snapshot; `cmd/luna-brain` separately buildable, one-shot replay; enhanced `lunactl` system snapshot/brain replay.
- Hardened `internal/policy.Decode` by reusing `internal/strictjson`: recursively rejects duplicate JSON object keys, unknown fields, trailing data, bounded sizes; new nested-duplicate regression tests and replay positive/negative tests. Existing policy validation/preview preserved.
- Ran Go unit tests and vet, standalone Brain replay/snapshot, and separate `linux/amd64`+`linux/arm64` cross-builds with network-disabled Go tooling. **Not a kernel/firewall/Cloud or malware scanning test**; no live infrastructure touched.

## 2026-10-14 — P0 unhooked nftables text compiler

- Implemented `internal/nftpreview.Compile` and `lunactl firewall policies nft-preview`: policy validation → stable source alias expansion/order → scoped `table inet luna_p0_preview` **regular chains without hooks**, JSON plan reporting **no kernel effect, no syntax check**. No privileged executor, base-chain policy, default DROP, emergency route, Cloud or NetBird.
- Added deterministic tests for rule priority, protocol/family combinations, unsafe mismatch rejection, expansion caps, and malicious note non-interpolation. Existing parser/Brain/CLI tests and Go vet passed.
- Verified HOMESV nftables v1.0.9 availability but `unshare --user --map-root-user --net` not permitted. **Skipped `nft -c` instead of touching parent-host network namespace**, recorded isolated-VM grammar test as next dependency; see [compiler truth/verification](../architecture/p0-nft-unhooked-compiler.md).

## 2026-10-14 — P0 synthetic Flow Simulator (Go code and tests)

- Implemented `internal/flowsim` and `lunactl firewall policies simulate <policy> <flows>`, strict 256 KiB/128-case input and deterministic rule intent with IPv4/IPv6, TCP/UDP/ICMPv6, CIDR aliases, log nonterminal behavior and input/output default differences.
- Source-level unit/CLI tests cover source trust assumptions, malformed/duplicate cases, port/family rejection and **management-critical flow lockout warnings**; effective verdict is always **UNKNOWN** and the simulator never imports unrelated kernel/firewall state. [Full tested scope and gaps](../architecture/p0-flow-simulator.md).
- Kept real nft apply, Cloud, device enrollment and NetBird absent. P0 and G-STANDALONE remain OPEN/BLOCKED; no production firewall or existing Sentinel/Light Remote data modified.

## 2026-10-14 — HOMESV direct GitHub SSH and local control-gate code

- Inspected HOMESV's existing per-repository deploy key and confirmed `git ls-remote` initially failed solely on missing SSH GitHub `known_hosts` entry. Matched GitHub's published **Ed25519 SHA256 host fingerprint** before writing the public host key; then verified remote branch reads with the existing local key. Subsequent commits are pushed **directly from HOMESV**, not relayed through VPS-ARM. Private key bytes were not printed, copied or stored in the repo.
- Implemented [`internal/controlgate`](../../internal/controlgate/control.go) and [CLI selftest](../../internal/controlgate/selftest.go): short-lived HMAC-authenticated, local-owner-only proposals pinned to site/actor/policy SHA256 and revision; nonce+request replay cache, clock window, fail-closed store capacity, and explicit denial of signed APPLY. [Exact implementation and limitations](../architecture/p0-local-control-gate.md).
- Added negative/concurrency tests for expiry/future, forged MAC, local identity, stale/different policy, weak key, duplicate nonce/request, and attempted enforcement. Normal Go unit tests, vet, local CLI selftest, catalog and amd64/arm64 builds passed. `go test -race` **not verified**: CGO requires gcc, which is unavailable on HOMESV. Run race instrumentation in CI/build runner later. No actual IPC/OS peer auth, privileged command, Cloud call, NetBird integration or firewall modification.

## Standing limitations

No production server, WebUI, autonomous firewall, WAF or NetBird connector. Development tests do not satisfy P0 or G-STANDALONE. Future handoffs are historical, Code Map/Git/test evidence take precedence.
