# ADR-020 — Go-native core and optional Python ML workers

Status: **DECIDED** (architecture direction; not an implemented Brain)
Date: 2026-10-13
Scope: Light Luna backend, Luna Brain, optional model analysis, process and privilege boundaries
Supersedes: the JavaScript `.mjs` Brain runtime proposal in [legacy ADR-010](../12-ADRS.md)
Primary sources: [architecture](../01-ARCHITECTURE.md), [Luna Brain](../03-LUNA-BRAIN.md), [security constraints](../07-SECURITY-SAFETY.md), [standalone roadmap](../11-ROADMAP.md), [Code Map](../project-book/LIGHT_LUNA_CODE_MAP.md)

## Decision

1. **Go is the main implementation language** for Light Luna's Linux agent/backend, CLI, deterministic Luna Brain core, policy compiler, Policy Gate, event correlation, telemetry, control API and constrained privileged executor. Brain and executor must be separate processes / trust boundaries, even though both use Go.
2. **Python is explicitly allowed where justified** by a concrete ML, statistics, model evaluation, offline data-science experiment or inference integration requiring the Python ecosystem. It is **not required for baseline Host/Edge installation** or routine firewall decisions. The Python component may be a separately packaged, unprivileged optional worker with pinned dependencies and resource budgets.
3. **React + TypeScript** remains the preferred **proposed WebUI** stack; **SQLite WAL** remains proposed local persistence. These preferences are not claims of shipped functionality.
4. JavaScript `.mjs` is **no longer the default Brain runtime**. It may still be proposed for a bounded, well-justified adapter/tool if future tests show an advantage; it is not a mandatory dependency of the firewall.
5. **Inference framework/runner is not decided**. Benchmark local inference or model-serving libraries when the relevant phase starts; external models are optional and never on the kernel data path.

## Why this direction

A single Go-native deterministic and security-critical core reduces deployment dependencies, encourages typed interfaces and cross-building for `amd64`/`arm64`, and helps audit the same capability boundaries across CLI, policy and Brain. Python remains valuable for mature numerical/ML tooling without making it part of the always-on critical path. This is an architectural tradeoff, **not a proven throughput/memory benchmark**.

## Worker and authority boundary

```text
Go collectors -> canonical events -> Go deterministic Luna Brain
                                        |
                                        +-> optional Python ML worker (non-root,
                                        |   bounded input/output/time/RAM/CPU)
                                        v
                               structured DecisionProposal
                                        |
                             Go independent Policy Gate
                                        |
                         Go constrained executor (future)
                                        |
                              owned nftables only (future)
```

- Python worker input is typed, bounded, redacted, versioned and treated as untrusted. It produces **evidence or recommendations**, never firewall commands or owner authorization. Errors/timeouts degrade to deterministic rules / insufficient telemetry, never create auto-bans.
- Worker runs **without root privileges**, arbitrary shell/network/file access, secret injection or direct privileged socket rights. Enforce OS limits and isolation as feasible; verify actual limits in tests before release.
- Do not embed `python -c` evaluation of event text; never execute model-generated scripts or external threat-feed instructions. Only a preinstalled, pinned, signed/verified worker may run.
- Use an authenticated, schema-versioned local IPC boundary (Unix socket or controlled process protocol, selected after design validation). No assumed direct access from worker to `lunalockd`.
- Config allows Python worker **disabled by default**; Go Brain, core firewall and unban/TTL processes must continue independently through worker crash, resource exhaustion and offline status.
- ML worker packaging must not silently add Python dependency to baseline Lite packages; document size, disk/network and architecture cost when candidate selected.

## Rejected as default

- Mandatory Node/`.mjs` main Brain process: not needed to ship the deterministic core.
- Mandatory Python core daemon: increases baseline runtime dependency and makes the firewall rely on optional ML.
- Rust/C++ rewrite of all subsystems: no present need; specialized inference engines remain possible later behind a constrained interface.

## Verification and status

- **Implemented at decision time:** P0 Go policy validator and inert CLI only (source baseline [Go module](../../go.mod), [CLI](../../cmd/lunactl/main.go)); **not** a real Luna Brain, Python worker or executor.
- **Not verified:** worker protocols, CPU/RAM budgets, model FP rates, throughput, distro compatibility and package sizes.
- **Next:** define `DecisionProposal` and worker contracts, prototype only with synthetic corpus, benchmark Go-only vs optional Python worker; require independent policy-gate tests before any live enforcement.

## Consequences

Keep the core build and security-critical decisions Go-native; permit optional Python extensions with an explicit review of necessity, deployment impact, privilege limits, evidence provenance and test results. **P0–P4 standalone-first remains unchanged**: no NetBird code/probing until G-STANDALONE passes and the owner separately approves.
