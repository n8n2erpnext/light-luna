# 04 — WAF, IDS/IPS & Threat Intelligence

## Three protection layers

| Layer | Responsibility | Planned backend |
|---|---|---|
| L3/L4 | Stateful ACL, per-IP/port rate and timed bans, NAT | nftables/conntrack |
| L7 | HTTP inspection, bot behavior, virtual hosts, request anomaly scoring | WAF adapter, candidate Coraza + OWASP CRS / reverse proxy integration |
| Network IDS/IPS | Signatures/protocol anomaly on visible traffic | Optional Suricata with explicit SPAN/TAP/inline architecture |

An L4 firewall cannot match encrypted HTTP hostname/path without TLS termination; a WAF only sees HTTP traffic at the termination point or other explicitly decrypted trusted location. A passive one-NIC Edge controller cannot inspect arbitrary switched LAN packets without mirror/feed.

## WAF MVP

- HTTP request/response metadata, policy by site/route/method, OWASP CRS signature set, anomaly score and exclusion controls.
- Modes `off` / `observe` / `enforce`; start observe, measure false positives on real workloads, signed rule pack versioning and rollback.
- Rules for known malicious paths, command-injection/SQL injection detection in HTTP traffic (not blanket guarantee), request limits, header/body constraints, bot/rate challenge where connector supports it, WAF-only site-scoped temporary blocks.
- Structured evidence: normalized URI (sensitive fields redacted), reason, rule ID, matched zone, trusted originating IP, reverse-proxy chain, policy version, action and response status.
- **Standalone priority:** generic HTTP termination integration for Nginx/Coraza or other vendor-neutral proxy/WAF deployment. Generic Traefik integration can be evaluated without NetBird. NetBird Reverse Proxy telemetry, vendor extension points and capability probes are **post-G-STANDALONE only**. **Do not claim native NetBird proxy WAF APIs until proven.**
- WAF protocol limitations: WebSocket upgrade, streaming, large request bodies, HTTP/2 and HTTP/3/QUIC need specific test cases. Never silently block unsupported protocols when merely observing.

## Future NetBird Proxy boundary (parked, no implementation in initial development)

**No NetBird WAF/proxy development, API probing or integration test until standalone Gate G-STANDALONE is accepted and owner approval granted.** This section is technical research for that later phase. NetBird Reverse Proxy (beta) supports HTTP/L4 exposures, authentication, IP/country restriction and CrowdSec IP reputation; it does not imply a complete OWASP-CRS WAF control surface. When this integration phase eventually opens, evaluate read-only audit/proxy events first; add a verified supported insertion point or sidecar WAF on termination path before promising inline WAF enforcement. Respect CrowdSec `off/observe/enforce` service modes and bouncer cache semantics; avoid duplicate conflicting ban sources.

## CVE engine design

1. Collect *read-only* OS package inventory and application dependency manifests/SBOM (`CycloneDX`/`SPDX` proposed) and map them to the **actual deployed image/build** with hash/digest and collection time.
2. Normalize identities: CPE/PURL, distribution/vendor package+revision, runtime dependencies; map effective service exposure.
3. Scheduled bounded imports of CISA KEV, NVD, EPSS/vendor advisories; verify source signatures where offered, ETag and pagination/full coverage, cache freshness and importer errors.
4. Match affected version ranges with distro vendor backport advisories and deployment reachability; distinguish `POTENTIALLY_AFFECTED`, `CONFIRMED_APPLICABLE`, `NOT_APPLICABLE`, `UNKNOWN`.
5. Recommendation: upgrade, isolate, WAF mitigation or compensating rule **only when independently validated**, owner-approved, with rollback. **Automatic patching is not enabled by default**.
6. Signed threat-feed updates are staged/tested in shadow; never import external text as instructions or run remote scripts.

## Network baseline and device health

Packet/connection counters, deny rate, scan distribution over disjoint windows, authentication errors, resource pressure, unusual egress and TLS/cert anomalies can all be signals. Avoid attributing packet drops to threat without service impact/expected traffic baselines. Multi-source corroboration is required for severe action.

## Community behavior intelligence and collective immune-style memory (ADR-027)

[ADR-027](adr/ADR-027-collective-threat-immunity.md) extends threat intelligence beyond IP reputation to **behavioral scenarios, independently verified actor indicators, time-window signatures, counterexamples, CVE/technique context and local mitigation outcomes**. The [full design](architecture/federated-immune-threat-intelligence.md) follows CrowdSec-inspired log/parser/scenario/decision separation and vetted signed distribution. A device's locally blocked source is **not** confirmed maliciousness and must not cause a global ban. Opt-in privacy-scoped reports are quarantined/reviewed in Cloud and distributed as versioned watch/curated intelligence, locally evaluated by Go Brain; any enforcement remains per-site Policy Gate and owner policy. Cloud reputation is never sufficient by itself to block protected admin/CDN or an entire subnet. This is research/design only, not a running feature.

## Operational guardrails

False positive analyst review, owner allowlist, CDN/healthcheck scanner classification, proxy trust chain verification, per-site vs global enforcement scope, human-accessible unblock function, threshold and per-minute command budget, short default TTL, evidence replay test, local privacy/retention and audit tamper detection.
