# Light Luna P0 — Owner Key Lifecycle and Revocation Contract

**Status:** DRAFT / SOURCE-AUDITED SECURITY CONTRACT — LIFECYCLE MODULE **NOT IMPLEMENTED**, NOT OWNER AUTHORIZATION, NOT RELEASE.

**Date:** 2026-10-14. **Scope:** Core P0, separate reviewer test-signing identity, key generation/revocation/rotation and anti-rollback acceptance criteria. **Supersedes:** none. **Audited baseline:** `feat/p0-standalone-foundation@efd1562135173c8762c9358ffef0aff3e4913301` (HOMESV). Current HEAD must be verified independently. **Origin:** owner-requested recovery of separate offline draft dated 2026-10-14, not prior implementation.

**Source:** [offline cryptographic verifier](../../internal/controlgate/owner_review.go), [offline reviewer journal](../../internal/controlgate/owner_review_journal_linux.go), [Gate replay journal](../../internal/controlgate/journal_linux.go), [existing offline tests](../../internal/controlgate/owner_review_test.go), [Core-only regression test](../../internal/controlgate/owner_review_authority_regression_test.go), [parent architecture](p0-offline-owner-review-separation.md).

## 1. Verified existing boundary

`OfflineReviewVerifier` pins two distinct Ed25519 *test* public keys and verifies the Agent intent and a detached reviewer signature scoped by request ID, nonce, site, policy hash/revision, operation and TTL. The optional journal stores review IDs and intent digests, rejects already-consumed records across normal restart, and is **not** wired into `Boundary.Evaluate`. An accepted signature is **`CRYPTOGRAPHIC_TEST_MATCH_ONLY`**. `HumanOwnerAuthenticated=false`, `GrantsApply=false`, `KernelWrites=false`, `DurableReplay=false`. All Gate `APPLY_POLICY` attempts remain `REJECTED`.

**Known gap:** there is **no** versioned reviewer key registry, trustworthy owner enrollment/consent ceremony, rotation/revocation authority, trusted private-key custody, persistent monotonic generation anchor, rollback/clock defense or production Executor. Agent `ActorRole="OWNER"` is a self-assertion. A valid reviewer signature proves only possession of a test key, not owner permission.

## 2. Proposed schema and trust model (not source)

Future `lightluna.owner-key-lifecycle/v1alpha1` state shall include strictly bounded `siteId`, `keyId`, positive `generation`, `algorithm=Ed25519`, canonical lowercase `publicKeySha256`, enum state `UNENROLLED | ACTIVE_TEST_KEY | REVOKED_TEST_KEY | ROTATION_PENDING`, `effectiveUtc` and separately authenticated transition provenance. Never store a private key, raw firewall IP/packet, webhook, password or owner credential in this document/state.

**Schema validity or a self-signed rotation by a compromised current reviewer key does not confer trust.** A real owner authority must be enrolled independently of Agent, present visible policy impact, control custody/recovery/revocation and persist authoritative anti-rollback state outside the self-replaceable local journal. A local in-memory counter is not enough across restarts. The real trust anchor and consent mechanism require separate owner design approval; they are not selected by this contract.

## 3. Required invariants

- Agent key and reviewer key are never equal. Reject wrong key length/algorithm, duplicate or recycled key fingerprints, malformed ID, unsupported state and reintroduction of any retired/revoked key.
- Generation must **strictly increase** on rotation or revocation; reject zero/stale/equal/overflow transitions, wrong site or key ID and rollback to an old generation. A new key must not trust an old self-contained snapshot without independent freshness proof.
- Revocation invalidates **all pending review evidence**, including valid signatures created before revocation. Rotation invalidates old-key signatures while retaining replay history across allowed key changes.
- The decision boundary must atomically serialize `Check` vs `Revoke/Rotate`: after revocation commits, no stale-key acceptance. Preserve Agent intent hash and review-ID replay history, with no silent memory-only fallback when a journal fails.
- Missing, corrupt, truncated, symlinked, competing-writer or unsyncable registry is **FAIL-CLOSED**. Clock rollback, reboot epoch uncertainty and persistent-state replay are **unverified** without a protected monotonic anchor. Do not interpret ordinary `fsync` as proof of power-loss resistance.
- Separate cryptographic test success from authenticated human owner consent. `APPLY_POLICY` stays unavailable independent of registry test status, and no review token may be passed to privileged Executor as a grant.

## 4. Mandatory negative-test matrix — proposed, NOT all implemented

| ID | Input | Expected result | Status |
|---|---|---|---|
| K01 | Agent public key identical to reviewer | Reject setup | Existing verifier test PASS |
| K02 | Invalid/unsupported/duplicate public key | Reject | Partial existing tests; registry not built |
| K03 | Old key signed valid review after revoke | Reject | NOT IMPLEMENTED |
| K04 | Old key signed valid review after rotate | Reject | NOT IMPLEMENTED |
| K05 | Rotate back to retired key fingerprint | Reject | NOT IMPLEMENTED |
| K06 | Generation unchanged/lower/overflow | Reject | NOT IMPLEMENTED |
| K07 | Restore stale key snapshot after revoke | Reject without independent freshness | NOT IMPLEMENTED |
| K08 | Review created before revoke, presented afterward | Reject pending evidence | NOT IMPLEMENTED |
| K09 | Rotation + same Agent intent under new review ID | Replay still rejected | Existing same-key replay test PASS; rotation case pending |
| K10 | Untrusted Agent-supplied replacement key | Never enroll from Agent payload | NOT IMPLEMENTED |
| K11 | Cross-site or wrong reviewer identity signature | Reject | Existing offline verifier tests PASS |
| K12 | Changed nonce, policy hash/revision or request after signing | Reject | Existing partial tests PASS |
| K13 | Future/expired TTL or backwards clock | Reject/outside verified time | Existing time-bound regression PASS; rollback anchor pending |
| K14 | Concurrent review vs revoke | No stale acceptance after committed revoke | NOT IMPLEMENTED |
| K15 | Two competing registry writers | Fail closed | Existing journal lock test only; registry pending |
| K16 | Corrupt/partial/missing registry, sync failure/quota | Fail closed, no RAM fallback | Existing review journal test only; registry pending |
| K17 | Restart after rotation without protected freshness proof | Reject/unknown, not a trusted live generation | NOT IMPLEMENTED |
| K18 | All cases incl valid dual signatures | `HumanOwnerAuthenticated=false`, `GrantsApply=false`, `KernelWrites=false`, `DurableReplay=false` | Existing verifier and regression tests PASS |
| K19 | Gate receives Agent-signed APPLY | `REJECTED`, `ExecutorStatus=NOT_AVAILABLE`, `KernelWrites=false` | Existing Gate and regression tests PASS |

The test statuses are source-based checks at this phase; they **do not** establish K01–K19 full acceptance. A later lifecycle registry implementation must rerun the full matrix and verify it in nonroot isolation, race tests, simulated snapshot rollback and separately reviewed recovery logic. No kernel/firewall operations are required for the proposed key-state unit tests.

## 5. Staged Core work and blockers

**A.** A future authorized purely in-memory, non-enforcing transition model and property/negative tests for K01–K06, K08–K09, K14. This is **not present**. A previous attempt to create a reviewer lifecycle module was blocked by tool safety checks; the blocked source action was **not retried** in this increment.

**B.** A separately reviewed, nonroot key-state journal and failure handling for K15–K17; must not recycle the Gate replay file or pretend a local journal is a trusted owner identity anchor. Not present.

**C.** Independent owner identity and consent UI/ceremony, on-device key custody/recovery, signed receipts, trust revocation, boot epoch and rollback-resistant monotonic generation. Requires security design and explicit owner approval before any real Executor link. Not present.

**Release truth:** P0 OPEN, G-STANDALONE BLOCKED. No production human owner authentication, privileged executor, safe kernel policy transaction, NetBird/Cloud/telemetry or notification integration. All source test paths remain non-enforcing.
