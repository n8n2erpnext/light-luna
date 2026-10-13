# Luna Brain V2 — Evidence, Knowledge, Reasoning and Defensive Decisions

Status: proposed implementation contract based on [ADR-022](../adr/ADR-022-rebuild-luna-brain-governed-knowledge-migration.md); NOT IMPLEMENTED
Date: 2026-10-13
Scope: Brain V2 domain boundaries, knowledge/weight semantics, scenario execution and architecture around the Brain
Primary evidence: [legacy source audit](../history/audits/SENTINEL_LUNA_COGNITION_AUDIT_2026-10-13.md), [original Brain design](../03-LUNA-BRAIN.md), [security model](../07-SECURITY-SAFETY.md), [threat intel ingestion](threat-intelligence-knowledge-pipeline.md)
Code authority: none yet; P0 contains only Go CLI policy validation/preview

## 1. Mission and epistemic invariant

Luna is a **bounded defensive reasoning system**, not a root shell, arbitrary command runner, guaranteed predictor or automatically trusted language model. It should understand observable behavior and the network/topology layer that produced it; form competing attack/benign hypotheses; test them against local assets, historical baselines and vetted global threat knowledge; request missing evidence; and generate a **typed recommendation with a forecast horizon**. Independent Policy Gate decides whether any action is permitted.

A CVE announcement is not evidence a node is vulnerable. A traffic spike is not proof of DDoS. A source/IP/ASN appearing across sites is not proof of one attacker; a verified source address is not a unique human identity. Prediction is conditional risk assessment, not certainty of an attack.

## 2. Architecture / clean component contracts

```text
Event inputs: proxy/WAF logs, host journald/auth, nft counters, conntrack, CPU,
              optional eBPF observer, network flows, installed software inventory
                   |
                   v
 [Go Collector + Topology/Source Provenance + bounded local queue]
                   |
                   v
 [EventV1 Normalizer] -----> [Local Evidence Ledger / audit / retention]
                   |                               |
                   +--> [Feature Windows/Baseline] <+
                   |           |                 |
                   +--> [Local Knowledge Cache] <-- [KnowledgePack verifier]
                   |           ^                         ^
                   |           | signed/hashed subset     |
                   |       [Threat Intel & Asset Matcher] [OPTIONAL Cloud]
                   v
 +-------------------- Go Luna Brain (unprivileged) --------------------+
 | deterministic detectors: new anomaly, attack technique, vulnerability |
 | scenario graph: preconditions -> signals -> alternate explanations     |
 | hypothesis selection + contradictions + missing evidence               |
 | temporal/state machine + confidence bands + baseline drift            |
 | short/long-term memory: evidence-linked, reversible, site-scoped       |
 | optional unprivileged Python ML worker: ranking only; timeout fallback |
 +--------------------------+-------------------------------------------+
                            |
                       DecisionProposalV1
                            v
                  [Independent Go Policy Gate]
                   deny / observe / notify / manual
                   (future scoped-auto only after approvals)
                            |
                      [Restricted Executor]
                   nftables / WAF / future XDP
                            |
                 outcome verification + audit
                            |
                  [labelled outcome feedback]
                            |
        replay/evaluation -> reviewed knowledge or model promotion
```

Go remains the core; Python ML is optional/off packet path. WebUI React/TypeScript, local SQLite and future Cloud are **consumers/management layers**, not Brain authority. eBPF/XDP requires [ADR-021](../adr/ADR-021-ebpf-observability-xdp-optional.md); nftables remains the principal firewall.

## 3. Port the TRI_REGION_COGNITION idea, not the template bulk

Preserve three distinguishable types, with explicit reasoner references:

- **KNOWLEDGE**: reviewed facts and known attack mechanics (IP/proxy trust, conntrack, technique patterns, CVE affected version, CISA exploitation status). Every claim has `origin`, `publisher`, `source_hash`, `observed_at`, `valid_until`, `scope`, `license`, `review_state` and links to contradictions/retractions.
- **REASONING**: conditional rules/hypothesis graphs with supporting features, counterexamples, missing evidence, falsification tests and precision limits. A rule is **one** logical statement with version, not thousands of mechanically generated phrasing variants.
- **CONTEXT**: deploy topology, service inventory, app/proxy/network boundaries, protected routes, incident budgets, local behavior baselines and non-bypassable safety invariants. Customer context stays **local** unless explicitly shared under a separate consent policy.

Add an independent **EXPERIENCE** ledger: observed feature summary → hypothesis → action or abstention → measured outcome → reviewed label → time decay/retraction. Do not let experience overwrite fact/contract without an authorized promotion.

Retrieval should use IDs, tags, structured indexes and knowledge graph edges with provenance; future optional embedding similarity is just a retrieval aid, never evidence authorization.

## 4. “Weights” must be typed by meaning

| Quantity | Meaning | May self-update? | May authorize firewall? |
|---|---|---|---|
| `rule_priority` | Operator-declared deterministic policy precedence | No, owner-policy change only | Only by independent gate |
| `source_reliability` | Curated publisher trust, signature/freshness/coverage quality | Reviewed/versioned changes | No |
| `feature_strength` | Magnitude of one measured signal (e.g. baseline deviation) | Measured per local window | No |
| `evidence_independence` | Are signals from distinct observers/time windows? | Computed from provenance | No |
| `hypothesis_support` | Qualitative/scored relative explanatory support | May adapt after *labelled* eval | No |
| `forecast_risk` | Calibrated *if measured*, else `UNVALIDATED` | Only through offline gated promotion | No |
| `model_parameters` | Optional ML model bytes/version; distinct from rules | Only signed offline release | No |
| `action_authority` | Owner scope, protected routes, TTL, rate caps, approval mode | Never learned | **Only** Policy Gate |

The legacy blend `0.65 old + 0.35 new`, `0.94^age` decay, `0.25/0.30` histogram TV and fixed sensitive-path scores are **audit baselines**; they are not portable calibrated probabilities or trusted Cloud weights. Keep them in legacy shadow replay and compare against independently labelled benign/attack cases before accepting any new feature weighting.

### Proposed scoring stages (specification, not deployed formula)

1. **Coverage gate:** source observed? fresh? complete? site attribution? topology known? If not, return `INSUFFICIENT_TELEMETRY` or `REQUEST_MORE_EVIDENCE`, with explicit missing fields.
2. **Evidence representation:** normalized `[feature, value, window, source, trust, scope, correlation_group, freshness]`; record benign alternatives and contradictions.
3. **Independent corroboration:** group same-source repeating evidence so two copies do not count as two votes. Score can grow only when new independent evidence arrives, not from repeated alert messages.
4. **Competing hypothesis graph:** compare normal deployment/crawler, flash crowd, configuration error, upstream outage and specific attack techniques with measured signatures. Prefer an abstention when explanations remain indistinguishable.
5. **Site-specific baseline:** compare like-for-like windows, service/route/traffic class, seasonality and recent deployments. Emit drift signals, not malicious verdicts from drift alone.
6. **Contextual threat exposure:** join CVE/advisory with **exact installed component/version, vendor affected-range/backport, reachable route and relevant telemetry**; allow `UNKNOWN`, `NOT_APPLICABLE_VERIFIED`, `POTENTIALLY_EXPOSED`, `EXPOSED_VERIFIED` as distinct statuses.
7. **Candidate rank:** multi-feature severity and service-criticality may order investigations, but numeric attack probabilities remain **unavailable until calibrated**. Keep `score_kind=UNCALIBRATED_RANK` vs `CALIBRATED_PROBABILITY` explicit.
8. **Response mapping:** decide between `NO_ACTION`, `INSUFFICIENT_TELEMETRY`, `WAIT_MORE_EVIDENCE`, `REVIEW_RECOMMENDED`, `PREPARE_MITIGATION`, `BAN_RECOMMENDED`, `INCIDENT_CRITICAL`; the last three are proposals, never executable commands.
9. **Policy separation:** require authenticated owner policy, safe source identity, effect scope, protected admin lists, TTL and rollback readiness. A Cloud hint, AI output or published CVE cannot bypass this gate.

**No universal numerical coefficients are approved for Brain V2 yet.** Weights must be derived from labelled benchmarks, with per-class/site coverage, false-positive tradeoff and versioned comparison against the deterministic no-ML baseline; use calibration measures (reliability curves/Brier score) only when labels and event base rates are valid.

## 5. Scenario & temporal threat reasoning

Scenarios are versioned **defensive case graphs**, not offensive playbooks and not direct nft commands. Each scenario describes:

- Threat family/ATT&CK technique IDs, environment applicability and **source/license**.
- Preconditions (e.g. public admin endpoint reachable, installed package affected), with explicit unknown state.
- Ordered **observables**: probing, repeated auth failure, anomalous connection fan-out, exploit-path requests, traffic impacts.
- Independent evidence requirements (socket, proxy, host resource, WAF, network telemetry), alternative benign explanations and falsifiers.
- Hypothesis stage (`RECON_SUSPECTED`, `ABUSE_POSSIBLE`, `IMPACT_CONFIRMED` etc.), lead indicators and prediction horizon.
- Proposed **non-disruptive preparations**: increase evidence collection, warn admin, verify patch coverage, stage short rate-limit proposal, request owner preauthorization. Live traffic mitigation uses a separate approval/rollback track.

Example **synthetic**: new KEV advisory + verified affected publicly exposed route + repeated suspicious paths in disjoint windows + app auth failures may raise `PREPARE_MITIGATION` early; without inventory/origin verification, only `INVESTIGATE` is justified. No assumption that all four indicators identify the same actor.

Model failure tests: one burst counted twice, fake X-Forwarded-For, benign scanner, deployment spike, resource outage without attack, 8-second long poll, stale KEV, package backport fixed despite upstream version string, aged cached inventory, false threat-feed positive, tenant crossover and manipulated logs.

## 6. Short/long-term memory and Cloud learning

- **Hot local:** bounded ring/time window, keyed by site/service/telemetry source; jitter and dropped-evidence accounting.
- **Cold local:** SQLite summaries, event IDs, validated recurring patterns and owner labels with TTL/retention. Avoid persisting full request bodies, credentials, raw traffic payloads. Site-owned histories never mix for enforcement.
- **Global curated pack:** canonical network/domain facts + ATT&CK/CVE/KEV/EPSS source links, signed with version, freshness, retraction and revocation.
- **Cloud experience exchange (future P7):** opt-in privacy-safe feature/campaign summaries with source diversity, benign outcome/negative labels, independent corroboration and poisoning defense. Deduplicate near-identical deployments/sensors; report data gaps and provenance coverage. A single compromised tenant cannot globally blacklist IP ranges or change executable policy.
- **Model releases (optional):** separate offline training, bounded clean data, holdout/time-forward/site holdout evaluation, signed model artifact, staged shadow rollout, per-site drift/rollback.

Cloud packs are read-only advice. The local Go Brain can work offline with its last verified pack; core nftables/TTL must work without Brain or Cloud.

## 7. Initial contracts and invariants to implement

Contract candidates in [knowledge-claim schema](../../schemas/knowledge-claim-v1alpha1.schema.json) and [decision proposal schema](../../schemas/decision-proposal-v1alpha1.schema.json); these only specify interchange fields and still require independent semantic validation. No JavaScript/Python text can be cast to root or shell commands.

Boundary invariants:
`EvidenceV1 -> KnowledgeClaimV1 / ScenarioV1 -> Go Brain -> DecisionProposalV1 -> Policy Gate -> constrained executor -> VerificationReceiptV1`.

Backpressure, untrusted publisher, privacy, review, cooldown, repeated alert deduplication, cross-tenant isolation and consent must be tested before staging.

## 8. Build order, verification and blockers

1. **P0 portable schemas + Go input validation + synthetic replay** with fail-closed unit tests. No enforcement.
2. **P1 local passive collectors + baseline + outcome ledger**, compare original Sentinel shadow logic, preserve human labels, quantify drift.
3. **P2 Brain deterministic detector/scenario graph + feed/inventory correlation**, Shadow proposals only; WAF read-only findings as evidence.
4. **P3 staged operator mitigations after rollback, site-scope and labelled FP gates**; optional ML worker can be tested separately.
5. **P4 disconnected multi-node/Edge capability**, fully usable without Cloud or NetBird.
6. After G-STANDALONE and owner consent, NetBird P5 independent roadmap; Cloud aggregation/multi-tenancy P7 when privacy/legal/reliability are separately proven.

Do not block foundational firewall P0/P1 on optional Brain Cloud or ML. A useful early warning means **better-justified, earlier recommendations with low measured false positives**, not a blanket guarantee of preventing unknown attacks.
