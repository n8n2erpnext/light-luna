# Sentinel Luna — Source and Cognitive Architecture Audit

Status: verification/evidence snapshot; historical after later deployments
Date: 2026-10-13
Scope: **read-only** source/runtime inspection on privately owned VPS-ARM; sanitized findings for the public Light Luna repository
Related decisions: [ADR-022](../../adr/ADR-022-rebuild-luna-brain-governed-knowledge-migration.md), [Brain V2](../../architecture/luna-brain-v2.md)
Evidence tier: inspected live implementation + latest source-only Git snapshot and dated handoffs; NOT a validated security model or attack detection benchmark

## Identity and operating surfaces

The existing **Luna** and **Selene** are independent Python micro-Sentinel observers assigned to separate websites. Source groups (private service paths; list filenames not sensitive log contents):

| Component | Verified files | Role |
|---|---|---|
| Observer | `micro-sentinel/bin/sentinel_cycle.py`, `sentinel.py`, `sentinel_behavioral_baseline.py`, `network_reasoner.py` | Network/proxy observation, topology reasoning, baseline and candidate pattern extraction |
| Memory/curriculum | `sentinel_memory.py`, `sentinel_ladder.py`, `cognitive_bridge.py`, `knowledge/cognition/*.jsonl` | Confidence decay, level-gated knowledge, indexed reasoning/knowledge/context priors |
| Site guardian | `sentinel-site-guardian/decision_engine.py`, `shadow_review.py`, `guardian_temporal_shadow.py`, `guardian_reasoner.py`, `learning_loop.py` | Evidence-bound site-scoped incident decisions, temporal corroboration and curriculum gaps |
| Intel | `sentinel-threat-intel/intake.py`, `correlate.py` | Fixed-source HTTPS fetch, CISA KEV/NVD/EPSS and security feeds; research correlation only |
| Reporting | `sentinel-guardian-reports/reporter.py` | Decision transition journal + bounded notifications |
| Source-only snapshot | `light-sentinel-decision` local Git repository | Guardian/intel source/tests; no GitHub remote or production deployment contract confirmed |

As observed during read-only inspection, existing *user* systemd Sentinel timers/services run with recent `ExecMainStatus=0`. This does not prove delivery, all feeds' freshness or decision quality. The source-only Git snapshot matched the inspected runtime Python source files in the sampled managed components, but it is **not a release pipeline**. No production configuration was modified.

## What the legacy “Brain” actually is

A **deterministic evidence-governed rule/curriculum/retrieval system**, *not* a trained foundation model. No torch/TensorFlow/ONNX/llama runtime imports or model weight files were found in scoped Sentinel modules. Absence in this scope does not prove none exist elsewhere on the host.

Three `knowledge/cognition/` JSONL regions inspected in one Luna instance:

| Region | Records (source inspection) | Approx bytes | Semantics |
|---|---:|---:|---|
| reasoning | **11,980** | 8.39 MB | hypothesis/falsification templates, tagged by network concepts |
| knowledge | **17,115** | 8.39 MB | curated-prior-labelled concepts, counterexamples and evidence surfaces |
| context | **5,863** | 4.20 MB | contextual constraints, review questions and authority boundary |
| total | **34,958** | ~20.98 MB | **Not 34,958 independently validated behaviors or trained weights** |

Source generator `build_sentinel_cognition.py` is a deterministic template-expansion loop that fills each zone until a target byte size by repeating combinations of network concepts, sensors, benign alternatives and safety questions. It mirrors the same curated JSONL to both agents. Therefore raw record count/file size must **never** be used as a proxy for diversity, detection accuracy or memory quality. The `cognitive_bridge.py` indexes tags to JSONL byte offsets and picks a deterministic candidate from a tag; it is **not** a neural model.

The architecture manifest shows byte budgets and a required reasoning + knowledge + context + live provenance combination before high-confidence claims. Level ladder has 14 topics, daily/unlock/reflection gates, NOT a graded model benchmark.

## Numeric heuristics and exact semantics (NOT calibrated ML weights)

| Existing source | Numeric logic observed | What it means / caution |
|---|---|---|
| `sentinel.py` | +0.55 sensitive/admin path, +0.12 denied/missing status, +0.10 unusual method; cap at 1 | Hand-authored per-event heuristic; `score` is **not probability of attack** |
| `sentinel_memory.py` | new confidence = `0.65×old + 0.35×new`; aging factor `0.94^days`; stable when seen ≥3 days and confidence ≥0.72, decay after ≥7 days, retire after ≥30 missing days | Retained pattern-strength heuristic; observations may be correlated, truth labels missing |
| `sentinel_behavioral_baseline.py` | Total Variation distance between HTTP method/status histograms; alert when method TV≥0.25 or status TV≥0.30, or newly occurring method/status; require ≥50 current and ≥200 historical events over ≥3 prior days | Host-level change detection, **not** proof of hostility; deployment/user-behavior changes may trigger |
| `guardian_reasoner.py` | Volume surge `current ≥40 rps` and `≥6× baseline`, plus independent error/latency/CPU connection-impact clues | Scenario reasoning thresholds; not all telemetry is actually available in current production |
| `guardian_temporal_shadow.py` | 5-minute recent vs 60-minute earlier rates; surge if ≥50 rpm, baseline≥1 rpm and ≥6×, with HTTP 5xx rise for incident review | Rate anomaly/proxy-observation heuristic; does not prove a DDoS event |
| `decision_engine.py` | `BAN_RECOMMENDED` only if suspicious hits≥8, paths≥6, denied/missing≥5, feature families≥2, separate 10-min buckets≥2, span≥600s | **Advisory only**, no enforcement; original client identity and scope remain unverified |
| `learning_loop.py` | Six topics gate missing evidence: original IP, installed software/CVE mapping, independent resource impact, false positives, temporal persistence, feed reliability | Deterministic curriculum: `READY_FOR_INDEPENDENT_EVALUATION` ≠ learned skill or granted capability |

The heuristic scores belong in a *legacy shadow baseline* for replay comparison, not live default Cloud-enforced risk weights. The existing `NO_ACTION` means no decision-grade evidence in an observed window, **not** site safety; stale/incomplete evidence returns `INSUFFICIENT_TELEMETRY`.

## Existing event memory (privacy-sensitive)

SQLite per agent uses `observations`, `learned_patterns`, `pattern_memory` and `memory_history` tables; index/timer handoffs document low-latency recent observation queries. Scoped inspection saw each live database roughly **700–720 MB**. DB rows include observed source identifiers, hosts, paths and evidence that may be private. **Do not copy them into Light Luna public Git or shared cloud knowledge.**

No live data rows or unverified IP addresses were exported by this audit; only schema, source code structure, sizes, counters and documented semantics were inspected.

## Current threat intel capabilities and limits

Existing fixed HTTPS feeds include CISA KEV, CISA advisories, NVD recent, FIRST EPSS (selected CVEs), Ubuntu notices, SANS/Cloudflare research, and Feodo/Spamhaus reputation. The intake is byte-limited, validates URLs against fixed hosts, disallows redirects, hashes fetched payloads, bounds RSS and records freshness. NVD's configured recent window requests ≤2,000 entries; a complete short window is **not** a complete historical CVE database. Reputation feeds identify context/C2 or high-risk prefixes, **not automatically an attack on the protected site**. CISA KEV marks exploitation in the wild; it does not establish local exposure or successful compromise.

Verified major gaps: full installed-app SBOM + distro backport mapping, trusted real ingress IP proof, independent upstream resource/latency pressure, owner-labelled benign/attack replay and calibrated false-positive rates. The existence of installed package counts or a headline alone cannot close these gates.

## Clone vs rebuild verdict

**Rebuild V2, curate/import governed logic and facts; do not clone deployment, historical traffic memory or imagined neural weights.** Reasons: Python services are tightly coupled to absolute legacy paths and website assignments; generated cognition corpus has repeated template combinations; heuristic constants are not validated probabilities; cloud sharing needs privacy and poison-resistant validation. Retain the old systems as shadow comparators.

No bulk migration, installation, daemon restart, config edit, rule load, real ban or package mutation was performed as part of this read-only source audit.

## Exact next verification gaps

1. Inventory class/scenario semantic deduplication and author/license/source provenance for candidate priors.
2. Curate a synthetic + owner-labelled controlled replay with benign negatives, and compare Go V2 against the legacy logic without linking real enforcement.
3. Define typed and versioned `EvidenceV1`, `KnowledgeClaimV1`, `ScenarioV1`, `DecisionProposalV1` and promotion/retraction state machines.
4. Separate feed freshness from affected-version/route verification, preserve local-only facts and human review of sensitive actions.
5. Implement opt-in Cloud sharing rules and signed knowledge pack lifecycle with rollback before any automatic distribution.
