# 11 — Delivery Roadmap, Test Gates and Release Discipline

**Current milestone:** P-1 architecture/documentation only. **No implemented firewall or shipping release.**

| Milestone | Deliverables | Exit gates |
|---|---|---|
| **P-1 — Product contract** | Docs, ADRs, threat model, ownership, compatibility assumptions, test plans | Architecture review; no mutation of production |
| **P0 — Local Firewall Foundation** | Go agent + typed executor, nftables compiler, aliases/rules/zones, local WebUI + CLI, API schemas, simulated Host/Edge chooser | Tests on isolated Linux netns; clean install/upgrade/uninstall; unknown firewall conflicts blocked |
| **P1 — Host GA candidate** | Stateful filtering, controlled timed bans, policy diff/simulation, auth/audit/rollback, IPv6 and Docker/NetBird coexistence | Admin lockout negative tests pass; external rescue works without Luna; signed packages amd64/arm64 |
| **P2 — WAF & Luna Shadow** | Proxy/WAF connector, OWASP CRS observe mode, replay corpus, threat-feed/SBOM engine, Luna recommendations only | Labeled FP measurements; no automatic enforcement; documented feature coverage |
| **P3 — Controlled Autonomy** | Approved scoped TTL-capped ban/rate/challenge/unban, outcome verification, site-specific policy, rollback | Canary field tests + owner approval; no unauthorized action in red-team exercises |
| **P4 — Edge Fleet & NetBird** | Read-only NetBird integration then gated writes, secure remote agents, multi-device WebUI, offline queues | IAM/scope, route/IP attribution, dropped network, multi-node regression |
| **P5 — Gateway & Hardware Reference** | NAT/forward/VLAN, inline IDS/IPS option, appliances, thermal/performance, recovery console | 1/2.5/10 Gbps measured device-specific throughput and packet loss; power loss recovery |
| **P6 — Cloud & OEM** | Signed intelligence, fleet multi-tenant, cloud privacy and licensing, OEM image only with tested hardware | Multi-tenant penetration testing, backup/restore, scale/failure model, vendor/licensing review |

**Order can be revisited**: NetBird read-only connector may start early for development, but is never a blocker for local firewall. WAF inspection throughput and gateway appliances do not gate Host MVP.

## Gate-specific test suites

- **Network correctness:** packet/flow fixtures, ICMP/TCP/UDP, IPv6, conntrack, interface/zone changes, DNS alias expiry, NAT hairpin where supported.
- **Coexistence:** nftables hook priority, NetBird `wt0` and Access Policies, Docker bridge/port publishing, UFW/firewalld, containerized proxy, cloud security groups.
- **Safety:** worst-case malicious rule, lost SSH, stale netbird client, model crash, CPU overload, disk full, interrupted upgrade, rollback path and expiry after reboot.
- **WAF:** OWASP CRS coverage, false positives on documented normal traffic, TLS termination, HTTP/2, streaming, WebSocket, large bodies and NetBird Proxy version probing.
- **Trust:** forged forwarded headers, NAT masking, VPN CGNAT address overlap, replayed signed command, compromised feed and adversarial log prompt injection.
- **Release quality:** multiarch native package, reproducible build, SBOM, provenance, signed updater, E2E setup wizard and independent agent health.

## Decision KPIs (target definitions, no invented results)

False-positive rate against independent labeled benign traffic, recall against vetted attack corpus, admin recovery time in controlled failures, packet path latency, throughput and drop rate by hardware mode, peak memory under backpressure, policy apply/rollback success, number of unverified IP attributions used in auto-ban (target zero), drift/anomaly explanatory quality.

## Initial epic backlog

1. Policy JSON Schema and network object model.
2. Kernel netns test rig and non-destructive compiler.
3. Privileged executor IPC and watchdog contract.
4. NetBird read-only capability discovery and non-interference test matrix.
5. WAF connector selection: Coraza/OWASP CRS feasibility on chosen proxy.
6. Luna evidence canonical schema, independent label/replay benchmark and data governance.
7. WebUI design system with moon-state and telemetry-unknown semantics.
8. Install/upgrade/uninstall cross-architecture packaging.
9. CI with threat-model checks, license/SBOM and dependency policies.
10. Formal owner review before real enforcement or infrastructure changes.
