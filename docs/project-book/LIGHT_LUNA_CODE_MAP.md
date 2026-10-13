# Light Luna Code Map

Edition: 0.1 source-derived P0 baseline
Status: verification snapshot of development branch, **not a production runtime map**
Date: 2026-10-13
Git: `feat/p0-standalone-foundation@9ae27b267e72e465ea7808c3936fabcf6b0cac05`; main `66920c251603b0b0c4bcd0d0cfc94ca8f7220d41`
Primary sources: [CLI](../../cmd/lunactl/main.go), [typed schema](../../internal/policy/model.go), [validation](../../internal/policy/validate.go), [preview](../../internal/policy/preview.go), [tests](../../internal/policy/policy_test.go)

## Reading contract (borrowed from LightBI)

For each surface ask: (1) reachable from what entry point? (2) owns, adapts, executes, or presents which truth? (3) branch/commit vs untracked state? (4) which real source/test to read next? File presence is not deployment proof. No background WebUI service, kernel supervisor, Luna process or executable policy apply exists in this baseline.

## Current executable P0 foundation (2026-10-14; source-level verification only)

[Local runtime foundation](../architecture/local-runtime-foundation.md) supersedes the **older snapshot below** for current source reachability. It is still **not deployed** and only operates on local files or synthetic replay. Git SHA: use the tip of `feat/p0-standalone-foundation` at evaluation time rather than the historical `9ae27b2` in the original sections.

| Entry point / chain | Actual imported module | Side effect | Verified tests |
|---|---|---|---|
| `cmd/lunactl system snapshot` | [offline.Snapshot](../../internal/offline/replay.go) → [guardian.Snapshot](../../internal/guardian/basic.go), [brain.Version](../../internal/brain/basic.go) | Print honest component health/version statuses; no host inspection | [offline tests](../../internal/offline/replay_test.go) + CLI |
| `cmd/lunactl brain replay testdata/offline-brain-replay.synthetic.json` | [strictjson](../../internal/strictjson/decode.go) → [foundation](../../internal/foundation/observation.go) → [offline.Replay](../../internal/offline/replay.go) → [brain.Evaluate](../../internal/brain/basic.go) → [gate.Preview](../../internal/gate/advisory.go) | Read a synthetic local file and output advisory JSON only, `kernelWrites=false` | [Replay tests](../../internal/offline/replay_test.go) |
| `cmd/luna-brain version/replay` | [separate CLI executable](../../cmd/luna-brain/main.go), same offline libraries | Separate binary compilation; one-shot simulation, **no persistent daemon or IPC isolation** | Go test, Linux amd64/arm64 cross-build, replay run |
| `cmd/lunactl firewall policies validate/preview` | [policy.Decode](../../internal/policy/model.go) → [strictjson](../../internal/strictjson/decode.go) → [Validate/MakePreview](../../internal/policy/validate.go) | Files + no-enforcement preview only | [Policy regression](../../internal/policy/duplicate_keys_test.go), [strictjson tests](../../internal/strictjson/decode_test.go) |
| `internal/guardian` | [basic.go](../../internal/guardian/basic.go) | `UNKNOWN` placeholder; **no kernel/process/file scanning** | [offline tests](../../internal/offline/replay_test.go) |

**Still absent:** actual agent service, local API, WebUI, Basic Brain trained model/persistent runtime, base KnowledgePack/signature validator, independent privileged executor, nftables compiler/runner, policy-apply/rollback watchdog, Cloud account/sync, eBPF/XDP attachment and Host Guardian scanner. Module/App version fields are *dev strings*, not signed independently upgradable packages. The original source map below is historical and should not be read as the current reachability list.

## Workspace topology

```text
light-luna/
  cmd/lunactl/main.go              Go CLI executable, only active entry point
  internal/policy/model.go         PolicyBundle data types + Decode reader
  internal/policy/validate.go      Deterministic semantic guardrails
  internal/policy/preview.go       Deterministic inert preview DTO
  internal/policy/policy_test.go   Unit/negative tests
  schemas/policy-v1alpha1.schema.json   Declarative schema, partial semantic complement
  testdata/host-baseline.json      Synthetic TEST-NET fixture, never live
  go.mod                           Go module and minimum language version
  docs/00-16*.md                   Architecture/decision/history library, no executable
  docs/project-book/              Governance / map / evidence bookmarks
  scripts/update_doc_catalog.py    Source catalog regeneration tool
```

## Executable reachability

```text
Go binary cmd/lunactl/main.go::main()
  -> run(os.Args[1:])
       -> "version": print version string; no network/kernel
       -> "firewall policies validate <file>"
            -> os.Open(file) -> internal/policy.Decode(reader)
              -> capped JSON decode -> Bundle.Validate()
              -> print VALID on success; no network/kernel
       -> "firewall policies preview <file>"
            -> os.Open(file) -> Decode -> MakePreview(Bundle)
              -> Validate + stable JSON hash + sort copy by priority
              -> JSON response (effects=none, enforceable=false)
       -> anything else: usage+error
```

No `apply`, `stage`, `rollback`, `ban`, `netbird` subcommand; no daemon or listener. The preview is **not** nft-ready syntax and does not query kernel rules or prove emergency reachability. [ADR-020](../adr/ADR-020-go-core-python-optional.md) records the planned Go-native Brain and optional Python ML-worker boundary, not currently reachable code.

## Modules, authority, change path and tests

| Surface | Reachability / authority | Key source | Test/evidence | Side effect |
|---|---|---|---|---|
| `cmd/lunactl` | CLI entry, presents validated results | [main.go](../../cmd/lunactl/main.go) | CLI Go run tested at SHA above | read local file, stdout/stderr |
| `internal/policy.Decode` | Called by CLI; parses bounded JSON | [model.go](../../internal/policy/model.go) | [policy_test.go](../../internal/policy/policy_test.go) | none |
| `Bundle.Validate` | Called by Decode + Preview, **owns P0 semantic constraints** | [validate.go](../../internal/policy/validate.go) | Negative checks/fixture | none |
| `MakePreview` | Called only by preview, **adapts** policy to inert view | [preview.go](../../internal/policy/preview.go) | Determinism and no input mutation test | none |
| JSON Schema | Declarative public contract; not invoked by CLI currently | [schema](../../schemas/policy-v1alpha1.schema.json) | JSON syntax check only; parity testing pending | none |
| Fixture | Test-only, not runtime policy | [host-baseline.json](../../testdata/host-baseline.json) | Test source | none |
| Documentation catalog | Local docs tooling; not runtime/production | [script](../../scripts/update_doc_catalog.py) | Catalog verify mode | updates catalog Markdown + JSON only |

## Deliberately absent / planned path (NOT REACHABLE)

```text
React+TypeScript WebUI (PROPOSED)
       -> local authenticated Go API (PROPOSED, no listener)
           -> policy gate / owner identity (PROPOSED)
               -> separate constrained Go lunalockd (PROPOSED)
                   -> nftables-owned rules + independent rollback watchdog (PROPOSED)
       -> Go Luna Brain (DECIDED DIRECTION; NOT IMPLEMENTED, unprivileged)
           -> optional non-root Python ML worker (DECIDED OPTION; NOT IMPLEMENTED)
       -> SQLite WAL (PROPOSED, not used)
```

The architecture deliberately keeps rule authorization independent of Luna Brain and Internet/cloud status. NetBird code is **not allowed** before G-STANDALONE plus owner's approval. See [architectural model](../01-ARCHITECTURE.md) and [hard order](../11-ROADMAP.md).

## eBPF/XDP planned hooks — NONE REACHABLE

[ADR-021](../adr/ADR-021-ebpf-observability-xdp-optional.md) records an **architecture decision only**. There are NO eBPF source objects, Go BPF loader imports, pinned maps, XDP link creation, tracepoint attachments, counters, security event streams or XDP drop path in current Light Luna code. Proposed future chain: kernel tracing hooks or NIC ingress XDP → optional least-privilege loader → Go collector → Go Brain → independent Policy Gate → constrained executor. Existing kernel XDP programs on development machines are **external** and are not Light Luna artifacts. Neither BPF/XDP nor NetBird enters standalone runtime as a required dependency.

## Future Brain V2 architecture contracts (DESIGN ONLY; NOT REACHABLE)

The VPS-ARM [source audit](../history/audits/SENTINEL_LUNA_COGNITION_AUDIT_2026-10-13.md) inspected the independent legacy Python observer/Guardian/intel services. They are **outside this Go module**, cannot be treated as Light Luna runtime and remain unchanged. [ADR-022](../adr/ADR-022-rebuild-luna-brain-governed-knowledge-migration.md) chooses new Go Brain V2, porting only reviewed knowledge and safe logic. Proposed [architecture](../architecture/luna-brain-v2.md) and [intel pipeline](../architecture/threat-intelligence-knowledge-pipeline.md) define candidate entry points and state/proposal contracts. The committed `schemas/knowledge-claim-v1alpha1.schema.json` and `schemas/decision-proposal-v1alpha1.schema.json` are **static schema designs** only: `cmd/lunactl` does not import/validate/execute them and no event feed/ML/Cloud component consumes them. Adding these files has **zero kernel, network or model runtime effect**.

## Dual-role Brain, local memory and Cloud (PLANNED ONLY)

[ADR-023](../adr/ADR-023-dual-brain-memory-cloud.md) is **PROPOSED**, not yet owner-finalized. [Memory budgets](../architecture/luna-memory-lifecycle.md) and [Cloud topology](../architecture/luna-cloud-intelligence.md) are sizing/architecture hypotheses; do not infer deployed client databases, memory quota enforcement, hemispheres, signed knowledge updater, Cloud API or tenant services from docs file presence. The only reachable source remains the Go non-enforcing P0 CLI. No legacy neural weights have been found in scoped Sentinel source, and no network sync occurs from Light Luna.

## Offline package and component lifecycle (DECIDED DESIGN; NO EXECUTABLE IMPLEMENTATION)

[ADR-024](../adr/ADR-024-offline-base-independent-component-updates.md) establishes a **mandatory packaged Basic Brain and knowledge baseline** for the first shipping Light Luna release. [Install/update design](../architecture/offline-base-component-update-lifecycle.md), [component release manifest](../../schemas/component-release-manifest-v1alpha1.schema.json), [installed state](../../schemas/component-install-state-v1alpha1.schema.json) and [fictional offline fixture](../../testdata/offline-component-state.example.json) are specification data only. There is **NO** componentized installer, updater, Brain process, signed manifest verifier, local WebUI version page, rollback manager or Cloud sync in current reachable Go code. The two schemas are not imported or semantically validated by `cmd/lunactl`. The only active command remains the inert policy CLI.

## Portable Cloud and migration design (ADR-025; NOT REACHABLE)

[ADR-025](../adr/ADR-025-portable-scalable-cloud.md) establishes a Cloud **product architecture decision**, not a runnable service. [Cloud architecture](../architecture/cloud-portability-and-scale.md), [cutover & DR runbook](../architecture/cloud-migration-and-recovery.md), [migration inventory schema](../../schemas/cloud-migration-inventory-v1alpha1.schema.json) and its [synthetic TEST_ONLY fixture](../../testdata/cloud-migration-inventory.synthetic.json) specify future Go API/worker, PostgreSQL, object-storage adapters and migration evidence. No Cloud API listener, database migration runtime, job outbox, backup archive, provider endpoint, signing service, DNS changes, production connection, S3 bucket or Kubernetes manifest currently exists in reachable source. Client/Brain/Cloud versions remain separately designed, not operational. Avoid inferring a deployed Cloud from the design files.

## Unified account/local Cloud remote configuration (ADR-026; NOT IMPLEMENTED)

[ADR-026](../adr/ADR-026-unified-account-remote-fleet-control.md) decides one Cloud OIDC human identity is used by enrolled local and Cloud WebUIs, while local emergency credentials and firewall authority remain independent. [Account + Fleet control design](../architecture/unified-account-cloud-fleet-control.md), [typed fleet change envelope](../../schemas/fleet-config-intent-v1alpha1.schema.json) and [synthetic fixture](../../testdata/fleet-config-intent.synthetic.json) specify a future local auth server, Cloud tenant/device enrollment, RBAC, outbound agent command polling, actual/desired configuration and verified receipts. **No auth handler, credential database, enrollment protocol, Cloud Fleet UI, queue, agent command processing or configuration apply exists in current P0 source.** The JSON Schema is not a reachable validation endpoint and forbids arbitrary shell structurally but cannot enforce identities by itself.

## Federated Threat/Behavior Memory (ADR-027; ARCHITECTURE ONLY)

[ADR-027](../adr/ADR-027-collective-threat-immunity.md) and [federated design](../architecture/federated-immune-threat-intelligence.md) specify a future CrowdSec-inspired behavior scenario and shared intelligence pipeline. The static `schemas/threat-signal-v1alpha1.schema.json` / `schemas/threat-bulletin-v1alpha1.schema.json` and `testdata/threat-*.synthetic.json` do **not** implement signal capture, client consent, event parsing, scenario execution, Cloud consensus, signing, caching, gossip/transport, global blocklist or automatic remote decisions. Only Go CLI policy validator/preview is reachable. The design strictly separates local block receipts from adjudicated attack labels; multi-tenant Cloud and distributed reputation remain P7 optional future work.

## Self-defense / Host Guardian / Cloud quarantine (ADR-028; DESIGN ONLY)

[ADR-028](../adr/ADR-028-self-defense-host-compromise.md), [Host Guardian spec](../architecture/host-self-defense-and-compromise-containment.md), [static integrity report schema](../../schemas/host-integrity-report-v1alpha1.schema.json) and [synthetic fixture](../../testdata/host-integrity-report.synthetic.json) define required future guardian/service boundaries. They provide **no runnable process scanner, eBPF/LSM/BPF attachment, XDP guard, malware file scanner, tamper monitor, attestation, Cloud contamination filter, revocation code, response/remediation or kernel privileges**. All previous runnable Go code remains only the P0 inert policy CLI. A successful schema test cannot detect malware or prove a device clean.

## Current verification and risks

HOMESV unit tests, vet, CLI validation/preview and amd64/arm64 cross-builds passed for baseline `9ae27b2`. No WebUI E2E, Linux namespace test, kernel compatibility check, privileged executor, secure install/uninstall, actual live recovery, independent release signatures or real ban/counter proof exists.

**Known P0 gaps:** duplicate JSON field detection, deeper JSON Schema semantic parity, real nft compiler, alias recursion/FQDN, conflict visibility, rule simulator, rollback/watchdog, SQLite and service authentication. Some restriction defaults are P0 validator design decisions, not universal firewall semantics.

## How to safely modify

- Parsing/input limits → `internal/policy/model.go`; add negative tests.
- CIDR, aliases, rules, policy constraints → `internal/policy/validate.go` + schema + fixture/test.
- Preview output/hash → `internal/policy/preview.go`, preserving non-enforcing contract.
- CLI surface → `cmd/lunactl/main.go`; do not add real apply without separately designed executor and owner-approved test gates.
- Docs and library → `docs/project-book/README.md` then governance; regenerate catalog and verify links.

## Evolution rules

Update this Code Map **after** reviewing changed entry points/imports, exact Git SHA and test evidence. Keep planned graph separate from reachable graph. On future production, introduce distinct production entry maps (UI, control API, executor, collector, brain) with deployment artifact identity; don't silently use branch HEAD as production version.
