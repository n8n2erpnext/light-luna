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
