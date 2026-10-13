# 12 — Architecture Decision Register

| ADR | State | Decision and rationale |
|---|---|---|
| ADR-001 | **DECIDED** | Brand **Light Luna**, not FrogWall/Light Sentinel; Light Family product |
| ADR-002 | **DECIDED** | Software packaged firewall first, compatible hardware appliance later |
| ADR-003 | **DECIDED** | One distribution, user-selected Host/Edge, later Gateway mode |
| ADR-004 | **DECIDED** | Edge can manage remote agents without two physical NICs; packet inspection visibility is topology-dependent |
| ADR-005 | **DECIDED** | Traditional firewall features are mandatory independent of Luna Brain |
| ADR-006 | **DECIDED** | Luna Brain learns and proposes, independent policy gate authorizes; auto-ban not enabled in initial baseline |
| ADR-007 | **DECIDED** | NetBird optional integration **deferred until independently working Light Luna has passed G-STANDALONE and owner authorizes connector work**; not firewall owner; never disable NetBird ACLs to bypass conflicts |
| ADR-008 | **DECIDED** | WAF and CVE intelligence included in roadmap, with honest layer/verification boundaries |
| ADR-009 | **DECIDED** | Moon/orbital abstract branding with cyan crescent; blood moon and eclipse statuses |
| ADR-010 | **PARTIALLY SUPERSEDED** | Original proposal: Go native agent/executor, JS `.mjs` brain, React WebUI, SQLite local. **The `.mjs` Brain default is superseded by [ADR-020](adr/ADR-020-go-core-python-optional.md)**; React/TypeScript WebUI and SQLite remain proposed |
| ADR-011 | **PROPOSED** | nftables Linux backend first, typed policy compiler and read-only coexistence audit |
| ADR-012 | **PROPOSED** | Coraza + OWASP CRS connector and optional Suricata module |
| ADR-013 | **PROPOSED** | `lunactl`, `lunalockd`, `luna-agent` internal naming, subject to package namespace review |
| ADR-014 | **PROPOSED** | Scoped Cloud signed threat intelligence and fleet management, offline local authority |
| ADR-015 | **OPEN** | Trademark/domain availability, software license, commercial pricing, OEM agreements |
| ADR-016 | **OPEN** | Minimum hardware profile validated through real benchmarks; inline packet path and acceleration strategy |
| ADR-017 | **OPEN** | Cloud tenant IAM/edition/licensing and NetBird API schema/version support matrix |
| ADR-018 | **OPEN** | Version 1 release definition and rollout site; production remains untouched |
| ADR-019 | **DECIDED** | **Strict standalone-first sequence:** design/build/test/package independent Host and Edge, prove G-STANDALONE; NO NetBird code/API probing/compatibility integration tests/dependency before gate and explicit owner approval, including so-called read-only connector. The current NetBird document is future research only |
| ADR-020 | **DECIDED** | [Go-native Backend/Brain; optional isolated Python worker](adr/ADR-020-go-core-python-optional.md). `.mjs` is no longer default Brain runtime; WebUI React/TypeScript and SQLite remain proposed; Brain/executor authority separated |

| ADR-021 | **DECIDED** | [Optional eBPF observability, later opt-in XDP mitigation](adr/ADR-021-ebpf-observability-xdp-optional.md): nftables remains primary; never replace unknown XDP; kernel/driver matrix, Policy Gate, rollback and staged isolated tests mandatory; no implemented runtime |

| ADR-022 | **DECIDED** | [Rebuild Go-native Luna Brain V2](adr/ADR-022-rebuild-luna-brain-governed-knowledge-migration.md), migrate reviewed/typed knowledge and evidence safeguards, **not live Sentinel DBs or assumed neural weights**; Cloud sharing opt-in and locally governed |

| ADR-023 | **PROPOSED** | [Dual-role Luna Brain, bounded local memory and optional Cloud intelligence](adr/ADR-023-dual-brain-memory-cloud.md). Analytical Go core + optional hypothesis/ML plane, 256/512 MiB and 2 GiB experimental data caps, local enforcement, signed relevant knowledge delta, opt-in Cloud learning; pending owner acceptance and benchmark |

| ADR-024 | **DECIDED** | [Offline-complete base installer and independent App/Brain/Knowledge/module updates](adr/ADR-024-offline-base-independent-component-updates.md): package Basic Go Brain+verified baseline knowledge; Cloud optional; distinct version ledgers, compatible ABI, signed offline bundles, staged rollback; advanced ML remains optional; no shipping implementation yet |

## Change process

Any change to a DECIDED architectural constraint requires a new ADR, migration impact, rollback plan and explicit owner acceptance. Proposed items are not existing implementations. Operational safety guarantees must be proven in tests, not marketing copy.
