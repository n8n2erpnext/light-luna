# 07 — Security Threat Model, Enforcement Safety, Recovery

## Assets and attackers

Protected assets: admin reachability, legitimate traffic, node identity/private keys, firewall policy, audit provenance, CVE data trust and tenant isolation. Attackers may control network packets, HTTP headers, logs, DNS names, threat feed entries, Cloud bearer tokens if compromised, or an untrusted plugin input. Internal failures include corrupted DB, hung AI process, replayed executor requests, bad rules or stale device connectivity.

## Trust boundaries

- Untrusted ingress → sensors: parse defensively, cap bytes/cardinality, normalize time, enforce stable source identity.
- Sensors and Internet feeds → Luna: **data, never instructions**. No direct system shell, no dynamic web command.
- Luna → Policy Gate: strictly typed proposal; no arbitrary command string.
- Policy Gate → privileged executor: Unix peer credentials/mTLS, signed or authenticated request, anti-replay, low TTL and scoped capabilities.
- WebUI → control API: RBAC, CSRF/session safety, MFA for risky actions, protected local emergency recovery.
- Cloud/NetBird → local agents: device identity and policy validation independently of network tunnel membership.

## Hard invariants

1. AI cannot modify owner allowlists, emergency recovery network, bootstrap keys or its own execution authority.
2. No blanket auto-ban of CDN/reverse-proxy IPs, private/control-plane paths or known admin endpoints without explicit emergency authorization.
3. Every ban has scope, evidence, author, target, TTL and idempotency key. Expiry and unban are deterministic and independent of Luna.
4. No global destructive nft command (`flush ruleset`), no uninstall of NetBird or forced disable of ACL backend to solve conflicts.
5. Rule changes require dry-run compiler, diff, kernel capability check, independently armed rollback watchdog, and post-apply validation.
6. Keep local console recovery even when Cloud, NetBird, WebUI, WAN and AI are unreachable. Controlled fail-open/fail-closed strategy must be explicitly selected per protected service and threat model, not implied globally.
7. Audit includes rejected actions, expired rules and rule version changes; logs are redacted, disk budgeted and tamper-evident where necessary.
8. A WAF/IDS detection is not by itself ground truth; exceptions, manual review and evidence-based confidence are required.

## Critical misuse cases to test

- Forged `X-Forwarded-For`, NAT/PROXY protocol mismatch, compromised upstream proxy.
- Noisy public IP NAT shared by many innocent users; false global ban.
- Malicious uploaded threat feed causing unbounded aliases/lockout.
- Privileged executor command injection or path traversal; remote API impersonation.
- NetBird peer re-enrollment, changed tunnel IP, failed TLS/relay; `wt0` chain ordering conflict.
- Luna hallucinating CVE risk from version alone or learning from false labels.
- Power outage mid-rule change; reboot while rollback timer active.
- Disk full and DB lock causing unusable UI; negative decisions should never create allow rules.

## eBPF/XDP additional safety boundary (future only)

[ADR-021](adr/ADR-021-ebpf-observability-xdp-optional.md) requires that any eBPF tracing/XDP loader be isolated from Luna Brain and user input. Read-only capability inspection is allowed in development, but attaching/detaching/replacing a BPF program is a kernel operation subject to explicit owner-approved isolated test gates. Existing kernel programs and libxdp dispatchers are **not** Light Luna-owned; never replace or detach them automatically. For future XDP drop, require bounded parser, traffic source attribution, independent authorization, protected management flows, explicit NIC/driver/kernel mode compatibility, resource cap, telemetry-loss handling, external rollback and verification of no unintended packet loss. eBPF verifier approval is not equivalent to semantic/firewall safety. Unsupported native XDP is not silently converted to generic mode. Keep nftables-only fallback and no upstream anti-saturation promises.

## Default authorization states

`OFF` / `OBSERVE` / `PROPOSE` / `CONFIRM_EACH` / `SCOPED_AUTO` (final only after all release gates). `SCOPED_AUTO` carries maximum TTL, protected service/target exceptions, rate caps and revocation switch. Default is `OBSERVE`, with all network ban actions disabled in design baseline.
