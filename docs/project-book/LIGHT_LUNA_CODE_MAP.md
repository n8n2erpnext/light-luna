# Light Luna Code Map

Edition: 0.1 source-derived P0 baseline
Status: verification snapshot of development branch, **not a production runtime map**
Date: 2026-10-13
Git: `feat/p0-standalone-foundation@9ae27b267e72e465ea7808c3936fabcf6b0cac05`; main `66920c251603b0b0c4bcd0d0cfc94ca8f7220d41`
Primary sources: [CLI](../../cmd/lunactl/main.go), [typed schema](../../internal/policy/model.go), [validation](../../internal/policy/validate.go), [preview](../../internal/policy/preview.go), [tests](../../internal/policy/policy_test.go)

## Reading contract (borrowed from LightBI)

For each surface ask: (1) reachable from what entry point? (2) owns, adapts, executes, or presents which truth? (3) branch/commit vs untracked state? (4) which real source/test to read next? File presence is not deployment proof. No background WebUI service, kernel supervisor, Luna process or executable policy apply exists in this baseline.

## KVM separate-kernel nft grammar and VM-only synthetic watchdog (2026-10-14 — LIVE LAB PASS)

- [`scripts/lab/vm/run-tests.sh`](../../scripts/lab/vm/run-tests.sh) checks LXD VM type, exact profile and kernel != HOMESV, KVM, no NIC; builds real [`cmd/lunactl/main.go`](../../cmd/lunactl/main.go) and extracts [`nftpreview.Compile`](../../internal/nftpreview/compile.go)'s **unhooked** script from [synthetic policy](../../testdata/host-baseline.json), asserts `kernelWrites=false` / syntax unchecked on host then sends bytes into VM.
- [`scripts/lab/vm/guest-nft-test.sh`](../../scripts/lab/vm/guest-nft-test.sh) runs guest `nft -c` on unhooked Light Luna script **without apply**, bad-syntax rejection, two transient netns/veth+ping packet baseline, a **synthetic—not product-generated—hook/drop+counter**, and independent VM watchdog process deleting only that isolated synthetic table; checks ping recovers. **Executed twice PASS** in real VM. Cleanup eliminates namespace and rules after each test. [Runbook](../../scripts/lab/vm/README.md); [precise trust boundary](../architecture/p0-lxd-kernel-nft-rollback.md).
- [`scripts/lab/vm/create-vm.sh`](../../scripts/lab/vm/create-vm.sh) creates future identical VM only after separately installed LXD, cached official VM image and pool; refuses existing VM/profile or NIC. [`delete-vm.sh`](../../scripts/lab/vm/delete-vm.sh) is opt-in destructive cleanup with exact VM marker and acknowledgement. **Actual deployed instance**: `light-luna-p0-nftvm`, Ubuntu 24.04, 2 vCPU/2048MiB, no NIC, KVM separate guest kernel. **No product executor or hooked nft policy can be reached through these scripts.**

## Networkless LXD distinct-UID test harness (2026-10-14 — EXECUTED CROSS-UID LAB)

**Reachable actual host operations:** `sudo snap install lxd --channel=5.21/stable` completed under explicit approved device capabilities; `lxd init --minimal` created default dir pool and unexpected `lxdbr0`, which was deleted after removing its default-profile NIC. `LUNA_LXD_USE_SUDO=1 LUNA_LXD_POOL=default bash scripts/lab/lxd/create-lab.sh` created a **no-NIC unprivileged container**. `PATH=/home/ubuntu/.local/opt/go1.27.2/bin:$PATH LUNA_LXD_USE_SUDO=1 bash scripts/lab/lxd/run-tests.sh` ran actual cross-user guest tests; terminal output ends with `LXD_DISTINCT_UID_POSITIVE_NEGATIVE_RESTART_PASS`. This is a **development container sharing kernel**, not a VM packet filtering or production service admission result.



- [`scripts/lab/lxd/check-host.sh`](../../scripts/lab/lxd/check-host.sh): **read-only** LXD prerequisites, daemon and storage/network inventory. The pre-install BLOCKED state is historical; LXD is now installed and the real guest run succeeded.
- [`create-lab.sh`](../../scripts/lab/lxd/create-lab.sh): **future explicit owner invocation** after LXD installation; refuses existing profile/instance, creates dedicated profile + root disk only, 2 vCPU/1536MiB, initializes without NIC and checks expanded config before start; **completed on HOMESV after LXD installation**.
- [`run-tests.sh`](../../scripts/lab/lxd/run-tests.sh) builds static `luna-gate`/`luna-agent` and pushes synthetic fixtures/binaries into a marker-checked LXD container; [`guest-test.sh`](../../scripts/lab/lxd/guest-test.sh) intends to create different **guest-only** UIDs/groups and test good Agent, outsider, wrong UID with copied *test key*, Gate UID pin, journal restart/corruption. **Real guest execution PASS** on separately mapped Linux UIDs; see verification markers in current section.
- [`delete-lab.sh`](../../scripts/lab/lxd/delete-lab.sh) permits explicit removal only of a marked instance with exact owner acknowledgement, no storage pool or default profile deletion. [Truth and conditions](../architecture/p0-lxd-cross-uid-lab.md). The LXD container **shares host kernel**, so this harness is **not** an nftables/eBPF acceptance environment.

## Ed25519 Agent/Gate signer separation + group Unix socket (2026-10-14 — DEV SOURCE ONLY)

- [`controlgate.SignEd25519`/`NewEd25519Boundary`](../../internal/controlgate/control.go) adds signature-based request auth with domain separation and no HMAC downgrade; [`NewEd25519BoundaryWithJournal`](../../internal/controlgate/journal_linux.go) reuses the restart replay journal. [Negative tests](../../internal/controlgate/ed25519_test.go) cover bad public key, mixed signature, tampering, replay, signed APPLY denial.
- [`identitykeys.Init/LoadPrivate/LoadPublic`](../../internal/identitykeys/ed25519_linux.go) stores private Ed25519 key **only** in Agent 0700 dir and public key **only** in Gate 0700 dir; 0600 file owner/link checks. [Key negative tests](../../internal/identitykeys/ed25519_linux_test.go).
- [`localipc.ListenGroup`](../../internal/localipc/group_linux.go) checks pre-existing Gate UID/IPC GID/0710 directory and 0660 socket, and [`CallExpectedGateUID`](../../internal/localipc/unix_linux.go) pins server kernel peer UID; [group tests](../../internal/localipc/group_linux_test.go). [`luna-gate serve-ed25519`](../../cmd/luna-gate/main.go) and [`luna-agent preview-ed25519`](../../cmd/luna-agent/ed_linux.go) are two separate Linux processes; [E2E](../../scripts/test_p0_ed25519_identities.sh) exercises same-UID demo, restart replay and denies kernel apply. **Not** actual distinct service-UID proof/installation/owner login. [Full contract](../architecture/p0-ed25519-agent-gate-identity.md).

## Two actual Linux processes and bounded replay journal (2026-10-14 — DEV SOURCE ONLY)

- [`cmd/luna-gate/main.go`](../../cmd/luna-gate/main.go): `init <private-dir>` → [`devkeys.Init`](../../internal/devkeys/keys_linux.go) creates new nonroot 0700 directory/0600 throwaway key; `serve <policy> <dir> <1..30s>` → `policy.Decode` → `devkeys.Load` → [`controlgate.NewBoundaryWithJournal`](../../internal/controlgate/journal_linux.go) → [`localipc.ListenPrivate/Serve`](../../internal/localipc/unix_linux.go). **Effects**: private temporary-files for test sessions, local Unix socket and append-only replay journal with fsync; no kernel/network firewall, privileged process or Cloud.
- [`cmd/luna-agent/main.go`](../../cmd/luna-agent/main.go): `preview|apply <policy> <dir> [id nonce]` → test key read + exact policy SHA → `controlgate.Sign` → `localipc.Call` → JSON receipt. Two separate OS PIDs share the same **test** UID/key; this is **not OS service isolation or authorized human OWNER**. A valid signed `APPLY_POLICY` is still `REJECTED`.
- [`journal_linux_test.go`](../../internal/controlgate/journal_linux_test.go), [`devkeys/keys_linux_test.go`](../../internal/devkeys/keys_linux_test.go) and [`scripts/test_p0_gate_processes.sh`](../../scripts/test_p0_gate_processes.sh) verify persistence after Gate restart, same replay rejection, corruption/locking/insecure path fail-closed and separate process IPC. [Source-level truth and limitations](../architecture/p0-two-process-gate-and-journal.md). Previous `lunactl` in-process selftests are retained and labeled accordingly.

## Linux Unix socket IPC transport (2026-10-14 — SOURCE IMPLEMENTED)

`cmd/lunactl.run` → `control ipc-selftest <policy>` → `policy.Decode` → [`localipc.SelfTest`](../../internal/localipc/selftest_linux.go) → [`ListenPrivate/Serve/Call`](../../internal/localipc/unix_linux.go) → kernel Unix domain socket + `SO_PEERCRED` matching UID + bounded length-prefix JSON → [`controlgate.Boundary.Evaluate`](../../internal/controlgate/control.go). [Linux socket tests](../../internal/localipc/unix_linux_test.go) additionally spawn an actual **different OS process** as client, deny mismatched UID, oversize frames, malformed JSON and insecure socket paths. **Only developer CLI has transport reachability; no persistent daemon/service or separated key holder, and no executor**. Requests are `ADVISORY_ONLY` or `REJECTED`, always `kernelWrites=false`; cross-user UID and lifecycle tests in isolated VM remain. [Verification contract](../architecture/p0-linux-unix-ipc.md).

## Local control gate and developer selftest (2026-10-14 — SOURCE IMPLEMENTED)

`cmd/lunactl.run` → `control selftest <policy-file>` → [policy.Decode](../../internal/policy/model.go) → [controlgate.SelfTest](../../internal/controlgate/selftest.go) creates one random per-run in-memory 32-byte HMAC key → [controlgate.NewBoundary/Evaluate](../../internal/controlgate/control.go). [Tests](../../internal/controlgate/control_test.go) check signed preview, MAC tampering, policy revision/content drift, TTL, site/actor origin, request/nonce replay and concurrency. Receipts are **ADVISORY_ONLY or REJECTED**, always `ExecutorStatus=NOT_AVAILABLE`, `KernelWrites=false`. There is **NO local control socket/HTTP API, verified OS peer, persistent replay, privileged executor, real rollback or Cloud request path**. [Exact contract](../architecture/p0-local-control-gate.md). Earlier historical baseline map sections must not be taken as the reachable tip.

## Pure-Go policy-intent flow simulator (2026-10-14 — SOURCE IMPLEMENTED)

Reachable `cmd/lunactl.run` command `firewall policies simulate <policy.json> <flows.json>` → [`policy.Decode`](../../internal/policy/model.go) → [`strictjson.Decode`](../../internal/strictjson/decode.go) → [`flowsim.Fixture.Validate/Simulate`](../../internal/flowsim/simulate.go) → JSON report with `kernelWrites=false`, per-flow `effectiveVerdict=UNKNOWN` and `externalFirewall=UNKNOWN`. [Tests](../../internal/flowsim/simulate_test.go) use a rich deterministic test policy and [synthetic CLI fixture](../../testdata/flow-fixtures.synthetic.json) to cover rule ordering, IPv4/IPv6 aliases, management lockout warnings, log nontermination, input default, unspecified output and malformed/overlimit cases. **No actual kernel/routing, conntrack, NetBird, rescue path or firewall effects.** [Implementation/acceptance contract](../architecture/p0-flow-simulator.md). Previous Code Map base sections are historical.

## Unhooked nftables policy compiler (2026-10-14 — SOURCE IMPLEMENTED)

`cmd/lunactl.main` → `firewall policies nft-preview <file>` → [policy.Decode](../../internal/policy/model.go) (strict JSON + semantic validation) → [nftpreview.Compile](../../internal/nftpreview/compile.go) → deterministic JSON `Plan`. [Tests](../../internal/nftpreview/compile_test.go) include priority order, mixed IPv4/IPv6, unknown family/protocol rejection, expansion limits and safety against note injection. The module **only generates unhooked regular-chain text**, cannot call `nft`, access a kernel or implement emergency rescue. It makes `syntaxChecked=false` and `kernelCompatibilityChecked=false` explicit. [Verification/limits](../architecture/p0-nft-unhooked-compiler.md) note that HOMESV unprivileged user/network namespaces are unavailable, therefore **no `nft -c` grammar/kernel test** was run. Both `lunactl` and `luna-brain` remain buildable Linux amd64/arm64. The older source map sections below describe the historical baseline and do not override this new reachable path.

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
