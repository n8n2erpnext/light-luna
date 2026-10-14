# Light Luna Local Observability API, Ban State and Multi-Channel Notifications

Status: **CANONICAL FUTURE ARCHITECTURE CONTRACT / NOT IMPLEMENTED**
Date: 2026-10-14
Scope: local Host/Edge telemetry, Prometheus scrape, Grafana dashboard, read-only ban/event API, asynchronous Telegram/Discord/Slack/email notification outbox
Source baseline: `feat/p0-standalone-foundation@de2063fe04eb09476017f9c58e16a3472a69c0a9`; future engineer must read actual Git SHA/tests
Supersedes: none
Primary sources: [ADR-029](../adr/ADR-029-local-observability-and-notifications.md), [ADR-019 standalone gate](../12-ADRS.md), [Host safety](../07-SECURITY-SAFETY.md), [Local runtime](local-runtime-foundation.md), [ADR-027 signal distinction](../adr/ADR-027-collective-threat-immunity.md)
Status boundary: ***all endpoints, adapters, databases, dashboards, credentials, services and metrics described below are proposed***. The current P0 repository only has non-enforcing local CLI/synthetic KVM test harnesses; there are NO real ban receipts from product executor and NO listener or outbound connector.

## P0 implemented pure synthetic metrics contract (2026-10-14; no HTTP, no bans)

**Status: SOURCE-ONLY synthetic fixture codec and renderer, NOT a running metrics exporter.** Implemented on HOMESV branch `feat/p0-standalone-foundation`, parent `f58fe0a449aa819bb0014a391057bc864ffc2fac` (confirm final SHA in Git). [Go source](../../internal/observability/synthetic.go) accepts a [versioned strict JSON schema](../../schemas/local-observability-synthetic-v1alpha1.schema.json) and renders deterministic [Prometheus text samples](../../internal/observability/synthetic_test.go) using [static inert fixture](../../testdata/local-observability-synthetic.sample.json). The renderer does **not** open sockets, inspect `nft` or OS telemetry, enqueue/send notifications, import a privileged executor, or assert an actual source IP was blocked. Callers would need independent verification before any real `lightluna_bans_active` metric may exist.

Immutable schema identity `lightluna.synthetic-observability/v1alpha1`; `origin=P0_SYNTHETIC_TEST_ONLY` and `authority=NO_KERNEL_OR_NOTIFICATION_AUTHORITY` are required constants. Ban aggregates accept only `family=ipv4|ipv6` and `state=proposed|unknown|rejected`; **`applied_verified` is deliberately absent**. Notification queue samples allow only four fixed channel enums (`telegram`, `discord`, `slack`, `email`) and represent **hypothetical queue counts, not actual sends**. Fixed six ban combinations and four queue series are serialized in deterministic order; no dynamic labels or actual policy authority. Output metric names are prefixed `lightluna_p0_synthetic_*`, preventing accidental conflation with the proposed live `lightluna_bans_active`.

Strict decoder reuses `internal/strictjson` for 16 KiB input cap, unknown key and nested duplicate-key rejection, invalid extra JSON, allowlisted enums, max 10 series, unique dimension combinations and per-count quota 1,000,000. Go tests verify deterministic output, negative raw IP/CIDR/URL/webhook/SMTP password/unbounded arbitrary label values, forged owner authority, `applied_verified` rejection, fixture decode and no real ban metrics. **This validates a P0 contract only**; it does not test actual Prometheus scrape, operational cardinality under concurrent host events, Grafana connectivity, authenticated detail API, notification outbox or external provider delivery.

## Goal and user-visible behavior

Light Luna's first-party local WebUI/CLI and opt-in Prometheus/Grafana integration should show (a) how many real bans are **present and independently verified applied**, (b) how many detections or policy proposals are *not applied*, (c) actions failed/reverted/expired, (d) packet/rule/detection trends with source coverage and uncertainty, and (e) service, Guardian, watchdog, queue/exporter and storage health. Operators drill into timestamped ban history and action receipts in a separately authorized local API. Notifications go only to owner-configured destinations and never constitute an enforcement command.

### Hard architectural split

```text
Local sensors / logs / nft read-side truth / approved policy audit [FUTURE]
       │ bounded typed observation events; uncertain -> UNKNOWN
       ▼
Local event normalizer + versioned append-only audit ledger [FUTURE]
       ├──> Trusted receipt reconciler (applied vs requested vs unknown)
       │       ├──> Aggregate read-only snapshot (in-memory/cache)
       │       │       ├──> GET /metrics (Prometheus text; local-only opt-in)
       │       │       ├──> Local status WebUI/CLI (future)
       │       │       └──> GET /v1/bans, /v1/events, /v1/status
       │       │                           (separate owner-scoped auth)
       │       └──> Bounded local event outbox (no ban/unban capability)
       │               ├── Telegram Bot API
       │               ├── Discord webhook
       │               ├── Slack incoming webhook
       │               └── SMTP mail
       └──> Policy Gate + privileged executor + independent TTL watchdog
                        [EXPLICIT OWNER AUTHORITY; NOT AN OBSERVABILITY INPUT]
Prometheus SCRAPES local /metrics via owner's protected collector
       ▼
PromQL / retention → Grafana dashboards/alert visualizations
(Optional Loki/Alertmanager: separate future integration, never required)
```

Transport outages, external auth failure, dashboards, emails and log shipping **cannot block packet decision/enforcement/rollback**. Collector and notifier cannot import executor privileges or write arbitrary policy. Noncooperating third-party tools must not acquire ban authority via metrics endpoints.

## Proposed API contracts (names/URLs subject to version review before implementation)

**Metrics endpoint**: `GET /metrics` (Prometheus exposition format, optionally negotiated OpenMetrics text) in a standalone local unprivileged status process. Disabled or Unix socket / `127.0.0.1` only by default; separate opt-in owner setting for scrape behind approved reverse proxy with mTLS, bearer auth + TLS, or local VPN AFTER independent host exposure review (NetBird specifically deferred). `HEAD` and management endpoints are not implicitly enabled. Export precomputed snapshots in bounded time/bytes; no live `nft list` or blocking DB scan inside scrape. Track failed collection honestly.

**Security status APIs** (future read-only and **not** a public Prometheus query API):
- `GET /v1/status`: version, mode Host/Edge, executor state `NOT_INSTALLED|UNKNOWN|READY|DEGRADED`, latest reconciled ledger timestamp, observation freshness, policy revision and counts with provenance.
- `GET /v1/bans?state=...&cursor=...&limit=...`: paginated owner-authorized, bounded entries with random stable opaque record ID (not an IP in query path), `subject` optional only with privileged read scope, `sourceType`, `decisionSource`, `decisionState`, `enforcementState`, `reasonCode`, `evidenceRef`, `ruleReceiptRef`, `firstSeen`, `appliedAt`, `expiresAt`, `revokedAt`, `rollbackVerifiedAt`, `confidence=UNKNOWN|...`; no fake `applied` based solely on a proposal.
- `GET /v1/events?after=...&limit=...`: cursor/keyset pagination of bounded, signed-or-origin-labeled local event records; clearly separate sensor observations, policy intent, actual applied/rejected, unban, system errors and outbound notification status.
- `GET /v1/notifications/status`: owner-only adapter health, queue depth, last successful send, retry/drop counts; never echo webhook URLs, tokens, recipient addresses or messages.

All detail endpoints require actual authenticated owner/session capability and local RBAC. Returned JSON has `schemaVersion`, `asOf`, `sourceTrust`, `dataFreshnessSeconds`, `effectiveEnforcementVerified`; enforce bounded max 100 records/page, size and time budget, cursor anti-tampering, audit query access and CORS deny-by-default. These are **design targets**, not implemented/public commitments.

### Ban/decision lifecycle truth table

| Local state | Meaning / display | Prometheus gauge membership | Notify? |
|---|---|---|---|
| `OBSERVED` | Telemetry only | Observations counter, no ban | Optional digest |
| `PROPOSED` | Brain suggested; not owner-approved | Pending proposals, not applied | Optional recommendation |
| `AUTHORIZED` | Policy Gate accepted typed intent | Pending operation, not yet verified | Owner choice |
| `APPLYING` | Executor attempting; no receipt | Unknown/pending, never count applied | Error escalation if stale |
| `APPLIED_VERIFIED` | Executor receipt **and** independent read-side verification consistent | Active enforced bans; expiry tracked | Ban applied event |
| `FAILED` / `REJECTED` | No proven enforcement | Error/reject counter | Important failure |
| `ROLLING_BACK` | Removal pending | **Still unknown until verified**, never silently cleared | Optional |
| `EXPIRED_VERIFIED` / `REVOKED_VERIFIED` | Removal/expiry confirmed | No active enforced ban | Unban/expiry; retraction if appropriate |
| `UNKNOWN` / `DEGRADED` | Missing/stale/mismatched receipt or untrusted state | Unknown gauge, do not miscount applied | Safety/degraded alarm |

Decisions from threat/Cloud intelligence and local detections are proposals until independently accepted by local Policy Gate. Observability must not promote uncertain proposals into real bans. Unknown-state incidents are **not hidden** by metric fallback to zero.

## Prometheus metric families — proposal v0 (small stable label enums only)

| Metric | Type | Fixed labels | Semantics |
|---|---|---|---|
| `lightluna_build_info` | gauge | `component,version` bounded signed version | 1 if reported; no user IDs |
| `lightluna_status_ready` | gauge | `component` bounded enum | 0/1 component available (not policy authority) |
| `lightluna_snapshot_last_success_timestamp_seconds` | gauge | `source` bounded | Unix time last trusted snapshot |
| `lightluna_detection_events_total` | counter | `scenario_class,outcome` enums | Trusted-classified observations |
| `lightluna_policy_decisions_total` | counter | `action,decision` enums | Proposed/authorized/rejected counts |
| `lightluna_bans_active` | gauge | `family,scope` small enums | **Only APPLIED_VERIFIED active**, not raw alerts |
| `lightluna_bans_unknown` | gauge | `family,reason` small enums | Incomplete/unreconciled ban state |
| `lightluna_ban_transitions_total` | counter | `transition,result` enums | Actual attempts/outcomes incl expiry/revoke |
| `lightluna_ban_expiry_next_timestamp_seconds` | gauge | none | Soonest known verified active ban expiry, if available |
| `lightluna_executor_operations_total` | counter | `operation,result` enums | Verified actuator attempts vs actual result |
| `lightluna_watchdog_rollbacks_total` | counter | `result` bounded enum | Audited rollback outcomes |
| `lightluna_guardian_state` | gauge | `state` fixed small enum | 1 for known state; UNKNOWN explicitly present |
| `lightluna_telemetry_dropped_total` | counter | `pipeline,reason` bounded | Input/output/backpressure loss |
| `lightluna_notification_queue_depth` | gauge | `channel` enum | Queued alerts, never message body |
| `lightluna_notification_attempts_total` | counter | `channel,result` enums | Sent/retry/terminal failure; not necessarily recipient read |
| `lightluna_notification_last_success_timestamp_seconds` | gauge | `channel` enum | Last successful transport response |
| `lightluna_observation_snapshot_duration_seconds` | histogram | `stage` bounded | Observe/calculate snapshot latency |

**Do not emit source IP, CIDR, URL, usernames, emails, endpoint addresses, webhook, rule IDs, event IDs, tenant/device IDs, process PID or unlimited scenario strings as labels.** Avoid `status` labels with arbitrary error text. Keep metric family names stable with versioned deprecation; never use current metrics to fabricate past time series. Underlying in-memory high-cardinality detail records are private and governed separately. In Grafana, configure provisioned read-only Prometheus data source and example dashboard panels: applied active bans; pending/unknown; IPv4 vs IPv6; ban expiry trends; policy rejects and rollback failures; detector error/traffic trend; exporter age and queue failures; Guardian unknown state. Grafana alerting may consume metrics but cannot apply policy; duplicated local owner notifications need independent dedupe.

Sample future PromQL (no assumptions about running exporter):
```promql
sum(lightluna_bans_active)
sum by (reason) (lightluna_bans_unknown)
sum(rate(lightluna_ban_transitions_total{result="error"}[5m]))
time() - max(lightluna_snapshot_last_success_timestamp_seconds)
sum by (channel) (lightluna_notification_queue_depth)
```

## Event/outbox state, provider adapters and delivery semantics

**Normalized event design `SecurityEventV1alpha1` (not yet a schema file):** `eventId` (random opaque internal), `source` (local component/provenance), `eventType` (`DETECTION|BAN_PROPOSED|BAN_APPLIED|BAN_FAILED|BAN_EXPIRED|BAN_REMOVED|RULE_REVERTED|GUARDIAN_DEGRADED|SERVICE_DOWN|UPDATE_FAILURE|HIGH_RISK_ALERT`), `severity` (INFO/WARN/HIGH/CRITICAL), `eventTime`, `observedAt`, `decisionRevision`, `receiptRef` (nullable/opaque), `enforcementVerified` boolean, `metadata` bounded whitelist and `dataSensitivity`. All actuation-related events link to **verified** executor/watchdog receipt and rule revision, not a free-text Brain judgment.

**Outbox design:** bounded durable queue, monotonic local insertion order, unique event ID + destination dedupe key, route rules by class/severity/mode, suppression/cooldown windows, digest aggregation (e.g. repeated scans), per-channel rate budgets, queued/sending/retry/sent/dead-letter/expired. Immediate HIGH/CRITICAL attempts have a bounded first-send budget and fallback local audit if network unavailable; lower priorities can batch. Exponential backoff with jitter, max age/attempts, bounded worker count and disk/memory limits. A provider 2xx/SMTP acceptance is only transport submission, not recipient viewing. Support redrive/manual retry with audit; never silently delete unsent critical evidence. No notification attempt may hold a firewall state-machine mutex while performing DNS/HTTP/SMTP work.

| Adapter | Transport | Security / rendering |
|---|---|---|
| Telegram | HTTPS Bot API `sendMessage` (outbound only) | Bot token + chat ID in secret store; escape Markdown/HTML; no arbitrary `parse_mode` from event |
| Discord | HTTPS webhook | Secret URL + owner-set destination; `allowed_mentions` none, explicit bounded embeds, no untrusted @everyone |
| Slack | HTTPS incoming webhook | Secret URL; bounded block/text layout, mention-safe escaping, scope to configured workspace |
| Email | SMTP STARTTLS or implicit TLS with certificate validation | Auth credentials stored encrypted/restricted, validated From/To, header and CRLF injection defense, bounded MIME body, no open relay |

Email should support configured relay rather than relying on VPS port 25; no reliance on specific provider at core. Each adapter must enforce HTTPS/TLS hostname verification, redirects disabled or tightly constrained, strict egress destination validation (SSRF/private metadata protection), timeout, size bounds, and secret redaction. Adapters are explicitly **outbound only**; a Telegram/Discord/Slack action/button cannot bypass local owner approval or become a firewall command.

### Example safe notification (template only)

> Light Luna • High • Block applied and verified • Host mode • Scenario: repeated authentication failures • Rule TTL: 15m • Occurred 2026-10-14T10:30+07:00 • Local evidence: [open in protected local WebUI]

Do **not** render the attacker IP/chat secrets by default. Include in-product link only if it is local owner-protected, not a publicly discoverable report. Ban count/decision status should be consistent between Prometheus, detail API and notifications at the same `asOf` revision, with explicit lag and unknown-state handling.

## Testing and release gating (future, no implementation claimed)

**Architecture acceptance now:** docs and versioned source contracts only, no listeners or provider calls. **P0 source-level future:** strict JSON schema validators and pure Go snapshots/event bus types, property tests that `APPLYING/UNKNOWN` never count `APPLIED_VERIFIED`; no dynamic labels or secret leaks; outage paths cannot call Policy Gate. **P1:** local-only authenticated Go exporter/API in an unprivileged service; Prometheus scrape and Grafana dashboard against a signed synthetic dataset; read-only kernel receipt reconciliation; external bind denied, rate/CPU limit verification; firewall remains independent with exporter SIGKILL/Prometheus down. **P2:** 4 adapter delivery mocks and real owner opt-in integration test with personal credentials outside repository; provider 401/403/429/500, network/DNS/TLS failures, duplicate webhook, queue overflow, retries and redrive, reboot/clock skew and critical alert suppression matrix; no log/metrics credentials.

**Do not claim complete until:** correct ban state matches actual verified executor/TTL receipts; privacy and high-cardinality tests pass; resource ceilings measured on target HOMESV/ARM; protected remote exposure independently reviewed; full offline standalone operation verified; no notification can alter policy or weaken emergency access; opted-in adapters deliver to **owner-approved** destinations; signed release and install/test evidence exist. Prometheus/Grafana/Telegram/Discord/Slack/email **are not yet integrated**.

## References

[ADR-029](../adr/ADR-029-local-observability-and-notifications.md) is the owner-decided feature direction. [Prometheus naming](https://prometheus.io/docs/practices/naming/), [instrumentation and cardinality](https://prometheus.io/docs/practices/instrumentation/), [security model](https://prometheus.io/docs/operating/security/), [Grafana native Prometheus source](https://grafana.com/docs/grafana/latest/datasources/prometheus/) and [optional Loki](https://grafana.com/docs/grafana/latest/datasources/loki/) informed non-product-specific operational defaults.
