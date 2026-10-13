# ADR-024 — Offline-Complete Installation and Independently Versioned Components

Status: **DECIDED product and update architecture; implementation NOT STARTED**
Date: 2026-10-13
Scope: Light Luna installation, base Luna Brain, Cloud/offline continuity, app/brain/module/knowledge/model release identities
Supersedes: the suggestion that Luna Brain can be absent from the initial installed product in [Packaging](../08-PACKAGING-HARDWARE.md). Optional **advanced ML** remains optional; the basic deterministic Luna Brain is mandatory in the product bundle.
Primary sources: [ADR-022](ADR-022-rebuild-luna-brain-governed-knowledge-migration.md), [ADR-023](ADR-023-dual-brain-memory-cloud.md), [offline install & lifecycle specification](../architecture/offline-base-component-update-lifecycle.md), [Cloud intelligence design](../architecture/luna-cloud-intelligence.md), [Safety](../07-SECURITY-SAFETY.md)

## Decision

A published Light Luna installer **must ship an immediately functional, offline-complete application framework plus a compact, Go-native deterministic Basic Luna Brain, and an authenticated, versioned base knowledge/scenario pack**, sufficient for baseline protection and explainable observation *without ever enrolling in or reaching Cloud*. Network/cloud integration is an enhancement, **not an activation requirement**. Firewall/nftables/TTL/recovery also remain operational if the Brain fails.

Version every release track independently, including at minimum:

- **`appVersion`** — app/control API, WebUI/CLI/agent packaging (and coordinated core protocol compatibility).
- **`brainVersion`** — Go deterministic Luna Brain engine implementation and reasoning contract (separate executable/process and update lifecycle).
- **`moduleVersions`** — individual optional modules, e.g. WAF, IDS, eBPF observer, XDP guard or vendor feed adapter; version/rollback per module, explicit privilege scope.
- **`knowledgeRevision`** — signed curated facts/CVE/scenarios/rule hints. A knowledge update **does not** silently change Brain code or App version.
- **`modelVersion`** — optional local ML weights/inference artifact; omitted when not installed. A model update **does not** silently change reasoning authority.
- **`executorVersion` / `policySchemaVersion`** — version or ABI of privileged policy engine, data contract and recovery path; independent only where signed migration/compatibility safety is demonstrated.

For example `App 1.4.0 | Brain 1.2.3 | Knowledge rev 185 | WAF 0.6.1 | ML: none` is conceptually valid **only if a signed compatibility matrix explicitly permits this combination**. These are fictional example values, not shipped versions.

## Offline-first start and upgrade channels

- Installer contains baseline executable artifacts, verified base knowledge pack, safety invariants, signed/hashed provenance, and a *last-known-good* rescue/install path. **Do not ship an empty shell that downloads its Brain after setup.**
- First boot is local-only by default; no account/cloud required, all protection controls and built-in knowledge usable without Internet. Enrollment, remote telemetry and collective learning require separately authorized actions.
- Signed releases can arrive via Cloud or from a manually imported **offline bundle** using the same manifest/trust path. Cloud-down or never-enrolled nodes still accept owner-authorized offline upgrades.
- Per-track update policies: `OFF`, `NOTIFY`, `AUTO_KNOWLEDGE`, `SCHEDULED_SOFTWARE` / `MANUAL` with risk-appropriate defaults. Future privileged executable/loader/core changes require explicit staging and owner-approved maintenance; recommendations from Luna cannot rewrite updater settings.
- Subscribing to Cloud may deliver *compatible* security advisories, vetted knowledge deltas and independently signed Brain/module/app artifacts according to each local channel/policy. **Cloud must not unilaterally force code/kernel module updates or widen firewall authority.**

## Isolation and compatibility

- App/control API, Brain, privileged executor and optional workers are **separate trust boundaries**, even if written in Go. The Brain is a separately packaged, unprivileged executable with a versioned local RPC/IPC contract; avoid an untrusted dynamic Go plugin loaded into the privileged process.
- Each component release declares `componentId`, `componentVersion`, `protocolAbi`, compatible min/max versions of required peers, OS/kernel/architecture bounds, artifact hash, signing identity, change/revocation metadata and health tests.
- No assumption of combinatorial compatibility: check the *whole installed component graph* before activation. A Brain release requiring a higher API version stays staged, never crashes the current firewall. A `requiresApp>=...` transition may require a **coordinated multi-artifact transaction** while preserving independent version numbers.
- The installer/Updater is not Brain-controlled: Cloud advice cannot cause arbitrary package script execution. Artifact signatures and provenance are verified independently of TLS, with rollback protection, trusted root/key rotation, allowlisted install paths and no arbitrary instructions from CVE feeds.

## Non-disruptive lifecycle invariant

```text
fetch (Cloud or owner offline bundle) → verify signatures, hashes, revision/ABI
→ preflight disk/RAM/architecture/kernel + entire dependency graph
→ stage side-by-side into immutable revision dir
→ smoke/IPC contract tests using synthetic inputs (no live firewall mutation)
→ arm independent rollback + preserve owner policies/DB backups
→ atomically switch permitted subsystem during safe window
→ health/decision checks + independent admin reachability
→ commit or restore last-known-good; retain audit receipt
```

- **Knowledge-only:** validate schema/size/publisher/citations, prepare inactive snapshot, switch read-only pointer after check; never add kernel rules or grant permissions. Retain previous verified pack.
- **Brain-only:** stage isolated new Go Brain, validate API with App and memory schema, replay synthetic regression, switch supervisor endpoint/worker; current nftables policy and independent TTL/rollback do not restart. Brain failure → deterministic firewall continues, intelligence becomes degraded/unknown; roll back Brain separately.
- **App/UI/agent-only:** do not reset Brain memory or reclassify proposals as authorizations; compatible Brain stays running if process boundaries allow. Restart/reload only necessary components.
- **Optional module:** disable/restart independently, reject unknown privileges, test hook coexistence; *never hot replace a live XDP attachment or nftables owner blindly*. Low-level executor/module upgrades are separately high-risk and need approval/recovery proof.
- **Model-only:** update signed model artifacts only when optional worker present; old approved model remains available until acceptance.
- **Schema/data migration:** transactional compatible forward/back strategy, snapshot/restore, pin compatible releases and no unsafe down-migration; if incompatible, abort activation and report required coordinated update.

## Operational guarantees to test, not claim

No Cloud: initialize, enforce owner-approved preinstalled policy, learn bounded local baselines, make explainable observations, show offline status, update via verified offline bundle. No Brain: nftables filtering, existing owner policies, expiring TTL and recovery survive. Interrupted update: both App and Brain independently recover known-good revisions; no corrupt DB or lost management path. Different version numbers should display independently in CLI/WebUI and support bundle diagnostics.

**Implementation is absent today:** actual repo baseline is P0 `lunactl` non-enforcing preview/validator only; no installer, Brain executable, signed update system or active Cloud sync has been written. Do not advertise these guarantees as implemented.

## Scope and precedence

Retain P0–P4 standalone-first; this ADR defines **future stable packaging/runtime contracts**, not approval to implement/probe NetBird or run Cloud sync. P5 NetBird still requires G-STANDALONE and owner authorization; Cloud fleet/intelligence architecture remains a future separately gated implementation. Data/resource quota proposals in ADR-023 remain benchmarks, not fixed product release requirements.
