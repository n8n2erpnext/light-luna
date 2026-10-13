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

## Standing limitations

No production server, WebUI, autonomous firewall, WAF or NetBird connector. Development tests do not satisfy P0 or G-STANDALONE. Future handoffs are historical, Code Map/Git/test evidence take precedence.
