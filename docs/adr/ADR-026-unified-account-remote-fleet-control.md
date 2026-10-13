# ADR-026 — One Cloud Identity for Local/Cloud UI, Owner-Governed Remote Configuration

Status: **DECIDED product architecture; NOT IMPLEMENTED**
Date: 2026-10-14
Scope: Light Luna local WebUI, Cloud Console, identity/enrollment, remote device observation/configuration, owner authority and offline break-glass
Supersedes: none; refines [Cloud & Fleet](../06-CLOUD-FLEET.md), [WebUI/API](../09-WEBUI-CLI-API.md), [Safety](../07-SECURITY-SAFETY.md)
Primary sources: [Unified identity & fleet control](../architecture/unified-account-cloud-fleet-control.md), [ADR-024 offline foundation](ADR-024-offline-base-independent-component-updates.md), [ADR-025 portable Cloud](ADR-025-portable-scalable-cloud.md)
External comparison: [CrowdSec Console enrollment](https://docs.crowdsec.net/u/getting_started/post_installation/console/), [console decision management](https://docs.crowdsec.net/u/console/decisions/decisions_intro/)

## Owner decision and practical interpretation

**One Light Luna user identity across local and Cloud UIs**. A user with a Light Luna Cloud account signs in to the enrolled device's local WebUI using the same **identity** via standard OIDC login when connected, and can sign in to the Cloud Console to view device state, inspect actual configuration and *propose/apply supported scoped configuration changes remotely* according to the device's consent and permissions. Separate browser sessions are allowed; local never asks the Cloud account password directly. Cloud is a management surface, not the local security authority or a privileged shell.

**Preserve previous binding offline requirement**: a freshly installed never-enrolled node can be set up and operated with a **local bootstrap owner** without Cloud/Internet. Upon later Cloud enrollment, that bootstrap owner may link a Cloud subject to the existing local ownership after mutual approval, without deleting offline recovery. An enrolled device permits bounded offline authentication based on a pre-established device-local identity/session where safe; a **separate, securely protected emergency local owner recovery path** always exists for extended outages/revocation. Cloud identity is preferred for normal local usage when reachable, not a mandatory online authentication request per action.

## Identity separation and authority

- **Human identity**: durable `account_id` and OIDC `(issuer, sub)`, not email string; Cloud username/email changes do not transfer device ownership. Hosted IdP or self-hosted OIDC component behind an abstraction; stable issuer and subject identifiers must survive VPS/provider moves per ADR-025.
- **Tenant/organization**: independently scoped `tenant_id`, memberships, optional organizations/roles; devices belong to exactly one active owner tenant at a time. Changing org ownership requires explicit device-local approval/re-enrollment and revokes prior owner credentials.
- **Device identity**: keypair generated on the device, enrolled with challenge + short lived one-time code, proof-of-possession and owner approval **on both sides**; no bearer Cloud password stored on device. Device key rotation/revoke/recover separate from user login.
- **Sessions**: OIDC Authorization Code + PKCE for interactive web; app/backend-for-frontend session cookies are HttpOnly, Secure, SameSite scoped and CSRF protected. Cloud browser tokens/session cookies **never** become root/executor credentials or sent raw to the device.
- **Authorization**: server-side RBAC **AND device-local capability policy**. Roles e.g. Owner, Administrator, Security Operator, Viewer, with finer permissions per tenant, device, site, configuration category and severity; high-risk operations require recent MFA/re-auth and additional local approval. Role alone does not authorize kernel actions.
- **Offline**: local stored bound principal record and short bounded cached authorization may permit approved non-destructive local operations after disconnection. Expired or revoked-unknown Cloud sessions must not retain privileged remote rights indefinitely; fall back to locally verified emergency credentials. Recovery credentials stored via platform-appropriate hardened hash/key protection, never duplicate Cloud password and never open a secret-free bypass.

## Cloud Console functionality (after enrollment + permissions)

The owner can view enrolled devices, online/offline/stale state, security status, alerts, version tracks (App/Brain/Knowledge/modules), actual firewall/WAF/Brain configuration summary, policy revision, recent signed config receipts and pending remote changes. Telemetry upload is minimal and configurable; Cloud never gets raw secrets/keys, broad topology, unredacted payloads or raw logs by default.

A remote operator may create a **typed declarative desired change** (e.g. enable existing benign sensor, change approved baseline sampling within budget, create scoped timed block proposal, edit supported policy subset) through Cloud UI and preview **diff, target device/site, predicted reachability effects, required local approval, TTL and expected revision**. Remote control is opt-in *per device* and may be disabled even if enrollment/status sync remains active.

Cloud **queues intent**, not executable OS commands. Local Go agent establishes an **outbound authenticated** long-poll or authenticated duplex channel and fetches actions; **no inbound public port or remote shell is required**. Device validates signature/authenticity, expiry/replay/lease, local owner capability and current configuration revision; dry-run/Policy Gate/rollback independent of Cloud; produces authoritative applied/rejected/deferred receipt. Cloud then displays `desired` and `reported/actual` separately.

Remote local state remains authoritative. Concurrent local changes invalidate stale Cloud proposals using expected revision + compare-and-swap; conflicts return `CONFLICT` with redacted diff, never silently overwrite local. Owner-defined `LOCAL_ONLY` controls (break-glass admin, management route, signing trust, local emergency policies, device unlink/recovery) cannot be modified remotely even by Cloud Owner.

## Action risk policy

| Risk | Default Cloud permission | Local decision |
|---|---|---|
| Read summary, version and reported configuration | Allowed after enrollment/role; minimize data | Subject to local telemetry sharing settings and freshness |
| Download vetted knowledge, review CVE/alerts | Allowed by signed update policy | Local independent verifier/compatibility |
| Low-risk configuration change | Propose/queue only when remote management enabled | Local authorization, revision match, typed validation |
| Timed site-scoped ban/rate limit or sensitive firewall change | Needs explicitly granted device capability, MFA/re-auth, short TTL and operator preview; default `CONFIRM_EACH` | Independent local Policy Gate, management reachability protection, rollback watchdog |
| Global deny/allow, protected admin path, trust-root changes, new kernel capability/XDP attach, device ownership takeover | **NOT remotely authorized in baseline** | Local physical/console owner recovery and separate safety procedure |
| Arbitrary shell, arbitrary scripts, SSH credentials, Python model command | **Forbidden Cloud fleet action** | Outside product's typed management protocol |

The presence of a valid Cloud session or successful device poll never makes an unsafe firewall command safe. Luna Brain recommendations require separate Policy Gate and cannot self-grant Cloud or local rights.

## Transport, synchronization and mobility

- Cloud fleet API can be scaled as stateless Go replicas with tenant-scoped durable device status, policy proposals, outbox and receipts in PostgreSQL. No sticky session, transient container RAM or single-VPS coupling.
- Agents use stable owner-controlled DNS and versioned `/v1` sync contract, bounded poll/backoff/jitter, idempotent action IDs, acknowledgments and event sequence. Transport identity survives Cloud move (ADR-025); no device reinstall or re-pairing after VPS migration.
- Local state machine `PENDING_CLOUD → DELIVERED → LOCAL_VALIDATING → LOCAL_APPROVAL_REQUIRED | STAGED → APPLYING → VERIFIED_APPLIED | REJECTED | CONFLICT | EXPIRED | ROLLED_BACK`. `DELIVERED` is never presented as applied.
- Cloud downtime: existing firewall, Brain, UI, local policies and TTL/recovery continue; remote actions pending until device reconnect or expire, with **no automatic execute of stale high-impact actions**.
- Security audit requires human/tenant/device identity, origin (local/cloud), idempotency ID, operation scope/risk/TTL, expected and actual revision, approver, evidence, local receipt and rejection reasons; no secrets in Cloud logs.
- Device status freshness is explicit, stale/offline shown distinctly; cached Cloud config is a *snapshot*, not a promise about current firewall state.
- Disabling Cloud management, unenrolling the device, tenant role revocation and device key compromise require separately tested auth/key and pending-action invalidation flows.

## Reference precedent and differences

CrowdSec's Console supports Security Engine enrollment, centralized security overview and, on supported plans/configurations, console-to-engine decisions via polling. **Light Luna expands the desired state model** to explicit config diff + local authoritative receipt, granular RBAC and owner-managed high-risk gates. This ADR does not copy CrowdSec implementation or promise parity with every paid Console feature.

## Acceptance / implementation status

Required future tests: cloud login to local device, never-enrolled offline bootstrap, 30-day offline emergency recovery, cloud user revocation/session expiry, same-email different-sub collision, tenant isolation, OAuth redirect/session compromise, stolen device/cloud token, local/cloud configuration revision conflict, replay/expired action, dangerous rule rollback and lost admin connectivity, VPN independence, Cloud migration with pending actions/ack continuity, pause remote control while status remains linked.

**No local account provider, OIDC flow, Cloud account, enrollment service, remote command queue, management agent, WebUI or firewall executor has been implemented in Light Luna.** The P0 baseline remains a non-enforcing Go CLI. This decision does not authorize deploying authentication/Cloud or live nftables changes; P0–P4 standalone-first and G-STANDALONE remain in force. Cloud runtime can be implemented after the relevant release gate; local typed contracts may be designed/tested earlier.
