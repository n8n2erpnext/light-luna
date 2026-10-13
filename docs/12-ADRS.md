# 12 — Architecture Decision Register

| ADR | State | Decision and rationale |
|---|---|---|
| ADR-001 | **DECIDED** | Brand **Light Luna**, not FrogWall/Light Sentinel; Light Family product |
| ADR-002 | **DECIDED** | Software packaged firewall first, compatible hardware appliance later |
| ADR-003 | **DECIDED** | One distribution, user-selected Host/Edge, later Gateway mode |
| ADR-004 | **DECIDED** | Edge can manage remote agents without two physical NICs; packet inspection visibility is topology-dependent |
| ADR-005 | **DECIDED** | Traditional firewall features are mandatory independent of Luna Brain |
| ADR-006 | **DECIDED** | Luna Brain learns and proposes, independent policy gate authorizes; auto-ban not enabled in initial baseline |
| ADR-007 | **DECIDED** | NetBird optional integration, not firewall owner, no disabling NetBird ACLs to bypass conflicts |
| ADR-008 | **DECIDED** | WAF and CVE intelligence included in roadmap, with honest layer/verification boundaries |
| ADR-009 | **DECIDED** | Moon/orbital abstract branding with cyan crescent; blood moon and eclipse statuses |
| ADR-010 | **PROPOSED** | Go native agent/executor, JS `.mjs` brain, React WebUI, SQLite local |
| ADR-011 | **PROPOSED** | nftables Linux backend first, typed policy compiler and read-only coexistence audit |
| ADR-012 | **PROPOSED** | Coraza + OWASP CRS connector and optional Suricata module |
| ADR-013 | **PROPOSED** | `lunactl`, `lunalockd`, `luna-agent` internal naming, subject to package namespace review |
| ADR-014 | **PROPOSED** | Scoped Cloud signed threat intelligence and fleet management, offline local authority |
| ADR-015 | **OPEN** | Trademark/domain availability, software license, commercial pricing, OEM agreements |
| ADR-016 | **OPEN** | Minimum hardware profile validated through real benchmarks; inline packet path and acceleration strategy |
| ADR-017 | **OPEN** | Cloud tenant IAM/edition/licensing and NetBird API schema/version support matrix |
| ADR-018 | **OPEN** | Version 1 release definition and rollout site; production remains untouched |

## Change process

Any change to a DECIDED architectural constraint requires a new ADR, migration impact, rollback plan and explicit owner acceptance. Proposed items are not existing implementations. Operational safety guarantees must be proven in tests, not marketing copy.
