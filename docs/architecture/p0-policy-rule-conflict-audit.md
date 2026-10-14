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

**Bounded completeness:** hard ceilings of 8,192 rule pairs, 100,000 prefix comparisons and 128 findings. If exhausted, `Complete=false`, `Status=INCOMPLETE_REVIEW_REQUIRED`, and `LimitedBy` indicates why; never convert partial inspection to a false clean audit. Analyzing an otherwise valid large policy does not silently approve it. Source-list unions may cover more than an individual prefix; the analyzer is deliberately conservative and does not claim exhaustive semantic equivalence.

CLI source: `lunactl firewall policies analyze <policy.json>` prints strict JSON `lightluna.rule-conflicts/v1alpha1`. `Enforceable=false`, `KernelWrites=false` for all completed or limited reports. **No HTTP listener, privileged IPC, nft call, file-write, actual rule application, Gate permission, signed owner approval, watchdog modification, NetBird or Cloud integration.**

## Offline verification and nonclaims

Source tests: [ICMP CIDR/alias matrix](../../internal/policy/policy_test.go) tests valid IPv4/IPv6, mismatched literals, mixed-family aliases and `any`; [conflict tests](../../internal/policy/conflicts_test.go) cover broad drop masking narrow allow, narrow exception before broad drop, prior `log` nontermination, TCP wildcard and disjoint ports, protocol/direction isolation, IPv4/IPv6 disjoint, CIDR alias expansion, deterministic priority order, pair budget and finding budget.

On HOMESV this slice passed `go test -count=3 ./...`, `go vet ./...` and a real CLI invocation using [existing policy fixture](../../testdata/host-baseline.json) that reported `COMPLETE_NO_FINDINGS` and both false authority flags. The fixture having no findings **does not establish firewall reachability**. No KVM test, host nft write or product install occurred for this source-only increment.

**Future Core work:** validate independent rescue/admin exceptions and owner-authenticated approval semantics; improve multi-rule union, source-family normalization and static conflict diagnostics as appropriate; stage real host/edge transactions only through a separately reviewed least-privileged executor and isolated rollback/recovery test. P0 OPEN; G-STANDALONE BLOCKED. ADR-029 monitoring and notifications frozen.
