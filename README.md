# Light Luna

**Orbit. Observe. Defend.**  
**Intelligent Software Firewall & Autonomous Defense**  
*A Light Family product by THAI DUY Digital*  
**Status:** Architecture / documentation baseline, 2026-10-13. **No runtime product, firewall installation, cloud service or auto-ban is claimed.**

Light Luna is a packaged **software firewall** for Linux servers and future compatible security appliances. It combines traditional, predictable firewall controls with **Luna Brain**, a bounded, evidence-driven security decision system. It works without Cloud, NetBird or Luna Brain; those features provide additional intelligence and fleet connectivity.

## Product principles

1. **Firewall first:** deterministic rules, aliases, stateful filtering, zones, interfaces, NAT where applicable, schedules, audit and recoverable configuration; AI is optional.
2. **Install anywhere, select deployment mode:** **Host** (local protection), **Edge** (log/telemetry controller with remote agents, not necessarily inline) and **Gateway** (routing/forwarding appliance, optional for compatible hardware). Mode is a configuration, not a separate codebase.
3. **Defense in depth:** L3/L4 firewall + layer-7 WAF/reverse-proxy adapters + optional IDS/IPS + Luna Brain.
4. **Autonomy with bounded authority:** detect → investigate → decide → independent policy gate → execute → verify → learn; every execution has provenance, expiry, replay protection and rollback.
5. **Offline-first safety:** no dependency on an external API or LLM to permit routine traffic; existing policies and ban expirations continue if Cloud fails.
6. **Non-invasive coexistence:** never flush other components' firewall tables, interrupt Docker/NetBird/UFW/Cloud security groups, or disable NetBird ACL enforcement as a workaround.
7. **Privacy and provenance:** source IP, proxy identity, version-to-CVE applicability, learning labels and telemetry gaps must be explicit; no default sharing of raw logs or customer IP histories to Cloud.
8. **Releasable product:** two-architecture signed packages, clean install/upgrade/uninstall, service watchdog, safe migrations, rollback, supply-chain attestations and compatibility matrix.

## Naming

**Product:** Light Luna (two words). **Brain:** Luna Brain. **Cloud:** Light Luna Cloud. **CLI:** `lunactl`. **Privileged executor:** `lunalockd` (working name). **System agent:** `luna-agent` (working name). **Appliance:** Light Luna Box. **Firmware/project image:** Light Luna OS (future, only after hardware feasibility).

## Read this first

| Document | Purpose |
|---|---|
| [Product vision](docs/00-VISION-BRAND.md) | Name, lunar metaphor, personas, market/product boundary |
| [System architecture](docs/01-ARCHITECTURE.md) | Modes, planes, interfaces, authority and data flow |
| [Firewall specification](docs/02-FIREWALL-ENGINE.md) | Rules, aliases, NAT, compiler, transactional enforcement |
| [Luna Brain](docs/03-LUNA-BRAIN.md) | Observation, decisions, learning, confidence and trust |
| [WAF, IDS & CVE](docs/04-WAF-THREAT-INTEL.md) | L7, traffic inspection, threat feeds, software provenance |
| [NetBird integration](docs/05-NETBIRD.md) | Stack/API research, access model and compatibility |
| [Cloud & Fleet](docs/06-CLOUD-FLEET.md) | Remote controllers, security, lifecycle and privacy |
| [Security & safety](docs/07-SECURITY-SAFETY.md) | Threat model, guardrails, lockout rescue, abuse defense |
| [Packaging & hardware](docs/08-PACKAGING-HARDWARE.md) | Linux deployment, future appliance, upgrade/recovery |
| [WebUI, API & CLI](docs/09-WEBUI-CLI-API.md) | Operator UX, RBAC, control API and execution tool |
| [Moon-state design](docs/10-MOON-DESIGN.md) | Brand and incident severity language |
| [Roadmap & test gates](docs/11-ROADMAP.md) | Milestones, acceptance criteria and anti-scope-creep |
| [Architectural decisions](docs/12-ADRS.md) | Decided / proposed / unresolved |
| [Research sources](docs/13-REFERENCES.md) | Public official documentation links and caveats |
| [Project handoff](docs/14-HANDOFF-2026-10-13.md) | Decisions, artifacts, GitHub blocker and next session |

## Initial topology

```text
          Optional Cloud Intelligence / Fleet Controller
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

           Optional NetBird Connector: peer/group/policy inventory,
             secure management path; NEVER owner of Luna policy.
```

## Current boundaries

- This repository starts as **design and technical documentation**; no firewall code, WAF engine, enforcement daemon, infrastructure, domain, logo asset, or deployment is claimed to exist.
- Existing Luna/Selene Sentinel work is **research input**; do not copy live logs, agent memory databases, credentials, user data or production automation into this repository.
- Existing production Light Remote, NetBird, VPS-ARM, HOMESV and VPS-AMD are **out of scope for rule changes** until a lab and owner-approved rollout plan are ready.
- Licensing, product name/trademark/domain availability, commercial Cloud terms, and vendor integration obligations remain subject to review.
