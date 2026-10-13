# 09 — WebUI, CLI & Typed APIs

## Operator UX

**Home (standalone first):** status/phase, protected nodes/services, local and remote Agent reachability over generic transport, moon-state confidence, current enforcement mode, last policy revision, signed update freshness and safe recovery link. NetBird reachability is a **future optional integration field**, not a required initial widget.

**Firewall:** rule editor/diff/reorder, aliases and nested groups, interfaces/zones, input/output/forward, NAT/port-forward (gateway contexts), schedules, connections/counters, timed bans/unbans, imported vendor firewall conflicts.

**Security:** Luna reasoning timeline, evidence provenance, incident score/uncertainty, scope, external source validation, independent critic/reviewer, WAF exceptions, IDS alerts and CVE exposure assessment.

**Fleet (standalone):** Light Luna-owned device pairing, remote Agent status, edge topology, owner-defined IP/DNS/mTLS transport and signed updates. **NetBird subpage (post-G-STANDALONE only):** read-only peer inventory, routing/masquerade and access-policy diff; do not build the NetBird subpage until the independent firewall works and the owner approves integration.

**Operations:** health, resource budgets, audit, backup/restore, break-glass, diagnostics bundle scrubber, upgrade/uninstall with rollback preview.

## Proposed Components & Updates page

[ADR-024](adr/ADR-024-offline-base-independent-component-updates.md) requires an independent version ledger: App, Basic Brain, Executor, knowledge pack/revision/freshness, each module and optional ML model. Show Cloud status `never enrolled/offline/connected/degraded` independently of whether local protection is running. Provide per-track update policies, separate available/current/staged/rollback versions and a **full compatibility preflight**, not a universal single "Update All" that silently changes firewall privileges. Proposed CLI `lunactl version --components`, `lunactl update status/check/plan/import/apply/rollback` are not implemented. See [lifecycle spec](architecture/offline-base-component-update-lifecycle.md).

## Moon-state icon (not a sole safety signal)

`MOON_CALM` cyan crescent; `MOON_WATCH` yellow; `MOON_ALERT` amber/orange; `BLOOD_MOON` red for confirmed high-severity threat; `ECLIPSE` for missing telemetry/control-plane visibility (unknown, not necessarily worse threat). **Show a separate threat severity and telemetry confidence indicator**; never imply red cannot occur while stale, and never imply green when sensors are blind. Icons/text/shape must be accessible without color vision.

## Initial CLI namespace proposal

```text
lunactl version
lunactl status [--json]
lunactl mode show
lunactl firewall rules list
lunactl firewall aliases list
lunactl firewall policies validate <file>
lunactl firewall policies diff <file>
lunactl firewall policies simulate <file> --scenario <scenario>
lunactl firewall policies stage <file>        # future gated capability
lunactl firewall policies apply <staged-id>   # future gated capability
lunactl firewall rollback <revision>          # break-glass via local auth
lunactl ban list
lunactl ban propose --target <ip> --scope <service> --ttl 5m
lunactl unban propose --decision <id>
lunactl luna status
lunactl luna explain <decision-id>
lunactl netbird inspect --read-only   # FUTURE ONLY; absent from standalone releases
lunactl audit recent
lunactl diagnostics --redact
```

**NetBird-specific commands and routes MUST NOT be implemented, shipped or exposed in the standalone development phases.** They are listed to preserve possible future API namespace only. **These are planned commands, not implemented binaries.** Real apply/ban/unban tools will not appear as unguarded root shells. Use stable machine-readable structured output and shell completion when implemented.

## Internal API

HTTP over Unix socket for local UI/CLI, mTLS transport for remote agents. OpenAPI 3.1, schema-generated clients, device identity and scoped RBAC. Mutating calls require idempotency keys, policy revisions, lease TTL and signed audit receipts. Use server-side auth; no browser-exposed privileged API key or device shell token.

```text
GET    /v1/status
GET    /v1/firewall/rules
GET    /v1/firewall/aliases
POST   /v1/policies/validate
POST   /v1/policies/simulate
POST   /v1/policies/stage          # future, authorized
POST   /v1/policies/commit         # future, authorized
POST   /v1/policies/rollback       # emergency, elevated
GET    /v1/decisions
GET    /v1/decisions/{id}/evidence
GET    /v1/integrations/netbird   # FUTURE ONLY, post-G-STANDALONE read-only
GET    /v1/audit
```

## UI security

Local WebUI binds loopback/explicit management interface by default, MFA/re-auth for high impact changes, short sessions, auth rate limits, CSRF defense, CORS allowlist, CSP, signed assets, secure download headers, redacted evidence and admin visibility to actual network policy conflicts.
