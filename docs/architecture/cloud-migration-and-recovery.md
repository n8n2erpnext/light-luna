# Light Luna Cloud Migration and Disaster Recovery — Operator Runbook

Status: **PROPOSED FUTURE RUNBOOK, NOT PRODUCTION-READY NOR EXECUTED**
Date: 2026-10-14
Scope: planned VPS-to-VPS / provider-to-provider / Compose-to-k3s moves and disaster recovery for optional Light Luna Cloud
Primary sources: [ADR-025](../adr/ADR-025-portable-scalable-cloud.md), [portable architecture](cloud-portability-and-scale.md), [Cloud identity and trust](luna-cloud-intelligence.md), [client update requirements](offline-base-component-update-lifecycle.md)
External technical references: [PostgreSQL PITR](https://www.postgresql.org/docs/current/continuous-archiving.html), [PostgreSQL logical replication](https://www.postgresql.org/docs/current/logical-replication.html), [Kubernetes declarative workloads](https://kubernetes.io/docs/concepts/overview/working-with-objects/), [The Update Framework](https://theupdateframework.github.io/specification/v1.0.36/)

## 0. Read before touching anything

**DO NOT RUN THIS AS A SHELL SCRIPT.** The workflow describes future, separately authorized migration. Current Light Luna Cloud does not exist; this runbook does not authorize changes to VPS-ARM, VPS-AMD, existing Light Remote, ERPNext, NetBird, DNS, database, VM host/firewall or real customers. Test first on isolated synthetic fixtures.

The owner-defined endpoint must remain stable for clients. Production database is **single writer**. A 'copy containers + DNS switch' without DB WAL, object store, release signing and in-flight job reconciliation is **not** a successful migration.

## 1. Migration model selection

- **Cold planned move (default first):** announce brief write-maintenance window, block new mutating Cloud API requests, drain workers, stop DB writes, consistent backup/restore, verify checksums and client contract before DNS/LB switch. Simpler, more downtime, easier to audit.
- **Warm planned move (later, tested):** seed target PostgreSQL with physical streaming or compatible logical replication; sync object hashes and jobs, run target read-only shadow checks, **fence** source writes, catch up replication and verify checkpoint/LSN, promote *one* target primary, switch routing. Higher complexity; no guaranteed zero downtime.
- **Disaster recovery:** restore last verified offsite basebackup + continuous WAL to last usable recovery point; restore objects and identity/signer state. Communicate actual RPO and possible lost acknowledgments; reconcile idempotent client uplinks safely. A disaster cannot promise RPO=0 by default.

Maintain exact schema/major versions, extensions and compatibility during physical replication; logical replication has limitations around DDL/sequences, initial sync, replication slots and conflicts. The runbook must choose tested method with evidence, not assume any are automatic.

## 2. Pre-migration inventory (export sanitized record)

Record, for both sites, without secrets in Git:

| Group | Required inventory |
|---|---|
| Workloads | container name, OCI digest, source Git SHA, semantic service revision, expected owner, CPU/RAM cap, readiness path |
| PostgreSQL | major/minor version, schema migration head, extensions, table/index scope, encrypted backup/WAL archive location, lag, basebackup verified |
| Object | provider endpoint type, bucket/key-space, inventory counts and SHA-256 digest verification, retention and object revision manifest |
| Cryptographic trust | signing root key IDs and custody policy, online pack signing service, encrypted escrow escrow location, certificates, pending revocations |
| Devices/tenants | counts and stable IDs, device token/key rotation and revocation audit, consent status, last manifest/ack sequence |
| Async work | outstanding leases, failed/retry/dead-letter jobs, feed cursors, outbox/inbox high-water marks, aggregation dedup |
| Network | canonical DNS name, cert renewal responsibility, reverse proxy/WAF/CDN and egress allowlist, TTL current |
| Recovery | primary and destination backup integrity, tested restore time, rollback gate, communications and operator signoff |

The inventory/receipt is the source of truth for controlled cutover. **Do not record actual credentials, unredacted customer IPs or private logs** in public reports.

## 3. Prepare target (without production switch)

1. Acquire fresh target environment with compatible OS/CPU architecture, storage, TLS ingress and isolation; use pinned IaC/Compose or future Kubernetes declarations. Deploy the *same signed/pinned OCI image digests* to target and verify health. No device enrollment/tenant split.
2. Confirm target schema and service versions can accept source DB and object data. Run DB migrations only through a single explicit schema migration process with expand/contract plan. No auto run of incompatible irreversible DDL.
3. Restore from a **verified consistent database backup** and use WAL replay where PITR applies, not `pg_dump` alone for recent mutations. Verify critical table counts, FK consistency, consent flags, revoked token state, release revision heads, tenant isolation and audit journal high-water marks.
4. Copy immutable objects into target S3-compatible storage, independently verify bytes/digests/count and signed manifest history. Keep **old and new object destinations available** until client offline/retry grace and evidence-based cutover expiry; don't delete old storage prematurely.
5. Prepare new API/worker without accepting tenant writes; test authenticated device requests and knowledge pack checksum/signature through a private staging hostname/mTLS test environment. No production key material in logs.
6. Verify KMS/signing continuity *using an approved safe key management procedure*: offline trust root stays protected, online signing keys transferred/rotated under an explicit operator policy. Never automatically dump private signing keys into plain backups or add them to Docker env history.
7. Set up offsite **new-target** backups/monitoring before turning on target writes. A successful migration without backup from the new site is incomplete.

## 4. Freeze, final sync and cutover

1. Announce defined maintenance window and expected read-only/unavailable Cloud API behavior; local device firewall continues.
2. Mark source `MAINTENANCE_NO_WRITES`: reject/new-queue mutating requests as planned, drain worker leases and pause schedule publishers; ensure writes to DB/outbox and object pointer publication have quiesced.
3. Capture source DB transaction checkpoint/LSN (or a verified consistent final dump for cold move) and jobs/outbox/inbox cursor high-water marks. Replicate remaining changes. **Prove no missing WAL/changes and no unexpected target writes**.
4. FENCE old primary/worker/ingress from write access using tested technical enforcement (not merely an operator note). Prove target alone is allowed to write. For warm move, promote target exactly once after lag=0 and transactional continuity checks.
5. Verify signed current release/knowledge manifest and all referenced objects on destination, including hash, expiry, revoked releases, chosen compatibility matrix and old trusted root IDs.
6. Shift stable DNS/LB endpoint to new ingress. Handle old clients with cached DNS/TLS and retry semantics; avoid auth/redirect assumptions and measure actual cache retention.
7. Canary synthetic devices → controlled operator test clients → wider users. Validate revocation, per-tenant read/write, opt-in upload idempotency, device ack sequence, feed job cursors, signed knowledge delta and client no-reinstall.
8. Turn on target workers/schedulers **after** verifying only one active lease owner/fenced scheduler. Ramp quotas and compare backlog/lag.
9. Record evidence: timestamp, operator approvals, old/new build digests, DB checkpoint, object digests, key version, observed reconnect errors, p95/p99, downtime window and data-loss estimate.

Do not treat a successful HTTP 200 from `/health` as proof of data/identity correctness.

## 5. Rollback and split-brain policy

- **Before target accepts writes:** the old environment may be reinstated if its database is still authoritative, once DNS/LB and source fencing are controlled; clearly prevent clients from talking to two writers.
- **After target accepts writes:** do **not** flip DNS back to an old un-reconciled PostgreSQL writer. First choose a new coordinated migration direction, export/synchronize target writes or restore target to verified safe state; sign off the data correctness/replay plan and capture audit.
- Lost or stale acknowledgements are retried with idempotency keys and scoped sequence/cursors. Potentially revoked devices must **stay revoked**, even after recovery; never allow old backups to resurrect a revoked signing/token authority without review.
- A broken signing chain or missing last-known-good pack is a hard stop. Cloud is allowed to remain offline; clients still use local signed packs and their independent firewall.
- If replicated blobs aren't durable or wrong hash, keep previously trusted manifest active; do not publish broken target revision.
- If new Cloud has unhealthy tenants/backups/keys, freeze rollout; escalate incident with exact gap rather than pretending rollback is trivial.

## 6. Disaster backup requirements

- PostgreSQL: schedule encrypted base backups plus continuously archived WAL to a separate failure domain; verify availability and restore as part of automated/offline drills. Backup DB logical dumps can aid migration but **do not replace WAL/PITR**. Monitor archive failures and runaway WAL disk usage.
- Objects: immutable versioned known-good packs/models/release targets copied or independently stored with checksums; archive manifest mapping DB revision to object digest and signing identity.
- Secrets: separate encrypted credential inventory/escrow with access and rotation controls. Exclude root private keys from standard DB backups. Recovery keys, tenant/device issuer identifiers and update-signing trust chain must be recoverable by an explicit secure procedure.
- Deployment: source-controlled pinned IaC/manifests, sanitized environment template, schema migrations, config version, SBOM/container digests and tested recreation steps.
- Keep at least one **offsite** restorable copy that is unavailable to a compromise of the production VPS account. Use backup immutability or independent credentials where feasible, and verify deletes/rotation retention.
- Recovery inventory includes all authoritative pieces **and external systems**: DNS ownership, TLS issuer, object endpoints, email/notifications, scheduled feeds and signed update publishing.

## 7. Acceptance tests and metrics (targets not yet set)

A migration trial passes **only if**:
- No client reinstall or account/device re-enrollment, same device/tenant/consent/revocation state, stable public API protocol and local offline firewall unaffected.
- Authoritative DB integrity, outbox dedup, feed cursor continuity and object checksum/pack signature cross-checks pass before reopening writers.
- Repeated signed pack downloads work from cache/staging/new storage without changing App or Brain version unexpectedly.
- No two primaries writable together in failure injection and rollback never resuscitates a revoked identity.
- Restore another VPS/provider from encrypted backup with **no original application process/filesystem dependency**; record time and measured data loss.
- Test `Compose -> Compose`, `ARM -> AMD`, `Compose -> k3s` with synthetic data and observe release same source SHA/image digest. Test source power off mid-move, network loss, stale DNS, duplicate uploader, full disk, corrupted blob, missing WAL and signer unavailable.

Suggested measurable fields (not promised numbers): **RPO** (minutes/successfully retained transactions), **RTO** (duration from unavailable to all required functions recovered), write freeze duration, device reconciliation success %, p95 API latency, DB lag, object integrity failures, tenant leakage count, queue DLQ count, stale signing metadata, restore test last-run date. Define actual SLOs only after first recovery exercises.

## 8. Safe staged readiness tasks

Even before Cloud runtime exists, review portability in code review: no hard-coded provider or host path in core code; stable versioned endpoints; signed content-addressed packs; event idempotency; independent signer; portable config and schema migration. Actual infrastructure provisioning, changing DNS/keys and production migration are **future separate operator tasks** requiring owner approval and dedicated rehearsal results.
