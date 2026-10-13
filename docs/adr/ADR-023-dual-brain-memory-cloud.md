# ADR-023 — Dual-Role Brain, Bounded Local Learning and Optional Cloud Intelligence

Status: **PROPOSED — design recommendation awaiting owner acceptance**
Date: 2026-10-13
Scope: Light Luna cognition/runtime, local storage limits, model lifecycle, future client–Cloud exchange
Primary sources: [Brain V2](../architecture/luna-brain-v2.md), [Sentinel cognition audit](../history/audits/SENTINEL_LUNA_COGNITION_AUDIT_2026-10-13.md), [memory budget](../architecture/luna-memory-lifecycle.md), [Cloud topology](../architecture/luna-cloud-intelligence.md)
Related decisions: [ADR-020 Go/Python](ADR-020-go-core-python-optional.md), [ADR-022 Brain rebuild](ADR-022-rebuild-luna-brain-governed-knowledge-migration.md)

## Motivation

The original TRI_REGION_COGNITION corpus has Knowledge, Reasoning, Context **data regions**. It does not have neural hemispheres; corpus size and rule repetition are not model quality. The owner requested review of a dual-brain organization, core model training, growth bounds and future client–Cloud relationship.

## Proposed cognitive division

- **Analytical / deterministic hemisphere (Go, always available):** normalize and assess telemetry quality; maintain independent time windows/baselines; asset/CVE applicability; graph constraints, exact rule matching and negative controls; produce typed evidence and deterministic judgments.
- **Hypothesis / reasoning hemisphere (Go, optional enrichment):** retrieve versioned Knowledge/Reasoning/Context priors and relevant scenarios, compare competing interpretations, identify precursors, quantify *uncertainty*, propose low-disruption prevention. Future Python ML inference may assist; never required.
- Both consult one **typed Evidence/Knowledge workspace** and an independent contradiction/critique stage. Shared memory is not shared trust or executable authority. The **Policy Gate is independent of both hemispheres**, separately privileged only in its executor interface.
- Do not infer a neurobiological left/right split, require a heavyweight LLM or let two models mutually invent ground truth.

## Proposed model strategy

- **Do not train or ship a required neural foundation model in the core firewall.** The primary algorithm is rule/baseline/evidence-state-machine Go code with synthetic and owner-labelled replay.
- Permit local bounded baseline/pattern learning; it is not online gradient retraining and not permission promotion.
- Train a small **optional** anomaly/ranking model only when labelled data, leakage-resistant time/site holdouts, workload/cost benchmarks and drift/FP measurements support improvement over Go baseline.
- Maintain separate `heuristic_revision`, `knowledge_pack_revision`, `model_revision` and `policy_revision`. An ML score cannot imply a probability until out-of-sample calibration. Model deployment requires signed reviewed release and shadow/canary/rollback.

## Proposed local memory and storage discipline

The installation has a **hard managed-data cap** per profile, configurable by owner, independent of total host disk. See [memory lifecycle](../architecture/luna-memory-lifecycle.md). Proposed Starter 256 MiB / Standard 512 MiB / Extended 2 GiB are **planning targets, not measured minima**; pack sizes, SQLite WAL, runtime RSS, C/BPF maps and third-party log files need separate limits/tests.

Storage follows raw-short-lived -> time-bucket summaries -> stable de-duplicated approved patterns -> expire/compact. Prioritize journal safety, current rules, rollback and validity metadata; when full stop noncritical ingestion/training, mark evidence incomplete, retain reserved space, never silently infer safety. Only relevant filtered CVE/technique packs are cached locally, not all historic feeds.

## Proposed Cloud authority

Clients make realtime defensive decisions **locally**. Future Cloud operates threat feed adapters, component/CVE normalization, curated scenario and controlled model training/evaluation; publishes signed, hashed, expiring, rollback-protected `KnowledgePack` deltas. **No mandatory per-node LLM or central packet-decision API.** A lightweight Cloud intelligence curator is useful; optional AI research workers may run asynchronously but may not issue enforceable commands.

Device → Cloud defaults to **minimal operational status**. Contributing de-identified aggregate behavior/outcomes to collective learning is **opt-in**, with private tenant isolation, consent/revocation, delayed/limited event upload, attack/benign labelled examples, protection against fake sites and duplicate sensor votes. Cloud reputation cannot directly alter local nftables/XDP/WAF and Cloud cannot self-grant owner firewall permissions.

## Conditions for accepting/implementing

1. Approve storage targets only after bench on low-resource `amd64` and `arm64`, including growth over simulated months and power-loss/database recovery.
2. Specify `EventV1`, `EvidenceV1`, `ScenarioV1`, `KnowledgePackV1` + scoring semantics and negative controls; see [Brain V2 design](../architecture/luna-brain-v2.md).
3. Define per-tenant/device auth, signed delivery, malicious rollback/freeze detection, compromise/revocation, network outages, repeat uploads, throttling and retention.
4. Do **not** implement Cloud dependencies in P0–P4 standalone Host/Edge. Reuse generic transport interface without NetBird; optional Cloud implementation is future gated work.
5. Preserve legacy Luna/Selene data/service and its observed-only safety status; no blind migration of private logs or neural-weight claims.

Status: **architecture proposal only**. No Brain hemispheres, ML weights, Cloud service, cache limits or signed-pack updater are currently implemented in Light Luna.
