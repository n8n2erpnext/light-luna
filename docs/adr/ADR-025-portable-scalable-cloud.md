# ADR-025 — Portable and Horizontally Scalable Light Luna Cloud

Status: **DECIDED product architecture and portability contract; NO CLOUD IMPLEMENTATION**
Date: 2026-10-14
Scope: optional Light Luna Cloud infrastructure, multi-node scaling, platform/VPS migration, data/identity continuity, disaster recovery
Supersedes: none. Refines [ADR-014/017 cloud directions](../12-ADRS.md) and the [Cloud Intelligence Proposal](../architecture/luna-cloud-intelligence.md); respects [ADR-023](ADR-023-dual-brain-memory-cloud.md) PROPOSED sizing and [ADR-024](ADR-024-offline-base-independent-component-updates.md) DECIDED offline client/version separation.
Primary sources: [Portable Cloud Architecture](../architecture/cloud-portability-and-scale.md), [Migration Runbook](../architecture/cloud-migration-and-recovery.md), [current Cloud & Fleet brief](../06-CLOUD-FLEET.md), [Roadmap](../11-ROADMAP.md).

## Decision

**Light Luna Cloud must be relocatable across independently owned VPSs, providers, on-prem machines and future k3s/Kubernetes environments without changing the public client protocol, resetting enrolled device identities, losing signed knowledge version history, or requiring clients to reinstall.** Build and test migration/restore as product acceptance rather than relying on the availability of one vendor or machine.

- Start with **one modular Go Cloud application** exposing stateless public Device/API endpoints, plus independently schedulable background workers, PostgreSQL as transaction source of truth, and a provider-neutral S3-compatible object-storage adapter for immutable knowledge/release artifacts.
- Initially deploy OCI images with declarative Compose profiles on a VPS. Design from day one for **multiple replicas** (API/sessionless/auth, idempotent workers) and **future k3s** with environment-scoped manifests. The Go application must never depend directly on Docker socket, Cloudflare APIs, VPS-specific paths or Kubernetes APIs for business logic.
- **PostgreSQL is the initial authoritative primary**; scale APIs and workers horizontally without creating multiple database writers or replacing PostgreSQL with vendor services. Backups include a consistent base backup + WAL archive for PITR, verified logical export for portability, and **regular restore tests**. PostgreSQL native physical/logical replication may later support low-downtime migration after correctness/compatibility review; a dump alone is not a PITR plan.
- **Object storage** uses immutable content-addressed keys, portable schemas and an explicit provider adapter. R2, MinIO or other S3-compatible implementations are options, **not locked dependencies**. Verify provider feature differences (multipart, metadata, retention, IAM, consistency, API behavior) by tests before assuming a drop-in switch. Artifact hashes/signed manifests independent of the object host.
- **Jobs must survive the move** with DB-backed transactional outbox/inbox, idempotency keys, leased worker ownership, safe retries/backoff, dead-letter policy and resumable cursors. Redis or broker is a performance option, never the sole unbacked source of business truth.
- **Device identity, signed update trust root, tenant identity and API protocol remain stable** across host migration. Network endpoints use an owner-controlled DNS name independent of a server IP; signed update roots are never replaced just because VPS moves. No client-side hardcoded VPS IP or NetBird requirement.
- **Cloud is entirely optional** to core firewall function and packet decisions. Client retains Basic Brain, base knowledge and owner policy offline. Cloud updates and opt-in telemetry continue where safely resumed after maintenance. Neither Cloud failure nor migration grants remote policy execution rights.

## Separation between scalable and stateful workloads

| Plane | Scales horizontally? | Durable truth | Move contract |
|---|---|---|---|
| Public API / auth / device manifests | Yes; nonsticky if possible | PostgreSQL/session token verification | Recreate replicas from pinned image |
| Signed immutable pack download | CDN/storage scale | object blobs + signer-approved manifest in DB | Replicate/copy blobs, checksum and URL indirection |
| Feed/advisory ingestion and normalizers | Yes via partitioned work/leases | DB cursor + raw source hash / object blobs | Restart with exactly-once business effect via idempotency, not exactly-once delivery assumption |
| Curator/scenario candidate + optional ML | Worker pool, bounded tenant budgets | reviewed DB state and versioned object artifacts | Portable job inputs and output IDs; no worker-local memory as authoritative |
| Tenant/device knowledge state and approvals | Read replicas possible; **one active writer primary** | PostgreSQL and PITR | Fence old writer before cutover |
| Signing key custody | No automatic replica/move without security approval | independent key trust/escrow/rotation process | Verify continuity/dual-control transfer or pre-rotate trust roots |

## Migration/restore invariants

1. A deployment can be created from **version-pinned OCI images, declarative environment configuration, DB schema migration version, immutable object metadata and a separately secured secret inventory**. No production secrets, raw telemetry, key material or credentials in public Git.
2. State inventory distinguishes PostgreSQL, object buckets, external DNS/TLS/CDN, signing identities, device enrollment root, cron/leases/worker offsets, monitoring, and optional analytics/model artifacts. **All stateful assets require backup and testable restore ownership.**
3. Only one writable production control-plane primary at a time. Planned move requires read/freeze or reviewed incremental DB catch-up, drain/reconcile in-flight jobs and outbox, verify object hashes, fence old services, then switch traffic. Prevent split-brain and **do not casually roll back DNS to old writers after new writes**.
4. Client polls and uploads must be idempotent/resumable: monotonic knowledge sequence, cursor handling, TTL, `Retry-After`/rate budgets, jitter, last-ack IDs, no duplicate attribution and no re-enrollment when cloud endpoint changes.
5. Keep App/Brain/module/Knowledge revisions independent, maintain unchanged component signatures/hashes across host moves. If signing root rotation is needed it is a separate controlled trust operation with offline clients considered.
6. Redundant encrypted backups are kept **outside the origin host/provider** and restore is demonstrated to a new clean environment; monitor PITR archive lag, restore errors and data integrity. Offline export/import of signed packs still functions without Cloud.
7. Do not promise zero downtime or zero data loss without actual proof. Controlled maintenance can target RPO=0 **only with a proven write fence and synchronized catch-up**; unplanned failure recovery RPO/RTO are empirical SLOs after backups and restore drills.

## Growth path / anti-premature complexity

- **Stage 0:** one VPS/Compose, 1 Go API + workers, 1 PostgreSQL writer, durable backups offsite, object storage. No required Kubernetes/multi-provider integration.
- **Stage 1:** stateless Go replicas behind load balancer, separate jobs, pool limits, dedicated storage, dashboards and tenancy quotas. DB remains single primary initially.
- **Stage 2:** PostgreSQL standby/failover and optional read replicas, proven PITR and low-downtime primary transfer, external durable queue only if observed limits justify it.
- **Stage 3:** k3s/Kubernetes for stateless workloads and workers, same OCI images/config/data formats. DB may stay externally managed/single-purpose; Kubernetes does not itself make state portable or highly available.

The scale thresholds and machine sizes remain **benchmark questions**; do not claim a given VPS can serve a fixed device count from this ADR alone.

## Required acceptance before any Cloud launch

- Restore onto a clean second host/provider with synthetic tenants, devices and objects; verify exact tenant/device IDs, no re-enrollment, DB consistency, artifact hash/signature and `appVersion/brainVersion/knowledgeRevision` continuity.
- Change of public ingress address while local client is offline, reconnect/ack/resume without losing local firewall or duplicating uploads.
- Fail mid-migration: disk-full, missing WAL segment, storage blob mismatch, signing key unavailable, worker double-exec, DB schema incompatibility, DNS cache delay and attempted split brain — **stop safely**.
- Repeat rehearsal for VPS-ARM→VPS-AMD (isolated nonproduction), Compose→Compose clean host and Compose→k3s; report measured downtime, RPO/RTO, recovery logs and resource cost. The first infrastructure portability test need not deploy Cloud to the owner’s live servers.
- Validate supply chain, image multiarch `amd64/arm64`, tenant isolation, backups encryption, secret transfer, outbound fan-out/cost limits and privacy boundaries.

## Status and relationship to roadmap

This records the **owner's architectural requirement**, not an instruction to migrate current production or stand up Cloud today. No Light Luna Cloud API, database, object store, storage backup, tenant service or k3s stack is implemented. P0–P4 standalone-first and G-STANDALONE remain unchanged; NetBird integration remains separately gated, and Cloud is optional at later stages. See [Migration Runbook](../architecture/cloud-migration-and-recovery.md) for future operator execution/verification; it is not an executable script.
