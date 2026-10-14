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

The ephemeral test-only replay state is mutex protected, bounded to **1,024 outstanding reviews**, and tracks **both review IDs and Agent intent hashes**. Duplicate review IDs and attempts to present another review ID for the same Agent request are denied. When capacity is full, it fails closed. This is **not durable replay protection** across process restart and does not prove clock rollback resistance. It is intentionally not integrated with `ReplayJournal` or `Boundary.Evaluate`.

## Outputs and the never-APPLY guarantee

A passing pair returns a structured `lightluna.review-assessment/v1alpha1` record with `status=CRYPTOGRAPHIC_TEST_MATCH_ONLY` and `cryptographicMatch=true`. It **always** reports `humanOwnerAuthenticated=false`, `grantsApply=false`, `kernelWrites=false` and `durableReplay=false`; failures return `REJECTED`, never auto-fallback to Agent authentication. These assessment fields are not accepted by any current production/apply path.

The existing `controlgate.Boundary.Evaluate` remains unchanged and continues to **reject every `APPLY_POLICY`** even if a separate experimental review-verifier test accepts both signatures. A valid reviewer signature **cannot override** the missing privileged executor, independently authenticated owner, policy safety, TTL/rollback and rescue gates. Nothing opens sockets, writes files, calls nft, attaches XDP, talks to NetBird or sends notifications.

## Test evidence and explicit gaps

[`owner_review_test.go`](../../internal/controlgate/owner_review_test.go) verifies independent pinned public keys, no-key-reuse, correct dual-signature match, current Gate signed APPLY rejection, replay review ID and same Agent intent under a different review ID, wrong site/request/reviewer, cross-operation, changed nonce/policy hash, future/expired/outside-agent TTL, bad/mixed/HMAC Agent signature, wrong reviewer/Agent key, test purpose escalation, noncanonical digest, bounded replay capacity, and non-authoritative output.

These are **offline unit tests**, not a human owner authentication ceremony, product credential provisioning, durable review approval, real privileged executor, crash-safe Gate transaction or security audit. Until separately implemented and owner-reviewed, `APPLY_POLICY` remains entirely unavailable. P0 OPEN; G-STANDALONE BLOCKED; monitoring/notification and NetBird/Cloud remain deferred.

## Next Core gate

Specify real local owner identity and consent capture (independent of Agent), hardware/user key lifecycle and revocation, durable one-time review binding to a trusted policy hash, signed audit receipt, monotonic boot/session binding and explicit privilege separation. All must be audited and tested **before** adding a Gate-to-Executor operation that changes kernel policy. Do not retrofit the inert reviewer fixture as a production owner token.
