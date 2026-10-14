# P0 nftables Compiler — Unhooked Preview Only

Status: **source-implemented, unit-tested, NOT kernel-verified and NOT an installable firewall**
Date: 2026-10-14
Scope: deterministic, bounded, non-privileged conversion from existing `PolicyBundle` to **unhooked regular-chain text**.
Primary sources: [compiler](../../internal/nftpreview/compile.go), [tests](../../internal/nftpreview/compile_test.go), [CLI](../../cmd/lunactl/main.go), [policy](../../internal/policy/model.go), [security](../07-SECURITY-SAFETY.md), [P0 notes](../15-P0-IMPLEMENTATION-NOTES.md).
External reference: [nftables configuring base and regular chains](https://wiki.nftables.org/wiki-nftables/index.php/Configuring_chains).

## What now runs

```sh
go run ./cmd/lunactl firewall policies nft-preview testdata/host-baseline.json
```

Execution path: `cmd/lunactl` opens bounded local input → `internal/policy.Decode` (strict keys + semantic validation) → `internal/nftpreview.Compile` → JSON `Plan` to stdout. This has **no nft executable invocation**, no local IPC, no socket, no process attach, no kernel changes and no environment capability checks. The result reports `kernelWrites=false`, `syntaxChecked=false`, `kernelCompatibilityChecked=false`, and `UNHOOKED_TEXT_PREVIEW_ONLY`.

The emitted textual script uses an isolated name `table inet luna_p0_preview` and two **regular** chains (`preview_inbound` and `preview_outbound`). There is **no `type filter hook input/output`, no `priority`, no `policy drop`, no `jump`/ `goto`, no dynamic shell content**. Under nftables semantics, regular chains see no packets unless invoked by a separate base chain. They are provided strictly as a syntax candidate for later validation, **not as an applicable policy**. Even if the text were manually imported it must never be treated as a Light Luna production firewall.

## Supported P0 subset

- Input and output only, with stable ascending rule `priority` (the policy-level priority determines text order, not a kernel hook priority).
- `tcp`/`udp` optional destination port, `icmp` (IPv4 only) and `icmpv6` (IPv6 only). Unsupported address/protocol family combinations fail rather than generating ambiguous nft text.
- `ip saddr` and `ip6 saddr` for strict canonical CIDRs; `any` source and validated `cidr-set` aliases (sorted and expanded to individual rules).
- `accept`, `drop`, `reject`, `log` are **preview tokens only**; comment stores the validated rule UUID. User-provided note text is never interpolated.
- Upper bounds: 8,192 expanded statements and 1 MiB output; 1 MiB PolicyBundle input and strict JSON are enforced earlier.

**Deliberately unimplemented:** chain hooks and kernel hook priorities, default inbound decision, emergencyAccess implementation/verification, state tracking/conntrack, loopback or established exceptions, effective Docker/UFW coexistence, real interface attribution, IPv6 ICMP/ND correctness in a live network, dynamic feeds, port-set references, simulation with actual traffic, timed bans, backend executor, stage/commit/rollback and external management rescue. No `nft -c` syntax validity or Linux kernel compatibility is currently proven.

## Verification 2026-10-14

- `GOTOOLCHAIN=local GOPROXY=off CGO_ENABLED=0 go test -count=1 ./...` and `go vet ./...` passed on HOMESV, including unhooked-script determinism, sorted rule priority, strict source family, cap rejection and malicious note non-interpolation.
- `lunactl firewall policies nft-preview testdata/host-baseline.json` emits two unhooked rules and all three explicit non-claim warnings. Existing `validate`/`preview` still work.
- HOMESV has nftables `v1.0.9`, but `unshare --user --map-root-user --net` is **not permitted** for this session. To comply with [library safety rule](../project-book/LIBRARY_RULES.md), we **did not run nft -c in the parent host namespace**, did not apply any table, and left `syntaxChecked=false`.
- No CI, isolated netns, multi-host or production packet-flow result exists. A future disposable VM/privileged isolated runner must execute `nft -c` **inside its own network namespace** and report syntax evidence before this output can be considered verified.

## Next safe P0 implementation slices

1. Isolated runner VM test: `nft -c`, IPv4/IPv6, L4 constraints, existing table conflict and portable grammar. Do **not** run on live host kernel namespace.
2. Pure-Go flow fixture simulator with explicit `UNKNOWN` for kernel/coexistence; controlled external SSH recovery negative tests later in a dedicated VM.
3. Independent limited-privilege executor/IPC and watchdog/rollback, **not yet authorized for production**. Never pass this unhooked preview text into an apply path.
4. Finish release/App+Brain module version contracts, signed base knowledge and offline installer gates. Cloud and NetBird stay optional/deferred.
