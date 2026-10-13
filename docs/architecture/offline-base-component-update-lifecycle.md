# Offline Base Installation and Per-Component Update Lifecycle

Status: **DECIDED architecture specification from ADR-024; NOT IMPLEMENTED**
Date: 2026-10-13
Scope: shipping bundle, independent versions, local/cloud/offline updates, compatibility gates and rollback
Primary sources: [ADR-024](../adr/ADR-024-offline-base-independent-component-updates.md), [Packaging](../08-PACKAGING-HARDWARE.md), [Cloud intelligence](luna-cloud-intelligence.md), [Safety](../07-SECURITY-SAFETY.md), [component manifest schema](../../schemas/component-release-manifest-v1alpha1.schema.json)

## 1. Installer = real firewall plus Basic Brain, not a downloader

A **future release** contains, as one verifiable installer/offline archive, at minimum:

| Installed component | Required for base installation? | Purpose / authority |
|---|---|---|
| `light-luna-agent` App Shell | YES | Local API/auth, CLI/UI, signed update client, resource health, local lifecycle |
| `lunalockd` restricted nftables executor | YES for firewall product | Scope-limited netfilter owner, atomic policy actions and independent recovery watchdog |
| `light-luna-brain` BASIC Go process | **YES** | Deterministic baseline, multi-window reasoning, decision proposals, no elevated network authority |
| `knowledge-base` embedded verified snapshot | **YES** | Initial curated defensive facts, network semantics, selected safe scenarios, publisher/source metadata and version |
| `light-luna-ui` local static UI | YES in normal interactive edition | Use fully offline; account/cloud login never required |
| `waf`, `ebpf-observer`, `xdp-guard`, extra vendor feeds | NO | Optional modules independently installed and authorized |
| Python ML worker + model | NO | Optional expensive analysis; default Go Brain does not depend on it |
| Cloud enrollment key or login | NO | Created only when owner explicitly enrolls |

The basic packaged knowledge contains reviewed deterministic network/behavior foundations, safety invariants, attack classes/counterexamples, known safe test scenarios and a publisher freshness timestamp. It **must not falsely claim all world CVEs are current**. Optional bounded official source fetch after installation (without proprietary Cloud enrollment) can refresh supported facts, but offline-only must remain meaningful.

Installer/repair should check architecture/kernel, install ownership, reserve disk/rollback, configure minimal privileges, verify signatures and install atomic versions without live packet-policy mutation until owner-approved safe setup. Base Brain process must be health monitored but if it is unavailable the firewall remains enforced by the independent executor/kernel and the UI reports intelligence degraded.

## 2. Version ledger — independently track five axes

```text
Installation ID: local-only
App              : 1.4.0 (example)
Brain            : 1.2.3 (example; independent executable)
Policy Executor  : 1.1.2 (example; separately signed, privileged)
Knowledge        : 185   (example global immutable revision, signed)
Modules:
  WAF            : 0.6.1 (example)
  eBPF observer  : not installed
  XDP guard      : not installed
ML Model         : not installed
Protocol ABI     : 2     (example)
Cloud            : disconnected | connected | degraded | never-enrolled
Last verified knowledge timestamp: ...
```

Numbers above are **illustrative fictional values**, not release announcements.

`App` semver changes application binaries/interfaces. `Brain` semver changes deterministic detection/reasoning implementation or its contract. `Knowledge` monotonically versioned content revisions change priors/CVEs/scenarios, **not Brain program version**. `Model` weights use their own immutable version/measurement ID; optional module binaries have their own semver/ABI, and privileged executor changes are high-risk. Policy owner revision stays separate from any of these; a Cloud release never silently edits owner policy.

**Never infer compatibility because semver digits match.** Individual releases declare strict feature/ABI requirements and compatible peer version ranges. Validation must include architecture, kernel capabilities, schema format, install state, running capabilities and *all* dependencies. E.g. App 1.4.0 and Brain 1.2.3 are only allowed if the signed manifest says they interoperate.

## 3. Lifecycle after installation

```text
INSTALL OFFLINE BUNDLE
      |
      +-- Local nftables/owner policies + independent TTL/rollback
      +-- Basic Go Luna Brain + signed base knowledge
      +-- Local UI/CLI, monitoring, no Cloud enrollment
      |
   Owner chooses Cloud connection (optional)
      |
   Device authenticates, fetches signed update root/manifest
      |
   Planner compares EACH TRACK independently:
   - appAvailable vs appInstalled
   - brainAvailable vs brainInstalled
   - executorAvailable vs executorInstalled
   - modules by id/version
   - knowledgeRevision and relevant selected CVEs/scenarios
   - model version if ML enabled
      |
   Preflight compatibility + tenant/site/role + budgets
      |
   Download candidate delta with caps → verify → stage
      |
   No new compatible update? Maintain last known good
   Candidate compatible? Update only permitted component
      |
   Health + rollback lease + audit + independent status
```

The Cloud is NOT part of any packet decision or emergency recovery path. A client can remain disconnected forever and still inspect/audit/update from owner's **offline signed bundle** via the same local verifier. Periodic local threat intelligence fetch is an optional feature, not cloud availability proof.

## 4. Trust, update authority and keys

1. Separate **software code-signing policy** (App/Brain/Executor/native module) and **knowledge publishing policy** (CVE/scenario pack) where feasible; a curator/ML worker is not a code publisher. Require signature threshold/review for kernel-impacting code and any privileged executor.
2. Trust bootstrap in installer pins a root release key. Validate metadata freshness/anti-freeze, signatures, artifact hashes, target paths and rollback/revocation. Consider TUF-style signed metadata and key rotation rather than trusting HTTPS alone. No real keys are specified in this proposal.
3. Untrusted advisory RSS/JSON/model output cannot request App/Brain upgrades, patch installation, auto-ban, executable post-install scripts or privilege changes. Installers and modules have a typed allowlist with static artifact hashes.
4. Device enrollment authenticates client → Cloud for distribution, but local owner policy determines auto vs manual updates. Device security critical advisory may **notify**; it does not automatically override its owner-configured protected update windows.
5. Direct owner-imported offline bundle must pass **the same** signature and compatibility validation as Cloud-delivered content; there is no weaker offline escape hatch.
6. Knowledge rollback due to erroneous content must not be blocked by simplistic revision monotonicity: accept only a **newer signed revocation/rollback authorization metadata** that names safe target revision, never an unauthenticated downgrade. Keep audit chain and last-known-good.

## 5. Update examples (proposed operational expectations)

| Event | Which revision changes? | Which components restart? | Safety |
|---|---|---|---|
| New CVE/KEV/vendor advisory relevant to device | `knowledgeRevision` only | No firewall/Brain restart; atomic read-only pack swap | Unverified CVE local applicability stays UNKNOWN |
| Luna reasoning engine improvement | `brainVersion` | Brain process only, if compatible | Kernel owner rules stay; fallback and Brain rollback |
| New WebUI/API behavior | `appVersion` | UI/API or agent only as needed | Brain/version ledger untouched |
| WAF connector update | `moduleVersions.waf` | WAF adapter only where isolation permits | Connector scoped, not blanket proxy restart |
| Optional AI model improvement | `modelVersion` | ML worker only | Go Brain continues if ML unavailable |
| Privileged nftables executor change | `executorVersion` | Controlled privileged process lifecycle | Maintenance window + independent rollback + host reachability |
| Brain requires newer App protocol | Two independently versioned packages in one atomic migration plan | Coordinated controlled restart | Abort/rollback as one planned graph, never partially activate |

## 6. Proposed update state machine

`IDLE → CHECKING → VERIFIED → STAGED → PREPARED → ACTIVATING → HEALTH_CHECK → COMMITTED` or `ROLLED_BACK` / `FAILED_SAFE`.

- **CHECKING:** obtain metadata or offline bundle without trust; block unknown path/scheme/size, feed prompt injection.
- **VERIFIED:** publisher signatures, hashes, freshness, compatible machine and full dependency DAG; reject downgrade unless signed revocation authorization.
- **STAGED:** unpack in restricted inert staging dir with quotas, no untrusted install hook or shell script.
- **PREPARED:** protect owner policies, reserve audit/disk/rollback, migrate against copy, run synthetic Brain schema handshake.
- **ACTIVATING:** only switch the selected component(s) with a supervised, atomic pointer/update mechanism; privileged executor never replaced by unprivileged Brain.
- **HEALTH_CHECK:** independent process/ABI/evidence and network reachability; do not report successful simply because a process started.
- **COMMITTED:** update revision ledger atomically and produce receipt; prune older caches ONLY after approved retention.
- **ROLLED_BACK/FAILED_SAFE:** restore prior graph, preserve last-known-good kernel policy/TTL, mark degraded component and halt repeated bad updates.

Crash during every state must be replayable/idempotent; upgrade report includes each component's `from→to` revision, individual result, dependency graph, whether protection continued and reason on failure. Rollback authorization for corrupted/signed bad knowledge must be explicit and logged.

## 7. Storage limits and update budgets

Installed baselines have a bounded base knowledge pack and **relevant knowledge deltas**, not all NVD history. Data quotas follow [local memory design](luna-memory-lifecycle.md), but code binaries, old Brain revisions, signed packages and optional ML models require a **separate, explicitly configured update/download/rollback disk budget** and a hard safety reserve. Do not count static executable size as knowledge bytes. The 256 MiB/512 MiB/2 GiB *proposed data* profiles are **not installer size estimates**.

Schedule download during low utilization, bandwidth and battery/power policies; verify compressed and decompressed size; no use of live root filesystem beyond declared storage roots. When no free safe space, delay update and keep running known-good versions.

## 8. Display and diagnostics

Expected future WebUI *About / Components & Updates* and CLI commands:

```text
lunactl version --components   # proposed, not implemented
lunactl update status          # proposed
lunactl update check --component brain
lunactl update plan            # show full compatibility and restart graph
lunactl update apply --component knowledge
lunactl update import <signed-offline-bundle>
lunactl update rollback --component brain --to <installed-revision>
```

Show `App version`, `Brain version`, `Executor version`, `Knowledge revision`, module list, `Model version`, protocol ABI, last verified data timestamp, Cloud enrollment/connectivity and separately chosen auto-update policies. Different version axes must **never** be collapsed into one marketing version.

These commands are **future design**, not current usable CLI features.

## 9. Release validation suite

- First install with network physically disabled: Basic Brain starts from embedded knowledge, local UI/CLI works, firewall works only after owner-configured policies; no cloud login/NetBird dependencies.
- Reboot after 30/90 days offline; bounded memory, honest stale-CVE status, deterministic default detectors, manual signed offline knowledge update.
- Knowledge rev +1 with unchanged App/Brain; Brain +1 with unchanged App/knowledge; App +1 alone; module +1 alone; opt-in model update alone.
- Full dependency matrix mismatches: new Brain requires newer App ABI; reject partial graph, preserve known-good; coordinated migration only if atomic.
- Signed bogus/expired/rollback replay/compromised source, decompression bomb, disk full, power failure at each updater state, CPU/memory spike, writable-symlink/bundle traversal, network loss.
- Brain segfault or restart during active firewall policy; no packet gate dependence, independent TTL/unban and emergency admin access retained.
- Both `amd64`/`arm64` plus targeted kernels and version/skew matrix; optional modules without hard-required platform support must fail as `UNSUPPORTED` not fatal.
- Assert installing with Cloud connected cannot silently expand owner authority or turn `OBSERVE` into `SCOPED_AUTO`.
- Release must provide build/reproducibility/SBOM/signing artifacts and exact traceability from component to Git SHA.

## 10. Status and sequencing

This is a **design of a future complete installer/updater**, not an implemented binary/update client or signed package. P0 source currently includes an inert Go policy CLI only; the Brain and independent executor must be built and verified before offline installer acceptance. Preserve P0–P4 standalone Gate G-STANDALONE and deferred NetBird work; future Cloud features are optional work, never a way to skip standalone testing.
