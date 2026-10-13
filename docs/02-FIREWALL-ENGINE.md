# 02 — Traditional Firewall Engine: Rules, Aliases, NAT

## First-class functions

- **Interfaces and zones:** wan, lan, dmz, netbird, container, loopback, custom VLAN; discover read-only, explicit assignment, zone precedence and interface renaming safeguards.
- **Stateful Rules:** direction input/output/forward; src/dst IP4/IP6, L4 protocol/ports, interface, zone, conntrack state, owner/service tag, schedule, action (`accept`, `drop`, `reject`, `log`, `rate-limit`), counters, priorities and comments.
- **Aliases:** host/net aliases, nested CIDRs, IP4/IP6 sets, port sets, FQDN references with resolver TTL and DNS failure policy, static and signed dynamic threat feeds; graph cycle prevention and maximum expansion sizes.
- **NAT:** SNAT, DNAT/port forwards, masquerade, hairpin only when routing; IPv4 first, documented IPv6 constraints; warning for double-NAT/NetBird route masquerade.
- **Dynamic actions:** temporary bans scoped to interface/zone/service, timed nft sets, unban, renew/shorten, exceptions, incident ID, audit receipts.
- **Policies:** default action, rule ordering, object ownership and schema migrations, staged diff, rollback snapshots, exported/imported signed policy bundles.

## Stable rule identity

Assign opaque UUIDs, revision hashes and origin (`owner`, `luna`, `feed`, `integration`) to every rule; never identify one by mutable numeric nft handle alone. Generate comments/ownership tags. A user rule must not be silently mutated by Luna or a cloud feed.

## Example proposed policy (illustrative; NOT compiled or applied)

```yaml
apiVersion: lightluna.io/v1alpha1
kind: PolicyBundle
metadata:
  name: host-baseline
  revision: 1
spec:
  mode: host
  defaultInbound: drop
  emergencyAccess:
    enabled: true
    networks: ["192.0.2.0/24"] # documentation test network ONLY
  aliases:
    - id: trusted-admin
      type: cidr-set
      values: ["192.0.2.14/32"]
    - id: protected-services
      type: port-set
      values: [22, 443]
  rules:
    - id: 6cdac70a-94b8-465e-856c-ddac843491b2
      priority: 100
      direction: input
      source: trusted-admin
      protocol: tcp
      destinationPort: 22
      action: accept
      note: "Example only; real rescue access must be independently verified"
  autonomous:
    mode: shadow
    allowRealBans: false
    maxBanTtlSeconds: 300
```

## Compiler safety contract

1. Enumerate current kernel rules, firewall owners, Docker/NetBird/UFW/firewalld state (read-only) and classify conflicts.
2. Parse + strict validation, canonicalize IP/CIDR/protocol syntax and forbid ambiguous aliases or unbounded expansion.
3. Produce nft-ready plan, **do not run it** in source compiler; check kernel compatibility in disposable namespace; emit human-readable diff and predicted flow simulation.
4. Before apply, arm independent out-of-process rollback lease and execute canary connectivity tests for both public service and authorized admin path.
5. Only after authorization, privileged executor applies atomic changes to **Light Luna-owned** tables/sets and records their kernel state. Never `flush ruleset` or erase vendor rules.
6. Verify live effect using independent source and audit receipts; if health checks fail, restore prior known-good revision regardless of Luna process health.
7. Explicit resolve of collisions: multiple base chains/hooks may evaluate independently. Rule order and `accept` in one chain does **not** guarantee acceptance if a later base chain drops. Mixed `iptables-nft`, native nftables, Docker and NetBird require coexistence tests.
8. Uninstall cleans **only** verified owned objects. If identity mismatch or drift, warn and leave unknown rules intact.

## Mandatory lockout protections

- Maintain separate rescue/management control interface with explicit origin and lifetimes, tested independently.
- Ensure DNS, NTP, update control, cloud API and NetBird Management/Signal/Relay endpoints are considered by policy; never blindly bypass all security for `wt0`.
- Reserve local break-glass console and an external rollback timer (e.g., privileged helper or watchdog) that does not depend on the WebUI, Luna, control-plane network or new firewall rules.
- Do not let LLM submit raw nft commands. Typed executor API: `validate`, `stage`, `commit`, `rollback`, `addTimedBlock`, `removeTimedBlock`, `status`, `getAudit`.
- Enforce hard limits per time interval and source cardinality to prevent a compromised brain spamming kernel state.

## Conflict policy

Initially deploy in **coexist/observe** mode. Managed takeover of existing UFW/firewalld policies is a distinct project phase with automatic backup and reversible migration. `NB_NFTABLES_TABLE` may identify NetBird-managed table; never assume a fixed table name without inspecting the installed client. NetBird `--disable-firewall` is not an acceptable normal integration method because it disables its own peer-side ACL enforcement.

## Testing

Use ephemeral Linux network namespaces, simulated `wt0`, veth pairs, Docker networking, inbound/outbound/forward IPv4+IPv6, connection reuse, port forwarding, DNS alias expiry, netlink restart, conflicting native chains, auto-ban TTL, crash/reboot/uninstall, lost NetBird and emergency rescue. CI can run unprivileged syntax/unit tests; privileged network integration tests run on isolated runners only.
