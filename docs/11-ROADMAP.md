# 11 — Delivery Roadmap, Test Gates and Release Discipline

**Current milestone:** P0 non-enforcing CLI/policy validation **plus the first executable offline Brain/Guardian/Gate simulation skeleton** exist on the development branch; no implemented kernel firewall, live host scanner, signed base pack or shipping release. See [current runtime foundation](architecture/local-runtime-foundation.md) and [Code Map](project-book/LIGHT_LUNA_CODE_MAP.md). See [Code Map](project-book/LIGHT_LUNA_CODE_MAP.md).

| Milestone | Deliverables | Exit gates |
|---|---|---|
| **P-1 — Product contract** | Docs, ADRs, threat model, ownership, compatibility assumptions, test plans | Architecture review; no mutation of production |
| **P0 — Local Firewall Foundation** | Go agent + typed executor, nftables compiler, aliases/rules/zones, local WebUI + CLI, API schemas, simulated Host/Edge chooser | Tests on isolated Linux netns; clean install/upgrade/uninstall; unknown firewall conflicts blocked |
| **P1 — Standalone Host candidate** | Stateful filtering, controlled timed bans, policy diff/simulation, auth/audit/rollback, IPv6 and generic Docker/UFW coexistence | Admin lockout negative tests pass; external rescue works without Luna; signed packages amd64/arm64; no NetBird dependency |
| **P2 — WAF & Luna Shadow** | Proxy/WAF connector, OWASP CRS observe mode, replay corpus, threat-feed/SBOM engine, Luna recommendations only | Labeled FP measurements; no automatic enforcement; documented feature coverage |
| **P3 — Controlled Autonomy** | Approved scoped TTL-capped ban/rate/challenge/unban, outcome verification, site-specific policy, rollback | Canary field tests + owner approval; no unauthorized action in red-team exercises |
| **P4 — Standalone Edge & Fleet** | Secure first-party remote agents and multi-node WebUI over ordinary owner-provided IP/DNS/mTLS transport, offline queues; no NetBird integration | Host and Edge work independently, deployable packages, local+remote policy actions, restarts/offline resilience, demonstrated end-to-end recovery |
| **P5 — Optional NetBird Integration (GATED)** | Only after standalone readiness acceptance and owner go-ahead: read-only NetBird connector, then separately approved advanced features | Confirm standalone readiness evidence, peer/API scope, route/IP attribution, noninterference, disconnected operation and multi-node regression |
| **P6 — Gateway & Hardware Reference** | NAT/forward/VLAN, inline IDS/IPS option, appliances, thermal/performance, recovery console | 1/2.5/10 Gbps measured device-specific throughput and packet loss; power loss recovery |
| **P7 — Cloud & OEM** | Signed intelligence, fleet multi-tenant, cloud privacy and licensing, OEM image only with tested hardware | Multi-tenant penetration testing, backup/restore, scale/failure model, vendor/licensing review |

**HARD ORDER — NOT OPTIONAL:** P0–P4 focus on a working **independent** firewall/Edge system. NetBird is *not* permitted as an early parallel implementation, even read-only. Do not add its client/server packages, API calls, credentials or NetBird-specific automated integration tests to standalone milestones. Public research docs may be retained without code. WAF and core Luna work belong to independent development. Hardware gateway can be evaluated separately when core is stable.

### Gate G-STANDALONE — required before NetBird engineering

1. A real, packaged Light Luna build installs, starts, upgrades and uninstalls without NetBird or Cloud, on both `amd64` and `arm64` reference Linux environments.
2. The Host firewall runs end to end: create rules/aliases, simulate, apply in isolated test environment, observe counters, remove/rollback without breaking host access. WebUI and `lunactl` control it locally.
3. Edge Mode runs without NetBird: remote agent enrollment/authentication, authorized command execution and telemetry over explicitly configured generic transport. No transit packet visibility claims.
4. The independent enforcement safety gate, TTL expiration, crash/reboot recovery, installer hygiene and emergency rescue pass documented negative tests.
5. CI and reproducible test reports capture pass/fail plus versions, and the owner explicitly accepts **G-STANDALONE** and authorizes NetBird integration work.

Passing only a unit test or completing docs does **not** satisfy the gate. Until all conditions pass, mark NetBird `DEFERRED` and do not begin its development.

## Offline installer / component release gate (ADR-024)

[ADR-024](adr/ADR-024-offline-base-independent-component-updates.md) makes a **Basic Go Luna Brain and verified baseline knowledge mandatory in the first shipping standalone install**; only larger ML/reasoning enhancements remain optional. App/Brain/Executor/Knowledge/modules/model have independently versioned, signed, rollbackable delivery and explicit compatible ABI. **Design the package boundaries in P0**, prove offline installer + Basic Brain + firewall/recovery in P1 release acceptance, then test App-only / Brain-only / Knowledge-only and coordinated incompatible update transactions. Cloud transport is an optional distribution channel and may be built later; do not implement required Cloud enrollment in P0–P4. Manual signed offline updates must work without Cloud.

## Optional eBPF/XDP capability track (ADR-021)

This track is **not a mandatory standalone acceptance dependency** and does not authorize real firewall changes:

- **Early P0/P1 research:** documented Linux feature matrix and read-only inventory of current BPF/XDP attachments (never detach or replace). Specify data contracts and eBPF userspace loader API as a proposed extension; prefer nftables-only as baseline.
- **P1/P2 laboratory observer:** bounded tracing/perf-ring/bpf-ringbuf candidate; synthetic events, kernel/driver/arch checks and clean detach in disposable VM/netns. Observe and validate only, no traffic drops.
- **P2/P3 experimental XDP:** XDP_PASS-only counter in isolated lab, then shadow attack classifier. Validate native driver vs generic SKB, IPv4/v6, fragments, drops, CPU and legitimate-service latency; coexistence with existing XDP dispatchers must fail safely.
- **P3+ opt-in live mitigation (separate approval):** scoped L3/L4 XDP_DROP only after policy-gate authorization, independent rollback and external admin reachability pass. Unavailable on unsupported platforms; standalone nftables must still work.
- **Never claim Cloudflare's >8M pps case as a Light Luna result.** XDP does not solve saturated upstream links, nor replace conntrack/WAF. See [ADR-021](adr/ADR-021-ebpf-observability-xdp-optional.md).

## Gate-specific test suites

- **Network correctness:** packet/flow fixtures, ICMP/TCP/UDP, IPv6, conntrack, interface/zone changes, DNS alias expiry, NAT hairpin where supported.
- **Standalone coexistence:** nftables hook priority and existing vendor-owned tables, Docker bridge/port publishing, UFW/firewalld, containerized proxy, cloud security groups; **NetBird-specific `wt0`/ACL interoperability tests occur only after G-STANDALONE**.
- **Safety:** worst-case malicious rule, lost SSH, disconnected remote management path, model crash, CPU overload, disk full, interrupted upgrade, rollback path and expiry after reboot.
- **WAF:** OWASP CRS coverage, false positives on documented normal traffic, TLS termination, HTTP/2, streaming, WebSocket and large bodies; NetBird Reverse Proxy compatibility probing is deferred to P5.
- **Trust:** forged forwarded headers, NAT masking, VPN CGNAT address overlap, replayed signed command, compromised feed and adversarial log prompt injection.
- **Release quality:** multiarch native package, reproducible build, SBOM, provenance, signed updater, E2E setup wizard and independent agent health.

## Decision KPIs (target definitions, no invented results)

False-positive rate against independent labeled benign traffic, recall against vetted attack corpus, admin recovery time in controlled failures, packet path latency, throughput and drop rate by hardware mode, peak memory under backpressure, policy apply/rollback success, number of unverified IP attributions used in auto-ban (target zero), drift/anomaly explanatory quality.

## P0 evidence integrity: FD-held network namespace denial (2026-10-14 — isolated KVM PASS)

[Source + proof](architecture/p0-vm-autonomous-ttl-watchdog.md): read-only guest inventory and rollback reviewer now deny a "clean" kernel state if any bounded visible process FD holds a foreign nsfs inode, even when the named namespace has been unlinked and a rollback DONE exists. Five added negative tests PASS; full Python VM suite 91 PASS; exact no-NIC KVM live FD experiment returns fail-closed during FD lifetime and clean only after FD close (active journal unchanged). This prevents a known false-negative ahead of future **ADR-029 Prometheus/Grafana ban-status reporting**; a bounded proc inventory still cannot certify absence against hostile root/kernel. P0 OPEN, G-STANDALONE BLOCKED.

## P0 synthetic security events and non-delivering channel plans (2026-10-14)

[Go events module](../internal/observability/events.go), [schema](../schemas/local-security-events-synthetic-v1alpha1.schema.json), [test fixture](../testdata/local-security-events-synthetic.sample.json) and [unit tests](../internal/observability/events_test.go) now validate **synthetic-only** event types and compute deterministic `SIMULATED_NOT_SENT` route plans for Telegram/Discord/Slack/email with a hard fan-out cap. Strict input and anti-forgery checks refuse real `BAN_APPLIED`, actual notification-send claims, raw IP/credential fields and automatic owner authority. **Source-only**: live audit storage, verified ban reconciliation, metrics HTTP endpoint, authenticated ban API, durable outbox, real provider transports, Grafana and signed owner auth remain open. P0 OPEN / G-STANDALONE BLOCKED.

## Active Core P0 — actual Gate wire authenticator hardening (2026-10-14)

[Actual Gate source](../internal/controlgate/control.go) now requires canonical lowercase hex encoding of Agent HMAC or Ed25519 authentication material. Pre-fix RED tests prove uppercase spellings of valid bytes previously yielded `ADVISORY_ONLY`; post-fix GREEN tests prove deny-before-replay and legitimate canonical-preview/replay-denial behavior. The existing separate Linux Agent/Gate restart regression passes in both modes. This is a **non-enforcing development security invariant**, not human authorization, transaction durability, snapshot protection or a real nft apply test. Kernel/privileged Executor unchanged; P0 OPEN / G-STANDALONE BLOCKED, ADR-029 parked, NetBird deferred.

## Active Core P0 — detached Agent signature spelling hardening (2026-10-14)

[Source and negative test proof](architecture/p0-offline-owner-review-separation.md) now require canonical lowercase hex for Agent Ed25519 signatures in the **detached offline reviewer**, preventing an uppercase textual alias of valid bytes from being acknowledged as test evidence. Red/green nonroot journal and in-memory tests prove rejection does not spend replay budget, and canonical test evidence retains replay denial. This is **not owner authorization, key revocation, Gate integration, privileged execution or a firewall transaction**. No policy apply or production kernel operations were added. Core P0 remains OPEN; G-STANDALONE BLOCKED, ADR-029 parked, NetBird deferred.

## Active Core P0 — restored offline reviewer journal semantics (2026-10-14)

[Source and tests](architecture/p0-offline-owner-review-separation.md) now reject semantically invalid SHA-256 replay digests when reopening the separate nonroot test journal and preserve consumed Agent intent replay across an externally re-pinned **test reviewer key**. This is a **source-only negative verification hardening**, not supported owner rotation/revocation, authenticated consent or a protected anti-rollback counter. No Gate/Executor privilege boundary or kernel action enabled; ADR-029 integrations remain parked, NetBird deferred. P0 OPEN / G-STANDALONE BLOCKED.

## Active Core P0 — owner key lifecycle contract and no-authorization regression (2026-10-14)

[Versioned reviewer key lifecycle contract](architecture/p0-owner-key-lifecycle-revocation-contract.md) records future K01–K19 acceptance obligations for revocation, rotation, generation anti-rollback and real independent owner enrollment. **Design only; no production lifecycle code or owner grant.** [Standalone deny-only tests](../internal/controlgate/owner_review_authority_regression_test.go) strengthen existing verifier coverage for absent/mismatched test public key, time boundary and self-asserted owner role. Rotation/revocation implementations and anti-rollback proofs remain **NOT IMPLEMENTED**. Gate `APPLY_POLICY` continues to reject; no nft or executor modification. P0 OPEN / G-STANDALONE BLOCKED, ADR-029 parked.

## Active Core P0 — optional detached reviewer journal replay across restart (2026-10-14)

[Core offline architecture](architecture/p0-offline-owner-review-separation.md): test-only review evidence can now be recorded in a **separate locked, private, fsynced nonroot journal**, restoring review-ID and Agent intent digest replay state across ordinary process restart. Corrupt, inaccessible, concurrently owned or unsyncable journal refuses verification. **Owner login/presence, clock rollback, disk power-loss proof, real Gate authorization and kernel APPLY remain NOT IMPLEMENTED**, and the reviewer artifact is not wired to any privileged process. Next Core steps must establish independent real owner enrollment/consent semantics and separately verified safe transaction/rollback before any executor. P0 OPEN, G-STANDALONE BLOCKED; monitoring/notification and NetBird remain deferred.

## Active Core P0 — separate Agent identity from owner review (2026-10-14)

[Inert Core verifier](architecture/p0-offline-owner-review-separation.md) pins two distinct test public keys and checks signed review evidence bound to the full Agent request, with exact site/nonce/policy/revision/operation/TTL and bounded in-memory replay. It **does not establish authenticated human-owner consent** and never grants APPLY; all existing Gate/Executor denial remains intact. Next mandatory steps are independently verified owner identity/consent, durable authorization with clock/replay safeguards, least-privilege executor and isolated kernel/rollback tests. No product Gate action is unlocked by an offline fixture. P0 OPEN / G-STANDALONE BLOCKED; ADR-029 still parked.

## Active Core P0 — secure journal creation boundary (2026-10-14)

[Development source + evidence](architecture/p0-two-process-gate-and-journal.md): nonroot Gate replay journal creation distinguishes new/existing file, syncs **new inode then owner directory** before accepting requests, and retains exclusive locking and corrupt-journal fail-closed rules. Go tests ×3, go vet, HMAC/Ed25519 test process roundtrips and 91 Python offline tests PASS. This is **not** evidence of hardware power-loss safety, owner-authenticated APPLY, installer or signed production executor; next Core work is actual owner authorization model, transactional receipts and safe isolated recovery. ADR-029, Cloud and NetBird remain deferred. P0 OPEN / G-STANDALONE BLOCKED.

## Active Core P0 — bounded exact CIDR-set source-union coverage (2026-10-14)

[Core audit and test evidence](architecture/p0-policy-rule-conflict-audit.md): a source-only Go helper now verifies coverage of later CIDRs by a single earlier rule's **union** of multiple same-family canonical CIDRs; eliminates a conservative false-negative on fully adjacent /25 or /33 aliases. Tests guard missing gaps, IPv4/IPv6 separation and comparison budgets. This does not expand Policy Gate authority, real ban capabilities, signed-owner approval or G-STANDALONE. Continue only on core; ADR-029 stays frozen.

## Active Core P0 — rule family consistency and static rule conflict audit (2026-10-14)

[Implemented source/verification](architecture/p0-policy-rule-conflict-audit.md): input policy validation now rejects mismatched ICMP family including CIDR-set alias entries, and `lunactl firewall policies analyze <file>` exposes a bounded pure-Go analyzer for definite later-rule shadowing vs reviewable different-action overlap. No claim about a live host firewall or actual kernel packet verdict; budget exhaustion reports `INCOMPLETE_REVIEW_REQUIRED`. **Next core priorities** remain signed local owner Gate/Executor boundary, independent crash-safe rollback and rescue, standalone Host/Edge packaging, test/release gates. **ADR-029 integrations are frozen.** P0 OPEN / G-STANDALONE BLOCKED.

## Planned standalone observability and owner notification integrations (2026-10-14 — decided, not implemented)

[ADR-029](adr/ADR-029-local-observability-and-notifications.md) and [contract](architecture/local-observability-and-notifications.md) preserve **future** Prometheus/Grafana metrics, read-only ban/event API and opt-in Telegram/Discord/Slack/SMTP notification requirements. **OWNER PRIORITY CORRECTION (2026-10-14): ARCHITECTURE ONLY / IMPLEMENTATION FROZEN.** Do not develop the exporter, dashboard, notification outbox, adapters, synthetic observability/event schemas or integrations further during the current core milestones. Two inert Go synthetic prototype modules already committed remain **parked and not runtime-connected**; they are NOT evidence of P0 core readiness. Focus exclusively on policy/validation, trusted Gate/Executor and owner authorization, deterministic rollback/recovery, Host/Edge standalone functionality, test and packaging. Revisit ADR-029 implementation **only after standalone core readiness is demonstrated and the owner explicitly reprioritizes it**; no assumption that the previous P1/P2 estimates are active work. No NetBird integration before G-STANDALONE plus separate owner approval.

## P0 guest-local fsync fault injection for immutable synthetic receipts (2026-10-14)

[Evidence and limitations](architecture/p0-vm-autonomous-ttl-watchdog.md): simulated receipt file/dir fsync `EIO`, pre-open `ENOSPC`, orphan/truncated JSON and no-overwrite evidence are now covered by [10 focused tests](../scripts/lab/vm/test_rollback_fsync_faults.py), PASS offline and PASS in **exact no-NIC KVM** using guest temporary directories. Active journal hash did not change; no nft rules were installed or altered by this fault suite. This is bounded failure-path behavior, **not** simulated power loss at the kernel or proof of durable receipt consensus. P0 blockers remain actual crash/power-failure matrix, signed owner authorization, boot-time safe handling of ambiguous epochs, protection against FD-held namespaces and production installer/executor. G-STANDALONE BLOCKED; NetBird deferred.

## P0 read-only recovery-evidence classification (2026-10-14 — no-NIC KVM verified)

[Source and test evidence](architecture/p0-vm-autonomous-ttl-watchdog.md): a new **read-only** rollback reviewer differentiates prepared-only, kernel-observed-absent, journal-terminal-but-DONE-missing, coherent DONE, cross-boot and invalid/foreign kernel evidence. 23 new offline unit tests PASS (76 total); actual KVM current missing-DONE and archived missing-DONE cases **exit 3**, while a newly completed normal TTL test produces an observation-consistent primary DONE (exit 0) that **does not grant owner success/permission**. No automatic repair, production signed receipts, kernel/disk atomic commit, FD-held namespace completeness or power-failure guarantee. P0 OPEN; G-STANDALONE BLOCKED.

## P0 immutable guest-only rollback observations and terminal-before-DONE crash (2026-10-14)

[Verification architecture and evidence](architecture/p0-vm-autonomous-ttl-watchdog.md): within the existing single-writer `flock`, primary and fallback emit non-overwriting, bound and fsynced `prepared → kernel-absent → journal-terminal` **observation** records before their DONE. Twelve new offline negative/order tests PASS. One real NIC-less KVM `os._exit(73)` after terminal journal but before DONE correctly produced a **missing-DONE review status** with all three observations intact, not forged success; ACK-gated archival preserved failed and earlier post-cleanup evidence. A normal primary TTL KVM suite and both fallback service-death cases were rechecked and PASS. Remaining release blockers: power-loss/torn write investigation, signed owner authority, durable epoch leases, real execution and coexistence, installer, CI and fail-safe admin reachability. **Not** a production crash-safe WAL; P0 OPEN and G-STANDALONE BLOCKED.

## P0 cooperative rollback write fencing in isolated KVM (2026-10-14 — bounded lab PASS)

[Implementation and evidence](architecture/p0-vm-autonomous-ttl-watchdog.md): both VM PID1-managed synthetic rollback workers now hold a shared bounded `fcntl.flock` across current journal check, fixed guest-netns nft deletion, terminal journal fsync and DONE receipt. Ten offline process/security tests PASS; real same-deadline dual-arm experiment produces only one primary DONE, and both postcommit SIGKILL takeover cases still recover synthetic IPv4/IPv6 TCP/UDP. Deliberate holder death after test table removal again proves **journal/kernel divergence remains possible despite exclusive lock** and is preserved without a fake success. The missing production capability is **durable single-owner epoch + recovery/receipt protocol** with signer verification, crash-point/power-cut matrix and host coexistence. P0 OPEN; G-STANDALONE BLOCKED; NetBird deferred.

## P0 dual-arm and crashpoint transaction consistency (2026-10-14 — VM-only detection PASS)

[Bounded evidence, implementation, remaining blockers](architecture/p0-vm-autonomous-ttl-watchdog.md): healthy primary and prearmed fallback finished with only **primary DONE**, while fallback separately recorded `PRIMARY_ALREADY_TERMINAL_NO_TAKEOVER`. At a deterministic fault after fallback's **fixed guest-netns nft delete** but before journal fsync, evidence showed journal **`COMMITTED_SYNTHETIC` with no nft table** and no false `ROLLED_BACK` or success DONE. Operator-acknowledged archival preserved the inconsistency. This is not transaction atomicity: a future single-writer lock/owner fencing, durable pre/post receipts, recovery authority and disk/power-failure matrix remain P0 blockers. G-STANDALONE remains BLOCKED.

## P0 independent fallback for guest synthetic service death (2026-10-14 — VM lab PASS)

[Source, boundaries and verification](architecture/p0-vm-autonomous-ttl-watchdog.md): the new separately armed PID1-managed guest-netns fallback survives a primary watchdog or namespace keeper SIGKILL, validates exact synthetic nft table structure, restores guest-private IPv4/IPv6 TCP/UDP, and records a distinct synthetic-only receipt. Two runs of both postcommit SIGKILL cases PASS after a strict nft JSON guard; 12 rule-whitelist unit tests PASS. This updates the prior observed *missing backup* gap **only for a no-NIC disposable VM experiment**. It does not pass the P0 release gate without owner-authenticated privileged executor, signed receipt recovery, service installer, concurrent journal safety, real host coexistence and crash/reboot/power-loss verification. NetBird remains deferred.

## P0 service death and missing fallback rollback (2026-10-14 — lab detection PASS)

[Source and verification](architecture/p0-vm-autonomous-ttl-watchdog.md): a no-NIC KVM VM experiment with 3 real SIGKILL cases demonstrates `keeper_precommit` leaves a noncommitted `VALIDATED` journal, while `keeper_postcommit` and `watchdog_postcommit` leave actual guest-netns synthetic nft drops **active after the monotonic deadline**. No false DONE or `ROLLED_BACK` receipt was generated. A scope-bound EXIT trap and independent manual archive cleared the synthetic **lab** footprint and preserved evidence without claiming an autonomous rescue. This is an **unresolved P0 product capability gap:** independent fallback supervision/recovery and owner-authenticated policy control still need to be designed and tested. G-STANDALONE remains BLOCKED.

## P0 actual forced KVM reboot / stale journal inspection — lab PASS (2026-10-14)

[Verified source and bounded experiment](architecture/p0-vm-autonomous-ttl-watchdog.md): a no-NIC independent-kernel VM ran a fixed synthetic guest-netns ICMP DROP with positive packet counter and journal `COMMITTED_SYNTHETIC`, then only that VM was forcibly stopped/restarted. Linux boot ID changed; read-only inventory observed no root nft/named/foreign-process netns, but the stale journal was **not** marked `ROLLED_BACK`. Cross-boot transition was denied and an explicitly acknowledged separate operator step archived evidence unchanged. 14 new offline inspection tests + existing 5 boot-ID tests PASS, and full KVM regression passed after inspection/archive. **Not a production reboot-rescue system or verified power-loss rollback; P0 and G-STANDALONE remain OPEN/BLOCKED.**

## P0 VM automatic TTL watchdog — verified lab slice (2026-10-14)

[Actual autonomous expiry design and evidence](architecture/p0-vm-autonomous-ttl-watchdog.md): VM-only PID1 `guest-auto-watchdog.py` validates an inert signed-**NO** owner-policy test manifest, kernel/mount isolation, exact PID birth ticks and target netns inode, fsyncs READY before synthetic commit, uses monotonic TTL, removes only fixed VM netns `luna_p0_stage`, and fsyncs a DONE receipt after journal `ROLLED_BACK`. Independent systemd namespace keeper prevents controller-owned anchor lifetime failures. **Full regression VM suite PASS 3 consecutive runs** after root-causing cleanup and namespace failures; IPv4/IPv6 TCP/UDP recovered without explicit rollback command. This **does not** prove real owner approval, production executor, reboot/power loss rescue or host firewall safety. P0 and G-STANDALONE remain OPEN.

## P0 typed VM staging, negative tests and deterministic separate-service revert (2026-10-14)

[Source implementation and measured VM evidence](architecture/p0-vm-synthetic-stage-revert.md): `vmstage.Build` returns a deterministic, non-executing policy-hash/revision staging manifest and TTL bound (1–10s), `lunactl vm stage-intent` prints it. A VM-root-only guest journal rejects wrong TTL/ID, out-of-order and duplicate commit; packet fixtures prove IPv4/IPv6 plus TCP/UDP accept → **synthetic-only** drop with counters → **explicit independent systemd service** rollback and restored traffic. Full suite **PASS two consecutive reruns** after explicit reconciliation of interrupted test states. **Autonomous TTL watchdog NOT ACCEPTED** (systemd timer nondeterministic in current harness); production executor, real owner approval, crash/reboot continuity, rescue path, Docker/UFW coexistence and G-STANDALONE still open.

## P0 KVM separate-kernel nft preview syntax + synthetic packet rollback (2026-10-14)

[Live VM acceptance evidence](architecture/p0-lxd-kernel-nft-rollback.md) is **PASS twice**. LXD VM `light-luna-p0-nftvm`, guest kernel different from HOMESV, **no NIC**, uses only ephemeral network namespaces. Go-generated **unhooked** policy preview passed `nft -c` kernel syntax check, malformed syntax rejected; synthetic hook blocked ICMP with counter then independent test watchdog restored ping after a stand-in controller was killed. These prove parser compatibility for the synthetic baseline and VM netfilter/rollback feasibility **only**. Real Light Luna hooked-rule compiler, executor, rescue canary, reboot/power-loss rollback, Docker/UFW/VPN coexistence and G-STANDALONE remain **OPEN/BLOCKED**. No production firewall changes.

## P0 LXD cross-UID isolated identity lab — deployed and tested (2026-10-14)

**Latest verified result:** LXD 5.21.8 installed on HOMESV; Ubuntu 24.04 `light-luna-p0-lab` created as **unprivileged, no-NIC, no-nesting** container (2 vCPU, 1536MiB, 256 processes). Live different-UID Gate/Agent/outsider/rogue-with-test-key signed-request matrix **PASS** via `scripts/lab/lxd/run-tests.sh`: authorized synthetic preview advisory-only, APPLY denied, wrong Gate UID denied, ungrouped outsider denied, wrong UID with copied synthetic test private key denied, restart replay rejected, corrupted journal refused. Initial LXD setup bridge `lxdbr0` was removed and its default-profile NIC deleted. This unblocks *P0 local Linux peer-UID lab evidence only*: **does not establish G-STANDALONE, Linux kernel nftables behavior, power-loss journal integrity, Cloud or NetBird readiness**.



[Current lab plan/evidence](architecture/p0-lxd-cross-uid-lab.md) and [five networkless LXD scripts](../scripts/lab/lxd/README.md) now define a dedicated LXD profile, no NIC, bounded CPU/RAM and a different-UID Agent/Gate/outsider/rogue-with-test-key matrix. **Historical pre-install checkpoint superseded:** the first attempt was denied; a subsequent approved Light Remote call installed LXD and executed the container tests. The verified deployment above is authoritative. `bash -n` validates source syntax only. Next real milestone is trusted local LXD installation/init with no bridge, then run tests and capture real results; a container is **not** kernel-isolated, so nftables/BPF/rollback requires a separate disposable VM.

## P0 Ed25519 request authentication and group Unix socket contract (2026-10-14)

[Agent/Gate Ed25519 contract](architecture/p0-ed25519-agent-gate-identity.md) is now a runnable separate mode: `luna-gate serve-ed25519` holds a pinned Agent **public** key and replay journal; `luna-agent preview-ed25519` holds only Agent **private** signing key and validates a configured Gate UID. `ListenGroup` enforces restrictive 0710 directory/0660 socket GID and Gate UID, while `Serve` checks the allowed Agent UID. Ed25519/HMAC mixed-signature and downgrade attempts are rejected. **HOMESV proof uses same UID/GID; no real Linux account separation**, human owner login, permanent service, signed response, firewall executor, rollback or kernel changes. Next milestone is a VM multi-UID access matrix with a truly separate key broker and owner consent.

## P0 separated Agent/Gate processes and replay journal (2026-10-14 — safe source slice)

[Separated Agent/Gate developer proof](architecture/p0-two-process-gate-and-journal.md) introduces real manually launched `luna-gate serve` and one-shot `luna-agent preview/apply` Linux executables under the same **unprivileged** UID. Their `internal/devkeys` creates 0700/0600 throwaway test credentials; `internal/controlgate` now supports a strictly bounded locked/fsynced JSONL replay journal and fail-closed startup on corruption. [E2E script](../scripts/test_p0_gate_processes.sh) demonstrates a signed preview, replay denial, signed apply rejection, restart replay refusal, corruption lockout; no daemon installed or live firewall changed. **G-STANDALONE remains BLOCKED**: production distinct UID key custody, owner auth and consent, durable power-loss/journal recovery semantics, isolated nft+watchdog acceptance still absent. Next P0 increments should target different-UID/sandboxed service identity and VM safety tests, not Cloud/NetBird.

## P0 Linux Unix-domain IPC foundation (2026-10-14 — developer-only)

[Linux Unix IPC implementation](architecture/p0-linux-unix-ipc.md) adds `internal/localipc` with actual local socket framing, Linux kernel `SO_PEERCRED` same-UID verification, private socket directory/mode, bounded timeouts and concurrency, and a callable `lunactl control ipc-selftest`. The test suite includes **real cross-process socket traffic**, invalid frames, duplicate nested JSON, bad UID and secure path permission checks. Authentication remains ephemeral HMAC with **memory-only** replay and no real local owner account; **all APPLY_POLICY intents remain rejected**, no privileged service/host firewall or production reachability. Next gate: dedicated service-user and session key provisioning, durable replay, isolated VM nft syntax and independent watchdog/rollback before any enforcement.

## P0 local authenticated control intent proof (2026-10-14; no IPC/executor)

[Local control boundary](architecture/p0-local-control-gate.md) is implemented in `internal/controlgate` with HMAC-SHA256, domain-separated typed intent, pinned local actor/site/policy digest/revision, strict expiry and bounded replay protection against repeated nonce/request IDs. `lunactl control selftest <policy>` demonstrates ephemeral authentication, tampering/replay rejection and a **denial-only APPLY_POLICY** path. This is an **in-process test**, not OS peer authentication, durable replay state, rescue reachability, privileged executor or production device account support. Next: isolated local IPC framing/credential verification (without privileges), then a separately guarded executor and independent rollback in a disposable VM. Do not promote P0 to G-STANDALONE.

## P0 pure-Go flow simulator implementation (2026-10-14)

[Flow simulator](architecture/p0-flow-simulator.md) is now callable from `lunactl firewall policies simulate <policy-file> <synthetic-flow-file>`. It evaluates current policy's **declared** TCP/UDP/ICMP/ICMPv6 flow intent, ordered rules, CIDR alias, nonterminal log and inbound default; output no-match remains `UNSPECIFIED`. It marks management-critical denial `POTENTIAL_LOCKOUT` but never claims actual route reachability; every `effectiveVerdict` and external firewall state stays `UNKNOWN`. JSON size/event caps, negative tests and Go vet pass. **Next P0 exit prerequisites:** isolated kernel syntax/packet tests and independent trusted rescue/rollback, not Cloud/NetBird.

## Additional P0 compiler implementation — source verified, not kernel verified (2026-10-14)

[Unhooked nft preview compiler](architecture/p0-nft-unhooked-compiler.md): `lunactl firewall policies nft-preview <fixture>` emits deterministic bounded `table inet` regular chains **without any base-chain hooks, default policy, or executor**; IPv4/IPv6 family mismatches and unbounded alias expansion are rejected. Its JSON status truthfully reports no kernel writes and `syntaxChecked=false`. HOMESV has nftables 1.0.9 but cannot create an isolated unprivileged network namespace; **do not run `nft -c` against the parent machine**. The next gate is an isolated disposable VM/nft grammar test, then a safe flow simulator and independent auth/rollback; this is NOT a standalone firewall.

## Verified first runtime scaffold (P0 source, not standalone acceptance)

The [implemented Go foundation](architecture/local-runtime-foundation.md) now supplies recursive duplicate-key rejecting JSON, typed synthetic observation replay, separate `luna-brain` and `lunactl` build targets, Guardian missing-evidence status, and a Policy Gate advisory-only preview. Next priority remains the independently safe **nftables compiler and kernel-namespace validation**, rather than Cloud/XDP or live monitoring. This source implementation does not satisfy G-STANDALONE or release-level Brain/Host Guardian requirements.

## Luna Brain V2 and security knowledge maturation (non-blocking plan)

- **P0:** define versioned `KnowledgeClaimV1` / `DecisionProposalV1` contracts, safe local input validation, synthetic replay, source provenance. Do **not** copy private Sentinel memory into public code or enable model enforcement.
- **P1:** optional offline shadow migration of reviewed Sentinel facts/counterexamples, deterministic Go baseline, site-scoped evidence and temporal memory; independent benign-label evaluation.
- **P2:** signed/source-hashed intelligence ingestion (CISA KEV, NVD, FIRST EPSS, MITRE ATT&CK, vendor advisories), exact SBOM+backport+route matching, scenario reasoning and early **recommendations** only; missing proof must produce `UNKNOWN`.
- **P3:** qualified response preparation and independently authorized bounded mitigation only after false-positive, rollback, provenance and owner gates. Optional Python model worker can be benchmarked, never required.
- **P4:** standalone Edge nodes exchange authorized, site-scoped evidence without Cloud/NetBird.
- **P7 (future):** optional Cloud collective knowledge, signed/versioned pack distribution and opt-in anonymized validated experience; privacy review, poison resistance and rollback required. No automatic global bans from repeated reports.

See [Brain V2](architecture/luna-brain-v2.md), [Threat Pipeline](architecture/threat-intelligence-knowledge-pipeline.md), [ADR-022](adr/ADR-022-rebuild-luna-brain-governed-knowledge-migration.md) and the [legacy source audit](history/audits/SENTINEL_LUNA_COGNITION_AUDIT_2026-10-13.md). Brain V2 remains **design only**, not a dependency that blocks the basic firewall.

## Cloud portability acceptance gate (design now; later optional Cloud release)

Before Cloud implementation, freeze the [ADR-025 portability contract](adr/ADR-025-portable-scalable-cloud.md): stable client DNS/API, OCI multiarch image, provider-neutral config and object API, single-writer PostgreSQL/state migrations, idempotent jobs, reproducible Compose profile and secure signed KnowledgePack custody. Follow the [portable architecture](architecture/cloud-portability-and-scale.md) and [operator migration rehearsal](architecture/cloud-migration-and-recovery.md). A Cloud release is **not accepted** until tested backup restore to a clean VPS, device-ID/tenant/consent/revocation continuity, no reinstall, blob hash/signature check, safe old-writer fencing and tested rollback under failure. Validate Compose→Compose first, then Compose→k3s only when useful. Keep P0–P4 independent of Cloud; G-STANDALONE/NetBird stage order is unchanged.

## Local/Cloud single account and safe remote management (ADR-026)

[ADR-026](adr/ADR-026-unified-account-remote-fleet-control.md) decides the finished product should use one Cloud user identity for its enrolled local WebUI and Cloud Console, permit locally controlled offline owner bootstrap/recovery and expose actual security/configuration state for remotely authorized Cloud management. The [architecture contract](architecture/unified-account-cloud-fleet-control.md) records local OIDC session mapping, separate tenant/device identity, enrollment consent, outbound polling, desired-vs-reported configuration with revision CAS, RBAC, local Policy Gate and receipts. Build secure **local authentication and policy APIs** first during standalone phases; remote Cloud endpoints/agent integration are future optional Cloud work after standalone acceptance, without NetBird dependency. Test Cloud takeover, session expiry, queued stale commands and remote-admin lockout before launch.

## Collective defense and immune-style sharing (ADR-027)

[ADR-027](adr/ADR-027-collective-threat-immunity.md) approves the behavior + subject + scenario learning direction. [Federated defense flow](architecture/federated-immune-threat-intelligence.md) defines:
- **P0 design only:** typed bounded `ThreatSignal` / `ThreatBulletin` contracts and local scenario source/hash trust; synthetic anti-poisoning/false-positive fixtures.
- **P1/P2 local only:** Go event parsers + behavior windows/scenario/counterexamples in `OBSERVE`/shadow; source identity proof and owner labels.
- **P3 gated local actions:** independent Policy Gate and audited scoped TTL/recovery, not global automatic bans.
- **P7 optional Cloud:** opt-in contributions, source validation/dedup/corroboration, signed relevant watch and curated behavior packs, revocation/appeals, anti-Sybil, latency/FPR benchmarks, safe migration. Existing signed knowledge channel can transport these; no new central packet decision dependency.
- **Cloud → device:** scenario and watch intelligence can quickly prime recognition but never bypass local authorization. A single local block is not distributed as confirmed maliciousness.

## Light Luna Self-Defense and Host Guardian (ADR-028)

[ADR-028](adr/ADR-028-self-defense-host-compromise.md) adds an essential safety gate: Light Luna must check its **own integrity and privileged boundaries**, collect defensible Linux host-compromise signals, and refuse to let compromised clients poison Cloud/other clients. [Host Guardian design](architecture/host-self-defense-and-compromise-containment.md) and static [report schema](../schemas/host-integrity-report-v1alpha1.schema.json) separate measured evidence and Cloud trust. Phase scope:
- **P0 design/tests:** typed host integrity/quarantine states, signed artifact + source hash contracts, no executable remote payloads, synthetic negative tests.
- **P1 baseline:** independent self-integrity checker/supervisor with least-privileged components, bounded local process/log/integrity observations; no invasive root scan required for first firewall correctness gate.
- **P2 host telemetry and scanning:** optional observe-only eBPF process/file/socket hooks and **optional bounded on-demand file scan** (candidate ClamAV/fanotify later). Validate support, benign false positives and CPU/I/O budgets; on-access blocking separately gated.
- **P3 safe containment:** local owner Policy Gate, limited lateral/C2 egress or process-specific protection, preserved admin/recovery, TTL, crash rollback, lab only before approval.
- **P7 Cloud contamination resistance:** every contribution untrusted, per-device integrity/trust/revocation, quarantined intake, independent signing, no sample upload by default, no automatic cross-client bans. Test compromised signed device/forged consensus/tenant breakout and trusted recovery.
- Root/kernel compromise is **outside the guarantees of on-host self-detection**; require degraded trust and external recovery path. No user or production device is to be scanned/isolated under this design task.

## Initial epic backlog

1. Policy JSON Schema and network object model.
2. Kernel netns test rig and non-destructive compiler.
3. Privileged executor IPC and watchdog contract.
4. Standalone Host/Edge setup, local+remote agent transport, offline retry and independent rollback tests; **no NetBird connector work**.
5. WAF connector selection: Coraza/OWASP CRS feasibility on chosen proxy.
6. Luna evidence canonical schema, independent label/replay benchmark and data governance.
7. WebUI design system with moon-state and telemetry-unknown semantics.
8. Install/upgrade/uninstall cross-architecture packaging.
9. CI with threat-model checks, license/SBOM and dependency policies.
10. Formal owner review before real enforcement or infrastructure changes.
11. **Post-G-STANDALONE backlog only:** NetBird version/API read-only discovery, scoped permissions and noninterference tests; start only on owner approval.
