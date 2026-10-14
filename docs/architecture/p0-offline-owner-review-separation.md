# P0 Offline Agent vs Reviewer Signature Separation — Denial-Only Contract

Status: **IMPLEMENTED PURE GO DEVELOPMENT SOURCE AND NEGATIVE TESTS; NOT HUMAN OWNER AUTHORIZATION**
Date: 2026-10-14
Scope: independent test-signer cryptographic separation from an Agent intent; no IPC/wiring or enforcement
Supersedes: none
Parent code: `feat/p0-standalone-foundation@52ece0f10ac0e162d912c22a273dc0557f59aa05` on HOMESV (verify final SHA in Git)
Sources: [Go evidence verifier](../../internal/controlgate/owner_review.go), [negative tests](../../internal/controlgate/owner_review_test.go), [existing Gate](../../internal/controlgate/control.go), [Gate replay journal](../../internal/controlgate/journal_linux.go), [original Agent/Ed25519 architecture](p0-ed25519-agent-gate-identity.md), [local Gate architecture](p0-local-control-gate.md)

## Reason for separate identities

An Agent's Ed25519 signature over a request proves only possession of the **Agent's private signing key**; setting `actorRole="OWNER"` is not independent proof of owner presence, identity, intent or informed approval. It also does not establish a legitimate, safe and recoverable host firewall transaction. The Gate's current `APPLY_POLICY` denial is the correct default.

The source-only [OfflineReviewVerifier](../../internal/controlgate/owner_review.go) is deliberately an **experimental detached proof checker**, not a second stage of permission. It holds **two different pinned Ed25519 public keys** (Agent and separate *test reviewer*), site, Agent actor and reviewer ID. It refuses identical Agent/reviewer keys, unrecognized HMAC or mixed request envelopes, invalid or tampered Agent signatures, wrong reviewer signature and noncanonical signature encoding. The verifier receives these keys from its caller in memory; there is **no enrollment, human login, key lifecycle, credential vault, transport, real owner device, WebUI confirmation or production source of authority**.

## Typed evidence and exact scope

The canonical review schema `lightluna.offline-review/v1alpha1` declares purpose `TEST_REVIEW_EVIDENCE_NOT_OWNER_APPROVAL`, opaque review ID, reviewer/site/request IDs, operation, issuance and expiry, plus a domain-separated SHA-256 of the **entire canonical Agent intent** (including nonce, policy name/hash/revision and TTL). Review signing uses a *different* Ed25519 domain from Agent request signing. The maximum review TTL is **2 minutes** and it must be fully nested in the Agent intent's valid window. The checker requires exact pinned site/reviewer, request ID, operation and SHA-256 match and rejects not-yet-valid/expired review. It verifies both signatures before recording an in-memory one-time result. An Agent digest mutation cannot be reused with another policy revision, operation or nonce.

The original **RAM-only** test verifier is mutex protected, bounded to **1,024 outstanding reviews**, and tracks **both review IDs and Agent intent hashes**. Duplicate review IDs and attempts to present another review ID for the same Agent request are denied. When capacity is full, it fails closed. The optional Linux test journal described below now reuses `ReplayJournal` for **ordinary restart replay evidence only**; neither constructor is integrated with `Boundary.Evaluate` or establishes trusted owner approval, hardware durability or clock rollback resistance.

## Optional separate offline replay journal (2026-10-14 Core P0 continuation)

**Implemented Linux nonroot source + tests only; never wired to product Gate or local IPC.** [`owner_review_journal_linux.go`](../../internal/controlgate/owner_review_journal_linux.go) adds `NewOfflineReviewVerifierWithJournal`, which opens an **absolute, dedicated `offline-review.jsonl`** (not `replay.jsonl` or `replay-ed25519.jsonl`). This reuses the already verified [Gate replay journal](../../internal/controlgate/journal_linux.go) for private owner-only `0700` directory, `0600` single-link regular file, `O_NOFOLLOW`, nonblocking exclusive `flock`, bounded strict JSON replay, journal-first append + `fsync`, and file+parent-directory `fsync` when newly created. This test-only ledger stores **review IDs and domain-separated Agent intent hashes**, never plaintext signing keys, firewall policy content or credentials. It reconstructs both replay dimensions on restart and declines an over-capacity or malformed ledger rather than accepting a request without recording it.

The existing [offline verifier](../../internal/controlgate/owner_review.go) records the two replay dimensions to disk **before** returning `CRYPTOGRAPHIC_TEST_MATCH_ONLY`, while holding its mutex. Any journal write, sync, corruption, lock or quota failure returns `REJECTED` and does not silently fall back to in-memory operation. The original no-journal `NewOfflineReviewVerifier` remains an explicit **RAM-only test constructor**. Review-ID and same-Agent-intent duplicate replays remain denied even if the other review identifier changes. [`owner_review_journal_linux_test.go`](../../internal/controlgate/owner_review_journal_linux_test.go) verifies nonroot create/reopen, restart replay, changed review ID, independent new intent, corrupt tail, symlink, conflicting lock, closed file-descriptor write failure, concurrent same-intent race and existing Gate signed APPLY rejection.

**Authority boundary and nonclaims:** The journal makes the test evidence **restart-persistent under ordinary successful fsync semantics**, but `DurableReplay=false` remains fixed in the assessment because hardware power loss, clock rollback, host compromise and full transaction/receipt durability are **not proved**. It is **not** a real authenticated owner-approval ledger and is not connected to an executor, local login, privilege change or kernel firewall operation. It shares a tested format/helper with Gate but not the Gate journal file or an authorization path. Additional tests are required before treating any implementation as production ready.

## Outputs and the never-APPLY guarantee

A passing pair returns a structured `lightluna.review-assessment/v1alpha1` record with `status=CRYPTOGRAPHIC_TEST_MATCH_ONLY` and `cryptographicMatch=true`. It **always** reports `humanOwnerAuthenticated=false`, `grantsApply=false`, `kernelWrites=false` and `durableReplay=false`; failures return `REJECTED`, never auto-fallback to Agent authentication. These assessment fields are not accepted by any current production/apply path.

The existing `controlgate.Boundary.Evaluate` remains unchanged and continues to **reject every `APPLY_POLICY`** even if a separate experimental review-verifier test accepts both signatures. A valid reviewer signature **cannot override** the missing privileged executor, independently authenticated owner, policy safety, TTL/rollback and rescue gates. Nothing opens sockets, writes files, calls nft, attaches XDP, talks to NetBird or sends notifications.

## Test evidence and explicit gaps

[`owner_review_test.go`](../../internal/controlgate/owner_review_test.go) verifies independent pinned public keys, no-key-reuse, correct dual-signature match, current Gate signed APPLY rejection, replay review ID and same Agent intent under a different review ID, wrong site/request/reviewer, cross-operation, changed nonce/policy hash, future/expired/outside-agent TTL, bad/mixed/HMAC Agent signature, wrong reviewer/Agent key, test purpose escalation, noncanonical digest, bounded replay capacity, and non-authoritative output.

These are **offline unit tests**, not a human owner authentication ceremony, product credential provisioning, durable review approval, real privileged executor, crash-safe Gate transaction or security audit. Until separately implemented and owner-reviewed, `APPLY_POLICY` remains entirely unavailable. P0 OPEN; G-STANDALONE BLOCKED; monitoring/notification and NetBird/Cloud remain deferred.

## Next Core gate

Specify real local owner identity and consent capture (independent of Agent), hardware/user key lifecycle and revocation, durable one-time review binding to a trusted policy hash, signed audit receipt, monotonic boot/session binding and explicit privilege separation. All must be audited and tested **before** adding a Gate-to-Executor operation that changes kernel policy. Do not retrofit the inert reviewer fixture as a production owner token.
