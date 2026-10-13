# 11 — Delivery Roadmap, Test Gates and Release Discipline

**Current milestone:** P0 **non-enforcing CLI/policy validation scaffold** exists on the development branch; no implemented kernel firewall or shipping release. See [Code Map](project-book/LIGHT_LUNA_CODE_MAP.md).

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
