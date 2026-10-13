# Luna Cloud Intelligence — Federated Knowledge, Local Authority

Status: **PROPOSED logical/physical design; NO CLOUD SERVICE IMPLEMENTED**
Date: 2026-10-13
Scope: client sync, collective threat knowledge, advisory processing, tenant privacy, scaling path
Primary sources: [ADR-023](../adr/ADR-023-dual-brain-memory-cloud.md), [Brain V2](luna-brain-v2.md), [knowledge feed pipeline](threat-intelligence-knowledge-pipeline.md), [standalone roadmap](../11-ROADMAP.md)

## 1. Role separation

**Local clients**: host/edge Go engine and conditional optional Luna Brain autonomously maintain firewall, telemetry, baseline, policy, TTL, emergency recovery and signed last-good knowledge offline. They never ask Cloud whether a packet may pass. Cloud disconnect does not suspend firewall.

**Cloud intelligence plane**: retrieve and normalize official advisories, validate publisher provenance/revisions, map to technique/asset facts, deduplicate/retract, accept *opt-in* privacy-minimized field evidence, curate scenario/rule/model candidate in private staging, publish validated knowledge-pack revisions and provide operations health metadata. Cloud cannot modify local owner permissions or issue kernel commands.

**Cloud Brain?** Not required as a single live model. Start with deterministic **Intelligence Curator** workflows and governed review. Optional *separate* analysis/ML/reasoning workers may generate candidate global patterns offline, compare against holdouts and promote reviewed versions. No mandatory Cloud inference for clients, and no cross-tenant raw packet pooling.

## 2. Logical system

```text
    NVD/CISA KEV/FIRST EPSS/vendor advisories/MITRE ATT&CK
                              |
                     Feed Intake adapters
                 timeouts, fixed hosts, cursors
                              |
              Canonical Threat & Provenance Store
                      + verification/QA
                              |
       Tenant-safe evidence aggregator (OPT-IN UPLOADS)
                              |
                 Scenarios + Curator + Review
             (future optional offline AI/ML worker)
                              |
                  Approved KnowledgePack Builder
                  hash, sign, revision, expiry
                              |
                     Object Storage / CDN
                              |
               Device API: Manifest / Delta / Ack
                        TLS + device auth
                              v
       +------------------- LOCAL --------------------+
       | Verify signature / trust root / revision    |
       | Last known good; atomic swap; anti-rollback  |
       | CVE relevance + local asset provenance       |
       | Go Brain; local policy+executor isolation    |
       | Offline outbox for approved aggregate only   |
       +----------------------------------------------+
```

**Two independent channels**: (a) DOWNLINK knowledge/software updates, (b) UPLINK optional aggregate experience. Consent for downlink cannot imply consent for telemetry upload.

## 3. Minimal scalable implementation proposal

A small **modular Go Cloud service** can initially serve device enrollment/auth, signed manifest/distribution and scheduled feed processing with separate worker processes. Prefer ordinary deployable containers; do not require Kubernetes/k3s for initial users. Data roles:
- **PostgreSQL**: feed revisions, assets by tenant where consented, approvals, device metadata, revision/manifests, cursors and aggregated labels; strict application authorization and scoped queries.
- **Object storage** (S3-compatible/R2): immutable versioned signed pack files, large normalized advisories or optional model artifacts.
- **Queue/outbox**: transactional table + bounded worker first; later Redis Streams/NATS/Kafka only when measured throughput/failure recovery justify separate cluster.
- **Stateless Go API**: horizontally scalable over DB/object storage, bounded auth/rate limits and audit. Deployable on one VPS at first, later moved/scaled to stronger VPS or k3s without changing client protocol.
- **Optional analysis/research workers**: no privileged actuator and no access to customer raw data outside reviewed privacy scope; owner or governed reviewer approves new global packs.

Cloud **storage and compute** grow with tenants/retention. That is separate from bounded client storage; enforce per-tenant quotas/cost/rate limits and review ability to reject poisoning before scale.

## 4. Device→Cloud consent/identity payloads

- Device enrollment: explicit owner/device approval, per-device identity key, scoped revocable credential. Bootstrap pins a trusted update-signing root; use modern TLS; mTLS *or* signed device requests with short-lived tokens. No NetBird dependency. Device identifiers should be pseudonymous outside tenant management.
- Status-only (default minimum): software version, last knowledge revision and freshness, coarse device health, supported feature flags; avoid private IP/topology if not needed.
- **Opt-in telemetry contribution**: per-technique aggregate counts/time-bucket/site pseudonyms, measured outcome, verified owner label, evidence quality/freshness, contradictions and signatures; no raw access logs, credentials, request bodies or full software inventory without separate authorization.
- Cloud returns only scoped manifests, relevant knowledge deltas, security notices and revocations; **not executable firewall commands**. Suggested local mitigation is revalidated by local independent Policy Gate and the owner policy.

## 5. Update contract and transport

Candidate endpoints (names **proposal**, not deployed API):
- `GET /v1/intelligence/manifest?channel=stable&cursor=...`: available signed root, snapshot, targets/revision + compat/schema information.
- `GET /v1/intelligence/packs/<digest>`: immutable hashed/signed binary/compressed knowledge artifact (range/resumable download; bounded download).
- `POST /v1/telemetry/aggregate-batches`: **opt-in** bounded event/outcome upload with idempotency ID, high-water timestamp, tenant/device scope, consent version.
- `POST /v1/intelligence/acks`: status of verify/apply/reject plus pack revision (not raw secrets).
- `POST /v1/devices/health`: bounded optional health heartbeat, not packet decisions.

Every pack has `pack_id`, `schema_version`, `created_utc`, `expires_utc`, `sequence`, `minimum_client_version`, `scope`, `publisher_revision`, `sha256`, `signature`, `approved_by`, `source_digests`, `dependency_versions`, `rollback_pointer`. Require trusted-root key rotation/revocation, anti-replay/cursor monotonics, downgrade rejection, rollback/freeze/mix-and-match defenses (e.g. [TUF](https://theupdateframework.github.io/specification/v1.0.36/)) and verified atomic activation. TLS alone is insufficient signing for cache/CDN compromise.

Payloads are compressed with decompressed-size cap and strict JSON/CBOR/Protobuf schema/limits. Deploy `stable`, `candidate`, `revoked` states and compatibility matrices; a source advisory withdrawal can revoke a knowledge assertion. Downgrade of signed packs is only an explicit controlled rollback, not silent install.

## 6. Collective learning with anti-poisoning controls

A shared observation is a **hypothesis**, not proof. Verify originating tenant/devices, proxy-IP provenance, independent time windows and source diversity; dedup identical detections from cloned images and repeated reports. Attackers may create multiple fake accounts/clients and submit spoofed patterns, so count identities/tenants **not votes** without evidence. Separate raw untrusted submissions from reviewable labelled candidate knowledge, tenant/private data and approved global packs. Track counterexamples, QA/known benign scanners, false-positive resolution and model drift. A review cannot grant enforcement permissions.

Global models may produce technique likelihood *only after independently labelled calibration*. Never portray EPSS (internet-wide CVE exploitation probability) as local-host compromise probability. Collect opt-in outcome labels and maintain data deletion/revocation capabilities.

## 7. Failure, offline and scale

- Client disconnect: continue with last **trusted, still-valid** knowledge, tag stale intel as stale, preserve nftables rules and TTL; no global 'green' from stale data. If pack expired, do not silently treat it as current.
- Feed intake failure: quarantine incomplete feed and continue prior properly labelled snapshot; no mass empty feed.
- Cloud compromise: signed pack verifier, key rotation, staged distribution, quarantine/rollback; no ability to run arbitrary commands on hosts.
- Client full disk: bounded outbox/knowledge pack and last-safe reserve, no uncontrolled retry.
- Privacy and tenancy: end-to-end auth, export/delete on owner request, encryption, per-tenant separation, poison review, rate limit.
- Data plane offline: no packet cloud roundtrip. Region migration: stateless API + PostgreSQL controlled backup + object storage versioning and reproducible deployment.
- Early stage do NOT deploy cloud or replace optional NetBird. Do not mix public threat feeds and private client knowledge in the same collection without typed scope.

## Cloud relocation, replica scaling and keys (owner architecture decision)

[ADR-025](../adr/ADR-025-portable-scalable-cloud.md) requires a provider-neutral, reproducibly deployed and **tested relocatable** Cloud. [Portability & scale](cloud-portability-and-scale.md) defines the single-writer PostgreSQL + immutable object storage and stateless Go API/worker scaling boundaries; [migration and disaster recovery runbook](cloud-migration-and-recovery.md) defines write fencing, PITR and device/signing trust continuity. A future signed KnowledgePack should be served through a stable owner-controlled DNS and content hash, independent of a specific provider, VPS IP, bucket host or orchestrator. Cloud service migration **does not change** `appVersion`, `brainVersion` or `knowledgeRevision` by itself. These are contracts, not running services.

## 8. Milestone and evidence to ship

P0–P4: freeze schemas/extension interfaces but **no Cloud runtime dependency**. P2 build local official-feed reader/matcher in bounded shadow mode; P3 guarded local response with owner approval; P4 Edge offline/peer isolation; P7 Cloud intelligence exchange if separately approved. Bench dry run: offline 30d, signed-pack rollback/revocation, private telemetry leak attempt, sybil/poison attack, 100/1000/10000 simulated clients, DB queue crash, rate limiting, Cloud region move and expired CVE updates.

No backend, endpoint, device provisioning, sign key, model service or production Cloud deployment exists under this design yet.
