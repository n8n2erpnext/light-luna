# Light Luna Document Library Map

Status: canonical navigation | Date: 2026-10-13
Scope: document ownership and historical path mapping

## Entry route

[Governance](LIBRARY_RULES.md) → [Project Book](LIGHT_LUNA_PROJECT_BOOK.md) → this map → [Code Map](LIGHT_LUNA_CODE_MAP.md) → [Source Catalog](SOURCE_CATALOG.md) → original code/verification. Use [Worklog](LIGHT_LUNA_WORKLOG.md) for chronology.

## Current legacy paths (do not move yet)

| Existing path | Classification / authority |
|---|---|
| [00 Vision & Brand](../00-VISION-BRAND.md) | Design decisions + positioning |
| [01 Architecture](../01-ARCHITECTURE.md) | Architecture direction, recommended stack still partly proposed |
| [02 Firewall Engine](../02-FIREWALL-ENGINE.md) | Desired policy/enforcement contract, **not implemented** |
| [03 Luna Brain](../03-LUNA-BRAIN.md) | Proposed intelligence system, **not implemented** |
| [04 WAF/Threat Intel](../04-WAF-THREAT-INTEL.md) | Future integration design, **not implemented** |
| [05 NetBird](../05-NETBIRD.md) | **Deferred research only**, post-G-STANDALONE |
| [06 Cloud & Fleet](../06-CLOUD-FLEET.md) | Future services, **not implemented** |
| [07 Security](../07-SECURITY-SAFETY.md) | Mandatory security invariants and test intent, not proven release |
| [08 Packaging](../08-PACKAGING-HARDWARE.md) | Proposed packaging/hardware targets |
| [09 WebUI/CLI/API](../09-WEBUI-CLI-API.md) | Proposed interface; limited CLI now exists |
| [10 Moon Design](../10-MOON-DESIGN.md) | Proposed visual/semantic design |
| [11 Roadmap](../11-ROADMAP.md) | **Canonical phase order** and G-STANDALONE gate |
| [12 ADR register](../12-ADRS.md) | Decisions vs proposals vs open questions |
| [ADR-020 Go/Python language roles](../adr/ADR-020-go-core-python-optional.md) | Go-native core; Python optional unprivileged ML worker; `.mjs` proposal superseded |
| [ADR-021 eBPF/XDP capability](../adr/ADR-021-ebpf-observability-xdp-optional.md) | Optional Linux tracing + early ingress mitigation; no code or live attachment; kernel coexistence rules |
| [ADR-022 Brain V2 knowledge migration](../adr/ADR-022-rebuild-luna-brain-governed-knowledge-migration.md) | Build new Go Brain; port only reviewed facts/rules; no legacy live DB/weight copy |
| [ADR-023 Dual Brain + Memory + Cloud](../adr/ADR-023-dual-brain-memory-cloud.md) | Proposed roles, model training policy, node disk caps and optional Cloud; **pending owner acceptance** |
| [ADR-024 Offline Base + Component Versions](../adr/ADR-024-offline-base-independent-component-updates.md) | DECIDED Basic Brain mandatory, offline-ready installer, App/Brain/Knowledge/module independent releases |
| [ADR-025 Portable Cloud & Scaling](../adr/ADR-025-portable-scalable-cloud.md) | DECIDED cloud portability contract; provider migration, independent scaling, state and signing continuity |
| [ADR-026 Shared Account and Remote Fleet](../adr/ADR-026-unified-account-remote-fleet-control.md) | DECIDED local Cloud identity / offline recovery / Cloud read-config and remote owner-gated changes |
| [ADR-027 Collective Threat Immunity](../adr/ADR-027-collective-threat-immunity.md) | DECIDED scenario, behavior and scoped indicator-based shared defensive learning; no global auto-ban from unverified local block |
| [ADR-028 Self-Defense & Compromise Containment](../adr/ADR-028-self-defense-host-compromise.md) | DECIDED self-integrity, Host Guardian, optional file scanner, compromised-client Cloud trust quarantine |
| [ADR-029 Observability + Multi-channel Notifications](../adr/ADR-029-local-observability-and-notifications.md) | **FUTURE ARCHITECTURE ONLY, IMPLEMENTATION FROZEN BY OWNER 2026-10-14:** Prometheus/Grafana + Telegram/Discord/Slack/SMTP; source-only Go prototypes parked; focus exclusively core, no endpoint/adapters implemented |
| [Local Observability + Notifications Architecture](../architecture/local-observability-and-notifications.md) | Contract for verified ban state/API, low-cardinality metric families, security/retention, bounded event outbox/provider routes and phase gates; **partially realized only as inert Go P0 schema/renderer** |
| [P0 Synthetic Observability Go library](../../internal/observability/synthetic.go) | **SOURCE-ONLY:** pure synthetic metric text + strict aggregation; not a live exporter, applied ban or notification provider |
| [Synthetic Observability JSON schema](../../schemas/local-observability-synthetic-v1alpha1.schema.json) | Fixed synthetic state/channel enums, no raw IP or credential fields |
| [Synthetic Observability test fixture](../../testdata/local-observability-synthetic.sample.json) | Static P0 lab data; no device or customer observations |
| [P0 Synthetic Security Events / simulated notifications](../../internal/observability/events.go) | **SOURCE-ONLY:** bounded validation and deterministic four-channel plans, all marked `SIMULATED_NOT_SENT`; no outbox or network sends |
| [Synthetic Security Event JSON schema](../../schemas/local-security-events-synthetic-v1alpha1.schema.json) | Strict batch/event contract, no IP/credential/real ban authority fields |
| [Synthetic Security Event test fixture](../../testdata/local-security-events-synthetic.sample.json) | Inert proposed/unknown events; no real device traffic |
| [Local Runtime Foundation](../architecture/local-runtime-foundation.md) | IMPLEMENTED P0 separate CLI/Brain build targets and synthetic offline Brain→inert Gate replay; no real firewall/host sensor/service |
| [Unhooked nftables Compiler](../architecture/p0-nft-unhooked-compiler.md) | IMPLEMENTED source-only nftables regular-chain text preview, no hook/kernel apply or nft -c syntax proof |
| [Synthetic P0 Flow Simulator](../architecture/p0-flow-simulator.md) | IMPLEMENTED pure-Go synthetic declared-policy matcher and admin lockout *warning*; all kernel verdicts UNKNOWN |
| [P0 Core Rule Conflict Audit](../architecture/p0-policy-rule-conflict-audit.md) | **IMPLEMENTED source/CLI only:** ICMP source family validation + bounded pure-Go precedence/shadow/overlap review; one-rule CIDR-set union coverage, no authority or nft write |
| [P0 Local Control Gate](../architecture/p0-local-control-gate.md) | IMPLEMENTED in-process HMAC/revision/nonce gate and denial-only apply selftest; HMAC/Ed25519 canonical hex encoding negative tests now PASS in real Gate Go verifier; no production owner-auth/executor |
| [P0 Linux Unix Socket IPC](../architecture/p0-linux-unix-ipc.md) | IMPLEMENTED real kernel UDS/SO_PEERCRED transport and cross-process Go tests; demo is NOT production-separated Agent/Gate and cannot enforce firewall |
| [P0 Separate Agent/Gate and Replay Journal](../architecture/p0-two-process-gate-and-journal.md) | IMPLEMENTED two separate dev binaries, same-UID temporary shared secret, bounded fsynced replay journal across restarts; **new journal fsync(file)+fsync(parent directory)** verified in nonroot development tests; NOT production service identity, power-loss proof, owner authority, executor or rollback |
| [P0 Ed25519 Agent/Gate Identity](../architecture/p0-ed25519-agent-gate-identity.md) | IMPLEMENTED asymmetric Agent private/Gate public test boundary and group IPC peer UID/GID contract; same UID test only; no cross-user authorization or executor |
| [P0 Separate Offline Reviewer Proof](../architecture/p0-offline-owner-review-separation.md) | **SOURCE + OFFLINE TESTS ONLY:** distinct Agent/reviewer signatures; separate nonroot fsynced review journal with restart replay guard, **canonical SHA-256 digest reload validation** and negative re-pinned key tests; crypto match **never human owner authorization / APPLY** |
| [P0 Owner Key Lifecycle and Revocation Contract](../architecture/p0-owner-key-lifecycle-revocation-contract.md) | **DRAFT / NOT IMPLEMENTED:** 19 key-rotation/revocation, generation anti-rollback and independent consent acceptance cases; new regressions cover existing deny-only verifier only |
| [P0 LXD Cross-UID Lab](../architecture/p0-lxd-cross-uid-lab.md) | **DEPLOYED/PASS:** LXD 5.21.8, Ubuntu 24.04 no-NIC unprivileged lab on HOMESV; real separate-UID Ed25519/Unix peer matrix passed, not kernel firewall or G-STANDALONE proof |
| [P0 LXD KVM Kernel nft Lab](../architecture/p0-lxd-kernel-nft-rollback.md) | **DEPLOYED/PASS:** KVM separate kernel, no NIC; unhooked Light Luna `nft -c` syntax + malformed-input negative, guest-only synthetic packet drop/counter and explicitly triggered rollback; automatic TTL timer NOT verified |
| [P0 VM Typed Stage and Explicit Revert](../architecture/p0-vm-synthetic-stage-revert.md) | **IMPLEMENTED/VM PASS x2 original:** inert typed manifest, synthetic staged IPv4/IPv6 TCP/UDP, guarded interrupted-journal recovery; original separate-service rollback required explicit trigger |
| [P0 VM Autonomous TTL Watchdog](../architecture/p0-vm-autonomous-ttl-watchdog.md) | **KVM lab:** primary TTL/fallback SIGKILL PASS; cooperative flock and three-phase immutable rollback observations; missing-DONE and coherent-DONE read-only classifier PASS; new 10-case simulated EIO/ENOSPC receipt fsync matrix PASS in guest temp directories; FD-held nsfs namespace blind spot now detected by read-only inventory (five offline negatives and actual no-NIC KVM verification), active journal SHA unchanged. No signed owner grant, real power-loss test or production rescue |
| [Host Guardian & Containment Architecture](../architecture/host-self-defense-and-compromise-containment.md) | Linux process/file/egress observation, integrity/attestation ceiling, safe local isolation and Cloud untrusted-ingress separation |
| [Host Integrity Report schema](../../schemas/host-integrity-report-v1alpha1.schema.json) | Design-only host status/device authentication vs trust, cloud contribution quarantine, no direct execution |
| [Synthetic compromised host fixture](../../testdata/host-integrity-report.synthetic.json) | Synthetic status with suspected compromise, no actual files/processes scanned |
| [Federated Immune Threat Intelligence](../architecture/federated-immune-threat-intelligence.md) | Local signals → Cloud vet/review → signed relevant bulletins → local early detection, retractions and consent |
| [ThreatSignal schema](../../schemas/threat-signal-v1alpha1.schema.json) | Design-only client opt-in behavior+outcome report; no proof that locally blocked source is malicious |
| [ThreatBulletin schema](../../schemas/threat-bulletin-v1alpha1.schema.json) | Design-only Cloud watch/curated scenario or actor indicator; knowledge not executable |
| [Threat signal fixture](../../testdata/threat-signal.synthetic.json) | Synthetic custom scenario/local block with UNVERIFIED label; no actor/IP disclosure |
| [Threat bulletin fixture](../../testdata/threat-bulletin.synthetic.json) | Synthetic single-source WATCH_ONLY bulletin; no global enforcement |
| [Unified Local + Cloud Account Architecture](../architecture/unified-account-cloud-fleet-control.md) | Identity, enrollment, RBAC, local Policy Gate, desired/actual sync and outbound secure agent |
| [Fleet Config Intent Schema](../../schemas/fleet-config-intent-v1alpha1.schema.json) | Design-only bounded typed remote change proposal, never shell execution |
| [Fleet Config Synthetic Fixture](../../testdata/fleet-config-intent.synthetic.json) | Synthetic Cloud owner/config intent; not live applied configuration |
| [Cloud Portability Architecture](../architecture/cloud-portability-and-scale.md) | Runtime/data boundaries, Compose→k3s, single-writer DB, signed object storage, API continuity |
| [Cloud Migration & Recovery Runbook](../architecture/cloud-migration-and-recovery.md) | Future operator cutover with backup/WAL, write fencing, identity and object verification |
| [Cloud Migration Inventory Schema](../../schemas/cloud-migration-inventory-v1alpha1.schema.json) | Static synthetic inventory/checklist contract, not a cutover authorization |
| [Cloud Migration Synthetic Fixture](../../testdata/cloud-migration-inventory.synthetic.json) | Intentionally unverified/TEST_ONLY mock, no live tenant/device data |
| [Offline Install & Component Update Lifecycle](../architecture/offline-base-component-update-lifecycle.md) | Detailed install payload, version ledger, trust, preflight, atomic rollout, rollback and test cases |
| [Component release manifest schema](../../schemas/component-release-manifest-v1alpha1.schema.json) | Design-only release-descriptor contract, not a trust verifier |
| [Component installed-state schema](../../schemas/component-install-state-v1alpha1.schema.json) | Design-only multi-version status/Cloud lifecycle contract |
| [Offline component state fixture](../../testdata/offline-component-state.example.json) | FICTIONAL disconnected install status, not a real release |
| [Local Memory Lifecycle](../architecture/luna-memory-lifecycle.md) | Proposed quotas, data expiration, evidence degradation and SQLite WAL reserve |
| [Cloud Intelligence](../architecture/luna-cloud-intelligence.md) | Proposed feed curator, signed packs, optional upload, tenant isolation, scale-ready Go/API architecture |
| [Luna Brain V2 design](../architecture/luna-brain-v2.md) | Detailed knowledge/reasoning/memory/score contracts and component topology |
| [Threat Intel & Knowledge Pipeline](../architecture/threat-intelligence-knowledge-pipeline.md) | CVE/KEV/EPSS/ATT&CK + asset-match + Cloud evidence and scenarios |
| [Legacy Sentinel cognition audit](../history/audits/SENTINEL_LUNA_COGNITION_AUDIT_2026-10-13.md) | Verified source layout, JSONL record counts, heuristic weights and migration risks |
| [Knowledge schema](../../schemas/knowledge-claim-v1alpha1.schema.json) | Proposed `KnowledgeClaimV1alpha1`; not yet validated by Go runtime |
| [Decision proposal schema](../../schemas/decision-proposal-v1alpha1.schema.json) | Proposed advisory-only proposal; no executor or code wiring |
| [13 References](../13-REFERENCES.md) | Dated external research bookmarks |
| [14 Initial handoff](../14-HANDOFF-2026-10-13.md) | Historical documentation-phase handoff; Git blocker superseded |
| [15 P0 implementation notes](../15-P0-IMPLEMENTATION-NOTES.md) | P0 code-slice handoff/notes, read with Code Map |
| [16 HOMESV P0 handoff](../16-P0-HOMESV-HANDOFF-2026-10-13.md) | Historical implementation checkpoint, not permanent code authority |

Root [README](../../README.md), [SECURITY](../../SECURITY.md), [CONTRIBUTING](../../CONTRIBUTING.md) and [go.mod](../../go.mod) are repository entry/build contracts, not free-form docs shelves.

## Future shelves (created only when required)

`docs/architecture/` for contracts and standalone phase proofs, `docs/adr/` for new ADRs, `docs/security/` for security protocols, `docs/design/`, `docs/release/` and `docs/history/` for historical agent/audit/verification materials. Do not manufacture empty placeholders or automatically move the legacy files.

## Generated source index

[Human index](SOURCE_CATALOG.md) and [machine index](source_catalog.json) include committed/working source and non-project-book documentation, with path and checksum. Current project-book files are intentionally indexed here and in its [README](README.md), not recursively in their own source catalog.

## Migration status

No legacy paths have been moved. **No implicit migration or broken-link rewrite is authorized.** Any future migration requires reference search, machine consumer analysis, old→new map, relative-link repair, verification, and a standalone Git commit.
