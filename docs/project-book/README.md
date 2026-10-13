# Light Luna Project Book

Status: canonical navigation | Date: 2026-10-13
Scope: governed documentation entry point for engineers and AI agents

Read in this order, **before editing any documentation or code**:

1. [LIBRARY_RULES.md](LIBRARY_RULES.md) — mandatory source authority, filing, traceability, and update rules.
2. [LIGHT_LUNA_PROJECT_BOOK.md](LIGHT_LUNA_PROJECT_BOOK.md) — verified project overview, decisions, P0 evidence and open gates.
3. [DOCUMENT_LIBRARY_MAP.md](DOCUMENT_LIBRARY_MAP.md) — index of canonical, proposed, research, and historical documents.
4. [LIGHT_LUNA_CODE_MAP.md](LIGHT_LUNA_CODE_MAP.md) — real entry points, call graph, authority boundaries and source paths.
5. [SOURCE_CATALOG.md](SOURCE_CATALOG.md) / [source_catalog.json](source_catalog.json) — source bookmarks, exact paths and SHA-256.
6. [LIGHT_LUNA_WORKLOG.md](LIGHT_LUNA_WORKLOG.md) — durable chronological changes.
7. [PROJECT_TRUTH_STATUS.json](PROJECT_TRUTH_STATUS.json) — machine-readable snapshot and unresolved gates.
8. [DOCUMENT_TEMPLATE.md](DOCUMENT_TEMPLATE.md) — use when a new durable document is justified.
9. [ADR-020](../adr/ADR-020-go-core-python-optional.md) — Go-native core plus optional, isolated Python ML worker (architecture decision; not implemented).
10. [ADR-021](../adr/ADR-021-ebpf-observability-xdp-optional.md) — optional future eBPF observation/XDP mitigation; not active.
11. [ADR-022](../adr/ADR-022-rebuild-luna-brain-governed-knowledge-migration.md) — rebuild Brain V2; port reviewed knowledge rather than cloning live data/imagined model weights.
12. [Brain V2](../architecture/luna-brain-v2.md) and [Threat Knowledge Pipeline](../architecture/threat-intelligence-knowledge-pipeline.md) — proposed, independently reviewable component contracts.
13. [Legacy source audit](../history/audits/SENTINEL_LUNA_COGNITION_AUDIT_2026-10-13.md) — VPS-ARM observations, heuristic numbers and implementation limits.
14. [ADR-023](../adr/ADR-023-dual-brain-memory-cloud.md) — proposed dual-role brain, bounded learning and optional Cloud; **not yet decided or built**.
15. [Local Memory](../architecture/luna-memory-lifecycle.md) and [Cloud Intelligence](../architecture/luna-cloud-intelligence.md) — proposed resource and sync contracts.
16. [ADR-024](../adr/ADR-024-offline-base-independent-component-updates.md) and [offline component lifecycle](../architecture/offline-base-component-update-lifecycle.md) — DECIDED release design: bundled Basic Brain plus independent App/Brain/Knowledge/module updates; not yet implemented.
17. [../../README.md](../../README.md) — public product introduction and constraints.

## Source precedence

**Exact code at Git SHA + test evidence + actual runtime status** > current validated contracts > accepted ADRs > contemporary design/handoff > archived research. These levels are not interchangeable. A passing unit test is not evidence of a working firewall.

## Continuity

The local, overwrite-only `.luna/CURRENT_CHECKPOINT.json` (ignored by Git) may describe the current session state. It is **not** canonical evidence and is never a substitute for Git, Code Map or release/acceptance artifacts.

## Publication

Repository: [n8n2erpnext/light-luna](https://github.com/n8n2erpnext/light-luna), intentionally **public** per owner's 2026-10-13 instruction. Do not commit secrets, live firewall exports, personal IP histories, agent memory databases or raw telemetry.
