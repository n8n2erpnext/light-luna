# 01 — System Architecture

## Operational modes (one distribution, selected during setup)

| Mode | Where | Observes | Enforcement | Requirements |
|---|---|---|---|---|
| Host | Linux VPS/server/NUC | Local host flows, services, SSH, proxy/application logs | Local nftables and local L7 adapters | 1 NIC works; no routing required |
| Edge | Mini-PC/NUC/VPS acting as out-of-band controller | Own host + received structured remote telemetry, logs, optional mirror feed | Local host; *remote agents* enforce their own rules | 1 NIC works for non-inline mode; packet visibility requires TAP/SPAN, gateway placement or forwarded logs |
| Gateway | Compatible Linux router/appliance | Traffic actually routed/bridged through it | forward/input/output, NAT, zones, WAF when TLS-terminating | 1+ interfaces technically possible with VLANs; practical WAN/LAN may use two NICs, trunk VLANs or virtual NICs |

**No universal claim that Edge sees all LAN traffic.** A passive box connected to a normal switch port sees only accessible traffic. Inline bridge/routed gateway requires explicit topology and L2/L3 tests. NAT is effective only on forwarded/routed traffic; no gateway side effects in Host/Edge merely because the UI exposes NAT configuration.

## Layer boundaries

1. **Data plane (always available):** kernel nftables/conntrack; packet movement is not blocked waiting on an LLM. Optional WAF in terminating proxy and IDS inspection path.
2. **Sensor plane:** host journald/auth, firewall counters, eBPF *optional and supported kernels only*, conntrack summaries, reverse proxy access records, resource health, route/interface changes; event source and trust chain retained.
3. **Decision plane:** canonical event schema → correlation → Luna Brain & deterministic detectors → severity, target scope, evidence IDs, uncertainty; rules execute independent of AI.
4. **Policy plane:** typed configuration, aliases, zone/rule compiler, owner allowlists, vendor-managed chain inventory, static validations and signed config version.
5. **Execution plane:** minimum-privilege `lunalockd` API, allowlisted typed operations (apply validated ruleset, add/remove expiring set element, service-level WAF action); cannot execute arbitrary bash or arbitrary paths. Execution independent of Luna.
6. **Control plane:** local WebUI, `lunactl`, authenticated local API, optional fleet controller; role and human approval separation.
7. **Intelligence/update plane:** local cache, scheduled bounded imports, signed update channel, CVE/advisory correlation; data quality and freshness gating.

## Recommended implementation (proposals, not commitments)

- **Go:** native Linux agent, typed rule compiler, executor service, telemetry buffering; cross-build Linux `amd64` and `arm64`.
- **JavaScript `.mjs`:** Luna orchestration, skill/evidence parsers and higher-level decision experiments; isolate using unprivileged process, bounded CPU/memory and deterministic schema. Final shipping brain model/runner may vary based on benchmark.
- **React/TypeScript:** responsive WebUI, shared schema-generated clients; no sensitive privileged operations in browser.
- **SQLite WAL:** single-node durable config, decision and audit index, event summaries (bounded retention); append-only integrity-checked audit chain where useful. Cloud relational DB only if needed for scale and tenancy.
- **unix domain socket + peer credentials:** local control interface; mTLS device identity for remote agents.
- **nftables:** first Linux enforcement backend; separate backend interfaces enable OS migration later rather than promising Windows/macOS firewall at initial release.
- **WAF:** proxy integration / Coraza + OWASP CRS; observe mode first. WAF placement explicit.

## Data flow

```text
network / proxy / host / NetBird -> collectors -> canonical events
                                                | provenance + timestamps
                                                v
                                       detector + Luna Brain
                                                |
                                    proposed decision with TTL/scope
                                                v
                                             Policy Gate
                                  / deny / require approval / allow \
                                                |
                                        typed executor request
                                                v
                                  nftables or WAF/IDS adapter
                                                |
                                       verify outcome and audit
                                                |
                                           label feedback
```

### Evidence quality

- Distinguish **observed source IP**, **trusted original client IP**, **proxy/container/NAT address** and **uncertain**. Ban-by-IP only when enforcement-point identity is trusted or policy specifically tolerates uncertainty.
- Preserve `observed_at`, `collected_at`, `source_id`, `site_id`, `scope`, `truncated`, `complete`, `trust_level`, `evidence_hash`, `reviewable_summary`.
- Distinguish genuine repeated scanning across disjoint windows from one burst; avoid counting repeated sampled reports as independent events.
- `NO_ACTION` means *no actionable evidence in the available observed window*, not “host is safe.” Stale/unknown evidence must not be styled green.

## Interfaces / future protocol packages

`EventV1`, `RuleV1`, `AliasV1`, `PolicyBundleV1`, `DecisionProposalV1`, `EnforcementReceiptV1`, `DeviceStatusV1`, `WafPolicyV1`, `EvidenceProvenanceV1`. Use OpenAPI + JSON Schema and version negotiation. Idempotency keys, anti-replay nonces, operation IDs and monotonic revision numbering required; downgrade unsupported schema to read-only.

## Isolation and multi-tenant scope

Site-level WAF rule actions must not unexpectedly impose global IP bans affecting unrelated hostnames. Multi-tenant Cloud decisions remain scoped by tenant/device/service and have policy-approved propagation, short TTL and explicit exceptions. Local config and last-known policy take priority if cloud unreachable.
