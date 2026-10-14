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
