# 06 — Cloud, Fleet & Trust

## Independence model

Light Luna installs, loads policy, enforces manual rules, expires temporary bans, audits actions and recovers **fully offline**. Cloud is optional and opt-in. Local ownership is authoritative for emergency restores. Loss of Cloud MUST NOT silently turn a local firewall permissive or halt scheduled unblock.

## Independently versioned Cloud and offline distribution (ADR-024)

[ADR-024](adr/ADR-024-offline-base-independent-component-updates.md) requires every install to contain a Basic Go Brain and curated base knowledge before optional Cloud enrollment. [Update lifecycle](architecture/offline-base-component-update-lifecycle.md) separates `appVersion`, `brainVersion`, `executorVersion`, optional module versions, `knowledgeRevision` and optional ML model version. Cloud distributes signed, compatible per-track updates/deltas under per-component owner policy; an offline signed bundle uses the **same trust verifier**. No background auto-updates to privileged executor or XDP module without a reviewed maintenance/rollback gate. The current code has no updater, Brain binary or Cloud sync.

## One account across local and Cloud Console; scoped device management

[ADR-026](adr/ADR-026-unified-account-remote-fleet-control.md) makes the owner's **Cloud account the standard identity for an enrolled local WebUI**, via OIDC redirect and local scoped session, and provides the Cloud Console with redacted **actual** status/config and typed **desired** configuration proposals. Local bootstrap/break-glass remains possible completely offline. Devices generate independent keys, enroll with two-sided owner approval, and poll an outbound authenticated Cloud queue, with no public inbound port or NetBird dependency. Remote access is an explicit opt-in distinct from Cloud enrollment and telemetry sharing. Every change is scoped, revision-checked, locally authorized by Policy Gate and acknowledged only after actual apply/rollback. Cloud logins do not confer root, protected admin policy changes or arbitrary shell. See [Unified Account & Fleet Management](architecture/unified-account-cloud-fleet-control.md). This is design, not a deployed portal/agent.

## Proposed Cloud services

1. **Device enrollment and ownership:** owner-activated short-lived invitation; device-bound keypair; certificate pinning/rotation; no automatic adoption of all peers found by NetBird.
2. **Fleet inventory and health:** per-node software version, code signature, last-seen, last policy revision, capabilities and protections, summarized security posture.
3. **Policy distribution:** signed bundles with site/tenant scopes, canary ring, staged diff, fail-safe rollback, audit history and per-device approval setting.
4. **Threat intelligence:** signed snapshots with freshness, source confidence, TTL, application selectors and privacy-safe event summaries; local context must decide relevance.
5. **Updates:** signed metadata and artifacts, SBOM and provenance, compatibility gates, minimum safe version, staged rollout, rollback on failed health.
6. **Operator notification and support:** alerts with receipts/delivery states, per-tenant RBAC and separation of security and billing data.

## Multi-tenant controls

Tenant+device keys, row-level isolation, encrypted secrets at rest, scoped service credentials, audit logs, signed replay-protected command envelopes, two-person approval for dangerous global policy distribution, rate/cardinality limits and clear billing/quotas. Cloud updates should never be accepted merely because they arrive via a trusted VPN; they require verified signing identity and policy version.

## Crowd intelligence privacy

Default **no upload of raw traffic payloads, secrets, private IPs, full request URLs or customer usernames**. Opt-in aggregated indicators with pseudonymization/minimum count thresholds, time-window bucketing, age-out policy and confidence; explicit opt-out and deletion lifecycle. Threat feeds and telemetry are untrusted input; do not let supplied content influence command invocation.

## Owner-decided Cloud migration & scale foundation (ADR-025)

[ADR-025](adr/ADR-025-portable-scalable-cloud.md) requires the optional Cloud to move VPS-ARM→VPS-AMD, other VPS providers or future k3s **without reinstalling clients, changing signed release trust, losing tenant/device identity or modifying public protocol**. [Portable architecture](architecture/cloud-portability-and-scale.md) chooses a **modular Go API + independently scalable workers**, PostgreSQL as one durable transaction writer, S3-compatible content-addressed signed objects and transactional outbox. Use pinned multiarch OCI images and Compose initially; add API replicas and move to k3s only when measured load justifies it. [Migration Runbook](architecture/cloud-migration-and-recovery.md) specifies off-host encrypted backup/WAL PITR, checksum-verified object copy, write-fencing, key custody, signed-pack continuity and measured restore. A cloud move is **never** a remote firewall rule change; original local Brain/policy remain independent.

## Scale and portability

Keep control plane stateless where possible, with transactional persistence, indexed event summaries, TTL-based data rotation, queue + backpressure, per-tenant quotas and idempotency. Define operator-controlled offline export/import from day one. Horizontal scale/K3s migration should be possible without redesigning device protocol; don't introduce K3s as a dependency for v1.
