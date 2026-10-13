# ADR-021 — eBPF Observability and Optional XDP Early Mitigation

Status: **DECIDED architecture direction; NOT IMPLEMENTED**
Date: 2026-10-13
Scope: Linux packet visibility, host/process telemetry and future L3/L4 mitigation
Supersedes: none
Primary sources: [Architecture](../01-ARCHITECTURE.md), [Firewall Engine](../02-FIREWALL-ENGINE.md), [Security](../07-SECURITY-SAFETY.md), [Roadmap](../11-ROADMAP.md), [Go/Python decision](ADR-020-go-core-python-optional.md), [Code Map](../project-book/LIGHT_LUNA_CODE_MAP.md)

## Decision

**nftables/conntrack remain Light Luna's primary, general-purpose Linux firewall and policy authority.** Add optional eBPF read-only observability where kernel, permissions and instrumented hook support are proven. Later consider XDP as a separately authorized, optional early ingress L3/L4 DDoS mitigation fast path for supported adapters, NEVER a blanket replacement for nftables, WAF, or Luna's independent policy gate.

This decision does not make XDP mandatory for the standalone Host/Edge product, does not authorize kernel program loading/attachment on production or HOMESV, and does not change the existing P0–P4 standalone-first/G-STANDALONE-before-NetBird requirements.

## Responsibilities and boundaries

| Layer | Desired purpose | Status |
|---|---|---|
| Go collectors + deterministic Luna Brain | Normalize, aggregate, correlate telemetry; produce bounded proposals | **Future**, not shipped |
| eBPF tracing (network/socket/process/selected syscall/LSM hooks) | Optional low-overhead host and connection observability with sampling, cardinality, privacy and source quality controls | **Future, observe-only initial mode** |
| XDP ingress | Early allow/pass/drop or bounded filter of verified L3/L4 attack signatures BEFORE much of the Linux networking stack | **Future, opt-in only** |
| nftables/conntrack | Stateful firewall, aliases, scoped timed blocks, NAT/gateway rules when topology permits | **Primary backend target; real enforcement not yet implemented** |
| L7 proxy/WAF | HTTP/TLS-terminated inspection at a compatible proxy | **Future, separate capability** |
| Go Policy Gate + separately privileged executor | Authenticate, authorize, audit, stage, verify and roll back owned actions | **Future, never bypassed by eBPF/AI** |

XDP is primarily an **ingress packet** hook. It generally runs before conntrack; an XDP rule cannot assume conntrack state or complete HTTP/TLS semantics. Tracing attachment and availability vary by BPF program type, hook, kernel, distro and security policy.

## Implementation technology boundary

- **Go remains primary userspace** for collectors, program lifecycle coordination, decision proposal handling, capability probing and security policy. Review `github.com/cilium/ebpf` / `bpf2go` as a candidate, not an approved package/dependency.
- **Restricted eBPF C + Clang/LLVM** is a practical future option to produce small verified BPF objects for kernel execution. Adding BPF-targeted C does **not** change Go as the main Backend/Brain language, and **no clang dependency is required at customer runtime** if verified objects can be packaged.
- CO-RE/BTF is a portability candidate, not universal support. Require feature probes and an explicit kernel/arch/driver/NIC matrix, including IPv4/IPv6, fragments, VLAN/encapsulation, GRO/LRO/MTU considerations and virtualization.
- Optional Python ML worker remains unprivileged and **off packet hot path**; it never loads/attaches eBPF, mutates maps for enforcement or connects directly to a privileged executor.

## Ordered implementation gates

1. **Design and read-only inventory (safe early task):** define BPF feature/support matrix and telemetry schemas; inspect *read-only* kernel/driver capabilities and existing attachments; never attach/detach/replace system-owned programs. Missing privileges/support means `unsupported` or `unknown`, not silent auto-enable.
2. **Isolated observer prototype:** compile a minimal **XDP_PASS-only** packet counter and/or bounded tracing program, use synthetic traffic on disposable veth/netns VM, benchmark overhead and verify clean detach. Prefer non-XDP tracing for general host observability.
3. **Observe-only integration:** Go collector receives bounded counters/events from BPF maps/ring buffers. No packet drops. Audit sampling loss, queue pressure, provenance, clock skew, CPU/RAM and privacy.
4. **DDoS shadow classifier:** independent labeled benign/attack replay, explicit match scope and version, explainable false-positive measurements, interface/driver feature tests; compare with nftables-only baseline.
5. **Explicit opt-in XDP enforcement after safety gates:** separate capability and owner permission; Go Policy Gate approves signed/typed versioned policy with scope, TTL, protected management traffic, rate cap and ban budgets; independent watchdog + external reachability/rollback. Never silently turn observe-only programs into a dropping fast path.

XDP is **optional** and these tasks must not block the core firewall's standalone acceptance. Adding XDP to an operational NIC is a high-impact network mutation, forbidden before an isolated recovery proof and explicit owner go-ahead.

## Existing XDP program coexistence (critical)

One network interface can already be managed by another application's XDP program or dispatcher. **Fail closed on ownership conflicts**: do not replace/detach/override any existing XDP program, BPF link, pinned map, `libxdp` dispatcher or unknown attach automatically. Multi-program chaining needs compatible kernel/dispatcher/protocol (e.g. libxdp); if incompatible, report unsupported and leave existing attachments intact. Pinning and link lifetimes require explicit uninstall/reboot/crash tests. Maintain a manifest of **Light Luna-owned** program IDs, owner identity, map pins, interface identity, revision and cleanup scope. Never claim `XDP_PASS` negates other XDP drops.

## Safe failure and rollback contract

- Dedicated BPF loader/executor capability isolated from Go Brain, Python worker, and unprivileged UI. Kernel verifier passing is **not** a proof that packet filtering semantics are safe or that false positives are absent.
- Program size/verifier cost, bounded parsing, tail calls/loops, map cardinality, per-CPU statistics, event export budgets, verifier failure, map-full, CPU pressure and offline behavior must be tested.
- A **coordinated** XDP + nftables policy plan must preserve permitted management/health traffic and document cross-layer ordering; no duplicate action or hidden global IP ban.
- Loader crash, model crash, daemon restart, network interface rename, NIC reset, package upgrade and reboot must preserve appropriate known-good state and restore admin connectivity in tests.
- When driver native mode is unsupported, do not claim same throughput from generic SKB fallback. Do not silently enable `xdpgeneric` on an operational interface; operator must accept reduced capability or choose nftables-only.
- XDP cannot fix a saturated upstream link if packets never reach the host. DDoS defense might require upstream/cloud mitigation, physical capacity and rate controls.
- No Cloud dependency, no auto-ban default, no NetBird API probing/integration during P0–P4.

## Benchmarks and acceptance

Publish measured PPS, CPU, p50/p99 service latency, packet drops, zero/negative false positive confidence bounds (as appropriate), dropped event telemetry, memory/map budget, driver/kernel/arch/MTU, IPv4/IPv6 equivalence and disruption/rollback times on representative native/generic/virtual NIC environments. Compare against nftables-only, no security protection speed claims from vendor benchmarks.

## Research sources (external; examples are not Light Luna performance)

- [Cloudflare L4Drop (2018)](https://blog.cloudflare.com/l4drop-xdp-ebpf-based-ddos-mitigations/) — recorded a >8M packets/second drop case on one Cloudflare server; **not** an expected Light Luna benchmark.
- [Cloudflare 2024 DDoS mitigation architecture](https://blog.cloudflare.com/how-cloudflare-auto-mitigated-world-record-3-8-tbps-ddos-attack/) — XDP/l4drop with packet sample + separate analysis.
- [Kernel eBPF userspace API](https://www.kernel.org/doc/html/v6.9/userspace-api/ebpf/index.html) — supported kernel extension/security model.
- [Linux AF_XDP documentation](https://www.kernel.org/doc/html/latest/networking/af_xdp.html) — native vs generic mode context.
- [cilium/ebpf Go Getting Started](https://github.com/cilium/ebpf/blob/main/docs/ebpf/guides/getting-started.md) and [BTF](https://github.com/cilium/ebpf/blob/main/docs/ebpf/btf.md) — Go + C/Clang generation candidate.
- [xdp-tools libxdp](https://github.com/xdp-project/xdp-tools/blob/main/lib/libxdp/README.org) — existing program chaining, attach ownership and kernel support limitations.

## Verification/status

**Documentation decision only.** No Light Luna BPF source, BPF maps, XDP links, kernel loader, per-NIC eligibility test, packet-drop data or rollback proof exists at the time of this ADR. Existing kernel XDP programs on any machine are third-party state and must not be re-used as Light Luna proof.
