# Light Luna Project Book

Status: canonical navigation | Date: 2026-10-13
Scope: governed documentation entry point for engineers and AI agents

Read in this order, **before editing any documentation or code**:

1. [LIBRARY_RULES.md](LIBRARY_RULES.md) — mandatory source authority, filing, traceability, and update rules.
2. [LIGHT_LUNA_PROJECT_BOOK.md](LIGHT_LUNA_PROJECT_BOOK.md) — verified project overview, decisions, P0 evidence and open gates.
3. [DOCUMENT_LIBRARY_MAP.md](DOCUMENT_LIBRARY_MAP.md) — index of canonical, proposed, research, and historical documents.
4. [LIGHT_LUNA_CODE_MAP.md](LIGHT_LUNA_CODE_MAP.md) — real entry points, call graph, authority boundaries and source paths.
5. [SOURCE_CATALOG.md](SOURCE_CATALOG.md) / [source_catalog.json](source_catalog.json) — source bookmarks, exact paths and SHA-256.
6. [LIGHT_LUNA_WORKLOG.md](LIGHT_LUNA_WORKLOG.md) — durable chronological changes.
7. [PROJECT_TRUTH_STATUS.json](PROJECT_TRUTH_STATUS.json) — machine-readable snapshot and unresolved gates.
8. [DOCUMENT_TEMPLATE.md](DOCUMENT_TEMPLATE.md) — use when a new durable document is justified.
9. [ADR-020](../adr/ADR-020-go-core-python-optional.md) — Go-native core plus optional, isolated Python ML worker (architecture decision; not implemented).
10. [ADR-021](../adr/ADR-021-ebpf-observability-xdp-optional.md) — optional future eBPF observation/XDP mitigation; not active.
11. [ADR-022](../adr/ADR-022-rebuild-luna-brain-governed-knowledge-migration.md) — rebuild Brain V2; port reviewed knowledge rather than cloning live data/imagined model weights.
12. [Brain V2](../architecture/luna-brain-v2.md) and [Threat Knowledge Pipeline](../architecture/threat-intelligence-knowledge-pipeline.md) — proposed, independently reviewable component contracts.
13. [Legacy source audit](../history/audits/SENTINEL_LUNA_COGNITION_AUDIT_2026-10-13.md) — VPS-ARM observations, heuristic numbers and implementation limits.
14. [ADR-023](../adr/ADR-023-dual-brain-memory-cloud.md) — proposed dual-role brain, bounded learning and optional Cloud; **not yet decided or built**.
15. [Local Memory](../architecture/luna-memory-lifecycle.md) and [Cloud Intelligence](../architecture/luna-cloud-intelligence.md) — proposed resource and sync contracts.
16. [ADR-024](../adr/ADR-024-offline-base-independent-component-updates.md) and [offline component lifecycle](../architecture/offline-base-component-update-lifecycle.md) — DECIDED release design: bundled Basic Brain plus independent App/Brain/Knowledge/module updates; not yet implemented.
17. [ADR-025](../adr/ADR-025-portable-scalable-cloud.md) — DECIDED portability, scale-out and VPS/Compose→k3s relocation guarantees to test, **not deployed**.
18. [Cloud architecture](../architecture/cloud-portability-and-scale.md) and [migration/recovery](../architecture/cloud-migration-and-recovery.md) — portable runtime/state and operator rehearsal plan.
19. [ADR-026](../adr/ADR-026-unified-account-remote-fleet-control.md) and [unified account/fleet design](../architecture/unified-account-cloud-fleet-control.md) — Cloud and local share identity after optional enrollment, with offline owner recovery and local approval. Not implemented.
20. [ADR-027](../adr/ADR-027-collective-threat-immunity.md) and [immune-style threat intelligence](../architecture/federated-immune-threat-intelligence.md) — owner-decided shared behavior/scenario awareness, never global ban from one report. Not implemented.
21. [ADR-028](../adr/ADR-028-self-defense-host-compromise.md) and [Host Guardian architecture](../architecture/host-self-defense-and-compromise-containment.md) — protect Light Luna itself, detect supported host compromise, quarantine untrusted Cloud inputs; NOT IMPLEMENTED.
22. [ADR-029](../adr/ADR-029-local-observability-and-notifications.md) and [observability/notifications contract](../architecture/local-observability-and-notifications.md) — **DECIDED future** Prometheus/Grafana metrics, read-only bans/events and asynchronous Telegram/Discord/Slack/SMTP; no runtime endpoints/notifications implemented.
23. [../../README.md](../../README.md) — public product introduction and constraints.

## Latest Core P0 guest file-read safety (2026-10-15)

- [Synthetic VM manifest/journal opened-inode checks](../architecture/p0-vm-synthetic-stage-revert.md) — [Python guest reader](../../scripts/lab/vm/guest-stage-journal.py) now uses nofollow descriptor open, `fstat` single-link regular-file verification and 4097-byte bounded read to defend against stale path checks/size growth. [Five focused test methods](../../scripts/lab/vm/test_stage_manifest_ttl.py), including three RED→GREEN negatives; VM Python offline suite **103 PASS**. This does not prove guest/Host kernel rollback; live LXD journal status UNKNOWN.

## Latest Core P0 guest manifest strictness (2026-10-15)

- [Guest metadata validation](../architecture/p0-vm-synthetic-stage-revert.md) now rejects `policyRevision` masquerading as `true`/`1.0` and forged `effect` claims. [Seven-case focused offline suite](../../scripts/lab/vm/test_stage_manifest_ttl.py) includes five RED→GREEN negative mutations; complete VM Python offline suite **98 PASS**. Real Go CLI manifest accepted by guest parser; read-only LXD socket access denied, **VM live state UNKNOWN**.

## Latest Core P0 VM-stage identity correctness (2026-10-15)

- [TTL-bound synthetic stage ID contract](../architecture/p0-vm-synthetic-stage-revert.md) — [Go builder](../../internal/vmstage/plan.go) and [guest manifest verifier](../../scripts/lab/vm/guest-stage-journal.py) now both bind synthetic `stageId` to policy digest, isolated table and TTL. [RED→GREEN Go regression](../../internal/vmstage/plan_test.go) and [five new VM offline negative tests](../../scripts/lab/vm/test_stage_manifest_ttl.py) prevent a stale-ID TTL change; full offline VM suite now 96 cases. No production rollback or kernel claims.

## Latest Core P0 source verification — exact required keys (2026-10-15)

- [Required PolicyBundle key case/presence audit](../architecture/p0-policy-rule-conflict-audit.md) — initial real Go CLI mismatch and ten-case prepatch RED established; [exact case-sensitive key checker](../../internal/policy/required_keys.go) passes post-edit Go regression GREEN ×3. The expanded [real CLI/schema suite](../../scripts/test_p0_policy_schema_contract.py) passes 59/59, retaining optional-omission positives. Full Go suite/vet, process scripts, offline VM tests and cross-builds PASS. Non-enforcing source only.

## Latest Core P0 code and test evidence (2026-10-15)

- [PolicyBundle null-tolerant Go decoding corrected](../architecture/p0-policy-rule-conflict-audit.md) — [`policy.Decode`](../../internal/policy/model.go) uses opt-in [`strictjson.DecodeRejectNull`](../../internal/strictjson/decode.go) to reject explicit JSON nulls, preserving all other callers. [Real CLI parity suite](../../scripts/test_p0_policy_schema_contract.py) now checks **47 bounded cases**; [Go policy regression](../../internal/policy/policy_test.go) observed RED→GREEN. No privileged policy apply.

- [Metadata revision int64 schema parity](../architecture/p0-policy-rule-conflict-audit.md) — [JSON Schema](../../schemas/policy-v1alpha1.schema.json) now caps `metadata.revision` at the Go `int64` maximum. The offline [contract script](../../scripts/test_p0_policy_schema_contract.py) runs **38 bounded scenarios** against real unprivileged `lunactl` with `--verify-cli`; [Go boundary regressions](../../internal/policy/policy_test.go) cover successful maximum and invalid overflow/zero/negative values. No owner or firewall authorization.

## Latest Core P0 code and test evidence (2026-10-14)

- [Schema→real Go CLI contract verification](../architecture/p0-policy-rule-conflict-audit.md) — rejects NUL in policy notes; [reusable 32-case suite](../../scripts/test_p0_policy_schema_contract.py) optionally offline-builds `lunactl` and checks every JSON fixture through the actual non-enforcing Go validation command. Bounded parity only; Go remains authoritative.

- [Conflict analyzer machine-readable warnings](../architecture/p0-policy-rule-conflict-audit.md) — `COMPLETE_NO_FINDINGS`, reported overlaps and budget-incomplete results all carry unverified kernel, rescue and owner/Executor disclaimers in JSON; not an APPLY permit. [Go regression](../../internal/policy/conflicts_test.go) and development CLI validated.

- [Protocol/port policy schema parity](../architecture/p0-policy-rule-conflict-audit.md) — TCP/UDP explicit wildcard `destinationPort=0` accepted, while ICMP/ICMPv6 nonzero port rejected; [contract regression](../../scripts/test_p0_policy_schema_contract.py) verifies this alongside rescue requirements and reserved source aliases. Go is authoritative and there is no active firewall executor.

- [Deny-default rescue declaration schema parity](../architecture/p0-policy-rule-conflict-audit.md) — public JSON Schema now rejects `drop/reject` policies without enabled emergencyAccess and at least one CIDR, matching existing Go validation; [repeatable contract check](../../scripts/test_p0_policy_schema_contract.py) covers explicit positive/negative cases. Declaration-only, no verified rescue reachability or firewall APPLY.

- [Actual Gate HMAC/Ed25519 canonical signature hardening](../architecture/p0-local-control-gate.md) — development source and RED→GREEN negative tests on HOMESV confirm lowercase hex-only authenticators, denying alternate text representations before replay state changes; no human owner authorization, privileged executor or kernel enforcement.

- [Policy rule precedence and CIDR-set union audit](../architecture/p0-policy-rule-conflict-audit.md) — actual static Go policy analyzer, IPv4/IPv6 alias coverage, advisory-only boundaries; extension verified on HOMESV at local source commit `834100a6671c45404d6a52f5a0ef26441d2711f7`. Published Git SHA must be independently confirmed.
- [Core Code Map](LIGHT_LUNA_CODE_MAP.md), [chronological Worklog](LIGHT_LUNA_WORKLOG.md), [machine-readable Project Truth](PROJECT_TRUTH_STATUS.json) and [Source Catalog](SOURCE_CATALOG.md) are the canonical route to implemented modules, tests, and current release gates.
- **No standalone release or real kernel enforcement:** P0 OPEN, G-STANDALONE BLOCKED. Prometheus/Grafana/notification adapters are **future architecture only**, paused by owner priority.

## Source precedence

**Exact code at Git SHA + test evidence + actual runtime status** > current validated contracts > accepted ADRs > contemporary design/handoff > archived research. These levels are not interchangeable. A passing unit test is not evidence of a working firewall.

## Continuity

The local, overwrite-only `.luna/CURRENT_CHECKPOINT.json` (ignored by Git) may describe the current session state. It is **not** canonical evidence and is never a substitute for Git, Code Map or release/acceptance artifacts.

## Publication

Repository: [n8n2erpnext/light-luna](https://github.com/n8n2erpnext/light-luna), intentionally **public** per owner's 2026-10-13 instruction. Do not commit secrets, live firewall exports, personal IP histories, agent memory databases or raw telemetry.
