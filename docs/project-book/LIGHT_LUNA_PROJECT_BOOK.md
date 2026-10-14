# Light Luna Project Book

Status: current project synthesis (not a substitute for code)
Date: 2026-10-13
Scope: product direction, architecture boundaries, documented P0 reality, development order, evidence and open gates
Historical baseline: public `n8n2erpnext/light-luna` `main@66920c251603b0b0c4bcd0d0cfc94ca8f7220d41`; initial development foundation `feat/p0-standalone-foundation@9ae27b267e72e465ea7808c3936fabcf6b0cac05`. Latest locally verified Core audit extension `834100a6671c45404d6a52f5a0ef26441d2711f7` (HOMESV; Git remote must be checked independently).

## Core P0 detached owner-review proof, not an authorization (2026-10-14)

[Implemented offline verifier](../../internal/controlgate/owner_review.go), [negative tests](../../internal/controlgate/owner_review_test.go), and [owning architecture](../architecture/p0-offline-owner-review-separation.md): P0 now models two **cryptographically different test keys** for Agent vs reviewer, domain-separated signature evidence bound to the exact Agent intent including request nonce, policy hash/revision, operation and TTL. In-memory review-ID and Agent-intent replay checks have a bounded fail-closed quota. The result is **`CRYPTOGRAPHIC_TEST_MATCH_ONLY`**, **not owner login, human consent, Gate permission or proof of durable approval**. The actual Gate never consumes the evidence and still denies signed APPLY with zero kernel writes. This is Core-only offline source; no runtime provider, monitoring or NetBird change.

## Core P0 replay journal new-file fsync hardening (2026-10-14)

[Canonical implementation/evidence](../architecture/p0-two-process-gate-and-journal.md): development [replay journal](../../internal/controlgate/journal_linux.go) now distinguishes exclusive **new journal** creation from existing-file reopen and performs **file fsync → parent directory fsync** for a newly created journal before accepting requests. Any open/sync failure is fail-closed. [Unit tests](../../internal/controlgate/journal_linux_test.go) and real nonroot development [HMAC](../../scripts/test_p0_gate_processes.sh)/[Ed25519](../../scripts/test_p0_ed25519_identities.sh) roundtrips PASS after the change; Go test ×3, Go vet and 91 Python VM unit regressions PASS. **Do not conflate this with true durable kernel-policy transactions, power-loss recovery, product Gate installation or verified human owner consent.** APPLY remains denied, external integrations frozen. P0 OPEN; G-STANDALONE BLOCKED.

## Core P0 CIDR-set union analysis refinement (2026-10-14)

[Canonical policy audit](../architecture/p0-policy-rule-conflict-audit.md): the existing offline `AnalyzeConflicts` now detects full later-rule coverage by **multiple CIDR members of one earlier alias**, not merely an individual ancestor CIDR. [`conflict_union.go`](../../internal/policy/conflict_union.go) handles bounded sorted IPv4/IPv6 address intervals; [property and negative tests](../../internal/policy/conflict_union_test.go) show contiguous halves shadow, gap means partial overlap, and mixed IP families never falsely combine. No live Gate/Executor, kernel firewall or owner approval was changed; Prometheus/Grafana/notification integration remains frozen. P0 OPEN, G-STANDALONE BLOCKED.

## Core P0 policy family and rule precedence implementation (2026-10-14)

[Core contract](../architecture/p0-policy-rule-conflict-audit.md): `policy.Validate` now rejects `icmp`/IPv6 and `icmpv6`/IPv4 source mismatch for literals **and every CIDR-set member**. New [static analyzer](../../internal/policy/conflicts.go) identifies a later rule fully shadowed by an earlier terminal action or a partial opposing overlap under declared priority/port/source semantics, with deterministic bounded JSON and explicit incomplete/budget status. [CLI](../../cmd/lunactl/main.go) supports `firewall policies analyze <file>`; a valid intentional narrow allow before a broad drop is not automatically rejected. Verified by [Go tests](../../internal/policy/conflicts_test.go), policy ICMP family negative tests and inert sample CLI. **No kernel writes, no owner/executor authority, no real firewall rule, no packaging or G-STANDALONE acceptance.** This is the active Core workstream; observability and notifications stay frozen.

## Owner priority correction: CORE FIRST, observability integrations parked (2026-10-14)

**CURRENT DIRECTIVE:** Prometheus/Grafana and Telegram/Discord/Slack/email are requirements recorded for future architecture, **not tasks to implement now**. Existing `internal/observability/synthetic.go` and `internal/observability/events.go` are inert source-only prototypes that were built prematurely; **freeze all further development** there, do not connect/deploy them and do not treat them as P0 core progress. Resume development only when standalone core safety/functionality and delivery gates are demonstrated and the owner explicitly changes priority. **Current active scope:** typed policy/validation and conflict handling → local signed owner Gate + least-privilege executor → fail-safe transactional rollback, watchdog and crash recovery → Host/Edge standalone with offline local UI/CLI → isolated tests, signed packaging and G-STANDALONE evidence. No unrelated provider adapters, dashboards, metrics HTTP listeners, Cloud or NetBird implementation in this workstream. The prior source prototypes remain in Git history/current branch unchanged for now; this clarification does not request their deletion.

## P0 security event + notification-plan Go contract — **source-only** (2026-10-14)

Following [ADR-029](../adr/ADR-029-local-observability-and-notifications.md), [`internal/observability/events.go`](../../internal/observability/events.go) now defines **inert, strict** bounded synthetic security events and a data-only deterministic notification plan. [Schema](../../schemas/local-security-events-synthetic-v1alpha1.schema.json), [fixture](../../testdata/local-security-events-synthetic.sample.json), and [Go negative tests](../../internal/observability/events_test.go) guard against raw subject/user/credential fields, forged approved/applied/sent state, arbitrary label growth, duplicate IDs, oversized input and unlimited fan-out. Four enum channels (Telegram, Discord, Slack, email) generate `SIMULATED_NOT_SENT` plan entries only; simulated queue counts feed the P0-only Prometheus renderer, never production ban/notification metrics. **No HTTP endpoint, real ban, persistent outbox, notifier, SMTP/Telegram/Discord/Slack integration, Cloud or NetBird.** This does not advance the P0 release gate.

## P0 observability: Go synthetic-only contract, no exporter or notification transports (2026-10-14)

Following [ADR-029](../adr/ADR-029-local-observability-and-notifications.md), [`internal/observability`](../../internal/observability/synthetic.go) provides a **pure Go, P0 non-enforcing** strict synthetic snapshot parser and deterministic Prometheus text renderer, plus [JSON schema](../../schemas/local-observability-synthetic-v1alpha1.schema.json), [synthetic fixture](../../testdata/local-observability-synthetic.sample.json) and [negative unit tests](../../internal/observability/synthetic_test.go). Only bounded IPv4/IPv6 `proposed|unknown|rejected` counts and simulated Telegram/Discord/Slack/email queue counts are accepted; **actual bans, real users/IPs, webhook URLs, owner commands and dynamic label values are rejected**. Synthetic metric family names are deliberately different from future production ban metrics. No listener, Grafana dashboard, queue worker, secret storage, outbound HTTP/SMTP or kernel call is implemented. P0 and G-STANDALONE remain open/blocked.

## Owner-decided Prometheus/Grafana ban observability and alerts (2026-10-14; design only)

[ADR-029](../adr/ADR-029-local-observability-and-notifications.md) confirms future **Prometheus-compatible low-cardinality `/metrics`**, Grafana views for **verified applied vs pending/unknown/expired bans**, separately authenticated **read-only** local ban/event APIs and asynchronous opt-in notifications via **Telegram, Discord, Slack and SMTP mail**. The [architecture](../architecture/local-observability-and-notifications.md) fixes the separation of actual kernel/Policy Gate receipts from mere Brain proposals and requires secrets/redaction, bounded queues, retry/dedupe, no external dependency of standalone security enforcement and no NetBird work before G-STANDALONE. These are **future P1/P2 product features, not existing P0 HTTP services, ban exporters, dashboards, webhook/SMTP sends or live kernel rules**. Work on P0 crash/durable receipt safety remains higher priority and proceeds independently.

## P0 FD-held namespace detection before reporting ban/rollback state (2026-10-14)

[Canonical KVM proof](../architecture/p0-vm-autonomous-ttl-watchdog.md): read-only kernel inspection now scans bounded process FDs for hidden **nsfs** references, not only visible process network namespaces or `ip netns list`. In actual no-NIC KVM, a foreign namespace kept open by an FD but deleted from its named mount list was detected; the rollback reviewer **refused a clean result even with valid DONE**, returning `KERNEL_STATE_PRESENT_OR_UNKNOWN`. Closing that FD left the guest clean and journal SHA-256 unchanged. [Five added offline negatives](../../scripts/lab/vm/test_boot_inspect.py) / [rollback review cases](../../scripts/lab/vm/test_rollback_review.py) bring the combined VM suite to **91 Python tests PASS**. This is bounded guest-only visibility improvement, **not proof against malicious root or every kernel object**. P0 OPEN, G-STANDALONE BLOCKED.

## P0 synthetic journal boot identity boundary (2026-10-14)

The [VM watchdog architecture/verification record](../architecture/p0-vm-autonomous-ttl-watchdog.md) has a new source-level kernel `boot_id` guard in the synthetic staging journal and watchdog READY/DONE receipts. A subsequent transition or rollback using a different/unknown boot ID is denied rather than trusting recycled process identities. [Five isolated offline tests](../../scripts/lab/vm/test_boot_journal.py) PASS; the no-NIC KVM suite also returned exit 0 with the previous packet and TTL markers. **No VM reboot with an active policy, power-loss recovery, production transaction keeper or automatic boot reconciliation is proven.** Development continues at P0, standalone gate blocked.

## P0 bounded receipt fsync fault matrix — isolated guest filesystem (2026-10-14)

[Canonical results and limits](../architecture/p0-vm-autonomous-ttl-watchdog.md): [10 new fault tests](../../scripts/lab/vm/test_rollback_fsync_faults.py) inject simulated `EIO` at the file and directory fsync boundaries for `prepared`, `kernel-absent` and `journal-terminal`; `ENOSPC` before creating an observation; malformed orphan file and epoch no-overwrite checks. Offline PASS and **same 10/10 PASS in the real no-NIC KVM guest**, but only in temporary guest directories with a fake kernel-write decision, **no nft operations**. The guarded VM runner validates active journal hash unchanged and empty guest nft/named netns. A readable receipt that failed fsync remains **durability-ambiguous**: do not auto-recover or create DONE from it. No real power-loss/disk fault or production signed executor. P0 OPEN, G-STANDALONE BLOCKED.

## P0 read-only rollback receipt and journal reconciliation classification — VM lab (2026-10-14)

[Canonical bounded evidence](../architecture/p0-vm-autonomous-ttl-watchdog.md): new [independent reviewer](../../scripts/lab/vm/guest-rollback-review.py) checks guest-only journal/READY/immutable receipt hash-chain/DONE against boot ID and **read-only** nft/netns inventory. It distinguishes PREPARED kernel-unknown, KERNEL_ABSENT journal-incomplete, TERMINAL-without-DONE, invalid/mismatched receipt, cross-boot ambiguity and coherent synthetic DONE **without ever authorizing rollback or real firewall mutation**. [23 classifier tests](../../scripts/lab/vm/test_rollback_review.py) PASS; **76 total offline Python tests** PASS. Real no-NIC KVM read-only verification: live terminal-without-DONE exit 3, archived terminal-without-DONE exit 3, and after a full ordinary synthetic TTL regression PASS, coherent live primary observation exit 0 **with `successAcknowledged=false`**. This is not a production boot reconciler or signed owner receipt. P0 OPEN; G-STANDALONE BLOCKED.

## P0 immutable rollback observation receipts — KVM synthetic verification (2026-10-14)

[Verified architecture and fault result](../architecture/p0-vm-autonomous-ttl-watchdog.md): primary and fallback now create non-overwriting, epoch-bound **prepared → kernel-absent → journal-terminal** receipt files under a shared flock before their DONE. [Twelve offline tests](../../scripts/lab/vm/test_rollback_receipts.py) cover immutable ordering/hash/identity failures. A real no-NIC KVM fault after journal `ROLLED_BACK` but before fallback DONE **PASS for detection**: all three observation files survived, DONE was absent, the observer explicitly reported `TERMINAL_JOURNAL_DONE_MISSING_REVIEW_REQUIRED` and a separately acknowledged archive preserved the evidence. Early harness mode/timing failures were recorded rather than discarded. **No production signed intent, atomic journal+kernel commit, reboot recovery, or owner-authenticated executor has been built.** P0 OPEN; G-STANDALONE BLOCKED.

## P0 cooperative single-writer rollback fencing — KVM synthetic only (2026-10-14)

[Canonical implementation and verification](../architecture/p0-vm-autonomous-ttl-watchdog.md): primary and fallback now share a bounded VM-local `fcntl.flock`, covering **read latest journal → inspect fixed nft table → remove test rule → fsync journal → fsync DONE**. Inside the lock they revalidate stage/boot/history; loser must re-evaluate, not trust stale pre-lock state. Real keeper/watchdog SIGKILL rollback cases PASS after this change; a same-deadline primary/fallback run wrote one primary DONE and fallback REJECTED; process crash after test nft deletion still preserved `COMMITTED_SYNTHETIC` with no false success. [Ten offline fencing cases](../../scripts/lab/vm/test_rollback_fence.py) cover independent processes and malformed identities. **Advisory lab mutex only**, not a production transaction manager, owner authentication, durability protocol or global kernel write protection. P0 OPEN; G-STANDALONE BLOCKED.

## P0 dual-armed takeover and nft-delete/journal crash boundary — KVM lab PASS for detection (2026-10-14)

[VM-only verification record](../architecture/p0-vm-autonomous-ttl-watchdog.md): with both watchdogs prearmed and primary healthy, the primary rollback completed and fallback recorded `PRIMARY_ALREADY_TERMINAL_NO_TAKEOVER`, without duplicate DONE. A separate injected `os._exit(72)` after the fallback deleted the synthetic nft table but before it advanced the journal showed **kernel rule absent while `COMMITTED_SYNTHETIC` persisted**; the read-only observer produced `FAILURE_DETECTED_NOT_RECOVERED`, neither service emitted a success receipt, and separately acknowledged archival preserved the fault. This proves **observability and denial of false success**, not atomic crash-safe recovery, a generic concurrent-owner lock, or product enforcement. P0 OPEN; G-STANDALONE BLOCKED.

## P0 independent fallback prototype — synthetic KVM recovery PASS (2026-10-14)

[VM-only fallback architecture and evidence](../architecture/p0-vm-autonomous-ttl-watchdog.md): a second systemd-managed process **pins the guest synthetic network namespace independently** before rule commit. With exact boot/stage/keeper identity, monotonic deadline and nft JSON rule-shape checks, it rolled back the fixed private-netns table after real keeper **or** primary watchdog SIGKILL postcommit, restoring IPv4/IPv6 TCP/UDP before test EXIT cleanup. Both cases PASS in two executions, including a stricter table-shape guard; [12 parser tests](../../scripts/lab/vm/test_fallback_supervisor.py) verify rejection of altered rules. This is a **lab-only fallback proof**, still no production owner authentication, signed receipts, installed services, product policy apply or reboot/power-loss guarantee. P0 OPEN; G-STANDALONE BLOCKED.

## P0 service crash boundary — verified failures, no rescue (2026-10-14)

[Three isolated SIGKILL tests](../architecture/p0-vm-autonomous-ttl-watchdog.md) demonstrate the no-NIC KVM guest's independent namespace keeper or watchdog can die **before commit or after a verified synthetic commit**. Three failure-detection tests PASS: `keeper_precommit` left `VALIDATED` and no success receipt; both `keeper_postcommit` and `watchdog_postcommit` left the synthetic nft table active **after TTL** and the journal `COMMITTED_SYNTHETIC`, again without a false DONE. The guest EXIT trap explicitly removed the synthetic rule after observation, and a separately acknowledged archival tool preserved incomplete journal states and fault evidence **without forging `ROLLED_BACK`**. This proves a critical missing backup recovery path, **not automatic service recovery**. P0 OPEN; G-STANDALONE BLOCKED.

## P0 real KVM reboot fault injection — evidence only (2026-10-14)

[Read-only boot inspection, test runner and results](../architecture/p0-vm-autonomous-ttl-watchdog.md) now demonstrate a **real forced reboot** of HOMESV's exact no-NIC KVM laboratory while a **fixed, independently authored guest-netns synthetic rule** was active and the journal was `COMMITTED_SYNTHETIC`. After the VM boot ID changed, the inspector observed no remaining root nft/named/foreign-process netns state, but preserved the old journal and **rejected false ROLLED_BACK**. Separate explicit operator archival preserved old evidence without a rollback claim; the original KVM regression passed again afterward. [14 offline inspector tests](../../scripts/lab/vm/test_boot_inspect.py) and existing five boot-ID tests pass. This is **not** production automatic recovery, a real Light Luna policy apply, or a power-loss guarantee. P0 OPEN; G-STANDALONE BLOCKED.

## Latest verified P0 increment — independent auto-TTL rollback in KVM lab (2026-10-14)

[Autonomous VM watchdog contract](../architecture/p0-vm-autonomous-ttl-watchdog.md) is now implemented in the *synthetic-only* KVM laboratory: a PID1 service verifies exact keeper process birth/netns inode, fsyncs an arm-READY record before synthetic commit, expires on monotonic TTL, deletes its **fixed guest-netns** nft table, verifies removal and journals terminal `ROLLED_BACK`/DONE. A separate systemd namespace keeper survives termination of the controller-owned anchor. [Combined VM scripts](../../scripts/lab/vm/README.md) **PASS three consecutive post-fix full suites**, including restored TCP/UDP and IPv6 with no manual second rollback command. This is **not a Light Luna privileged executor, owner acceptance, kernel safety certification, host deployment or reboot-proof rescue**; P0 and G-STANDALONE still open.

## Latest verified P0 increment — VM staged synthetic transaction + explicit rollback (2026-10-14)

[Current stage verification](../architecture/p0-vm-synthetic-stage-revert.md) and [Code Map](LIGHT_LUNA_CODE_MAP.md): a typed non-executing `vmstage` JSON intent now binds a policy SHA/revision and synthetic VM table with TTL field. The VM-only journal validates stage ID/hash/TTL and rejects duplicate or out-of-order transitions. In a separate-KVM-kernel no-NIC guest, synthetic netns rule tests prove IPv4/IPv6, TCP/UDP packet drop with nft counter and **post-controller-death rollback by explicitly triggered independent systemd service**, restoring traffic. **Full combined VM suite PASS twice.** Automatic TTL firing remains unreliable and **NOT ACCEPTED**, so this is not a production independent rescue mechanism or G-STANDALONE. Interrupted states require explicit cleanup proof and guarded journal reconciliation; all prior failure evidence retained in the worklog. No actual Light Luna policy applied.

## Latest verified P0 milestone — separate kernel nft syntax and synthetic rollback PASS (2026-10-14)

The new [LXD KVM VM test](../architecture/p0-lxd-kernel-nft-rollback.md) uses a **different Linux kernel from HOMESV** (guest 6.8.0-142, host 6.8.0-146), 2 vCPU / 2048MiB, no NIC and ephemeral netns. **Twice successfully:** generated `nftpreview` unhooked text passed kernel `nft -c`; deliberately bad syntax failed; guest-only synthetic hooked nft rule blocked ICMP with counter evidence, then an independent watchdog deleted the synthetic table and packets resumed. **Light Luna-generated rules were not applied.** Still no real firewall executor, owner-rescue watchdog, crash/power-loss transactional rollback, real device enrollment or Cloud/NetBird integration. This passes the *VM nft syntax and synthetic packet rollback lab*, **not** P0 or G-STANDALONE.

## Latest verified P0 milestone — LXD different-UID IPC lab PASS (2026-10-14)

On HOMESV, Light Remote installed LXD snap 5.21.8, initialized storage and created a networkless Ubuntu 24.04 **unprivileged** container `light-luna-p0-lab` (2 vCPU, 1536MiB memory, no NIC). Minimal LXD setup unexpectedly created `lxdbr0`, which was removed after deleting its default-profile NIC **before** creating the lab container. Current `lxc network list` shows physical interfaces only. The lab pool uses a `dir` backend, so no per-instance disk quota is guaranteed.

Inside the container, **actual different Unix users** `lunagate` and `lunaagent` successfully completed Ed25519-signed advisory IPC. Other negative cases passed: denied signed APPLY, unexpected Gate UID, outsider without IPC group, **wrong UID even with copied synthetic test private key** (peer UID enforcement), replay after Gate restart and corrupt journal Gate-startup rejection. The exact test markers `LXD_DISTINCT_UID_POSITIVE_NEGATIVE_RESTART_PASS` and `LXD_CROSS_UID_TEST_RUN_COMPLETED` were observed. [Full environment and acceptance evidence](../architecture/p0-lxd-cross-uid-lab.md), [executable lab runbook](../../scripts/lab/lxd/README.md), [current Code Map](LIGHT_LUNA_CODE_MAP.md) and [Worklog](LIGHT_LUNA_WORKLOG.md).

**Scope:** This demonstrates user/group access separation and Gate's Ed25519/UID policy in one LXD container. LXD containers share the host kernel; it is **not** a valid isolated-kernel nftables/XDP proof, production owner login, full host compromise protection, or G-STANDALONE. The previous project-book snapshots below are historical and may describe earlier unbuilt milestones.

## Reading contract

Read [governance](LIBRARY_RULES.md), [library map](DOCUMENT_LIBRARY_MAP.md), [Code Map](LIGHT_LUNA_CODE_MAP.md), [truth status](PROJECT_TRUTH_STATUS.json), and exact source before changing implementation. Product documentation may describe intended capability, not shipped behavior.

## Identity and owner direction

Light Luna is an independent, installable Linux **software firewall first**, with a Basic Luna Brain required in the future full installer (ADR-024); enhanced ML, WAF/IDS, Cloud and later hardware modes remain optional. It should function without Cloud or NetBird. Owner confirmed **public GitHub repository is intentional** on 2026-10-13, superseding the earlier private-repository preference in historical handoffs.

## Non-negotiable stage order

[Roadmap](../11-ROADMAP.md) and [ADR register](../12-ADRS.md): P0–P4 independent Host/Edge, explicit gate **G-STANDALONE**, then P5 optional NetBird development **only after separately authorized by owner**. No early connector, read-only API probing, dependency, credentials or compatibility integration tests. Host/Edge packet visibility depends on topology; one-NIC passive Edge is not an automatic LAN packet sensor.

## Technology roles (implemented vs proposed)

- **Go — actual P0 implementation**: `lunactl` CLI, typed policy validation, inert preview. Target main backend, API and separately privileged executor. The executor **does not exist**.
- **React + TypeScript — WebUI proposal only**: browser dashboard, policy editor, observability; no frontend source or service exists in this branch.
- **Go-native Luna Brain — owner-approved architecture direction, NOT IMPLEMENTED**: deterministic analytics, evidence correlation and typed recommendations should use Go; even with Go shared across components, Brain cannot access the privileged executor. See [ADR-020](../adr/ADR-020-go-core-python-optional.md).
- **Python — optional worker only, NOT IMPLEMENTED**: selective, justified ML/stats/inference work using separate sandboxed unprivileged process, bounded CPU/memory/time and versioned inputs/outputs; no mandatory Python runtime for normal Host/Edge operation. `.mjs` is no longer the default Brain choice.
- **SQLite — local store proposal only**: no database schema or service yet.
- **nftables — future enforcement backend**: source-only **unhooked text preview compiler and synthetic policy-intent Flow Simulator exist**; no complete nftables apply compiler/transaction, watchdog, kernel change, service or real rules deployed.

Rust, a **mandatory Python runtime**, and a separate Node backend are not required by the P0 firewall. Python 3 currently runs only the docs catalog tool; a future optional ML worker is permitted but does not yet exist. React/TypeScript WebUI and SQLite remain proposals rather than delivered code.

## Verified engineering snapshot

[Code Map](LIGHT_LUNA_CODE_MAP.md) traces the exact Go CLI path. HOMESV is primary development workspace `/home/ubuntu/n8n2erpnext/light-luna`; Go 1.27.2 user-space toolchain used for tests. Development commit `9ae27b2` adds policy validator + preview + synthetic fixtures; recorded `go test ./...`, `go vet ./...`, CLI checks and amd64/arm64 cross-builds passed on HOMESV. **Those tests do not establish a working firewall or P0 acceptance.**

Previously the docs-only `main` commit `66920c2` was published using VPS-ARM deploy key. All production infrastructure/firewall remained untouched in those operations.

## Optional Linux eBPF/XDP direction (decision, not deployed)

[ADR-021](../adr/ADR-021-ebpf-observability-xdp-optional.md) adds **non-mandatory eBPF tracing** as a potential source of bounded host/network/socket/process telemetry and **optional XDP ingress DDoS filtering** after isolated tests and explicit enablement. nftables remains the primary firewall. No source object, attach, XDP_DROP, loader or benchmark exists in the current project. Do not confuse an existing kernel BPF attachment on HOMESV with Light Luna functionality. Kernel/driver support and coexistence must be checked before any isolated lab attachment, with no interruption of third-party XDP. Optional XDP is not a G-STANDALONE requirement; NetBird remains deferred.

## Brain V2 and collective security knowledge (agreed architecture, not built)

After read-only inspection of the independent VPS-ARM Sentinel services, the source-audited legacy “Luna” is **Python rules, hand-authored heuristic scores, SQLite observations, JSONL cognition templates and bounded learning curriculum**—not a demonstrated model with trained neural weights. [Evidence audit](../history/audits/SENTINEL_LUNA_COGNITION_AUDIT_2026-10-13.md) records the exact logic, confidence blend/decay and generated knowledge record counts. [ADR-022](../adr/ADR-022-rebuild-luna-brain-governed-knowledge-migration.md) chooses a new Go-native Luna Brain V2 and selective, reviewed knowledge migration rather than cloning the live system. [Brain V2](../architecture/luna-brain-v2.md) defines typed evidence/weights/scenario and action limits, while [Threat Knowledge Pipeline](../architecture/threat-intelligence-knowledge-pipeline.md) defines trusted CVE/KEV/EPSS/vendor+ATT&CK inputs, exact affected-asset match, signed knowledge snapshots and a future opt-in, privacy-preserving Cloud learning contract. Static schemas are provided for design but **are not Go runtime API**. No Brain V2 code, local threat intel matcher, neural training/inference, Cloud knowledge service or live response has been deployed.

## Pending review: dual-role Brain, bounded memory and Cloud intelligence

[ADR-023](../adr/ADR-023-dual-brain-memory-cloud.md) **PROPOSES** two logical Brain roles (Go deterministic analytical path and optional hypothesis/ML reasoning) using one evidence/knowledge workspace, with independent policy authorization. It proposes bounded data storage at experimental Starter 256 MiB / Standard 512 MiB / Extended 2 GiB; these are NOT measured deployment figures and exclude application binaries, other apps' logs and kernel map RAM. [Local memory design](../architecture/luna-memory-lifecycle.md) defines sampling/TTL/compact/pressure, while [Cloud intelligence](../architecture/luna-cloud-intelligence.md) defines a future optional feed curator, signed scoped deltas, consent-based aggregate upload and offline local authority. No Cloud Brain is required for per-packet decisions. These concepts must not block P0–P4 standalone operation or bypass the G-STANDALONE and NetBird restrictions.

## Owner-decided offline installation and separate component versions (NOT IMPLEMENTED)

[ADR-024](../adr/ADR-024-offline-base-independent-component-updates.md) **decides** the first shipping installer must contain **Basic Go Luna Brain plus verified baseline knowledge**, a usable standalone App/UI/CLI and separately restricted firewall/recovery, even with no Cloud connection. The Brain is **required inside the install bundle but not allowed to block firewall/TTL operation when it crashes**; heavy optional Python inference remains extra. [Lifecycle specification](../architecture/offline-base-component-update-lifecycle.md) defines separate App, Brain, Executor, per-module, KnowledgePack-revision and optional ML model versions, explicit compatibility graph, signed Cloud/offline distribution and isolated rollback. This supersedes earlier packaging text calling the Brain optional. These are release acceptance criteria only: no installer/Brain/update system exists in this P0 Go CLI branch. Store/budget proposals in ADR-023 remain unaccepted sizing experiments.

## Cloud portability and scale-out direction (DECIDED design, no deployment)

[ADR-025](../adr/ADR-025-portable-scalable-cloud.md) formalizes the owner's requirement that Cloud be movable among VPS-ARM, VPS-AMD, other providers and eventual k3s **without reinstalling devices, changing public API/trust identity or losing KnowledgePack revisions**. Start small with modular Go API + independent workers, **single PostgreSQL writer**, signed immutable object storage via a provider-neutral adapter, container images/config templates and offsite restorable backups. [Architecture](../architecture/cloud-portability-and-scale.md) defines later stateless replica scaling, while [operator runbook](../architecture/cloud-migration-and-recovery.md) protects old-writer fencing, Postgres WAL/PITR, object checksums, key custody, client sync/idempotency and safe failure. **No cloud service, backup, hosted storage or migration is implemented or operated.** The standalone-first gate and Cloud-optional device model remain unchanged.

## One-account local/Cloud identity and owner-managed remote control (DECIDED design)

[ADR-026](../adr/ADR-026-unified-account-remote-fleet-control.md) selects one human OIDC identity for Cloud and enrolled local WebUI accounts, plus **offline local bootstrap owner and protected break-glass**. Cloud Console can later show actual reported firewall/Brain/module status and enqueue versioned typed configuration intents via outbound device connection, not a direct inbound root/shell. [Architecture](../architecture/unified-account-cloud-fleet-control.md) keeps reported local config authoritative, desired Cloud revisions pending until acknowledged and local Policy Gate enforcing scope/revision/MFA/risk checks. Cloud enrollment, remote control and learning telemetry opt-in are different permissions. No Account/Cloud/device enrollment or remote-action runtime is yet implemented; P0 standalone remains sole reachable implementation.

## Collective immunity and behavior scenario sharing (DECIDED design, not deployed)

[ADR-027](../adr/ADR-027-collective-threat-immunity.md) introduces the owner's **immune-style federated defensive memory** requirement: when a device locally detects, labels and optionally mitigates behavior, a separately consented privacy-minimized report may be reviewed by Cloud to improve other devices' early recognition of the *behavior, scenario and scoped subject* rather than IP alone. [Architecture](../architecture/federated-immune-threat-intelligence.md) enforces source-hashed scenarios, independent corroboration, benign negatives, anti-poisoning and signed expiring/retractable ThreatBulletins. **One node's block is an outcome not ground truth or a transferable firewall order.** Clients decide locally, with independent Policy Gate/TTL, and offline Basic Brain survives Cloud outage. Nothing in this pipeline is a deployed service. ADR-023 memory cap sizing remains proposed, unrelated to actual metrics.

## Owner-decided Light Luna self-defense and compromised-client containment (not implemented)

[ADR-028](../adr/ADR-028-self-defense-host-compromise.md) adds a mandatory **Self-Integrity Guardian** and a path for host process/file/egress threat detection, safe local containment and Cloud quarantine of compromised submissions. [Architecture](../architecture/host-self-defense-and-compromise-containment.md) separates XDP ingress filtering from optional eBPF tracing, OS events, integrity checks and bounded on-demand malware scanning; kernel/root compromise can falsify host evidence, requiring a trusted external rescue/credential rotation. A compromised but authenticated device cannot directly distribute trusted global patterns, signed KnowledgePacks, software updates or peer actions. Nothing has been attached to kernel, scanned, quarantined or enforced in current Light Luna P0 source.

## Agent/Gate asymmetric identity and group IPC — source implementation (2026-10-14)

[Ed25519 trust boundary](../architecture/p0-ed25519-agent-gate-identity.md) adds a separate Agent private signing key and Gate public verifier, explicit config of Gate/Agent Unix peer UID and group socket permissions. It rejects mixed Ed25519/HMAC and invalid signatures, while retaining the old HMAC demo as a **separately constructed** Gate. Same-UID HOMESV development tests verify functional signed request flow, replay denial and wrong expected server UID. **Distinct Linux user isolation has NOT been tested**; bootstrap uses same user and test-only key files. No real login, user-group provisioning, privileged executor, signed server response or kernel changes. P0 and G-STANDALONE are not accepted.

## Two-process local Gate/Agent with restart replay protection (2026-10-14, development only)

[Current source contract](../architecture/p0-two-process-gate-and-journal.md) adds `cmd/luna-gate` and `cmd/luna-agent` as actual separate Linux executables run **manually, nonroot and under the same UID**. Their shared ephemeral test key is created in an owner-only scratch dir, and the Gate's bounded nonblocking exclusive-locked `replay.jsonl` synchronizes request ID/nonce before replying. [E2E script](../../scripts/test_p0_gate_processes.sh) proves a stale request is rejected after Gate restart and a corrupt journal refuses startup. This advances from previous one-process IPC selftest but **does not establish production distinct-UID isolation, real owner login, power-loss proof, a privileged executor, rollback or any live nftables behavior**. Signed apply is always denied. P0/G-STANDALONE still NOT passed.

## Linux local IPC developer proof (2026-10-14; real socket, no privileged action)

The implemented [Linux Unix socket transport](../architecture/p0-linux-unix-ipc.md) adds `internal/localipc`, `SO_PEERCRED` UID checks, a private `0700` socket directory and `0600` socket, strict length-prefixed bounded requests and replies, plus a cross-process Go integration test. `lunactl control ipc-selftest` uses **same-process goroutines** but a real kernel Unix socket. No persistent separate Agent/Gate daemons, OIDC ownership, durable replay/transaction, privileged executor, real nft apply, scanner or signed KnowledgePack exist. It **does not** satisfy P0 acceptance or G-STANDALONE. Control gate is denial-only for apply even with valid HMAC.

## Implemented local-control trust-boundary demo (2026-10-14; no executor)

[Control gate source and verification](../architecture/p0-local-control-gate.md) establishes a runnable **in-process** HMAC-SHA256-signed intent with local actor/site binding, policy content/revision match, short expiry and in-memory replay protection. CLI `lunactl control selftest` intentionally rejects replay, tamper and even correctly signed `APPLY_POLICY`. There is no device user login, Unix socket, OS peer identity, Cloud/Brain permission to enforce, persistent replay journal or privileged executor. This is an incremental testable safety contract, **not proof of policy authorization or firewall protection**. HOMESV direct GitHub access is verified through its own existing SSH deploy key (fingerprint-checked), and is the normal source-push route.

## Implemented synthetic Flow Simulator (2026-10-14, not a firewall)

[Flow Simulator implementation and verification](../architecture/p0-flow-simulator.md) adds a real pure-Go `internal/flowsim` entry called by `lunactl firewall policies simulate`; its input consists exclusively of synthetic cases. It evaluates policy intent for IPv4/IPv6, TCP/UDP and ICMP, priority and alias resolution, and flags potential administration lockout in a fixture. It **cannot** establish actual kernel behavior: every effective verdict and external firewall state is `UNKNOWN`; output default has no schema and is `UNSPECIFIED`. Unit tests/CLI/vet passed; no physical or virtual firewall/netns was changed. This does not advance G-STANDALONE to PASS.

## Verified unhooked nftables text compilation (2026-10-14, not a firewall)

[Source-level compiler](../architecture/p0-nft-unhooked-compiler.md) adds `cmd/lunactl firewall policies nft-preview` backed by the new unprivileged `internal/nftpreview` Go package. It renders only deterministic `table inet` regular chains without base hooks or default policy; it **cannot apply policy or prove admin recovery**. Unit tests and CLI output passed; `nft -c` was not run because HOMESV cannot create an isolated unprivileged network namespace and parent-host kernel testing is prohibited. Separate Go build targets, Brain synthetic replay and original policy validator remain unaffected. `kernelWrites`, `syntaxChecked` and `kernelCompatibilityChecked` are all false in preview. NetBird and Cloud remain deferred.

## Implemented P0 offline replay skeleton (2026-10-14, no enforcement)

The [local runtime foundation](../architecture/local-runtime-foundation.md) adds **source-reachable** separate build targets `cmd/lunactl` and `cmd/luna-brain` with distinct App/Brain developer version strings; versioned normalized synthetic observation/replay contracts, demonstrative deterministic Go Brain and non-executing Gate preview; Guardian explicitly reports UNKNOWN; shared strict JSON parser rejects duplicate nested keys and trailing/unknown inputs. It is not a running background Brain service, validated signed base KnowledgePack, actual device health scanner, firewall executor or Cloud agent. See [Code Map](LIGHT_LUNA_CODE_MAP.md) for exact entry points, test evidence, independent build scope and absence of actual enforcement.

## Safety boundary

Mandatory scope and TTL, owner emergency paths, audit, independent privileged policy gate, transactional rollback lease and separation of host/vendor nftables tables are [design requirements](../07-SECURITY-SAFETY.md), **not implemented test results**. No root host firewall modifications; no development code is deployed.

## Open gates

1. P0 strict input/schema negative/fuzz checks, inert compiler, isolated netns tests, independent watchdog contract, scoped CLI/API.
2. Dedicated HOMESV Git SSH authorization and CI workflow (pending separate verification).
3. Full package/install/upgrade/uninstall, isolated enforcement and recovery proof; benchmarked amd64/arm64.
4. P1–P4 Host/Edge release-quality implementation and tests.
5. G-STANDALONE owner signoff, then *separate* NetBird go-ahead.
6. Cloud/brain/WAF/hardware decisions require explicit verified evidence.

## Source bookmarks

- [Architecture](../01-ARCHITECTURE.md), [Firewall](../02-FIREWALL-ENGINE.md), [Security](../07-SECURITY-SAFETY.md), [Roadmap](../11-ROADMAP.md), [ADRs](../12-ADRS.md)
- [CLI source](../../cmd/lunactl/main.go), [policy](../../internal/policy/validate.go), [tests](../../internal/policy/policy_test.go), [fixture](../../testdata/host-baseline.json)
- [Initial handoff](../14-HANDOFF-2026-10-13.md), [P0 implementation record](../15-P0-IMPLEMENTATION-NOTES.md), [HOMESV handoff](../16-P0-HOMESV-HANDOFF-2026-10-13.md) — historical context, not runtime truth.
