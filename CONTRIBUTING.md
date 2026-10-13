# Contributing (P0 standalone-first development)

Documentation and engineering changes must follow [the Light Luna Library Rules](docs/project-book/LIBRARY_RULES.md), [Project Book](docs/project-book/README.md) and [Code Map](docs/project-book/LIGHT_LUNA_CODE_MAP.md). Label every claim as implementation evidence, accepted decision, proposed feature, or historical handoff. Place new durable documents in the governed shelf, update the source catalog with `python3 scripts/update_doc_catalog.py`, verify with `python3 scripts/update_doc_catalog.py --check --check-links`, and never bundle docs organization changes with unrelated runtime changes.


Propose changes with a short problem statement, affected threat boundary, risk/rollback implications, schema impact and test expectations. Respect the ADR register; label assumptions as PROPOSED/OPEN. Use synthetic IPs from RFC 5737 (192.0.2.0/24, 198.51.100.0/24, 203.0.113.0/24). No production credentials, raw traffic logs or agent memory databases. No privileged firewall mutations without an isolated network lab and explicit owner sign-off.
