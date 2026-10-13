# 08 — Packaging, OS & Hardware Compatibility

## Product is installable software first

Deliver signed/verifiable native Linux packages and an operator-friendly installer, independent of an appliance purchase. Target first: Ubuntu 24.04 LTS / Debian stable on `amd64` and `arm64`. Other distributions after kernel, netfilter and systemd capability testing. No assumption that one package works with every NAS, OpenWrt or proprietary router OS.

### Package component split (planned)

- `light-luna-agent`: unprivileged data collectors, local control plane and secure communication.
- `light-luna-executor`: constrained privileged `lunalockd`, kernel rule owner, watchdog and rollback lease.
- `light-luna-brain`: optional Luna Brain and local inference backend; separate resource profile.
- `light-luna-waf`: optional adapter/connector, not enabled without L7 ingress configuration.
- `light-luna-ui`: loopback/private-bound static WebUI with auth.
- `lunactl`: typed local and remote CLI.

Ship as `.deb` first with systemd service presets, checksums, SBOM and versioned schema. Support `install`, `upgrade`, `repair`, `export-policy`, `restore`, `uninstall` with clear ownership of changes. Never use `curl | sudo bash` as sole supported trusted install path; provide verified binary and installation instructions with signature checks.

## Hardware baseline profiles (tentative until benchmarks)

| Profile | Candidate | Assumptions |
|---|---|---|
| Host Lite | 2 vCPU, 2-4 GiB RAM | nftables, counters, log-based detection; AI optional/remote/off |
| Host Smart | 4 cores, 8 GiB RAM | local lightweight Luna + regular telemetry |
| Edge Observer | x86-64/ARM64 mini PC, 4-8 cores, 8-16 GiB | multi-node logs, optional sampled IDS; not full line-rate IPS guarantee |
| Inline Gateway/Box | NICs/VLANs tested, hardware offload optional | throughput, packet loss and thermal profile hardware-specific |

The values above are **planning hypotheses, not verified product requirements**. Resource budgets should be published with measured 1/2.5/10 Gbps traffic profiles, TLS/WAF inspection options, active conntrack count, Suricata ruleset size and local inference overhead.

## Appliance/OEM readiness

User-selectable modes are available in software; future reference hardware uses a compatibility certification matrix rather than a fork: NIC driver/kernel, PCIe, IOMMU/offloads, VLAN, bypass/watchdog, eMMC/NVMe wear, thermal, x86-64/ARM64, secure boot and physical console. Gateway placement behind an upstream router, as primary gateway, single-arm/VLAN, and out-of-band Edge each need topology templates and fault-injection tests.

## Update & recovery

- Signed artifacts and release manifest pinned by digest; staged/canary upgrades, watchdog and post-update health gates.
- Maintain `last_known_good_policy`, export/import, schema down-migrations where safe, transactional DB upgrades and automatic rollback on failed service health.
- Offline rescue bundle with console `lunactl rescue` and independent executor; apply only owned rule cleanup.
- Privilege and capabilities minimized, Linux service sandboxing, immutable executable filesystem, read-only packaged assets and jailed model/feedback storage.
- Packaging must not silently stop Docker, NetBird, UFW or fail2ban/CrowdSec; preflight detects conflicts and shows user an explicit plan.
