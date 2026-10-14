# Light Luna

**Orbit. Observe. Defend.**  
**Intelligent Software Firewall & Autonomous Defense**  
*A Light Family product by THAI DUY Digital*  
**Status:** Architecture baseline plus an experimental **non-enforcing P0 policy/CLI scaffold**, 2026-10-13. **No working firewall, nft apply, service, cloud, live ban or standalone acceptance is claimed.**

Light Luna is a packaged **software firewall** for Linux servers and future compatible security appliances. It combines traditional, predictable firewall controls with **Luna Brain**, a bounded, evidence-driven security decision system. It works without Cloud, NetBird or Luna Brain. **Development is strictly standalone-first**: implement, package and prove the independent firewall before initiating ANY NetBird connector development. NetBird is a later optional integration, not an early development track.

## Product principles

1. **Firewall first:** deterministic rules, aliases, stateful filtering, zones, interfaces, NAT where applicable, schedules, audit and recoverable configuration; AI is optional.
2. **Install anywhere, select deployment mode:** **Host** (local protection), **Edge** (log/telemetry controller with remote agents, not necessarily inline) and **Gateway** (routing/forwarding appliance, optional for compatible hardware). Mode is a configuration, not a separate codebase.
3. **Defense in depth:** L3/L4 firewall + layer-7 WAF/reverse-proxy adapters + optional IDS/IPS + Luna Brain.
4. **Autonomy with bounded authority:** detect → investigate → decide → independent policy gate → execute → verify → learn; every execution has provenance, expiry, replay protection and rollback.
5. **Offline-first safety:** no dependency on an external API or LLM to permit routine traffic; existing policies and ban expirations continue if Cloud fails.
6. **Strict development order:** standalone Host/Edge firewall + WebUI/CLI + installer + tests first; demonstrate operational standalone acceptance before starting NetBird integration. NetBird API research is parked reference material only until that gate.
7. **Non-invasive coexistence:** never flush other components' firewall tables, interrupt Docker/NetBird/UFW/Cloud security groups, or disable NetBird ACL enforcement as a workaround.
8. **Privacy and provenance:** source IP, proxy identity, version-to-CVE applicability, learning labels and telemetry gaps must be explicit; no default sharing of raw logs or customer IP histories to Cloud.
9. **Releasable product:** two-architecture signed packages, clean install/upgrade/uninstall, service watchdog, safe migrations, rollback, supply-chain attestations and compatibility matrix.

## Naming

**Product:** Light Luna (two words). **Brain:** Luna Brain. **Cloud:** Light Luna Cloud. **CLI:** `lunactl`. **Privileged executor:** `lunalockd` (working name). **System agent:** `luna-agent` (working name). **Appliance:** Light Luna Box. **Firmware/project image:** Light Luna OS (future, only after hardware feasibility).

## Read this first

**Documentation starts at [Light Luna Project Book](docs/project-book/README.md).** New engineers and AI agents must read its [Library Rules](docs/project-book/LIBRARY_RULES.md), [Project Book](docs/project-book/LIGHT_LUNA_PROJECT_BOOK.md), and [Code Map](docs/project-book/LIGHT_LUNA_CODE_MAP.md) before assuming a planned feature is implemented or changing docs. Source files and Git/test evidence remain the final authority.

| Document | Purpose |
|---|---|
| [Product vision](docs/00-VISION-BRAND.md) | Name, lunar metaphor, personas, market/product boundary |
| [System architecture](docs/01-ARCHITECTURE.md) | Modes, planes, interfaces, authority and data flow |
| [Firewall specification](docs/02-FIREWALL-ENGINE.md) | Rules, aliases, NAT, compiler, transactional enforcement |
| [Luna Brain](docs/03-LUNA-BRAIN.md) | Observation, decisions, learning, confidence and trust |
| [WAF, IDS & CVE](docs/04-WAF-THREAT-INTEL.md) | L7, traffic inspection, threat feeds, software provenance |
| [NetBird integration — DEFERRED](docs/05-NETBIRD.md) | Reference research only. NO implementation before standalone acceptance gate |
| [Cloud & Fleet](docs/06-CLOUD-FLEET.md) | Remote controllers, security, lifecycle and privacy |
| [Security & safety](docs/07-SECURITY-SAFETY.md) | Threat model, guardrails, lockout rescue, abuse defense |
| [Packaging & hardware](docs/08-PACKAGING-HARDWARE.md) | Linux deployment, future appliance, upgrade/recovery |
| [WebUI, API & CLI](docs/09-WEBUI-CLI-API.md) | Operator UX, RBAC, control API and execution tool |
| [Moon-state design](docs/10-MOON-DESIGN.md) | Brand and incident severity language |
| [Roadmap & test gates](docs/11-ROADMAP.md) | Milestones, acceptance criteria and anti-scope-creep |
| [Architectural decisions](docs/12-ADRS.md) | Decided / proposed / unresolved |
| [Research sources](docs/13-REFERENCES.md) | Public official documentation links and caveats |
| [Project handoff](docs/14-HANDOFF-2026-10-13.md) | Historical design-era handoff |
| [P0 implementation notes](docs/15-P0-IMPLEMENTATION-NOTES.md) | Actual implemented slice, commands, limitations and next tests |
| [HOMESV P0 handoff](docs/16-P0-HOMESV-HANDOFF-2026-10-13.md) | Dev environment, verified tests, safe next tasks and Git publishing |

## Implementation language direction

**Go** is the planned primary backend, CLI, deterministic Luna Brain and separated constrained executor language. **React + TypeScript** is the proposed WebUI. An **optional, unprivileged Python worker** may be added only for justified ML/statistics/inference tasks; baseline Host/Edge operation and safety never depend on Python or any LLM. JavaScript `.mjs` is no longer the planned default Brain runtime. See [ADR-020](docs/adr/ADR-020-go-core-python-optional.md). **Only the P0 Go CLI/validator is implemented today.**

## Offline-complete installation and independent releases

The planned product installer contains the local application, independently protected firewall executor, **Basic Go Luna Brain** and verified **base knowledge** from day one. It runs without Cloud enrollment, NetBird or an Internet connection; advanced Python ML and extra WAF/XDP modules remain optional. **App, Brain, executor, module, knowledge and optional model versions update independently** with signed compatibility checks and rollback, even if two components sometimes require a coordinated transaction. See [ADR-024](docs/adr/ADR-024-offline-base-independent-component-updates.md) and [lifecycle design](docs/architecture/offline-base-component-update-lifecycle.md). The repository currently contains only a non-enforcing P0 Go CLI; this is a **release architecture requirement**, not a ready installer.

## Proposed Brain runtime/memory and Cloud extension

[ADR-023](docs/adr/ADR-023-dual-brain-memory-cloud.md) evaluates Go deterministic analysis plus optional hypothesis/ML reasoning, independently checked by an outside Policy Gate. [Local learning budgets](docs/architecture/luna-memory-lifecycle.md) propose bounded telemetry and knowledge storage. [Cloud intelligence](docs/architecture/luna-cloud-intelligence.md) describes a **future** curator and signed, per-asset delta downloads; aggregate experience uplink is opt-in. Light Luna must always preserve offline core firewall authority. None of this is implemented or approved as release-capable functionality.

## One account for local and Cloud security operations (future)

Enrolled Light Luna devices will offer local WebUI login using the **same Light Luna Cloud account identity**, plus Cloud Console visibility into last-reported firewall, Brain and module status. Remote configuration is a permissioned **proposal**, delivered through an outbound device connection and committed only by local Policy Gate with scope, revision match, rollback and audit. Offline installation still works using a local bootstrap owner, and emergency local recovery remains independent of Cloud. Read [ADR-026](docs/adr/ADR-026-unified-account-remote-fleet-control.md) and [unified account/fleet architecture](docs/architecture/unified-account-cloud-fleet-control.md). No Cloud Console, enrollment or local login is implemented yet.

## Cloud portability and seamless relocation (future)

Light Luna Cloud must be architected so the same client protocol, device identities and independently signed App/Brain/Knowledge releases survive changes between VPS providers, CPU architecture and future k3s. Begin with modular Go API/workers, a single authoritative PostgreSQL writer, provider-neutral immutable object storage and offsite restore-tested backups; scale stateless replicas rather than coupling state to container disks. See [ADR-025](docs/adr/ADR-025-portable-scalable-cloud.md), [Cloud architecture](docs/architecture/cloud-portability-and-scale.md) and [migration runbook](docs/architecture/cloud-migration-and-recovery.md). **This portability requirement is decided; Cloud services are not implemented.**

## Real no-NIC KVM reboot-interruption detection (2026-10-14 — lab PASS only)

[Forced reboot and inspector evidence](docs/architecture/p0-vm-autonomous-ttl-watchdog.md) now demonstrate a **real restart** of the dedicated no-NIC `light-luna-p0-nftvm` while its **synthetic guest namespace** contained an ICMP-drop rule and journal `COMMITTED_SYNTHETIC`. A read-only inspector detected that boot ID changed and observed no remaining test namespace/root nft state; critically it **refused to turn the stale journal into a false `ROLLED_BACK`**. A distinct operator-acknowledged step archived unreconciled evidence before the existing KVM regression was rerun successfully. This does not establish product automatic boot recovery or a power-loss guarantee, and never applied actual Light Luna-generated nft rules. P0 OPEN; G-STANDALONE BLOCKED.

## P0 autonomous TTL rollback in isolated KVM VM (2026-10-14 — LAB PASS x3)

[VM-only independent watchdog acceptance](docs/architecture/p0-vm-autonomous-ttl-watchdog.md) now proves a **separate systemd service with monotonic expiry** can automatically remove a fixed synthetic nft table after TTL without a second controller call. Systemd-managed [namespace keeper](scripts/lab/vm/guest-keeper-smoke.sh) survives the original `lxc exec` anchor, and the watchdog verifies keeper PID birth time and netns inode, writes an atomic READY before any test rule commit, then records journal `ROLLED_BACK` and DONE after the deadline. **Three consecutive full KVM suites PASS**, including TCP/UDP/IPv6 traffic restored. **Lab-only synthetic nft**, no Light Luna product-generated hooked policy or production executor; VM reboot, keeper failure rescue and owner consent still untested.

## P0 typed VM staging and explicit separate-service rollback (2026-10-14)

[Stage/revert verification](docs/architecture/p0-vm-synthetic-stage-revert.md) adds source-level `vmstage.Build` and `lunactl vm stage-intent <policy> <TTL>`, producing an **INERT** policy-bound JSON manifest. The real, no-NIC KVM lab now demonstrates exact `STAGED → VALIDATED → COMMITTED_SYNTHETIC → ROLLED_BACK` journal order, denied invalid TTL/hash/out-of-order/repeat attempts, IPv4/IPv6 + TCP/UDP packet drops with counters and restored connectivity after a **separately triggered transient systemd service**. **Combined VM suite PASS twice; automatic expiry timer FAILS intermittently and is NOT accepted.** All nft writes use a fixed synthetic VM-only table in a private guest netns, never the product's unhooked policy compiler. Interrupted-stage journals must be explicitly reconciled after proving no active rules/namespaces; no product executor or owner authorization exists.

## Independent-kernel LXD VM nftables/rollback lab — PASSED (2026-10-14)

[VM kernel/nft evidence](docs/architecture/p0-lxd-kernel-nft-rollback.md) and [runnable VM scripts](scripts/lab/vm/README.md) document a real **KVM VM** on HOMESV with separate guest kernel `6.8.0-142-generic` (host `6.8.0-146-generic`), no NIC, 2 vCPU and 2048MiB RAM. **Executed twice**: `lunactl firewall policies nft-preview` generated an unhooked script accepted by guest `nft -c`; malformed nft syntax rejected; real ICMP passed, a **separate synthetic VM-only netns rule** dropped it with counter evidence, then a separate watchdog process removed the synthetic table and restored ICMP. No Light Luna-generated rule was applied. **No privileged production executor or actual owner rescue was built**; host firewall remained outside the test.

## Offline LXD identity lab — running, cross-UID tests PASS

**Updated 2026-10-14:** HOMESV now has LXD 5.21.8 and a running unprivileged Ubuntu 24.04 `light-luna-p0-lab` container (2 CPU, 1536 MiB, no NIC). The `lxdbr0` bridge created automatically by initial LXD setup was removed immediately; current default profile and dedicated lab profile contain **no NIC**. The [distinct-UID harness](scripts/lab/lxd/README.md) **ran successfully**: authorized Agent signed preview; denied signed APPLY; denied ungrouped outsider and wrong UID even with copied synthetic key; denied replay after Gate restart and corrupt journal. Marker `LXD_DISTINCT_UID_POSITIVE_NEGATIVE_RESTART_PASS`. Earlier preparation-only notes in docs are historical. **LXD shares the host kernel**, so nftables/eBPF tests still require separate VM.



The [P0 no-NIC LXD lab plan](docs/architecture/p0-lxd-cross-uid-lab.md) and [reproducible lab scripts](scripts/lab/lxd/README.md) now cover separate Gate/Agent Linux UIDs, IPC group ACLs, an outsider and a wrong-UID insider with a deliberately copied **synthetic lab key**, plus replay after Gate restart. **Historical initial checkpoint superseded:** before the approved retry, the snap was absent and no lab was running; see the verified deployed results immediately above. The planned container has no NIC and shares the host kernel; do **not** use it for nftables/eBPF/host firewall tests. A separate VM is required for kernel safety work.

## P0 Ed25519 Agent/Gate identity separation — development contract

[Asymmetric identity boundary](docs/architecture/p0-ed25519-agent-gate-identity.md) now signs Agent requests with **Ed25519 private key** while Gate verifies using a pinned **public key only**, with strict rejection of HMAC downgrade/mixed signatures and support for explicit Linux peer UID/GID checks over group-mode `gate.sock`. The test `bash scripts/test_p0_ed25519_identities.sh` verifies two processes, wrong-Gate-UID refusal, replay after restart and corrupted-journal fail-closed. **HOMESV test still uses one Unix UID and temporary files**, not separate production service accounts or real owner authentication; neither CLI can apply firewall. Real cross-user tests must run in disposable VM.

## P0 separate Agent/Gate processes with restart replay protection (developer-only)

[Two-process development slice](docs/architecture/p0-two-process-gate-and-journal.md) adds separate `luna-gate` and `luna-agent` Linux binaries communicating over real local Unix sockets, with a **test-only same-UID shared HMAC key** and a bounded `fsync`-before-reply replay journal that rejects old request IDs/nonces **after Gate restarts**. `bash scripts/test_p0_gate_processes.sh` creates/tears down isolated nonroot temporary sessions, validates preview, denies signed apply, checks replay across restart and verifies corrupt journal **fails closed**. **Not production identity, not separately sandboxed service users, not power-loss-certified transactional storage, no firewall executor.** The prior `lunactl` IPC selftest still exists.

## P0 Linux Unix socket IPC — bounded developer prototype

[Linux local IPC](docs/architecture/p0-linux-unix-ipc.md) now has a **real Unix socket and SO_PEERCRED UID checks**, 0700 private socket directory / 0600 socket permissions, bounded 16KiB messages, 4KiB receipts, 3-second deadlines and per-process concurrency limits. `lunactl control ipc-selftest testdata/host-baseline.json` exercises the socket transport with an ephemeral HMAC key; unit tests additionally verify communication across **two different Linux processes**. **No privileged executor, durable replay, production user isolation or daemon exists**; every `APPLY_POLICY` stays denied and kernel writes remain impossible. NetBird/Cloud remain deferred.

## P0 authenticated local-control gate (in-process demo, no privileges)

The [local control gate contract](docs/architecture/p0-local-control-gate.md) now has actual Go source for **HMAC-authenticated, time-scoped, revision/hash-pinned** local intents, nonce/request-ID replay rejection and a fail-closed `APPLY_POLICY` decision. Run `go run ./cmd/lunactl control selftest testdata/host-baseline.json` for an ephemeral in-process preview/replay/tamper/apply denial demo. This is **not** a socket listener, durable session store, real local authentication, policy executor, rollback or kernel apply. HOMESV can access GitHub directly via its own deploy key after validating GitHub's host fingerprint; no ARM relay is required for source synchronization.

## P0 synthetic packet-flow simulator (source implemented, no kernel effects)

`lunactl firewall policies simulate testdata/host-baseline.json testdata/flow-fixtures.synthetic.json` now calls [`internal/flowsim`](internal/flowsim/simulate.go) to evaluate a bounded JSON fixture against ordered policy rules, label **potential management lockout**, handle IPv4/IPv6 aliases and record nonterminal `log` matches. It deliberately reports `effectiveVerdict=UNKNOWN` and `externalFirewall=UNKNOWN` in every result; **no kernel state, actual packet flow, Docker/UFW chain or conntrack is inspected**. See [verified P0 flow simulation contract](docs/architecture/p0-flow-simulator.md). No real firewall was changed.

## First nftables text compiler — safe non-enforcing P0 preview

A real Go [nft preview compiler](internal/nftpreview/compile.go) now converts validated Host/Edge policy rules into **unhooked nftables regular-chain text** via `lunactl firewall policies nft-preview testdata/host-baseline.json`. The result is JSON with deterministic hash, rule expansion counts, `kernelWrites=false`, `syntaxChecked=false` and explicit unsupported semantics. It never configures a base-chain hook, default DROP policy or emergency access route, and **must not be applied**. See [implemented contract and limits](docs/architecture/p0-nft-unhooked-compiler.md). HOMESV cannot create an isolated user/network namespace, so kernel nft syntax checking remains unverified; no live firewall was touched.

## First runnable Go foundation (P0, simulation only)

The new [standalone runtime foundation](docs/architecture/local-runtime-foundation.md) has two separately buildable binaries: `lunactl` (policy validator, local snapshot, offline replay) and `luna-brain` (synthetic Brain replay). Real Go packages now define bounded [observation contracts](internal/foundation/observation.go), a deterministic [demo Brain](internal/brain/basic.go), honest [Guardian UNKNOWN status](internal/guardian/basic.go), a non-executing [Gate preview](internal/gate/advisory.go), and [strict JSON](internal/strictjson/decode.go) rejecting duplicate nested keys. Use `go run ./cmd/lunactl system snapshot` and `go run ./cmd/luna-brain replay testdata/offline-brain-replay.synthetic.json` to verify **without network or kernel changes**. This is **not** a daemon, real malware scanner, nftables executor, signed knowledge installer or production detector.

## Host self-defense and protection against compromised clients (future)

[ADR-028](docs/adr/ADR-028-self-defense-host-compromise.md) requires a separate Host Guardian/self-integrity plane: detect tampered Light Luna binaries/config/updater, suspicious Linux process/file/network behavior and possible malicious egress/lateral spreading. Optional eBPF tracing and bounded file scanners provide evidence; **XDP alone cannot scan malware**, and root/kernel compromise cannot be reliably self-certified clean. [Containment design](docs/architecture/host-self-defense-and-compromise-containment.md) requires local policy/rollback and Cloud untrusted-intake quarantine so one compromised node cannot publish global signatures, poison collective knowledge, upload malware payloads by default or control other nodes. This is design only, not current antivirus/EDR functionality.

## Collective defensive immunity (future architecture)

[ADR-027](docs/adr/ADR-027-collective-threat-immunity.md) defines CrowdSec-inspired **behavior/scenario-centered intelligence sharing**, not just IP blocklists. A consented local detection and mitigation produces an evidence-linked signal; Cloud deduplicates, validates, correlates and publishes scoped signed/retractable watch or curated threat patterns for other clients. Their **local Basic Brain recognizes behavior earlier**, while local Policy Gate independently controls any ban. A client blocking an IP does **not** make that IP globally guilty. See [Federated Threat Intelligence](docs/architecture/federated-immune-threat-intelligence.md). This is not yet implemented.

## Brain V2: a knowledge-driven defensive system

Light Luna's differentiator is a governed security reasoning plane: **local observations + trusted public CVE/KEV/EPSS/vendor feeds + ATT&CK attack-technique models + exact installed-asset applicability + observed temporal behavior → evidence-bound incident/pre-attack recommendations**. Its collective Cloud knowledge is a future privacy-controlled, opt-in, signed/versioned subscription rather than a global ban authority. The existing Python Sentinel Luna is studied as a source of curated patterns and false-positive safeguards, **not** an existing pretrained neural model or production Light Luna Brain. The new Go-native Brain V2 is currently **architecture only**: [ADR-022](docs/adr/ADR-022-rebuild-luna-brain-governed-knowledge-migration.md), [Brain V2 design](docs/architecture/luna-brain-v2.md), [Threat Pipeline](docs/architecture/threat-intelligence-knowledge-pipeline.md), [source audit](docs/history/audits/SENTINEL_LUNA_COGNITION_AUDIT_2026-10-13.md).

## Optional eBPF/XDP roadmap

nftables remains the primary firewall. Linux **eBPF host/network/socket/process observability** is an optional future sensor, and **XDP early-ingress L3/L4 filtering** is an optional, explicitly enabled future anti-DDoS fast path. Neither is implemented; no XDP/BPF attachment or packet drop is authorized by this plan. Preserve all existing XDP program ownership and support nftables-only on unsupported NICs. See [ADR-021](docs/adr/ADR-021-ebpf-observability-xdp-optional.md).

## Mandatory stage gate — standalone first

**Stage A (ACTIVE PLAN):** Build and prove Light Luna as a self-contained firewall. Host and Edge must work without NetBird installed, without NetBird APIs, and without any NetBird dependency. Complete local rules, aliases, policy compiler, agent/executor, WebUI/CLI, packaging, recovery and standalone end-to-end tests first.

**Stage B (LATER, BLOCKED):** Only after a documented standalone working build passes the exit criteria in [`docs/11-ROADMAP.md`](docs/11-ROADMAP.md) **and the owner authorizes the next phase**, begin engineering and testing the optional NetBird connector. Until then: **no NetBird integration code, API probing, access tokens, compatibility tests or modification to the existing NetBird stack**. Existing public documentation research may remain in this repository as future design reference.

## Initial topology

```text
          Future optional Cloud Intelligence / Fleet Controller
                             |
                     mTLS / signed updates
                             |
WebUI / lunactl ---> Local Control API ---> Policy & Config Store
                             |                    |
              Sensors <--- Event Normalizer ---> Luna Brain
                             |                    |
                             +---> Policy Gate <---+
                                     |
                             Privileged Executor
                              /      |        \
                         nftables   WAF      IDS/Proxy adapters
                                     |
           Local Host / Edge Agent / Routed Gateway (chosen at install)

           [FUTURE / NOT PART OF STANDALONE BASELINE]
           Optional NetBird Connector: peer/group/policy inventory,
             secure management path; NEVER owner of Luna policy.
```

## Current boundaries

- This repository now includes an **experimental, non-enforcing policy validator and CLI preview only**. No nft compiler/apply, WAF engine, privileged daemon, installer, infrastructure, logo asset or working firewall exists.
- Existing Luna/Selene Sentinel work is **research input**; do not copy live logs, agent memory databases, credentials, user data or production automation into this repository.
- Existing production Light Remote, NetBird, VPS-ARM, HOMESV and VPS-AMD are **out of scope for rule changes** until a lab and owner-approved rollout plan are ready. NetBird integration development itself is **out of scope until standalone gate acceptance**.
- Licensing, product name/trademark/domain availability, commercial Cloud terms, and vendor integration obligations remain subject to review.
