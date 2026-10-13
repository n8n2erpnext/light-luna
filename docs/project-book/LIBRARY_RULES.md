# Light Luna Documentation Library Rules

Status: canonical governance | Date: 2026-10-13
Scope: all durable Light Luna docs, AI-generated records, codemaps and evidence
Primary sources: [LightBI library convention (adapted)](README.md), [Light Luna architectural decisions](../12-ADRS.md)

## 1. Read first, do not improvise

Before creating, editing, moving, renaming or deleting docs, read [README](README.md), [Project Book](LIGHT_LUNA_PROJECT_BOOK.md), [Library Map](DOCUMENT_LIBRARY_MAP.md), then search for the existing owning document. Prefer changing that owner over duplicating it.

## 2. Truth and provenance

Label each document as **canonical/current**, **decision**, **proposal/draft**, **verification/evidence**, **worklog/progress**, **handoff**, **historical/superseded**, or **machine evidence**. Date, scope, exact Git SHA, runtime target and source links belong near the top. Never promote an intended capability to implemented; distinguish code reachable from a development CLI from a packaged service running in production. Do not confuse branch/dirty-worktree claims with committed Git truth.

## 3. Reading protocol

For project model: Project Book. For module owner, entry point and side effects: Code Map **then actual source and tests**. For a path: Catalog. For a historical rationale: Worklog then the ADR/handoff. For a release: independent test, build, packaging and deployment evidence. If sources conflict, record the conflict rather than silently picking a convenient narrative.

## 4. Shelves and ownership

- `docs/project-book/`: governance, source catalog, current high-level truth, Code Map, Worklog.
- `docs/architecture/`: new lasting architecture/contracts and phase proofs.
- `docs/adr/`: future immutable one-decision-per-record ADRs.
- `docs/security/`: current security boundaries and test policies.
- `docs/design/`: product/web visual contracts.
- `docs/release/`: builds, packaging, signed release evidence.
- `docs/history/agent/handoffs/`: bounded continuity records; historical, never canonical by default.
- `docs/history/audits/`, `docs/history/verifications/`, `docs/history/progress/`: dated evidence and progress.

**Legacy file-path lock:** `docs/00-*.md` through `docs/16-*.md` remain in place; the [Library Map](DOCUMENT_LIBRARY_MAP.md) indexes them. Do not move or rewrite them for aesthetics. A later migration requires scanning all code/test/doc path consumers, a move map, repaired relative links and separate review.

## 5. Naming and new-doc test

Only create a document if there is a new durable decision, current contract, independent verification, or genuinely needed handoff. Avoid `FINAL`, `LATEST`, `TEMP`, duplicate summary, and one-file-per-chat. Use `ADR-NNN-topic.md`, `VERIFICATION-scope-YYYY-MM-DD.md`, `AGENT_HANDOFF_SCOPE.md` or a descriptive lowercase architecture filename. Do not renumber the historical ADR register.

## 6. Required header and body

New durable docs follow [DOCUMENT_TEMPLATE.md](DOCUMENT_TEMPLATE.md): status, date, scope, supersession, primary sources; purpose, context, finding, boundaries/non-goals, evidence (code/tests/SHA), actual status, source bookmarks and follow-up. Keep a clear distinction between statements verified by code and proposals. Write for an engineer or agent with **zero access to the authoring chat**.

## 7. Source links and catalog

Use relative repository links where possible; never fabricate a target. Upon creating a source file or doc, refresh both [SOURCE_CATALOG.md](SOURCE_CATALOG.md) and [source_catalog.json](source_catalog.json) using `python3 scripts/update_doc_catalog.py`. Project Book files are intentionally excluded to prevent self-referential churn and are indexed by [README](README.md) and [Library Map](DOCUMENT_LIBRARY_MAP.md). Never insert sensitive content into public catalog.

## 8. Code Map discipline

Track **entry point → imports/calls → authoritative module → side effects → tests → release status**. Reachable, planned, inactive, legacy and unverified must be different labels. Update Code Map when routes/modules/interfaces change; include tested SHA or state precisely that the map is a source-level snapshot. Static path existence alone is not runtime reachability.

## 9. Worklog and checkpoint

Keep one [LIGHT_LUNA_WORKLOG.md](LIGHT_LUNA_WORKLOG.md) for durable project chronology. Do not create a new worklog per session. Keep `.luna/CURRENT_CHECKPOINT.json` local, overwrite-only, short and secret-free. A handoff records exact objective, source SHA, work done, tests, blockers and next action; it cannot authorize live firewall testing, merge or release.

## 10. Safety and completion

Documentation actions do not grant permission to run `nft` changes, install daemons, probe NetBird APIs or alter production. P0–P4 standalone; NetBird integration prohibited until verified G-STANDALONE **and explicit owner approval**.

Before concluding: verify document owner/status, links and relevant paths, code/commit truth, catalog generation `--check`, `git diff --check`, tests when behavior changes, no secrets or raw logs, and no unrelated code mixed into a docs-only commit. Never erase historical evidence or conflate synthetic tests with production.
