# 03 — Luna Brain: Bounded Autonomous Security

## Brain V2 architecture and source evidence

[ADR-022](adr/ADR-022-rebuild-luna-brain-governed-knowledge-migration.md) selects a **new Go-native Brain V2** rather than cloning the existing Python Sentinel deployment. [Source audit](history/audits/SENTINEL_LUNA_COGNITION_AUDIT_2026-10-13.md) confirms the older cognition is primarily curated/generated JSONL templates, deterministic heuristics, SQLite observations and gated learning—not a verified neural model checkpoint. [Brain V2 design](architecture/luna-brain-v2.md) defines knowledge/weights, evidence authority and private/local/cloud contracts; [Threat Knowledge Pipeline](architecture/threat-intelligence-knowledge-pipeline.md) defines CVE/feed → installed affected asset → scenario → advisory path. **This is planned architecture: no Brain V2 process, model or enforcement path is implemented.**

## Proposed two-role Brain/memory design (not accepted implementation)

[ADR-023](adr/ADR-023-dual-brain-memory-cloud.md) evaluates the requested left/right-brain structure as two **logical software roles**: Go deterministic analytical detection and optional Go/Python hypothesis reasoning. Both read three knowledge regions and typed evidence; independent critic checks contradictions while Policy Gate remains separate. Proposed [bounded local memory](architecture/luna-memory-lifecycle.md) prevents growing SQLite history by quota, TTL and downsampling; future [Cloud](architecture/luna-cloud-intelligence.md) distributes verified signed knowledge but does not decide packets or require raw log upload. All profiles/implementations remain proposals subject to owner approval and tests.

## Mission

A decision engine, not a privileged shell. Luna learns *normal context*, correlates anomalies, tests hypotheses and issues structured proposals. A deterministic policy and executor approve/reject actions. Users can turn Luna OFF without disabling the firewall.

## Implementation direction (ADR-020)

Luna Brain's deterministic event normalization, anomaly detection, correlation, evidence aggregation and typed proposals are to be written **primarily in Go**. An **optional Python worker** is permitted for specific ML, offline evaluation, statistical analysis or supported inference integrations after capability, resource and packaging review; it is not a core runtime dependency. A Python process cannot mutate firewall state or access the privileged executor. It runs unprivileged with strict time/memory limits, a versioned input/output contract and no arbitrary script execution from untrusted logs or model output. Failure/absence falls back to Go deterministic operation and honest telemetry uncertainty. See [ADR-020](adr/ADR-020-go-core-python-optional.md). **Neither component is implemented as of the P0 CLI-only baseline.**

## Core loop

`COLLECT → NORMALIZE → BASELINE → DETECT → CORRELATE → HYPOTHESIZE → DECIDE → POLICY-GATE → EXECUTE/WAIT → VERIFY → LEARN`

## Roles

- **Luna (defender brain):** analyzes site/network/host context; recommends scoped actions with evidence and uncertainty.
- **Selene (optional shadow critic):** independently reviews Luna confidence, reproducibility, base rate and collateral damage; no independent execution authority.
- **Deterministic baseline engine:** simple known-signature rule, rate anomalies and threshold detections keep working without an LLM.
- **Policy Gate:** non-AI authorization against immutable policy bundle and provenance/health requirements.
- **Owner:** defines maximum authority, protected identities and sign-off procedures.

## Decision taxonomy

| Decision | Meaning | Real action? |
|---|---|---|
| `NO_ACTION` | No actionable signal within observed evidence; does not prove safety | No |
| `INSUFFICIENT_TELEMETRY` | Untrusted/stale/truncated/missing data | No |
| `WAIT_MORE_EVIDENCE` | Suspicious but confidence or duration insufficient | No |
| `REVIEW_RECOMMENDED` | Human triage warranted | Notification only |
| `RATE_LIMIT_RECOMMENDED` | Propose bounded mitigation at compatible proxy/WAF | Only if policy allows |
| `CHALLENGE_RECOMMENDED` | Propose site-level challenge at capable L7 proxy | Only if policy allows |
| `BAN_RECOMMENDED` | Candidate time-bound IP/service/site containment | Only if policy allows |
| `UNBAN_RECOMMENDED` | Current evidence, label or TTL supports removal | Only if policy allows |
| `INCIDENT_CRITICAL` | Material harm or high-confidence active incident | Escalation + pre-authorized actions |

## Typed proposal contract (draft)

```json
{
  "schema": "lightluna.decision/v1alpha1",
  "decision_id": "unique-id",
  "node_id": "node-a",
  "service_id": "site-b",
  "subject": {"kind":"source-ip","value":"203.0.113.42","identity_verified":false},
  "verdict":"WAIT_MORE_EVIDENCE",
  "confidence_band":"medium",
  "confidence_calibration_id":null,
  "evidence_refs":["sha256:evidence-bundle-hash"],
  "time_range_utc":{"from":"...","to":"..."},
  "action":{"type":"none","scope":"site-b","ttl_seconds":0},
  "uncertainties":["unverified original client address"],
  "execution_requested":false
}
```

`confidence_band` is a qualitative judgment until properly calibrated against independent owner-labeled outcomes. Never invent `98% accurate` from internal/synthetic checks.

## Learning and retained memory

- **Short-term observations:** ring buffer and 10-minute disjoint time windows; deduplicate observed events vs reports. Retain access only per privacy policy.
- **Crater Memory:** compact event → decision → observed effect → analyst label → corrected outcome summaries, versioned and revocable; no perpetual raw IP/log hoarding.
- **Learning agenda:** evidence-gated tasks for trusted IP provenance, software inventory/SBOM, CVE validity, independent health, labeling and temporal corroboration.
- **Rule promotion:** offline corpus tests, shadow observation, owner approval for higher privileges; no automated reward loop that optimizes “more bans” or false precision.
- **Concept drift:** track behavior changes after releases, legitimate crawlers and customers; automated de-escalation with outcome verification.
- **Prompt injection:** untrusted logs, DNS names, advisories, threat intel and web pages are data only. They cannot authorize tool calls, policy updates, scripts or new privileged capabilities.
- **Resource caps:** CPU/RAM/model context, concurrent inference, queue size, disk retention and time budgets configurable per profile; host protection continues during LLM outage.

## Threat-intelligence learning rules

Correlate NVD/CISA KEV/EPSS with installed artifact identifiers, full dependency tree, Linux distro package/backport status, reachable network paths and compensating controls. A version string match without distro advisory validation is **not** proof of exploitability. When data is incomplete, mark `UNKNOWN`, not `NOT_VULNERABLE`.

## Autonomy maturity

0. **Observe only** → 1. **Proposals with justification** → 2. **Human-confirmed reversible actions** → 3. **Pre-authorized, scoped, TTL-capped actions** → 4. **Advanced local learning with independent drift/fp monitoring**. Distinct modes and rollback gates; no silent promotion.

## Seed/research material

Existing Luna/Selene Sentinel observations are useful for design experience, but are **not** proof of commercial firewall training or independently verified field accuracy. Bring over only sanitized code ideas/fixtures after source/license/privacy review; leave real memory databases, logs and credentials in their current secured locations.
