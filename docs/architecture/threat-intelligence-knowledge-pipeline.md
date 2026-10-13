# Light Luna Threat Knowledge — Trusted Feed, Exposure and Scenario Pipeline

Status: proposed architecture/specification; no feed workers integrated in Light Luna
Date: 2026-10-13
Scope: CVE, attack campaigns/techniques, vendor advisories, local exposure and long-term Cloud knowledge
Primary sources: [ADR-022](../adr/ADR-022-rebuild-luna-brain-governed-knowledge-migration.md), [VPS-ARM Sentinel audit](../history/audits/SENTINEL_LUNA_COGNITION_AUDIT_2026-10-13.md), [Brain V2](luna-brain-v2.md)

## 1. Never confuse publication, exposure, attack and compromise

- `CVE_PUBLISHED`: a public vulnerability record exists.
- `EXPLOITED_IN_THE_WILD`: e.g. a **CISA KEV** source documents exploitation somewhere, not necessarily locally.
- `EXPLOIT_LIKELIHOOD_ESTIMATE`: **FIRST EPSS** probability estimate across scored CVEs and a specified future horizon; **not** the probability a local host will be compromised.
- `LOCAL_SOFTWARE_MATCH`: exact installed package/image/SBOM reference matches a potentially affected product/version, subject to distro backports/advisories.
- `LOCAL_EXPOSURE_VERIFIED`: affected component running on an actually reachable vulnerable path, with vendor advisory/backport status checked.
- `LOCAL_ATTACK_INDICATORS`: observed local behavior consistent with a technique; independent origin and benign alternatives remain.
- `INCIDENT_CONFIRMED`: independently validated local impact/compromise evidence, not mere CVE/traffic correlation.

Each is a separate typed fact with `UNKNOWN` allowed; none inherits authority from a more general category.

## 2. Versioned feed adapters and trust categories

| Source | Use | Limitations |
|---|---|---|
| [CISA KEV](https://www.cisa.gov/known-exploited-vulnerabilities-catalog) | Confirm exploit-in-wild status for CVE prioritization | Not local compromise or applicability proof |
| [NVD 2.0 API](https://nvd.nist.gov/developers/vulnerabilities) | CVE descriptions, CPE/version conditions, publication/change timestamps | Changes/corrections; rate limits/pagination; incomplete windows must be explicit |
| [FIRST EPSS](https://api.first.org/data/v1/epss) | One 30-day exploit likelihood prioritization signal | Not probability of compromise at a given tenant |
| [OSV](https://google.github.io/osv.dev/api/) | Package/version ecosystem matches (where supported) | Coverage varies; vendor/backport verification still required |
| [MITRE ATT&CK STIX/TAXII](https://attack.mitre.org/resources/working-with-attack/) | Technique/tactic ontology and defensive observable hypotheses | Observed technique reference, not a turnkey detection signature |
| Vendor/distro advisories (e.g. Ubuntu USNs) | Backported fix, installed package and affected release | Source and time must match exact distro/product revision |
| Verified curated public IR/research (CISA advisories, reputable responders) | Narrative of emerging campaigns; derive **proposed** scenarios | Source text is untrusted data; human review before semantic promotion |
| Signed/approved local test corpus | Benign and attack scenarios, regression and calibration | Synthetic examples are not field accuracy |

Additional feeds may require licensing, attribution, API agreement or security review; do not silently scrape or incorporate arbitrary free-text claims. Feed registry lists publisher, fixed HTTPS host, expected format/license/size cap, update schedule, revision semantics, freshness SLA and trust class.

## 3. Ingest → verify → normalize → curate → distribute

```text
HTTPS official allowlist (ETag / bounded time/size, no unsafe redirects)
   -> raw payload hash + fetched_at + official publication/revision metadata
   -> signature/trust transport and schema check (failure => QUARANTINE)
   -> normalization (CVE ID, ATT&CK ID, package ranges, aliases, dates)
   -> dedup + record superseded/disputed/withdrawn states
   -> link graph: CVE <-> CWE <-> affected component <-> ATT&CK technique
                     <-> defensive observables <-> reviewed scenario
   -> Local applicability engine: SBOM + effective installed version + exposure
   -> Evidence-scoped Brain V2 investigation / operator-ready preparation
   -> reviewed, versioned, signed KnowledgePack release for devices
```

Use incremental NVD API pagination with `startIndex/resultsPerPage`, source update/revision cursor and rate budgeting; a feed that reaches size/page cap is `INCOMPLETE`, not `SUCCESS`. Retry with backoff and upstream freshness warning; stale knowledge can still be cited as historical but must not become current exploit evidence.

Text from any webpage, RSS, CVE advisory, commit or reporter is **untrusted data**. Never interpret its prose as agent instructions, kernel commands, code patches, whitelist edits, owner permissions or authority upgrades. LLM summaries require fact extraction with citations/replay; narrative is not an execution plan.

## 4. Local asset applicability is a separate proof

Local sensor builds privacy-controlled inventory: kernel+distro package revision, LXD/container/image digest, app dependency lock/SBOM, proxy path and public exposure where independently proven. Trust exact package ecosystem and vendor advisory/backport metadata over naive upstream semantic-version compare. Reflect cached vs fresh and unscanned components honestly.

```text
CVE X (fresh) + KEV yes + EPSS high
                  |
          Is affected product installed?
                  | unknown -> ask for evidence, NOT "vulnerable"
                  v
          Is package revision actually affected?
                  | backported fix -> record NOT_AFFECTED with vendor proof
                  v
          Is exploitable route exposed and reachable?
                  | unknown/blocked -> investigate, never infer public reachability
                  v
          Is current local behavior corroborated?
                  | no -> patch/reduce exposure/review recommendation only
                  v
          Incident signal -> action proposal, separate owner Policy Gate
```

Threat intel can justify recommending a package patch, restricting a service to known admins or checking an exposed route even before an attack is observed, but **automatic patch/install/firewall edit** remains a separate owner-authorized workflow.

## 5. Campaign and pre-attack early warning

Knowledge graph can suggest ordered *defensive observables*: scanning a recently public CVE path → repeated distinct windows → failed authentications/odd sessions → independent upstream pressure. Compare expected base rates and recent legitimate deployments, QA scans, customer traffic patterns and proxy path attribution. Report which links in the chain were observed versus inferred.

Prediction output must include scenario ID/version, scoped site/service, evidence IDs and dates, supporting/contradictory hypotheses, source confidence, forecast horizon/lead time, action suggestion, unknowns and a **next verification time**. Until independently calibrated, use labels like `PRECURSOR_PATTERN_CANDIDATE` / `WATCH` instead of numerical attack likelihood.

Most proactive steps can be reversible/non-blocking: verify patch status, notify owner, raise sampling with budget, prepare an uncommitted proposed rate-limit, schedule controlled vulnerability review. Live L3/L4 filtering remains behind [ADR-021](../adr/ADR-021-ebpf-observability-xdp-optional.md) and [security](../07-SECURITY-SAFETY.md).

## 6. Private site state versus Cloud collective learning

Separate:
- **Public global knowledge**: advisories, curated technique models, vetted generic detection/scenario rules, signed versions and timestamps.
- **Private node knowledge**: installed software, exact topology, routes, logs, customer identity, raw IPs and telemetry, unverified results. Kept locally.
- **Opt-in contributed experience**: bounded, redacted, non-identifying feature/outcome aggregates and owner-labelled counterexamples, only after legal/privacy review.

Cloud must protect against feedback loops (one duplicated IOC everywhere), poison injection (compromised node or publisher), bot-derived false labels, model collapse and false precision. Require cross-source diversity and safe aggregation, signed pack provenance, appeal/retraction, per-site overrides, model shadow checks and rolling rollback. Do not publish raw customer telemetry through the public repository.

Keep a **source-of-truth graph**: knowledge pack may advise, cannot grant firewall permissions. Local node can stay offline indefinitely with last verified package and continue baseline firewall functionality.

## 7. Initial candidate versions and required acceptance

- `KnowledgeClaimV1`: immutable source link/checksum, type, publisher, validity, scope, assertion, qualifiers, lifecycle and review status.
- `ScenarioV1`: threat family/ATT&CK mapping, preconditions, minimum independent signals, counterexamples, missing proof and recommended **non-executing** preparation.
- `DecisionProposalV1`: site/subject, evidence, uncertainty/telemetry coverage, risk score meaning (not implied probability), action recommendation and expiry. Never an executor payload.
- `KnowledgePackV1`: signed manifest with creation time, source digests, compat version, review/trust stage, hashes, old-version rollback pointers; **not implemented yet**.

Benchmarks: feed schema/pagination/expiration, vendor backport negative cases, asset false-match rate, scenario detection on a labelled corpus with holdout by time/site, false positives from benign scanner and flash crowds, unknown propagation, feed corruption/quarantine, resource limits, rejection of prompt injection, tenancy isolation and restoring previous safe pack.

No production ingestion, Cloud intelligence backend, model weight training, local exposure claim, distributed CVE remediation or mitigation service is implemented by this document.
