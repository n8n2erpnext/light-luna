# Light Luna P0 Core — Policy Family Validation and Rule Conflict Audit

Status: **IMPLEMENTED IN DEVELOPMENT SOURCE / STATIC ADVISORY ONLY; NOT APPLIED**
Date: 2026-10-14
Scope: standalone Host/Edge policy rule precedence, CIDR-set sources, ICMP family validation, bounded non-enforcing analysis
Supersedes: none
Parent source baseline: `feat/p0-standalone-foundation@98342212f328d803fd6a7c9eeddf3b9dfde6fb4a` on HOMESV; confirm final commit SHA in Git/checkpoint
Primary sources: [Policy validation](../../internal/policy/validate.go), [conflict analyzer](../../internal/policy/conflicts.go), [unit tests](../../internal/policy/conflicts_test.go), [family negative tests](../../internal/policy/policy_test.go), [CLI](../../cmd/lunactl/main.go), [flow simulator](../../internal/flowsim/simulate.go), [nft preview](../../internal/nftpreview/compile.go)

## Core change, not an enforcement feature

**Validator fix:** `Bundle.Validate` now rejects `icmp` with an IPv6 source and `icmpv6` with an IPv4 source, for literal CIDRs **and every member** of a referenced `cidr-set` alias. `any` remains family-neutral, while the selected protocol establishes its IPv4/IPv6 family. Non-ICMP TCP/UDP rules can remain dual-stack where the existing schema permits. No new schema field or port-set selector was introduced. Validation is now consistent with existing `nftpreview.render` and `flowsim.Fixture.Validate`, so the CLI `validate/preview` cannot approve an ICMP source that the nft compiler rejects later.

**Pure static rule audit:** [`policy.AnalyzeConflicts`](../../internal/policy/conflicts.go) validates the bundle, sorts by ascending unique `priority`, then compares pairs with matching direction/protocol. An earlier terminal `accept/drop/reject` can mask a later rule; an earlier `log` is nonterminal and cannot shadow. CIDR-set and `any` selectors are expanded within fixed limits; IPv4-vs-IPv6 are never overlapping. TCP/UDP port 0 is a wildcard. Findings:
- `LATER_RULE_SHADOWED`: one earlier terminal rule certainly covers the later source and port (including same-action redundancy or a later `log`).
- `DIFFERENT_ACTIONS_OVERLAP`: different actions overlap, but full shadowing was not proven; **a narrow allow before a broad drop is a reviewable intended exception**, not automatically invalid or dangerous.
- Zero findings means **only no pairwise issue detected within this simplified declared-policy model**, not no lockout risk, verified real rescue route or safe kernel policy.

**Bounded completeness:** hard ceilings of 8,192 rule pairs, 100,000 prefix comparisons and 128 findings. If exhausted, `Complete=false`, `Status=INCOMPLETE_REVIEW_REQUIRED`, and `LimitedBy` indicates why; never convert partial inspection to a false clean audit. Analyzing an otherwise valid large policy does not silently approve it. A single earlier rule's CIDR-set source **now supports exact, bounded union coverage** of a later CIDR by multiple same-family subnets, but **coverage by several different earlier rules together remains unmodeled**; this does not claim exhaustive policy/kernel equivalence.

CLI source: `lunactl firewall policies analyze <policy.json>` prints strict JSON `lightluna.rule-conflicts/v1alpha1`. `Enforceable=false`, `KernelWrites=false` for all completed or limited reports. **No HTTP listener, privileged IPC, nft call, file-write, actual rule application, Gate permission, signed owner approval, watchdog modification, NetBird or Cloud integration.**

## Bounded CIDR-set union coverage — Core continuation (2026-10-14)

**Implemented source-only extension**, parent Git `feat/p0-standalone-foundation@3d5a88cc1a190b7c39263fc85ad04b0539d6894e`: [`conflict_union.go`](../../internal/policy/conflict_union.go) derives inclusive last addresses for canonical IPv4/IPv6 prefixes and tests whether a sorted union of earlier-rule source intervals covers a later prefix **without gaps**. [`compareSources`](../../internal/policy/conflicts.go) still enforces `MaxCIDRComparisons=100000` while collecting source intervals and returns `INCOMPLETE_REVIEW_REQUIRED` on exhaustion. Prefix-set size stays bounded by `Bundle.Validate` (at most 256 alias members), and no IP iteration over a full /0 or /32/128 range occurs.

Examples: earlier `192.0.2.0/25` + `192.0.2.128/25` and later `192.0.2.0/24` now report `LATER_RULE_SHADOWED`; missing a /26 yields `DIFFERENT_ACTIONS_OVERLAP` if actions differ. Likewise, two IPv6 /33 parts can cover a /32; mixed IPv4/IPv6 entries cannot bridge address-family gaps. [`conflict_union_test.go`](../../internal/policy/conflict_union_test.go) includes positive/negative policy cases, exhaustive combinations of four /30 pieces of a /28 and four /128 pieces of a /126, plus IPv4/IPv6 address-space endpoint and comparison-budget tests. These are **bounded static correctness tests only**.

Limitations remain: an `any` later rule is treated conservatively rather than proving two-family full coverage; combining **multiple earlier terminal rules** is not analyzed as a collective union; connection state, kernel hook priority, Docker/UFW and emergency-access viability remain UNKNOWN. The analyzer cannot authorize owner approvals, ban/enforcement, nft operations or a claim that remote rescue is safe.

## Offline verification and nonclaims

Source tests: [ICMP CIDR/alias matrix](../../internal/policy/policy_test.go) tests valid IPv4/IPv6, mismatched literals, mixed-family aliases and `any`; [conflict tests](../../internal/policy/conflicts_test.go) cover broad drop masking narrow allow, narrow exception before broad drop, prior `log` nontermination, TCP wildcard and disjoint ports, protocol/direction isolation, IPv4/IPv6 disjoint, CIDR alias expansion, deterministic priority order, pair budget and finding budget.

On HOMESV this slice passed `go test -count=3 ./...`, `go vet ./...` and a real CLI invocation using [existing policy fixture](../../testdata/host-baseline.json) that reported `COMPLETE_NO_FINDINGS` and both false authority flags. The fixture having no findings **does not establish firewall reachability**. No KVM test, host nft write or product install occurred for this source-only increment.

**Future Core work:** validate independent rescue/admin exceptions and owner-authenticated approval semantics; improve multi-rule union, source-family normalization and static conflict diagnostics as appropriate; stage real host/edge transactions only through a separately reviewed least-privileged executor and isolated rollback/recovery test. P0 OPEN; G-STANDALONE BLOCKED. ADR-029 monitoring and notifications frozen.
