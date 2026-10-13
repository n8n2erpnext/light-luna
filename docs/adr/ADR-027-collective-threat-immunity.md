# ADR-027 — Federated Threat Memory and Collective Defensive Immunity

Status: **DECIDED architectural direction; enforcement thresholds and performance targets PROPOSED; NOT IMPLEMENTED**
Date: 2026-10-14
Scope: Light Luna local-to-Cloud signals, threat actors/behavior/scenarios, community intelligence, fast distribution, privacy and local defensive authority
Primary sources: [Adaptive immune network architecture](../architecture/federated-immune-threat-intelligence.md), [CVE/knowledge pipeline](../architecture/threat-intelligence-knowledge-pipeline.md), [Brain V2](../architecture/luna-brain-v2.md), [Cloud architecture](../architecture/cloud-portability-and-scale.md), [Security](../07-SECURITY-SAFETY.md)
Related decisions: [ADR-022](ADR-022-rebuild-luna-brain-governed-knowledge-migration.md), [ADR-023](ADR-023-dual-brain-memory-cloud.md), [ADR-024](ADR-024-offline-base-independent-component-updates.md), [ADR-026](ADR-026-unified-account-remote-fleet-control.md)

## Owner requirement

Make Light Luna a **collective defensive immune network**: if one opted-in node detects and locally mitigates a credible threat, it records a portable *subject + behavior + scenario + evidence + outcome* label, shares only an authorized privacy-minimized signal with Cloud, and enables other relevant nodes to **recognize the behavior and candidate subject earlier** from signed curated intelligence. A reusable detection scenario is generally more valuable than a bare IP blacklist. This design must not make a single node or Cloud account a global firewall authority.

## CrowdSec primary-source lessons (verified at time of design)

- [Security Engine](https://docs.crowdsec.net/u/getting_started/intro/): acquisitions/logs → parsers → scenarios → alerts → decision profiles → remediation; detects behavior, not just static IP lists.
- [Scenario flow](https://docs.crowdsec.net/docs/log_processor/scenarios/intro/): filter, groupby (often by source IP), distinct, leaky bucket capacity/leakspeed and post-overflow whitelists.
- [Central API](https://docs.crowdsec.net/docs/central_api/intro/): opt-out-capable sharing of scenario name, scenario content hash/version, time, machine identifier and reported source IP; community distribution tied to enabled scenarios.
- [Community Blocklist](https://docs.crowdsec.net/docs/central_api/community_blocklist/): curated IP intelligence tailored to engine behavior; only unmodified official Hub scenario hashes are accepted for community-consensus contribution; custom/tainted scenarios are excluded.
- [Remediation Sync](https://docs.crowdsec.net/u/console/remediation_sync/): CrowdSec describes a **Premium beta** organization-wide decision synchronization capability. This is a precedent, **not** a claim that cross-community raw signals trigger universal blocks.
- [Remediation concepts](https://docs.crowdsec.net/docs/concepts/): local detection, decision and enforcement are separate roles.

Light Luna's proposal takes these as design inspiration, not a code/dependency copy or a false promise of zero false positives.

## Five objects, never conflated

1. **Event**: raw local observation, short retention, site/tenant-scoped.
2. **Signal** (`ThreatSignalV1alpha1`): bounded *observed behavior* and local action/outcome with source provenance, scenario code/content hash and scope. A local block is an **action**, not evidence of malicious ground truth.
3. **Subject**: possibly a verified public ingress IP/IP-prefix or other weak indicator, correlated with a bounded time/context window; **not a person**, not guaranteed unique, NAT/CDN/shared addresses can change. Fingerprints and headers are potentially spoofable.
4. **Scenario/Behavior Pattern**: explainable, reviewed versioned features + threshold/window + benign counterexamples + applicability, ATT&CK mapping if defensible; stored as versioned signed knowledge that clients may match with local telemetry.
5. **Threat Bulletin** (`ThreatBulletinV1alpha1`): Cloud-reviewed distribution statement: behavior/scenario revision + scoped optional actor indicator + evidence provenance and quality + freshness + lifecycle/expiry + *recommended posture*. It is **never an executable firewall rule** and is independently signed as part of a KnowledgePack.

## Signal lifecycle

```text
Local observation → normalized features / trusted ingress identity
  → deterministic behavioral scenario + corroboration
  → local independent Policy Gate (if action authorized)
  → verified local action receipt + evidence/outcome (NOT proof of guilt)
  → optional scoped, privacy-minimized signal outbox (consent separate)
  → Cloud source identity / signed official scenario-hash validation
  → dedup, diversity, benign-negative screening, poisoning review
  → CANDIDATE → CORROBORATED → CURATED or REJECTED/RETRACTED
  → signed relevant Threat Bulletin / Scenario KnowledgePack
  → opt-in subscribed clients verify hash/signature/scope/expiry
  → local early recognition/WATCH, investigation, optional safe preparation
  → local Policy Gate independently chooses scoped action if justified
  → opt-in outcome/counterexample feedback → review/retraction
```

**Fast recognition, not global automatic punishment:** other clients can use a freshly **verified** watch indicator/scenario for fast detection. A single client report is a candidate, not instant cross-tenant block authorization. Same-tenant owner-configured rapid sync is a separate opt-in, narrowly scoped, TTL-bound feature and still must pass every local Policy Gate. No globally shared emergency-deny bypass.

## Learning and memory boundaries

- **Local short-term memory:** source-verified time windows, unique path/method/action counts, recurrence, independent host/resource impact and benign context.
- **Local long-term memory:** deduplicated scenario matches, reviewed positive/negative outcomes, expiry and local trust budget, not infinite event rows.
- **Cloud collective memory:** canonical scenario/technique graph + weighted evidence provenance + independently corroborated subject observables, bounded retention, signed/versioned packs, reference to revoked packs and explicitly typed confidence (`UNVALIDATED` vs measured/calibrated).
- Cloud **must not** count ten cloned accounts or repeated alerts from one event as ten independent sources. Feed prominence, number of reports and AI confidence are not calibrated probability or proof of hostile intent.
- Offline nodes use base Brain and last trusted eligible pack, honestly mark stale indicators; no Cloud data plane dependency or forced updates. No changes to App/Brain code version when only the Threat Bulletin/Knowledge revision changes.

## Privacy and abuse restrictions

Separate checkboxes/consents for Cloud enrollment, remote config management and sharing community threat signals. Default **community contribution off** without informed authorization. Status/fleet sharing alone does not grant access to source IPs, URLs or traffic. Public IP/identifier contribution is potentially sensitive and requires clear notice/purpose/retention and opt-in; site/user identity is pseudonymized where useful and access scoped. No raw HTTP bodies, credentials, private topology, arbitrary request URLs or traffic logs as generic uploads. A Cloud publisher can issue **knowledge**, never `nft`/XDP shell commands or protected policy overrides.

## Action/reputation tiers (proposed product policy)

| Distributed verdict | Requirements | Remote receiving-node default |
|---|---|---|
| `UNREVIEWED_SIGNAL` | Single node, custom scenario, uncertain provenance | **No community distribution as authority**; local only |
| `CANDIDATE_WATCH` | Trusted source / preliminary corroboration and bounded context | Inspect more closely, local logging/telemetry within quotas; no global block |
| `VERIFIED_PATTERN` | Reviewed scenario or multiple independent corroborated signals, benign negatives tested | Signed behavior matcher can warn/challenge or suggest local rule subject to local policy |
| `CURATED_HIGH_RISK_INDICATOR` | Explicit publisher review, strict identity/scope TTL, QA and revocation | Local owner-enabled short-lived scoped mitigation **only after local Policy Gate** |
| `DISPUTED` / `RETRACTED` | False positive, expired/stale identity, publisher revocation | Withdraw candidate and mitigate safely under local audit; propagate withdrawal quickly |

The default cross-tenant posture **does not automatically import another customer's ban**. Local actions also need source authenticity, protected admin/CDN/VPN allowlists, competing benign explanations and service scope.

## Acceptance conditions and phases

- P0/P1: design/test local deterministic scenario engine with benign negative controls, subject provenance and typed `Signal/Scenario/Bulletin` contracts; do not build mandatory Cloud or alter live firewall.
- P2: shadow evaluation for feature vector, scenarios, dedup, near-identical attackers vs legitimate scanning, CDN/proxy attribution, ipv6, NAT/mobile churn and spoofed user-agent/TLS fingerprints.
- P3: local owner-approved bounded mitigations behind independent Policy Gate and TTL/rollback; Cloud signals still cannot bypass local enforcement.
- P7 optional Cloud: signed fast-lane watch feeds, multi-tenant opt-in aggregate contributions, anti-Sybil diversity, cross-source consensus, low-latency rollouts, canary/rollback, controlled appeal/retraction, rate/cardinality quotas, privacy tests and Cloud migration.
- Gate against malicious poisoning by forged clients, compromised telemetry parsers, attacker-triggered false positives, source-IP spoofing, fake cross-tenant consensus, one CDN IP serving many customers, malicious scenario updates and stale indicators.
- **Measure time-to-publish / time-to-recognition, detection accuracy and FP rates**, not just packets blocked or reported actor counts. Publish none as production metrics without real tests.

## Status

**Documentation architecture only.** No Light Luna behavior classifier, local event feedback, opt-in Cloud signal exchange, CrowdSec integration, crowdsourced blacklist, signed bulletin builder or multi-client replication has been implemented in the Go P0 CLI. CrowdSec itself is not added as a product runtime dependency. The standalone-first/G-STANDALONE requirements and separately authorized NetBird future gate remain unchanged.
