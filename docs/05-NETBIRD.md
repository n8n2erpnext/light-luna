# 05 — NetBird Integration Research and Architecture

> **DEFERRED — NOT IN INITIAL DEVELOPMENT SCOPE.** This document preserves research for a *future optional connector*. Do **not** start NetBird code, package dependencies, API discovery/probes, integration tests, credential setup or changes to the deployed NetBird stack until Light Luna first runs successfully as an independently packaged firewall and passes the standalone acceptance gate in [`11-ROADMAP.md`](11-ROADMAP.md), followed by explicit owner approval to start the integration phase. Standalone Host and Edge require no NetBird client, server or account.

Research date: 2026-10-13. This is **parked reference design based on published NetBird documentation**, not a live authorized test of the user's NetBird stack or a claim that every API feature works in the installed version. API probing is **deferred until after standalone acceptance and owner go-ahead**.

## What NetBird is and is not

NetBird is a WireGuard-based private network using **Client, Management, Signal and Relay** components. Management holds peer/network/policy state; Signal coordinates initial peer discovery; encrypted traffic is peer-to-peer or relayed if direct connection fails. Signal is not a packet inspection point. A NetBird peer on an Edge box does not give that box automatic visibility into other peers' app traffic.

Self-hosted versions differ: modern new deployments (v0.65.0+) use combined `netbird-server` container with `config.yaml`; legacy installations use separate `management`, `signal`, `relay`, `coturn` and associated `management.json`. The Reverse Proxy is a separate beta feature, with self-hosted Traefik TLS-passthrough dependency. **Do not assume a migration**, do not edit the installed stack during connector discovery.

## High-value Light Luna integration pathways

1. **Device discovery (read-only):** inventory peer IDs/names/IPs/connectivity/platform and map to registered Light Luna Agents via signed device identity, not just IP/hostname.
2. **Access context (read-only):** group membership, policies, posture checks, Networks, Network Resources and router/masquerade configuration. Display but do not reinterpret as global firewall policy.
3. **Secure control path:** Agent-to-Controller mTLS over NetBird mesh; NetBird encryption complements, but never replaces, per-agent authorization and command signatures.
4. **Audit corroboration:** correlate NetBird audit events and (where deployed and available) reverse proxy request metadata with Luna observations. Audit event availability differs by edition/version.
5. **Optional user-approved policy proposals:** present NetBird policy/group changes as explicit diffs requiring confirmation. Never synchronize `BAN` into NetBird groups/policies as a default, because an accidental edit may disconnect all admins or entire fleets.
6. **Reverse Proxy signals:** consume available HTTP/L4 proxy events and CrowdSec reputation verdicts; connect WAF at a documented L7 point, not a fictitious API.
7. **Gateway support:** NetBird Networks can reach network resources through routing peers; distinguish LAN source IP from masqueraded routing-peer IP and evaluate subnet return path.

## Published Management API endpoints to evaluate

| API (method/path examples) | Intended usage in first connector | Proposed privilege |
|---|---|---|
| `GET /api/peers` | Read peer inventory | Read only |
| `GET /api/groups` | Discover group scopes | Read only |
| `GET /api/policies` | Display access controls | Read only |
| `GET /api/networks` | Network inventory | Read only |
| `GET /api/networks/{networkId}/resources` | Visibility of routed resources | Read only |
| `GET /api/networks/{networkId}/routers` | Routing peers, masquerade and metrics | Read only |
| `GET /api/events/audit` | Correlate policy and device changes | Read only |
| `GET /api/events/proxy` | Reverse Proxy events **where supported** | Read only |
| `GET /api/events/network-traffic` | **Cloud-only experimental**; never rely on for self-host | Optional/no requirement |
| `POST/PUT /api/policies`, network resource writes | Future owner-confirmed sync only | Separate elevated role and audit |

Use service users and narrowly scoped token permissions where supported; official docs show `Authorization: Token <PAT>` or OAuth bearer. NetBird Cloud API example base is `https://api.netbird.io`; self-hosted URL derives from the deployment's configured management endpoint, not an assumed universal cloud hostname. Keep secrets in OS credential vault, secret file or supported secure store, never YAML committed to Git. Make request rate limits, pagination, idempotency, backoff and retries configurable. Do not expose a read-only credential as a write credential.

### Example *read-only* API contract (not run)

```http
GET /api/peers HTTP/1.1
Host: <configured-netbird-management-host>
Authorization: Token <redacted-service-token>
Accept: application/json
```

Never print token or raw response containing sensitive peer names/addresses in operator diagnostics by default. Record only aggregated / redacted samples.

## Critical kernel firewall coexistence

The NetBird Linux client **manages its own iptables/nftables filtering/forwarding and interface `wt0`**. It may create a table named `netbird` (configurable by `NB_NFTABLES_TABLE`). Light Luna must:

- Discover NetBird's active backend/ruleset and observe owner labels without editing them.
- Own distinct `table inet light_luna_*` / namespace. **Never flush global nft ruleset, NetBird table, Docker chains, UFW/firewalld assets.**
- Test actual kernel netfilter hook priority and final verdict across all active base chains; an `accept` in one chain is not an across-chains allow override.
- Preserve NetBird connectivity for management/rollback, with an **independent** break-glass route. Do not treat all `wt0` traffic as intrinsically trusted.
- **Never use `netbird up --disable-firewall` as a conflict shortcut.** NetBird documents that this disables peer-side ACL enforcement and users must then replace it themselves.
- Detect dynamic change/reconnect/upgrade drift. Maintain compatibility matrix for NetBird versions, kernel, iptables-nft/native nft, firewalld/UFW, Docker/LXD.

## NAT and source IP provenance

The NetBird Networks router may masquerade traffic; backend sees router IP rather than original peer. Conversely, disabling masquerade needs a valid return route and careful segmentation. The NetBird Reverse Proxy behind Traefik requires a properly configured PROXY protocol/trusted-proxy chain to identify original external client IP reliably. A malicious user-supplied `X-Forwarded-For` or `X-Real-IP` cannot be considered authoritative outside a validated trust boundary.

**Decision gate:** no global ban by a supposed original-IP identity unless the chain to the actual enforcement point is proven, logged and periodically rechecked. Test from a genuinely independent external test source before claiming provenance verification; single-host spoof-header tests alone are insufficient.

## NetBird Reverse Proxy, HTTP/WAF and CrowdSec

- NetBird Reverse Proxy beta supports HTTP (TLS termination) and L4 TCP/UDP/TLS modes. TLS passthrough does not expose HTTP body to Light Luna.
- It supports IP/country access restrictions and CrowdSec IP Reputation with per-service `off`, `observe`, `enforce` modes (as documented for supported self-hosted proxy instances).
- CrowdSec LAPI/stream bouncer can be co-deployed. Decide a *single owning enforcement tier* per action; Light Luna should not inject duplicate global blocks without correlation and user policy.
- WAF such as Coraza/OWASP CRS requires HTTP request visibility and documented compatible connector/insertion path. Merely reading NetBird proxy logs provides detection, not in-band HTTP blocking.
- Self-host reverse proxy requirements include Traefik and TLS passthrough. Avoid claiming full compatibility until tested with chosen NetBird version and architecture.

## Integration maturity/gates

**Start gate: POST-STANDALONE ONLY.** The milestones below are **not** the Light Luna product P0/P1. They are independent future NetBird-connector milestones and remain blocked until [`11-ROADMAP.md`](11-ROADMAP.md) standalone readiness acceptance and owner authorization.

**NB0 (FUTURE):** read-only connector; client feature probe, peer inventory, basic health and policy drift detection. Does not need write token.
**NB1 (FUTURE):** derive authorized fleet targets, privacy-safe aggregate events and management channel identity, sandbox network tests.
**NB2 (FUTURE):** explicit owner-approved NetBird policy suggestions and conflict simulation, negative security tests and rollback proof.
**NB3 (FUTURE):** optional advanced Reverse Proxy/CrowdSec adapter when vendor version/extension guarantees established.

## Tests required (ONLY after standalone readiness gate)

- NetBird connected vs disconnected, P2P vs relay; no peer traffic sniffing assumption.
- `wt0` filtering with nftables and iptables-nft; no ACL bypass/lockout on restart/upgrade.
- Docker/LXD forwarding, Networks vs deprecated Routes, route and masquerade combinations.
- API token scopes denied/expired/revoked, rate limits, 401/403/429, schema drift and pagination.
- Self-host combined and legacy stack, Traefik TLS passthrough, trusted source IP chain, multiple proxies.
- Admin lost peer connection while policy staged; rollback without Luna/NetBird/Cloud.

## Primary references

- https://docs.netbird.io/about-netbird/how-netbird-works
- https://docs.netbird.io/api/introduction
- https://docs.netbird.io/api/guides/authentication
- https://docs.netbird.io/api/resources/peers
- https://docs.netbird.io/api/resources/networks
- https://docs.netbird.io/api/resources/policies
- https://docs.netbird.io/api/resources/events
- https://docs.netbird.io/about-netbird/ports-and-firewalls
- https://docs.netbird.io/client/environment-variables
- https://docs.netbird.io/help/troubleshooting-client
- https://docs.netbird.io/manage/networks/masquerade
- https://docs.netbird.io/manage/network-routes
- https://docs.netbird.io/manage/reverse-proxy
- https://docs.netbird.io/selfhosted/maintenance/crowdsec
- https://docs.netbird.io/selfhosted/external-reverse-proxy
