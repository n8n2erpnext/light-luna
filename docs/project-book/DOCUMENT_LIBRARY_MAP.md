# Light Luna Document Library Map

Status: canonical navigation | Date: 2026-10-13
Scope: document ownership and historical path mapping

## Entry route

[Governance](LIBRARY_RULES.md) → [Project Book](LIGHT_LUNA_PROJECT_BOOK.md) → this map → [Code Map](LIGHT_LUNA_CODE_MAP.md) → [Source Catalog](SOURCE_CATALOG.md) → original code/verification. Use [Worklog](LIGHT_LUNA_WORKLOG.md) for chronology.

## Current legacy paths (do not move yet)

| Existing path | Classification / authority |
|---|---|
| [00 Vision & Brand](../00-VISION-BRAND.md) | Design decisions + positioning |
| [01 Architecture](../01-ARCHITECTURE.md) | Architecture direction, recommended stack still partly proposed |
| [02 Firewall Engine](../02-FIREWALL-ENGINE.md) | Desired policy/enforcement contract, **not implemented** |
| [03 Luna Brain](../03-LUNA-BRAIN.md) | Proposed intelligence system, **not implemented** |
| [04 WAF/Threat Intel](../04-WAF-THREAT-INTEL.md) | Future integration design, **not implemented** |
| [05 NetBird](../05-NETBIRD.md) | **Deferred research only**, post-G-STANDALONE |
| [06 Cloud & Fleet](../06-CLOUD-FLEET.md) | Future services, **not implemented** |
| [07 Security](../07-SECURITY-SAFETY.md) | Mandatory security invariants and test intent, not proven release |
| [08 Packaging](../08-PACKAGING-HARDWARE.md) | Proposed packaging/hardware targets |
| [09 WebUI/CLI/API](../09-WEBUI-CLI-API.md) | Proposed interface; limited CLI now exists |
| [10 Moon Design](../10-MOON-DESIGN.md) | Proposed visual/semantic design |
| [11 Roadmap](../11-ROADMAP.md) | **Canonical phase order** and G-STANDALONE gate |
| [12 ADR register](../12-ADRS.md) | Decisions vs proposals vs open questions |
| [ADR-020 Go/Python language roles](../adr/ADR-020-go-core-python-optional.md) | Go-native core; Python optional unprivileged ML worker; `.mjs` proposal superseded |
| [13 References](../13-REFERENCES.md) | Dated external research bookmarks |
| [14 Initial handoff](../14-HANDOFF-2026-10-13.md) | Historical documentation-phase handoff; Git blocker superseded |
| [15 P0 implementation notes](../15-P0-IMPLEMENTATION-NOTES.md) | P0 code-slice handoff/notes, read with Code Map |
| [16 HOMESV P0 handoff](../16-P0-HOMESV-HANDOFF-2026-10-13.md) | Historical implementation checkpoint, not permanent code authority |

Root [README](../../README.md), [SECURITY](../../SECURITY.md), [CONTRIBUTING](../../CONTRIBUTING.md) and [go.mod](../../go.mod) are repository entry/build contracts, not free-form docs shelves.

## Future shelves (created only when required)

`docs/architecture/` for contracts and standalone phase proofs, `docs/adr/` for new ADRs, `docs/security/` for security protocols, `docs/design/`, `docs/release/` and `docs/history/` for historical agent/audit/verification materials. Do not manufacture empty placeholders or automatically move the legacy files.

## Generated source index

[Human index](SOURCE_CATALOG.md) and [machine index](source_catalog.json) include committed/working source and non-project-book documentation, with path and checksum. Current project-book files are intentionally indexed here and in its [README](README.md), not recursively in their own source catalog.

## Migration status

No legacy paths have been moved. **No implicit migration or broken-link rewrite is authorized.** Any future migration requires reference search, machine consumer analysis, old→new map, relative-link repair, verification, and a standalone Git commit.
