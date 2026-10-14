# P0 Synthetic Policy Flow Simulator — Implementation and Verification

Status: **IMPLEMENTED IN DEVELOPMENT SOURCE — policy intent simulation only; NOT kernel packet verification or firewall enforcement**
Date: 2026-10-14
Scope: `PolicyBundle` Host/Edge P0 TCP/UDP/ICMP/ICMPv6 matching, bounded synthetic flows, strict input, potential owner-management lockout flag
Primary sources: [Flow simulator](../../internal/flowsim/simulate.go), [Go tests](../../internal/flowsim/simulate_test.go), [CLI](../../cmd/lunactl/main.go), [Policy model](../../internal/policy/model.go), [Validation](../../internal/policy/validate.go), [synthetic fixture](../../testdata/flow-fixtures.synthetic.json)
Related: [Unhooked nft compiler](p0-nft-unhooked-compiler.md), [Security](../07-SECURITY-SAFETY.md), [Roadmap](../11-ROADMAP.md)
Git: verify current development branch SHA; date and test conditions below are source-level, NOT deployed product validation.

## CLI and behavior

```sh
go run ./cmd/lunactl firewall policies simulate \
  testdata/host-baseline.json \
  testdata/flow-fixtures.synthetic.json
```

Call chain: `cmd/lunactl.run` → `internal/policy.Decode` (strict 1MiB bounded JSON, known fields, semantic policy validation) → `internal/strictjson.Decode` for 256KiB bounded synthetic flow JSON → `internal/flowsim.Fixture.Validate` → `internal/flowsim.Simulate` → inert JSON output.

No subprocesses, nft command invocation, firewall execution, network connection, process inspection or kernel writes are present in the simulator. `kernelWrites=false` and **`effectiveVerdict=UNKNOWN` for every case**: a policy match is **not** the effective Linux packet outcome.

## Exactly what is modeled

| Item | Implemented P0 behavior | Important limitation |
|---|---|---|
| Rule order | Policy `priority` ascending for `input` and `output` | Not nftables base-chain hook priority |
| Source IP | Canonical `netip.Addr`, match literal CIDR / `cidr-set` alias / `any`; IPv4 and IPv6 | No proxy forwarding, local interface, NAT or real flow path |
| L4 | Match `tcp`, `udp`, `icmp` (IPv4) and `icmpv6` (IPv6), optional TCP/UDP destination port | No source ports, TCP flags, fragment processing, ICMP types, conntrack handling |
| Actions | `log` records rule ID and continues; first `accept`, `drop` or `reject` terminates local policy intent | Other later nftables chains can still drop accepted packets |
| Defaults | No terminal match on input → `defaultInbound`; no terminal output match → `UNSPECIFIED` | No `defaultOutbound` in current schema, so do not invent one |
| Management | `managementCritical=true` + policy `drop/reject` → `POTENTIAL_LOCKOUT`; else `UNVERIFIED_REACHABILITY` | Real remote SSH/Cloud recovery is **never** proven from declared CIDR |
| Connection | `new/established/unknown` allowed fixture values | Stateful handling is not implemented; warns for non-new flows |
| Coexistence | Always mark `externalFirewall=UNKNOWN` | Does not enumerate Docker/UFW/NetBird/iptables or nftables base chains |

**Intent** can be `accept`, `drop`, `reject`, or `UNSPECIFIED`, but **not** an actual firewall verdict. `POTENTIAL_LOCKOUT` means the intended policy is likely to deny a supplied synthetic management-critical path, **not** verified host lockout. If `UNVERIFIED_REACHABILITY`, do not claim that an admin session survives; it may be blocked by other layers.

Fixture limits: 1–128 cases, ≤256KiB encoded JSON, unique bounded ASCII IDs, source IP must be valid non-multicast, non-unspecified unscoped address; reject IPv4-mapped IPv6, invalid direction/protocol, incompatible ICMP family, nonsensical ports and connection state. The IPs in public fixtures are **IETF documentation addresses**; not real devices.

## Verification recorded

- HOMESV Go unit tests passed, including admin IPv4/IPv6 matching, alias match, `log` non-terminal semantics, rule priority, TCP/UDP, default input drop, explicit output unknown, management risk and deterministic JSON.
- Negative tests passed for invalid source/direction, IPv4-mapped IPv6, multicast, wrong ICMP family, inconsistent ports, case duplication, excessive cases, unsupported schema.
- `go vet ./...` and CLI against `testdata/host-baseline.json` + `testdata/flow-fixtures.synthetic.json` passed. Cross-platform Go build validation must still be run and recorded for this specific commit.
- **No nft -c or namespace test**: HOMESV cannot create permitted isolated `unshare --user --map-root-user --net` namespace; no privileged tests were attempted against a live host. The standalone G-STANDALONE gate is still blocked.

## Gaps before privileged execution

1. Supply isolated disposable VM or other authorized independent Linux network namespace, prove nft syntax and no host chain interference with known kernel versions.
2. Add read-only discovery of existing firewall owners and actual interface/NAT/conntrack state with explicit privileges; until then retain `UNKNOWN`.
3. Model real rescue route constraints with independent remote canary/rollback watchdog; synthetic `managementCritical` is not a protected recovery mechanism.
4. Separate and authenticate Agent → policy compiler → restricted privileged executor, with transaction staging, deterministic rollback and lockout fail-safe **before any kernel writes**.
5. Only later connect Brain/Guardian/Cloud proposals through independent Policy Gate. No NetBird integration before G-STANDALONE and explicit owner consent.

This document is evidence of **synthetic deterministic policy evaluation**, not evidence of real packet filtering or host infection detection.
