# Light Luna Collective Defense — Distributed Immune-Style Threat Learning

Status: **architecture contract from ADR-027, NOT IMPLEMENTED**
Date: 2026-10-14
Scope: event-to-signal classification, subject+behavior/scenario intelligence, trusted distribution and consent, Cloud correlation, false-positive retraction
Primary sources: [ADR-027](../adr/ADR-027-collective-threat-immunity.md), [Threat Intel Pipeline](threat-intelligence-knowledge-pipeline.md), [Brain V2](luna-brain-v2.md), [Cloud](luna-cloud-intelligence.md), [Device Fleet/Account](unified-account-cloud-fleet-control.md)
Static contracts: [ThreatSignal](../../schemas/threat-signal-v1alpha1.schema.json), [ThreatBulletin](../../schemas/threat-bulletin-v1alpha1.schema.json). These are not enforced by running software.

## 1. Analogy and exact product behavior

**Local recognition = innate defense.** Go Brain analyzes trusted observations, known attack scenarios, and verified local policies while fully offline. **Shared curated behavior knowledge = acquired defense.** Cloud aggregates consented, provenance-aware observations across diverse clients, vets/retracts them and signs narrower-scoped knowledge packs, so subscribed clients can recognize similar campaigns **before their own services suffer the same sequence**. The immune analogy is helpful for product vocabulary; no neural model, human-like immune mechanism or preemptive identification guarantee is implied.

Crucial distinction: a **subject** (IP address, TLS fingerprint, autonomous system or bot-session descriptor) is not automatically a malicious person or globally stable entity. Other nodes may recognize the same *behavior* even when adversaries rotate IPs, and may see one source IP used for both benign and malicious clients behind NAT/CDN. Do not merge identities solely from a fingerprint.

## 2. Separate four pipelines

```text
(1) LOCAL EVIDENCE
proxy / auth / WAF / nft counters / optional eBPF → normalization/source provenance
   → time windows, event de-dup, distributed behavioral scenarios
   → local scenario match + benign alternates + confidence type
   → local Policy Gate independently decides observe/notify/scoped action
   → sealed local action/outcome receipt (block was an action, not ground truth)

(2) CONTRIBUTION [separate explicit opt-in]
selected provenance, behavior vector and *optional* public indicator
   → minimal data retention / size + privacy and consent check
   → bounded signed device upload (replay ID, scenario content hash/version)
   → Cloud quarantine, tenant/auth/source verification

(3) CLOUD KNOWLEDGE CURATION
candidate signals → de-dup at sensor/event/temporal fingerprint
   → independent-source and cluster diversity, provenance trust
   → applicability: scenario family, service exposure, WAF context
   → check known benign/owner-labelled negatives and hostile submissions
   → CANDIDATE_WATCH / VERIFIED_PATTERN / CURATED_HIGH_RISK / DISPUTED
   → signed TTL-bound bulletin, scenario and per-asset knowledge delta
   → CDN/object distribution, retractions and opt-out/complaint handling

(4) RECEIVER LOCAL LEARNING
local signed pack verifier → freshness/source/revision/compatibility
   → safe local watch matcher (no implied global ban)
   → corroborate against actual local traffic and policy/allowlists
   → optional local approved time-bounded mitigation
   → incident outcome/false-positive labels fed back only by consent
```

## 3. Feature hierarchy and scenario semantics

| Layer | Examples | Caveat |
|---|---|---|
| Observed subject | Verified ingress IPv4/v6, short-lived IP prefix, app session or challenge identity, optionally stable client key if the owner explicitly collects it | Spoofed forwarded headers, NAT/CGNAT, CDN/shared address and IP rotation; not a human identity |
| Behavior signature | Failed auth bursts, distinct sensitive paths, repeated HTTP method/status transitions, connection-rate/fan-out across windows, resource impact | A single count/URL/method is not a proof of hostility |
| Scenario graph | Scan → targeted endpoint abuse → escalating failed auth → independent service impact; variation per known technique | Include ordered timing, preconditions, match/falsify, trusted telemetry and benign alternatives |
| Threat context | CVE/KEV, affected component/vendor patch and exposed service, ATT&CK technique or curated IOC | Local asset/version/route verification is independent of vendor CVE headlines |
| Outcome | Local block/rate-limit executed, access restored, admin dismissed, proven false positive, independent review | 'Blocked' does not mean malicious; successful challenge also not proof |

Scenario V1 candidates use `scenario_id`, `scenario_revision`, immutable `scenario_hash`, bounded filters/windows/group/distinct keys, threshold metrics, original publisher and benign counterexamples. Historical CrowdSec Hub unmodified-vs-custom scenario distinction motivates two roles:
- **Reviewed/official** trusted content hashes can contribute *quality-rated* common scenarios to candidate consensus, but not automatically bypass review.
- **Custom/user-modified** local scenarios remain useful locally and may be shared only as **research hypotheses** after license/consent checks; never weighted as official community proof merely by displaying a familiar name.

No plan to load arbitrary community YAML or code into privileged local processes. Future scenario representation must use a typed declarative feature DSL with size/CPU caps, no reflection, command execution or kernel attachment.

## 4. Reputation model: think "warrants" with expiry and appeals

A Cloud bulletin is **not a universal arrest warrant**. It says: at time T, verified sources observed scenario S with evidence E affecting scope X, and subject indicator I may be useful until expiry U; uncertainty and known benign exceptions are included. The receiving node determines applicable context, service, its own observations and permissions. Subject labels:
- `UNKNOWN`: reported/seen, insufficient verified evidence.
- `SUSPECTED`: provisional time-limited watch candidate.
- `CORROBORATED`: reviewed diverse signal pattern, still not guaranteed malicious.
- `CURATED_HIGH_RISK`: explicit reviewer approval with strong identity/exposure/expiry constraints.
- `DISPUTED` / `RETRACTED`: evidence conflicts, source compromised, false positives or expiry.

Share *scenario knowledge* broadly and early once validated; share *specific public IP/IP-prefix indicator* narrowly with short TTL and app/context scope. Only permit a subscriber to activate an imported high-risk indicator with owner-approved local policy and safe scope, not Cloud-triggered global `nft add rule`.

## 5. Independent corroboration and anti-poisoning

- Authenticate device, tenant, scenario publisher, content hash, contribution consent and upload expiry. Machine counts alone are not source diversity: shared management, cloned VM, repeat CDN log, co-located sensors and account farms must be grouped.
- Count **independent observations**, not the number of alerts, accounts or IPs. Link correlated reports to one origin/attack campaign where possible; unknown correlation reduces confidence, it does not boost it.
- Require a *source-quality score type*, explicit alternative hypotheses, benign negatives and reported/confirmed outcome distinction. Attackers may provoke a chosen service into banning a benign CDN/monitoring IP then flood Cloud with copies.
- Official feed evidence remains separate from user-generated patterns; no malicious CVE text or forum post can become an AI instruction, signing grant or privileged command.
- Quarantine suspicious report spikes, contradictory data, invalid scenario hashes, missing provenance, repeated contributor churn and improbable geography/timestamps. Human/approved release gates for globally high-risk packs.
- Feedback loop incorporates owner false-positive dismissals and appeal outcomes with independent checks, not blind trust in outcome count. Rate/cardinality quotas, cost abuse defenses and multi-tenant isolation.

## 6. Distribution tiers and timing (targeted but no invented SLO)

```text
LOCAL IMMEDIATE: direct observation → independent local action if owner authorized.
FLEET FAST LANE: same tenant owner opt-in → bounded temporary watch notice and
                 explicit local gate; optional high-risk approval per device.
COMMUNITY WATCH: early candidate scenario indicator; detect/observe, no blanket ban.
CURATED KNOWLEDGE: signed reviewed verified scenario packs + scoped indicators.
RETRACTION PRIORITY: signed retractions/tombstones should distribute quickly and
                     remove stale watch/auto-derived decisions with local safety check.
```

Low latency is achieved by **short signed delta manifests + long-poll/ETag/cache + prioritized queue** as resource budgets allow; the host must not continuously poll every second or overload a small Cloud. A never-connected device has baseline knowledge, while Cloud subscribers may download relevant signed updates and report opt-in signals. The community watch feed and actual enforcement are different services/permissions.

Cloud Go ingestion → PostgreSQL source ledger/outbox → curated pack builder + signing boundary → immutable S3-compatible storage/CDN → stateless manifest API. Horizontally scale ingestion/curation independently of publication. Do not bind recipient to a particular VPS IP or use NetBird by default. Retain data within explicit tenant/system quotas, TTLs and publisher retention policies.

## 7. Threat-to-defense examples (non-operational, synthetic)

**Example A: brute force.** Node A observes repeated failed auth over separated windows with verified ingress and another independent impact signal, and its owner-authorized rule challenges/temporarily limits a source. It uploads a compact scenario outcome. Cloud validates the scenario revision, benign negatives and independent sources, then publishes a signed behavior notice. Node B, even if the actor uses a different IP, already knows to correlate the same behavior and can warn early when its local events match; it does not globally ban the original IP solely because A did.

**Example B: CVE campaign.** Official advisory plus affected local SBOM/route evidence informs a curated scenario of pre-exploitation signs. Cloud sends only relevant scenario data to nodes with matching components if known, or lets them request via private selectors. Node B recognizes a sequence of suspicious requests, proposes a safe temporary shield and asks for permission according to local policy. Nodes with patched/not-exposed software should not inherit automatic attack verdicts.

**Example C: false positive / shared infrastructure.** A monitoring probe or CDN egress accidentally matches one scenario and gets locally limited. The owner dismisses and labels it. Cloud rechecks/report provenance, lowers/withdraws bulletin and broadcasts a signed retraction. B does not continue acting on a stale indicator; the local site allowlist/protected admin route always wins.

## 8. Model and acceptance

The **Basic Go Luna Brain** should detect locally from bundled scenario knowledge with no Cloud. A future optional ML worker may rank/cluster candidates offline, not train continuously on unverified community labels nor obtain execution permissions. The labels and scenario sets evolve through separately versioned, curated knowledge packs; App/Brain versions only change when code changes.

Initial tests: 1/10/100 malicious simultaneous contributors, spam clones/replays, source-IP forgery, proxy NAT, IP reassignment, IPv6 churn, known benign scanner, CDN challenge, monitor traffic, broad legitimate login burst, XSS/path noise, stale/revoked pack, offline node reconnect, pack rollout rejection, spammed appeals, cross-tenant breach and Cloud downtime. Replay labels should include **attack, benign, ambiguous and insufficient telemetry**. Measure lead-time, FPR/FNR, scenario recall, ingest-to-watch latency, cost per 1K reports, safe rollout latency and retraction completion. Values will be measured later; none are published performance claims.

## Integrity-gated community contributions (ADR-028)

[ADR-028](../adr/ADR-028-self-defense-host-compromise.md) adds independent **device integrity and Host Guardian** as a quality/trust signal *before* a community threat contribution may even be considered for curation. Authentication proves enrollment key possession, not cleanliness. A compromised/suspected node's threat signals go to segregated untrusted intake with community-consensus weight **zero** pending external review; no raw suspicious files, executable scenario payloads or credential material may be uploaded. [Host containment architecture](host-self-defense-and-compromise-containment.md) separates local owner mitigation from Cloud quarantine, signed publisher and recovery. Even 'healthy observed' device submissions remain untrusted until source diversity, consent, falsification and independent review pass.

## 9. Relationship to phases

P0 contracts/schema + synthetic negative tests only; P1/P2 local scenario observer/replay and Brain baseline; P3 gated local mitigation; P7 opt-in Cloud collective pattern curation/distribution. Shared community Cloud cannot become a prerequisite for G-STANDALONE and does not authorize NetBird engineering. All proposed endpoints, signed pack types and Cloud processing remain **unimplemented**.
