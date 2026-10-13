# Cloud Portability, Scale-Out and State Boundaries

Status: **architecture contract from DECIDED [ADR-025](../adr/ADR-025-portable-scalable-cloud.md); NOT IMPLEMENTED**
Date: 2026-10-14
Scope: optional Light Luna Cloud infra/platform portability, public device API continuity, multi-tenant scale model
Sources: [Cloud intelligence pipeline](luna-cloud-intelligence.md), [Cloud & Fleet](../06-CLOUD-FLEET.md), [Offline component lifecycle](offline-base-component-update-lifecycle.md), [migration/recovery](cloud-migration-and-recovery.md)
Evidence: source design only; no cloud stack, endpoint, database or HA benchmark exists in the repo.

## Goals and architectural boundaries

- Run the same business logic and API on VPS-ARM, VPS-AMD, x86/ARM dedicated servers, multiple VPS nodes, and future k3s. Container images, migrations and release manifest are pinned and reproducible.
- Client always addresses **stable owner-controlled DNS and API version** rather than IP/NetBird device ID/hostnames. Device enrollment and per-device signing identity survive a machine move; device secrets never live in container image or public repo.
- Keep packet processing and emergency firewall recovery wholly local. Cloud is an optional intelligence/distribution/control *management* plane, not synchronous per-packet enforcement.
- Make vertical scale first an option (more CPU/RAM on one VPS), while stateless replicas, workers, database and blob storage can split into separate nodes without API redesign.
- **Do not** promise user-count capacity, zero downtime, perfect portability or absolute RPO/RTO until measured on restored/migrated test systems.

## 1. Logical architecture (vendor-neutral)

```text
                    Devices (offline capable Basic Brain)
                      HTTPS client identity + signed packs
                                    |
                     Owner DNS name + TLS ingress
                       (edge/CDN optional)
                                    |
             +---------- Stateless Go API -----------+
             | auth, tenant, device identity         |
             | version manifests, knowledge index   |
             | status/ack, opt-in summaries          |
             +----------------+----------------------+
                              |
                    PostgreSQL (1 writer)
                   + tenants, device state
                   + revision/approval journals
                   + feed cursors, outbox/inbox jobs
                     /                       \
             Leased Go workers           Object API (S3 subset)
           advisory/ingest/curator      content-addressed signed packs
           aggregate/review/sign       immutable models/artifacts
                     |                       |
         optional isolated Python        Object store: R2, MinIO,
              research/ML worker         other tested provider
```

Components must not rely on the same filesystem, process local queue, in-memory session or node-local cron state to maintain business correctness. `SIGTERM` and worker lease recovery need to be safe in container replacement.

## 2. Code/module deployment design

- **Go API binary**: stateless auth/policyless intel API; readiness checks validate ability to serve without forcibly rewriting DB.
- **Go job-worker binary**: same codebase initially acceptable, different `role` config and container entrypoint. Each job has tenant scope, bounded payload, lease and idempotency key, lease extension and DLQ.
- **Go migration tool**: versioned database schema, explicit expand/contract transitions, forward/back compatibility across rolling versions. Execute once with migration lock; no uncontrolled automatic migrations on every API replica boot.
- **Go pack-curator**: consumes normalized reviewed facts, builds immutable serialized pack and publishes signed manifest in an atomic DB transaction after blob availability/hash verification.
- **Optional Python research/model service**: separate bounded worker, no signer/DB admin credentials, no authority to approve knowledge or push kernel commands to clients.
- **Signing service**: cryptographic boundary distinct from threat-feed ingest and research; key escrow/rotation/operator process, approval workflow; never place root private keys in Go API environment.

Initially one repository/deployable product can include multiple binaries; do not multiply repositories/microservices unless independent scaling/security duties require it.

## 3. Data ownership and replication

| Data | Authoritative location | Portability mechanism | Important caveat |
|---|---|---|---|
| Devices, tenants, identity bindings, consent, review/approvals | PostgreSQL | Verified logical export/restore + continuous physical base backup/WAL | Single active writer and private tenant boundaries |
| Pack manifest, release/knowledge revision, version ABI graph | PostgreSQL + signed metadata in object store | DB-consistent snapshot + verify signed object digest | Do not publish pointer before blob is durable and verified |
| Immutable KnowledgePack and model artifact bytes | S3-compatible object provider | Keyed SHA-256; inventory, checksummed copy + retain two generations | API feature differences must be tested |
| Job/notification retry state | DB transactional outbox/inbox | Replay idempotently after restore with lease fencing | Never rely on Redis-only queue offsets as sole record |
| Analytics/aggregate opt-in events | DB partitions / optional later analytics store | TTL/batch export with tenant segregation | No raw sensitive payload by default |
| TLS certificates and public DNS | Managed operator infrastructure | Exportable documented configuration & DNS control | Certificate/ACME/account portability may require re-issue |
| Update signing root/online keys | Managed secure signing custody | Separately encrypted escrow + root rotation procedure | **Never commit/copy unencrypted signing keys into backups** |
| Deployment configuration | Git version-controlled templates | Environment variables/secret-file mounts | Public Git contains NO deployment credentials |

PostgreSQL architecture:
- Use schema-owned migrations `schema_migrations` with test fixtures and explicit forward/back compatibility.
- Base backup + complete archived WAL (outside source VPS) supports point-in-time recovery; `pg_dump` is a portable logical export **but not itself a PITR backup**.
- Later streaming replicas can provide HA read/standby; logical replication may reduce cross-version/platform migration downtime but demands sequence, DDL, conflicts and role/extension checks and **write fencing**. No casual bidirectional logical replication.
- Partition high-volume aggregate events; bound ingest per tenant and default TTL, avoid storing every packet or forever-retained HTTP raw logs.
- Transactions cover idempotency receipt/outbox insertion with business mutation. Workers treat deliveries as at-least-once and use dedup on effects.
- Connection pools enforce global max connections across all replicas; a proxy/pooler is optional based on load measurement.

Object store:
- SDK uses lowest required S3-compatible feature subset (Get, Put with verified checksum, Head, List, Delete only when retention permits, versioning if tested); no Cloudflare-specific API in domain logic.
- Stable logical object key `packs/sha256/<digest>` and optional metadata; immutability enforced by application and object permissions. CDN is a cache of signed bytes, not source of truth.
- Published manifests can use alternative mirrors. Clients verify signatures and artifact hashes independently of URL, CDN or TLS edge. During migration both old/new object stores may serve identical artifacts while DB pointer stays one authoritative revision.

## 4. Config and workload packaging

**Build once, deploy anywhere**:
- Version-pinned multiarch OCI images with reproducible build/SBOM/provenance and build-time source SHA; separate `linux/amd64` and `linux/arm64` images under a manifest list.
- Deploy with Docker Compose initially (`deploy/compose/` in future), with typed environment config, secret files, readiness probes, resource limits, unprivileged identities and named volumes *only for explicitly stateful components*. No app code mount, Docker socket or host network needed.
- Keep a **k3s/Helm or Kustomize** overlay as future validated deployment adapter (`deploy/k8s/`), not a second implementation of business logic. HTTP health/metrics and configuration fields remain equivalent.
- Config keys are transport/database/object-store interfaces, not provider IDs. Example candidate `LUNA_PUBLIC_API_BASE_URL`, `DB_DSN_FILE`, `OBJECT_ENDPOINT`, `OBJECT_BUCKET`, `TRUST_ROOT_BUNDLE_PATH`, `DEPLOYMENT_ID`; these are *proposed*, not current environment variables.
- No hardcoded `/home/ubuntu`, VPS IP, fixed Docker bridge, provider-specific object bucket URL or NetBird connection required.
- Infrastructure-as-code (e.g. OpenTofu/Terraform) is optional and platform adapters may vary; no state store credentials committed to Git. Maintain an exportable inventory of DNS, provider resources, storage policy and backup targets.
- Ephemeral writable paths only for cache/tmp; applications use bounded tmp and do not depend on local disks for irrecoverable state.

## 5. Scale stages (and why they do not imply Kubernetes now)

| Stage | Shape | Trigger | Correctness gate |
|---|---|---|---|
| S0 initial VPS | Compose: one API, worker, primary PostgreSQL and external/included S3 adapter | First Cloud customers; measure | Offsite backup + clean restore, TLS + signed packs |
| S1 separated functions | Multiple stateless API replicas behind LB; scale workers/ingest separately | CPU latency, queue lag, utilization under real measurement | Idempotent retries and DB pool capacity |
| S2 resilient state | DB dedicated machine/managed Postgres + standby, dedicated object store, data plane scaling | DB/IO/recovery risks | PITR+failover rehearsal, single-writer fencing |
| S3 k3s portability | Same OCI image and protocol, declarative workloads and worker HPA | Operational benefits exceed overhead | Clean Compose→k3s move with DB/object continuity |
| S4 regional/multi-region | Explicit DNS routing, CDN signed object mirrors, regional API+auth strategy | Verified latency/regional availability need | Single writer or carefully designed region sharding, no invented active-active guarantees |

Cloud service cost and CPU/GB/pps throughput must be measured per tenant, API p95/p99, queue lag, DB commits, R2 egress/storage, knowledge pack revisions and incident pattern cardinality before adding complexity.

## 6. Fixed client update protocol through server migration

- `/v1` API routes and signed pack schema remain stable; maintain explicit protocol/compatibility ranges independent of backend platform.
- Devices persist `device_id` (opaque), tenant binding, credential, trust root, knowledge revision and last committed sync cursor **locally**, not tied to VPS hardware ID.
- Server uses a stable canonical origin hostname owned by project. Client resolves DNS normally and supports cert rotation; never require reinstall after A-record/provider change.
- Retry uses bounded exponential backoff/jitter, idempotency key + ack cursor, good HTTP response codes incl. 429/503 + `Retry-After`; no unbounded offline outbox.
- Auth key rotation does not require mass re-enrollment. Revoked tokens/devices stay revoked after restore. Clock skew/expiry errors degrade intelligence, not local firewall.
- Root signing trust is not equivalent to Cloud API TLS; Cloud compromise cannot silently issue a new trusted Brain/App binary.
- Cloud version number remains separately managed from App/Brain/model/Knowledge revision; e.g. migrating S0→S3 without changing any client component version is valid.

## 7. Multi-tenancy and privileged boundaries

- All read/write endpoints and DB operations have tenant-scoped authorization; workers enforce scope not merely a request field. Per-tenant quotas for API, device count, storage, feed-processing and cloud learning.
- Data contributed to Cloud learning is opt-in, bounded, pseudonymized and quality-labelled; default device health shouldn't include internal topology/raw logs.
- Threat feeds, compromised clients and model-generated advice are untrusted; global intelligence review/approval is separate from publication-signing roles.
- No public anonymous bulk CVE scrape endpoint and no unbounded fanout from one advisory to all clients. Efficient cached signed manifests/deltas and gradual rollout rings with backpressure.
- Signing root key custody, trusted publisher policies and product incident response are provider-independent control assets; migration must not invalidate existing enrolled clients.

## 8. Portability verification matrix

1. Reproducible clean Compose install and health on isolated ARM64/x64 VM.
2. New API/worker node from image alone with zero copied application state.
3. Simulated 100/1,000/10,000 clients, signed pack checks, ack resumption and quotas (NO performance claim until benchmarked).
4. Staged migration from source DB/object → target with synced cursors, no lost/duplicate authorization, verified object SHA, immutable published revisions and signed root continuity.
5. Clean host migration plus k3s adoption uses unchanged device API URL/protocol and update trust.
6. Failure injection: DB outage, WAL gap, partial object copy, two active writers, long-lived worker leases, stolen/cloud signing token, old DNS caches and offline reconnect.
7. Post-cutover monitor release signing, recurring jobs, no secret leakage, no tenant crossover, endpoint redirection and backups from **new** site.

See [Migration/Recovery](cloud-migration-and-recovery.md) for safe procedure. This is a design-stage contract, not an operationally tested claim.
