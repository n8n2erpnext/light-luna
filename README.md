# Light Luna

**Orbit. Observe. Defend.**  
**Intelligent Software Firewall & Autonomous Defense**  
*A Light Family product by THAI DUY Digital*  
**Status:** Architecture baseline plus an experimental **non-enforcing P0 policy/CLI scaffold**, 2026-10-13. **No working firewall, nft apply, service, cloud, live ban or standalone acceptance is claimed.**

Light Luna is a packaged **software firewall** for Linux servers and future compatible security appliances. It combines traditional, predictable firewall controls with **Luna Brain**, a bounded, evidence-driven security decision system. It works without Cloud, NetBird or Luna Brain. **Development is strictly standalone-first**: implement, package and prove the independent firewall before initiating ANY NetBird connector development. NetBird is a later optional integration, not an early development track.

## Product principles

1. **Firewall first:** deterministic rules, aliases, stateful filtering, zones, interfaces, NAT where applicable, schedules, audit and recoverable configuration; AI is optional.
2. **Install anywhere, select deployment mode:** **Host** (local protection), **Edge** (log/telemetry controller with remote agents, not necessarily inline) and **Gateway** (routing/forwarding appliance, optional for compatible hardware). Mode is a configuration, not a separate codebase.
3. **Defense in depth:** L3/L4 firewall + layer-7 WAF/reverse-proxy adapters + optional IDS/IPS + Luna Brain.
4. **Autonomy with bounded authority:** detect → investigate → decide → independent policy gate → execute → verify → learn; every execution has provenance, expiry, replay protection and rollback.
5. **Offline-first safety:** no dependency on an external API or LLM to permit routine traffic; existing policies and ban expirations continue if Cloud fails.
6. **Strict development order:** standalone Host/Edge firewall + WebUI/CLI + installer + tests first; demonstrate operational standalone acceptance before starting NetBird integration. NetBird API research is parked reference material only until that gate.
7. **Non-invasive coexistence:** never flush other components' firewall tables, interrupt Docker/NetBird/UFW/Cloud security groups, or disable NetBird ACL enforcement as a workaround.
8. **Privacy and provenance:** source IP, proxy identity, version-to-CVE applicability, learning labels and telemetry gaps must be explicit; no default sharing of raw logs or customer IP histories to Cloud.
9. **Releasable product:** two-architecture signed packages, clean install/upgrade/uninstall, service watchdog, safe migrations, rollback, supply-chain attestations and compatibility matrix.

## Naming

**Product:** Light Luna (two words). **Brain:** Luna Brain. **Cloud:** Light Luna Cloud. **CLI:** `lunactl`. **Privileged executor:** `lunalockd` (working name). **System agent:** `luna-agent` (working name). **Appliance:** Light Luna Box. **Firmware/project image:** Light Luna OS (future, only after hardware feasibility).

## Read this first

**Documentation starts at [Light Luna Project Book](docs/project-book/README.md).** New engineers and AI agents must read its [Library Rules](docs/project-book/LIBRARY_RULES.md), [Project Book](docs/project-book/LIGHT_LUNA_PROJECT_BOOK.md), and [Code Map](docs/project-book/LIGHT_LUNA_CODE_MAP.md) before assuming a planned feature is implemented or changing docs. Source files and Git/test evidence remain the final authority.

| Document | Purpose |
|---|---|
| [Product vision](docs/00-VISION-BRAND.md) | Name, lunar metaphor, personas, market/product boundary |
| [System architecture](docs/01-ARCHITECTURE.md) | Modes, planes, interfaces, authority and data flow |
| [Firewall specification](docs/02-FIREWALL-ENGINE.md) | Rules, aliases, NAT, compiler, transactional enforcement |
| [Luna Brain](docs/03-LUNA-BRAIN.md) | Observation, decisions, learning, confidence and trust |
| [WAF, IDS & CVE](docs/04-WAF-THREAT-INTEL.md) | L7, traffic inspection, threat feeds, software provenance |
| [NetBird integration — DEFERRED](docs/05-NETBIRD.md) | Reference research only. NO implementation before standalone acceptance gate |
| [Cloud & Fleet](docs/06-CLOUD-FLEET.md) | Remote controllers, security, lifecycle and privacy |
| [Security & safety](docs/07-SECURITY-SAFETY.md) | Threat model, guardrails, lockout rescue, abuse defense |
| [Packaging & hardware](docs/08-PACKAGING-HARDWARE.md) | Linux deployment, future appliance, upgrade/recovery |
| [WebUI, API & CLI](docs/09-WEBUI-CLI-API.md) | Operator UX, RBAC, control API and execution tool |
| [Moon-state design](docs/10-MOON-DESIGN.md) | Brand and incident severity language |
| [Roadmap & test gates](docs/11-ROADMAP.md) | Milestones, acceptance criteria and anti-scope-creep |
| [Architectural decisions](docs/12-ADRS.md) | Decided / proposed / unresolved |
| [Research sources](docs/13-REFERENCES.md) | Public official documentation links and caveats |
| [Project handoff](docs/14-HANDOFF-2026-10-13.md) | Historical design-era handoff |
| [P0 implementation notes](docs/15-P0-IMPLEMENTATION-NOTES.md) | Actual implemented slice, commands, limitations and next tests |
| [HOMESV P0 handoff](docs/16-P0-HOMESV-HANDOFF-2026-10-13.md) | Dev environment, verified tests, safe next tasks and Git publishing |

## Implementation language direction

**Go** is the planned primary backend, CLI, deterministic Luna Brain and separated constrained executor language. **React + TypeScript** is the proposed WebUI. An **optional, unprivileged Python worker** may be added only for justified ML/statistics/inference tasks; baseline Host/Edge operation and safety never depend on Python or any LLM. JavaScript `.mjs` is no longer the planned default Brain runtime. See [ADR-020](docs/adr/ADR-020-go-core-python-optional.md). **Only the P0 Go CLI/validator is implemented today.**

## Brain V2: a knowledge-driven defensive system

Light Luna's differentiator is a governed security reasoning plane: **local observations + trusted public CVE/KEV/EPSS/vendor feeds + ATT&CK attack-technique models + exact installed-asset applicability + observed temporal behavior → evidence-bound incident/pre-attack recommendations**. Its collective Cloud knowledge is a future privacy-controlled, opt-in, signed/versioned subscription rather than a global ban authority. The existing Python Sentinel Luna is studied as a source of curated patterns and false-positive safeguards, **not** an existing pretrained neural model or production Light Luna Brain. The new Go-native Brain V2 is currently **architecture only**: [ADR-022](docs/adr/ADR-022-rebuild-luna-brain-governed-knowledge-migration.md), [Brain V2 design](docs/architecture/luna-brain-v2.md), [Threat Pipeline](docs/architecture/threat-intelligence-knowledge-pipeline.md), [source audit](docs/history/audits/SENTINEL_LUNA_COGNITION_AUDIT_2026-10-13.md).

## Optional eBPF/XDP roadmap

nftables remains the primary firewall. Linux **eBPF host/network/socket/process observability** is an optional future sensor, and **XDP early-ingress L3/L4 filtering** is an optional, explicitly enabled future anti-DDoS fast path. Neither is implemented; no XDP/BPF attachment or packet drop is authorized by this plan. Preserve all existing XDP program ownership and support nftables-only on unsupported NICs. See [ADR-021](docs/adr/ADR-021-ebpf-observability-xdp-optional.md).

## Mandatory stage gate — standalone first

**Stage A (ACTIVE PLAN):** Build and prove Light Luna as a self-contained firewall. Host and Edge must work without NetBird installed, without NetBird APIs, and without any NetBird dependency. Complete local rules, aliases, policy compiler, agent/executor, WebUI/CLI, packaging, recovery and standalone end-to-end tests first.

**Stage B (LATER, BLOCKED):** Only after a documented standalone working build passes the exit criteria in [`docs/11-ROADMAP.md`](docs/11-ROADMAP.md) **and the owner authorizes the next phase**, begin engineering and testing the optional NetBird connector. Until then: **no NetBird integration code, API probing, access tokens, compatibility tests or modification to the existing NetBird stack**. Existing public documentation research may remain in this repository as future design reference.

## Initial topology

```text
          Future optional Cloud Intelligence / Fleet Controller
                             |
                     mTLS / signed updates
                             |
WebUI / lunactl ---> Local Control API ---> Policy & Config Store
                             |                    |
              Sensors <--- Event Normalizer ---> Luna Brain
                             |                    |
                             +---> Policy Gate <---+
                                     |
                             Privileged Executor
                              /      |        \
                         nftables   WAF      IDS/Proxy adapters
                                     |
           Local Host / Edge Agent / Routed Gateway (chosen at install)

           [FUTURE / NOT PART OF STANDALONE BASELINE]
           Optional NetBird Connector: peer/group/policy inventory,
             secure management path; NEVER owner of Luna policy.
```

## Current boundaries

- This repository now includes an **experimental, non-enforcing policy validator and CLI preview only**. No nft compiler/apply, WAF engine, privileged daemon, installer, infrastructure, logo asset or working firewall exists.
- Existing Luna/Selene Sentinel work is **research input**; do not copy live logs, agent memory databases, credentials, user data or production automation into this repository.
- Existing production Light Remote, NetBird, VPS-ARM, HOMESV and VPS-AMD are **out of scope for rule changes** until a lab and owner-approved rollout plan are ready. NetBird integration development itself is **out of scope until standalone gate acceptance**.
- Licensing, product name/trademark/domain availability, commercial Cloud terms, and vendor integration obligations remain subject to review.
