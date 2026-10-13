# Luna Local Memory — Resource Budgets, Compaction and Retention

Status: **PROPOSED sizing and behavior contract; NOT IMPLEMENTED OR BENCHMARKED**
Date: 2026-10-13
Scope: host-level disk/RAM budgets, useful learning vs unbounded accumulation, CVE cache and rollback behavior
Primary sources: [ADR-023](../adr/ADR-023-dual-brain-memory-cloud.md), [legacy source audit](../history/audits/SENTINEL_LUNA_COGNITION_AUDIT_2026-10-13.md), [Brain design](luna-brain-v2.md)

## Distinguish storage categories

1. **Go application + static assets / packages:** one deployed binary/web bundle/optional C-BPF object set; not part of managed *data* quotas.
2. **Kernel state:** nftables sets/counters, optional eBPF maps and per-CPU data use kernel RAM, *not SQLite disk*. Set cardinality, byte budgets and TTL separately.
3. **Live event buffer / feature time windows:** bounded in-memory; queue backpressure and counted drops.
4. **Local SQLite WAL:** events, per-service aggregates, provenance, baseline, pattern/decision memory, protected audit receipts; WAL+journal/temporary files count against the cap.
5. **Knowledge cache:** signed/verified selected advisories, attack patterns, scenario packs, last-known-good revision and metadata.
6. **Optional ML models / Python workers:** separate opt-in downloaded artifact budget; not in default SKU.
7. **External host logs:** journald, reverse-proxy, Docker and other applications retain their independent storage policy; Light Luna cannot pretend these bytes are included.

## Profiles (proposed initial experiments, not device requirements)

| Profile | Managed-data hard cap | Intended use |
|---|---:|---|
| Starter | **256 MiB** | Host with low traffic and tight storage; reduced history / narrow intel pack |
| Standard | **512 MiB** | Default candidate for small server/NUC if benchmark allows |
| Extended | **2 GiB** | High telemetry density or longer investigation retention, owner configured |

Target RAM ceiling requires separate measurement and enforcement. **No measured Light Luna production RSS or storage rate is available.** Changing hard cap requires owner action; do not auto-expand whenever disk is free.

### Standard 512 MiB illustrative subquotas

| Compartment | MiB |
|---|---:|
| Selected knowledge / CVE / scenarios including last-known-good | 96 |
| Short-lived event summaries and baseline rollups | 160 |
| Pattern/experience memory (deduplicated) | 80 |
| Protected audit/decision/recovery data | 96 |
| Offline sync outbox + upload cursors | 32 |
| Filesystem/SQLite WAL/migration safety reserve | 48 |
| **TOTAL** | **512** |

These are allocation **budgets**, not promised actual runtime footprint. Config/schema and emergency rollback documents must have protected reservation, never be evicted as generic caches. For WAL, cap checkpoint growth and verify power loss recovery. Backups/snapshots require explicit external budget/destination.

## Lifecycle policies (starting hypotheses only)

- Current high-granularity features: minutes–hours (e.g. 24h) capped by bytes/cardinality; **no raw HTTP request bodies or credentials**.
- Per-service / pattern 1m–10m summaries: days (e.g. 7d), downsample to 1h–1d aggregates for longer windows (e.g. 30d).
- Deduplicated learned patterns: keep only evidence-linked validated facts, capped LRU/recency + per-site quotas, degrade to tombstone + count when expired; do not promote solely due to repeated copies.
- `CVE/KEV/EPSS` cache: metadata/hash/revision plus **asset-relevant** selected records, TTL per publisher; actively retract withdrawn/invalid rules. No full historical NVD mirror per client.
- Audit: bounded append-only chunks / tamper evidence; backfill off-device when authorized. Never silently overwrite active incident or recovery records. When audit reserve critically full, disable new privileged policy mutations (once executor exists) rather than falsely logging success.
- Stale/no telemetry: verdict `INSUFFICIENT_TELEMETRY`, not green/NO_ACTION; kernel's last verified firewall policy remains functional.

## Hard-limit pressure response

```text
<80%  NORMAL  -> ingest+aggregate within per-site quotas
80-95% PRESSURE -> reduce sample, checkpoint WAL, downsample rollups, cleanup expired safe cache
>=95% CRITICAL -> pause nonessential capture/ML and uploads; keep control, audit reserve, emergency timers
100% or failure -> reject new learning writes; no silent reset; alarm + explicit disk/path diagnostics
```

Percentage boundaries are candidate operational thresholds, not measured performance guarantees. Refuse to purge last-known-good signed knowledge pack, owner policies, rollback lease or active safety/audit records just to meet a cache target. Ensure external host disk headroom, SQLite WAL close/checkpoint and actual accounting via `du` + filesystem/quota primitives in future tests.

## Cloud interaction

Download **delta packs** from a signed manifest keyed by revision/hash, only relevant to local inventory, selected threat families and hardware capabilities. Avoid duplicate records by canonical key and `source_revision`; continue with pinned verified knowledge while offline. Do not have each client fetch the complete public CVE history; a future Cloud curator does heavier feed pagination and matching. A standalone client without Cloud may fetch bounded official feeds independently, without uploading telemetry, subject to local quotas/freshness rules.

Sharing/upload is opt-in and independently rate-limited; the outbox has bounded bytes, TTL, resume cursors, dedup IDs and reset/error semantics. When offline, drop expired low-priority statistics rather than growing indefinitely; never turn outbound upload backlog into kernel packet-path dependency.

## Tests before locking budgets

Simulate 30/90/365 days of different rates and event cardinality, burst scans, repeated identical feeds, overwritten vendor advisories, malicious huge packets/JSON, offline Cloud 30d, disk-full, WAL crash/recovery, upgrade/rollback and model workers disabled. Benchmark `amd64`/ARM low-RAM devices using actual record sizes, SQLite page/WAL growth, IO amplification, BPF map memory separately. Measure loss rate and false-positive impact at saturation. Revise proposed MiB from evidence before release.
