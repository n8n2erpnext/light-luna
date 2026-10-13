# Unified Light Luna Account, Local Console and Cloud Fleet Management

Status: **DECIDED architecture contract from ADR-026; NOT IMPLEMENTED**
Date: 2026-10-14
Scope: authentication, tenant/device enrollment, WebUI account identity, remote state/config management, safe local execution, offline fallback and Cloud migration
Primary sources: [ADR-026](../adr/ADR-026-unified-account-remote-fleet-control.md), [Cloud architecture](cloud-portability-and-scale.md), [Offline installer](offline-base-component-update-lifecycle.md), [Security](../07-SECURITY-SAFETY.md)
Related static contract: [Fleet command envelope](../../schemas/fleet-config-intent-v1alpha1.schema.json)
External precedent: [CrowdSec Console](https://docs.crowdsec.net/u/getting_started/post_installation/console/) and [polling decisions](https://docs.crowdsec.net/u/console/decisions/decisions_intro/)

## Product UX: one account, two consoles

**Local WebUI** and **Cloud Console** share the Cloud *human identity* (OIDC issuer+subject) after enrollment, not a synchronized password database, browser cookie, or root credential. No account duplication. Each UI has its own session and local/server authorization. A user signs in locally through a secure redirect to Cloud identity (Authorization Code + PKCE, state/nonce/redirect validation), then receives a locally scoped session. The Cloud uses the same IdP for its separate browser session. Google/email/password are identity-provider choices; don't embed external IdP business rules into device authorization.

**Never-enrolled / disconnected**: installer creates a local bootstrap owner with a recovery secret or hardware-protected equivalent; accounts and policies can be set up offline. Linking that owner to a newly created Cloud user requires proving owner control on the device and Cloud. Cached offline bound identity/session may allow limited, time-bound local actions, but only independent device-local emergency owner credentials can guarantee access during prolonged Cloud outage or tenant lockout. No globally shared fallback PIN or permanent cached superuser access.

### User journey

```text
Install from signed offline package
  -> local owner setup + emergency recovery
  -> basic firewall + Luna Brain + base knowledge work standalone
  -> (optional) Sign in / Create Light Luna Cloud Account
       -> OIDC identity verified, local owner approves linking
       -> local device generates key and short-lived enrollment challenge
       -> Cloud account owner accepts device; device proves key possession
       -> local owner confirms remote-management capabilities
       -> Cloud Console Device page displays ACTUAL last reported status
       -> Cloud operator proposes config changes -> local approves/validates
  -> (optional) Cloud disconnect -> local continues, state marked stale
```

No Cloud account required for installation or offline operation. Cloud-connected account is the **same identity** in local/Cloud UX, with a separate last-resort local recovery principal.

## Three identities, different secrets and lifecycles

| Identity | Bound to | Stored/validated | Not allowed |
|---|---|---|---|
| Human user | tenant + OIDC `(iss, sub)` | OIDC IdP verified by Cloud/local BFF, short session + local role mapping | No direct firewall root/kernel authority |
| Organization/tenant membership | human → tenant → role(s) | PostgreSQL authoritative Cloud; cached limited local grants | No cross-tenant access from email alone |
| Device identity | node-generated keypair + `device_id` + enrollment proof | Cloud DB revocable public-key binding; private key retained on local device | Not a human login, not shared across cloned devices |
| Emergency owner recovery | local host owner protected credential | local OS-protected storage, controlled rotation | Not a Cloud sync password or universal backdoor |

Device transfer or revocation invalidates prior membership and pending action queue; requires *local and new owner* explicit authorization. Cloud restoration to another VPS preserves durable user/tenant/device IDs and OIDC issuer binding (per ADR-025), not sessions with copied plaintext secrets.

## Cloud operator rights and local limits

Proposed RBAC baseline: **Owner, Administrator, Security Operator, Viewer**. Each permission further scoped by tenant, device, site, capability and severity; the local owner can forbid remote changes while permitting status sync.

| Action | Viewer | Operator | Admin | Owner | Local decision |
|---|---|---|---|---|---|
| View status/versions/freshness | yes | yes | yes | yes | Based on local approved telemetry |
| Inspect redacted reported policy, audit, Luna decisions | read | read | read | read | Cloud snapshot explicitly marked timestamp/revision |
| Propose bounded non-sensitive preference change | no | limited | yes | yes | Local typed validation and CAS revision |
| Propose short-lived site-scoped mitigation | no | permission-gated | permission-gated | permission-gated | MFA + local Policy Gate + safety rollback/approval |
| Configure device remote permissions / link ownership | no | no | no | limited | **Local owner confirmation mandatory** |
| Change protected admin access / trusted signing root | no | no | no | no remote | Local console recovery procedure only |
| Arbitrary exec / root shell / generic task automation | no | no | no | no | Not supported by Cloud fleet protocol |

**Owner consent has at least two distinct switches:** Cloud enrollment/status, and remote management capability. **Telemetry sharing for collective learning is a third opt-in toggle.** Never bundle consent implicitly with login, Cloud billing or software update subscription. Cloud account compromise must not automatically grant a remote ban or shell execution right.

## Desired vs actual configuration

- **Actual/reported** configuration comes from the local agent's signed/authenticated status receipt and is the authoritative device state (including reported revision, timestamp, error and coverage). Cloud cache snapshots can be stale.
- **Desired** configuration is a Cloud proposal/intent stored durably in a per-tenant/device outbox; it does **not** become actual on submission or delivery. It references an expected local policy revision and bounded operation/scope.
- **Local-only protected configuration** includes emergency owner credential, management access, Cloud authority limits, trust roots, executor privilege, local override, kernel XDP attachment and device ownership/unenrollment. These cannot be changed in Cloud.
- **Safe convergence**: device receives a typed proposal over outbound TLS/authenticated long-polling; checks target, sender/role, MAC/signature, nonce/sequence, issued/expiry, expected local revision, device-approved capability, site/TTL bounds and policy constraints; runs dry-run/diff and safety gate; reports `REJECTED`, `CONFLICT`, `AWAITING_LOCAL_APPROVAL` or staged verified outcome.
- **Conflicts**: when owner changes a local setting after Cloud proposal creation, its revision increments; proposal with stale `expectedLocalRevision` **must reject**, not merge or replace. Operator sees a redacted conflict diff and may prepare a new intent from current reported revision.
- **High-risk**: terminal confirmation or local console approval and independent firewall watchdog/rollback required; negative decision keeps current policy unchanged.
- **Offline pending**: queue contains TTL and idempotency; expired high-risk remote actions never execute when the device later reconnects.

### Envelope (design, not an implementation)

See [JSON Schema](../../schemas/fleet-config-intent-v1alpha1.schema.json) for `intentId`, tenant/device/site scopes, actor account and role, expected revision, time/expiry, operation category, action risk/authorization, local approval requirement, typed config patch and explicit `canExecuteShell=false`. It is a *structural schema only*; a future server/device verifier must independently enforce permissions, cryptography, policy semantics and anti-lockout rules.

### State machine

```text
CLOUD_EDIT_DRAFT
   -> HUMAN_REVIEWED & AUTHORIZED
   -> QUEUED_FOR_DEVICE (durable DB)
   -> DELIVERED_TO_DEVICE (transport ack, NOT applied)
   -> DEVICE_VERIFIES (principal scope, revision, ttl, replay, local policy)
     -> REJECTED | CONFLICT | EXPIRED
     -> AWAITING_LOCAL_APPROVAL (when required)
     -> STAGED (dry-run/rollback plan)
     -> APPLYING (local independent Policy Gate)
     -> VERIFIED_APPLIED | ROLLED_BACK | FAILED_SAFE
   -> device-signed/authenticated receipt updates Cloud ACTUAL state
```

Remote updates must be **idempotent** by `intentId`, device+revision and receipt revision; duplicate delivery or reconnect cannot duplicate a ban, drift authority or reset TTL. Avoid false convergence: Cloud desired revision stays "pending" until verified local receipt arrives. Conflict/denial and stale state are first-class UI statuses.

## Stable scalable transport, no open local public port

Cloud Go API/frontdoor and durable PostgreSQL inbox/outbox supports stateless replicas. Device-originated HTTPS **long polling** initially (or authenticated WebSocket later) carries queues/receipts; a connection is **outbound from device**, no inbound public port or NetBird integration required. Device messages include version/capability matrix, bounded idempotency and privacy-preserving telemetry status. Compression, total bytes, retry backoff/jitter, queue expiration and per-tenant/device rate limits prevent runaway resource use.

Future service API candidates (*not implemented*):
- `GET /v1/me`, `GET /v1/tenants/:tenant/devices` and `GET /v1/devices/:id/reported-config` (Cloud Console RBAC).
- `POST /v1/devices/:id/config-intents`: expected-revision + typed patch + approval. Cloud does not execute.
- `GET /v1/agent/intents?cursor=...`: authenticated device fetch long poll.
- `POST /v1/agent/intents/:intent/receipt`: device-authenticated outcome, actual revision and verification evidence.
- `POST /v1/agent/status`: bounded reported state and freshness, never user secrets.
- `POST /v1/device/enrollment` and approval flows: separate one-time authenticated enrollment.

Offline control still uses a local authenticated Unix socket/loopback WebUI through the *same underlying typed policy module* and guards, not a privileged Cloud direct RPC endpoint.

## Remote fleet UI navigation

- **Cloud Overview:** enrolled devices (online / offline / degraded / unknown), recent signals, vulnerabilities known vs locally applicable; no false live state.
- **Device:** OS capability, last seen, installed App/Brain/Knowledge/module versions, Brain health, telemetry freshness, snapshot local policy revision and Cloud last report time.
- **Policies:** local reported vs Cloud desired diff, scope/service, dry-run, pending/rejected/applied receipts, owner-only protected settings locked.
- **Security:** incident and CVE evidence, threat patterns and source confidence, local proposal timeline.
- **Updates:** independent signed version lanes, compatibility, health and rollback.
- **Account/Organizations:** enrollment/ownership, members/roles, device remote capability and per-tenant subscription; account/profile identity stays Cloud-based, not duplicated in device DB.

The local WebUI shares visual and domain components where practical but has a different connection and permission context; do not forward a Cloud browser session to local privileged endpoints.

## Attack / failure tests before shipping

No Internet at fresh install, Cloud login unavailable, 30-day outage, expired offline session and emergency login; same-email/different OIDC `sub`; cross-tenant IDOR; user revocation during queued command; account takeover MFA requirement; browser CSRF/open redirect/XSS; stolen/replayed device key or action envelope; device cloned without key rotation; per-tenant flood; forged status/receipt; ack lost/replayed; pending high-risk action expired offline; local/cloud revision drift; Cloud migrated to different VPS with identity and pending queues unchanged; failure mid-rule apply and admin lockout rollback; customer opt-out clears Cloud queue and revoke. Cloud commands must never include arbitrary shell, root exec or route-changing scripts.

**Status: architecture and schema only.** Light Luna presently has no local/Cloud WebUI or Auth runtime, enrollment, action queue, remote actor, device agent or firewall apply. G-STANDALONE stage still takes priority; no NetBird-specific code is authorized.
