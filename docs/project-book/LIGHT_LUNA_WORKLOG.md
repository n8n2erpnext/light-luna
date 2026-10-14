# Light Luna Worklog

Status: chronological worklog; not canonical runtime evidence
Date: 2026-10-13
Scope: durable decisions, verified engineering milestones and blockers

## 2026-10-13 — Documentation baseline, standalone-first lock

- Architecture, security, stages P0–P7 and explicit G-STANDALONE were written in [legacy docs](DOCUMENT_LIBRARY_MAP.md). NetBird connector remains deferred until G-STANDALONE and owner approval.
- Git `main` published to public `n8n2erpnext/light-luna` at `66920c2`; owner later confirmed repository may remain **public**.

## 2026-10-13 — HOMESV P0 source slice

- HOMESV became primary development environment. Feature branch `feat/p0-standalone-foundation@9ae27b267e72e465ea7808c3936fabcf6b0cac05`.
- Go CLI PolicyBundle validator and inert preview added. HOMESV `go test ./...`, `go vet ./...`, validate/preview, Linux amd64/arm64 builds passed. No firewall runtime or release.
- Source branch propagated to GitHub with SHA-256 verified Git bundle via VPS-ARM. GitHub PR creation via connector returned 403; remote branch exists. See [historical handoff](../16-P0-HOMESV-HANDOFF-2026-10-13.md).

## 2026-10-13 — Documentation governance adopted from LightBI

- Inspected read-only VPS-ARM LightBI document library in `LightBI-docs-cli-lightbi/docs/project-book/`: its `LIBRARY_RULES`, `DOCUMENT_LIBRARY_MAP`, `DOCUMENT_TEMPLATE`, `LIGHTBI_CODE_MAP`, Project Book and source catalog mechanism.
- Adapted a **new Light Luna-owned** library with [governance](LIBRARY_RULES.md), [navigation](DOCUMENT_LIBRARY_MAP.md), [Project Book](LIGHT_LUNA_PROJECT_BOOK.md), [Code Map](LIGHT_LUNA_CODE_MAP.md), machine truth status and generated SHA source catalog.
- Existing `docs/00`–`docs/16` paths kept unchanged. Future docs use functional shelves and authority labels; no LightBI code/doc copied wholesale.

## 2026-10-13 — Go/Python component decision

- Owner accepted the Go-native security/backend/Brain direction and explicitly allowed Python in components that benefit from its ML/data-science ecosystem.
- Recorded [ADR-020](../adr/ADR-020-go-core-python-optional.md): Go deterministic Brain remains independent of an optional sandboxed non-root Python worker. `.mjs` is no longer the default Luna Brain runtime.
- This is an **architecture decision only**; no Brain, worker, Python runtime dependency, firewall enforcement or production changes were made.

## 2026-10-13 — eBPF/XDP extension decision

- Reviewed [Cloudflare L4Drop 2018](https://blog.cloudflare.com/l4drop-xdp-ebpf-based-ddos-mitigations/), Linux kernel eBPF/XDP documentation, cilium/ebpf and libxdp multiprogram coexistence guidelines.
- Recorded [ADR-021](../adr/ADR-021-ebpf-observability-xdp-optional.md): nftables authoritative; optional read-only eBPF tracing; XDP opt-in ingress mitigation only after independent gate/recovery and owner approval; existing XDP programs must never be replaced.
- HOMESV read-only link/kernel inventory confirms the machine has Linux 6.8 kernel/BTF but **does not prove safe/native XDP support**. No Light Luna BPF program loaded, attached or built; do not publish user infrastructure-specific device details in the repository.

## 2026-10-13 — Legacy Luna source audit and Brain V2 redesign

- Performed **read-only** VPS-ARM inspection of the original Python Luna/Selene service sources, live systemd unit metadata, source-only local Git checkout, cognitive JSONL schemas/counts, SQLite schema/size and intelligence adapters. No live user data rows exported and no configuration changed.
- Verified three generated cognition zones: 11,980 reasoning, 17,115 knowledge, 5,863 context records; generator creates repeated template combinations to fill memory budget, not neural model weights. Source audit lists site-scoped decision thresholds, pattern scoring/decay and missing evidence gates.
- Chose [ADR-022](../adr/ADR-022-rebuild-luna-brain-governed-knowledge-migration.md): rebuild Go-native Brain V2, migrate reviewed/deduplicated facts and tests **without copying live production DB or generated bulk context**. Original Sentinel stays independent.
- Created [Brain V2](../architecture/luna-brain-v2.md), [Threat Knowledge Pipeline](../architecture/threat-intelligence-knowledge-pipeline.md) and two versioned advisory knowledge/proposal JSON Schema drafts (not wired to runtime). Future Cloud learns from opt-in reviewed and privacy-minimized evidence, not raw traffic or automatic global ban policies.
- Site-specific legacy production status and all security actions left unchanged. Brain implementation and G-STANDALONE remain open.

## 2026-10-13 — Design review: hemispheres, training, memory and Cloud

- Reviewed existing 3-region Sentinel corpus semantics against the requested two-brain metaphor; recommended **logical analytical vs hypothesis reasoning** roles rather than actual neural hemispheres, preserving independent critic and Policy Gate.
- [ADR-023](../adr/ADR-023-dual-brain-memory-cloud.md) remains **PROPOSED** pending owner approval. Training a core neural model is not a baseline prerequisite; small optional ML worker requires labelled benchmarking.
- Proposed bounded client storage profiles, detailed quota/compaction and stale-evidence rules in [Local Memory Lifecycle](../architecture/luna-memory-lifecycle.md).
- Proposed offline-first, signed update/delta and privacy-preserving future multi-tenant service in [Cloud Intelligence](../architecture/luna-cloud-intelligence.md). No Cloud service, model, database size caps or client sync actually implemented.

## 2026-10-13 — Owner approves offline Basic Brain and split release versions

- Clarified product contract: Light Luna installs as a **self-contained, Cloud-optional firewall with mandatory compact Go Basic Brain and baseline knowledge**, not a Cloud-dependent empty shell. Brain failure still cannot halt or loosen independent firewall policy/TTL.
- [ADR-024](../adr/ADR-024-offline-base-independent-component-updates.md) DECIDED separate `appVersion`, `brainVersion`, `executorVersion`, module versions, `knowledgeRevision`, optional `modelVersion`, compatibility matrix/ABI, signed Cloud/offline update bundles and per-component rollback.
- [Lifecycle](../architecture/offline-base-component-update-lifecycle.md) and two static JSON schema contracts plus a fictional offline-state fixture were created; these are design only and not a running updater, signed artifacts or executable firewall release.
- Preserve P0–P4 standalone-first, defer optional Cloud service and NetBird integration, and test first-install with no network before any release claim.

## 2026-10-14 — Cloud portability and migration-first architecture

- Owner required Light Luna Cloud to **scale and move VPS/provider/stack** with minimal client disruption. Recorded [ADR-025](../adr/ADR-025-portable-scalable-cloud.md) DECIDED portability acceptance, leaving specific providers/thresholds and recovery SLOs subject to testing.
- Designed [modular scalable Cloud](../architecture/cloud-portability-and-scale.md) (Go stateless API, workers, single Postgres primary, S3-compatible immutable objects, Compose-first → k3s) and [migration runbook](../architecture/cloud-migration-and-recovery.md) (restore/PITR, old-primary fencing, signing keys/device identities, no split brain and verified handover).
- Introduced a **synthetic schema + non-approved inventory fixture** for validating future migration preparation without recording customer credentials/logs. No Cloud instantiated, backups initiated or infrastructure changed.

## 2026-10-14 — Owner decides unified account and CrowdSec-style Cloud fleet management

- Reviewed official CrowdSec Console enrollment and polling-based decision-management docs as **product precedent**, not an exact implementation dependency or promise of identical features.
- [ADR-026](../adr/ADR-026-unified-account-remote-fleet-control.md) sets one Cloud-account human identity for enrolled local and Cloud WebUIs; preserve offline bootstrap owner/emergency recovery and avoid Cloud-password replication to client.
- [Architecture](../architecture/unified-account-cloud-fleet-control.md) defines explicit device enrollment and remote-control authorization, RBAC, device outbound polling, desired vs actual config, compare-and-swap revisions, local Policy Gate, approval, expiry, audit and conflict receipts; Cloud cannot root-shell local nodes.
- Added design-only typed fleet intent JSON Schema and synthetic intent fixture; no accounts, Cloud API, local WebUI or firewall changed.

## 2026-10-14 — CrowdSec behavior/scenario comparison and federated immune-style direction

- Reviewed [CrowdSec scenarios](https://docs.crowdsec.net/docs/log_processor/scenarios/intro/), [Central API contribution](https://docs.crowdsec.net/docs/central_api/intro/), [community blocklists](https://docs.crowdsec.net/docs/central_api/community_blocklist/) and [Remediation Sync beta](https://docs.crowdsec.net/u/console/remediation_sync/).
- [ADR-027](../adr/ADR-027-collective-threat-immunity.md) DECIDED owner direction: opted-in client sends bounded **subject + behavior + scenario + evidence and local action/outcome**, Cloud vets/deduplicates and issues **signed scoped watch/curated bulletins** so other nodes recognize attack-like behavior early. No automatic global ban from one client.
- [Federated design](../architecture/federated-immune-threat-intelligence.md) distinguishes local mitigation receipts vs proven maliciousness, official scenario hashes vs custom, public indicator provenance/TTL, corrections/retractions and anti-Sybil. Added two static candidate JSON schemas and synthetic fixtures. Nothing executed on host firewall or Cloud.

## 2026-10-14 — Owner requires Light Luna self-defense and host-compromise containment

- Investigated authoritative eBPF/Tetragon/Falco runtime security telemetry, Linux IMA/fs-verity integrity enforcement and ClamAV on-demand/on-access fanotify scanning. **XDP alone is a packet hook, not a host malware scanner**; severe root/kernel compromise can falsify all in-host sensors.
- [ADR-028](../adr/ADR-028-self-defense-host-compromise.md) DECIDED new guardian/threat model: tamper/signed update protection, limited eBPF process/file/network observation, optional bounded AV, safe local egress/lateral containment after Policy Gate, independent recovery; Cloud rejects/quarantines compromised contributions and never accepts raw malware/executable samples from clients by default.
- [Architecture](../architecture/host-self-defense-and-compromise-containment.md) and static HostIntegrityReport schema + synthetic compromised-device fixture added. No malware scanning, system installation, firewall attach, Cloud upload or code enforcement performed.

## 2026-10-14 — First executable P0 Go component skeleton

- Implemented a **genuine runnable but non-enforcing** foundation slice in [local runtime design/verification](../architecture/local-runtime-foundation.md): `internal/foundation` bounded event contracts; `internal/brain` deterministic, synthetic-only auth-burst observation; `internal/guardian` UNKNOWN coverage; `internal/gate` advisory-only/no-write; `internal/offline` first observable call chain and component snapshot; `cmd/luna-brain` separately buildable, one-shot replay; enhanced `lunactl` system snapshot/brain replay.
- Hardened `internal/policy.Decode` by reusing `internal/strictjson`: recursively rejects duplicate JSON object keys, unknown fields, trailing data, bounded sizes; new nested-duplicate regression tests and replay positive/negative tests. Existing policy validation/preview preserved.
- Ran Go unit tests and vet, standalone Brain replay/snapshot, and separate `linux/amd64`+`linux/arm64` cross-builds with network-disabled Go tooling. **Not a kernel/firewall/Cloud or malware scanning test**; no live infrastructure touched.

## 2026-10-14 — P0 unhooked nftables text compiler

- Implemented `internal/nftpreview.Compile` and `lunactl firewall policies nft-preview`: policy validation → stable source alias expansion/order → scoped `table inet luna_p0_preview` **regular chains without hooks**, JSON plan reporting **no kernel effect, no syntax check**. No privileged executor, base-chain policy, default DROP, emergency route, Cloud or NetBird.
- Added deterministic tests for rule priority, protocol/family combinations, unsafe mismatch rejection, expansion caps, and malicious note non-interpolation. Existing parser/Brain/CLI tests and Go vet passed.
- Verified HOMESV nftables v1.0.9 availability but `unshare --user --map-root-user --net` not permitted. **Skipped `nft -c` instead of touching parent-host network namespace**, recorded isolated-VM grammar test as next dependency; see [compiler truth/verification](../architecture/p0-nft-unhooked-compiler.md).

## 2026-10-14 — P0 synthetic Flow Simulator (Go code and tests)

- Implemented `internal/flowsim` and `lunactl firewall policies simulate <policy> <flows>`, strict 256 KiB/128-case input and deterministic rule intent with IPv4/IPv6, TCP/UDP/ICMPv6, CIDR aliases, log nonterminal behavior and input/output default differences.
- Source-level unit/CLI tests cover source trust assumptions, malformed/duplicate cases, port/family rejection and **management-critical flow lockout warnings**; effective verdict is always **UNKNOWN** and the simulator never imports unrelated kernel/firewall state. [Full tested scope and gaps](../architecture/p0-flow-simulator.md).
- Kept real nft apply, Cloud, device enrollment and NetBird absent. P0 and G-STANDALONE remain OPEN/BLOCKED; no production firewall or existing Sentinel/Light Remote data modified.

## 2026-10-14 — HOMESV direct GitHub SSH and local control-gate code

- Inspected HOMESV's existing per-repository deploy key and confirmed `git ls-remote` initially failed solely on missing SSH GitHub `known_hosts` entry. Matched GitHub's published **Ed25519 SHA256 host fingerprint** before writing the public host key; then verified remote branch reads with the existing local key. Subsequent commits are pushed **directly from HOMESV**, not relayed through VPS-ARM. Private key bytes were not printed, copied or stored in the repo.
- Implemented [`internal/controlgate`](../../internal/controlgate/control.go) and [CLI selftest](../../internal/controlgate/selftest.go): short-lived HMAC-authenticated, local-owner-only proposals pinned to site/actor/policy SHA256 and revision; nonce+request replay cache, clock window, fail-closed store capacity, and explicit denial of signed APPLY. [Exact implementation and limitations](../architecture/p0-local-control-gate.md).
- Added negative/concurrency tests for expiry/future, forged MAC, local identity, stale/different policy, weak key, duplicate nonce/request, and attempted enforcement. Normal Go unit tests, vet, local CLI selftest, catalog and amd64/arm64 builds passed. `go test -race` **not verified**: CGO requires gcc, which is unavailable on HOMESV. Run race instrumentation in CI/build runner later. No actual IPC/OS peer auth, privileged command, Cloud call, NetBird integration or firewall modification.

## 2026-10-14 — Real Linux UDS P0 local IPC with cross-process test

- Implemented [`internal/localipc`](../../internal/localipc/unix_linux.go) with owner-private directory and socket permissions, Linux peer UID verification (`SO_PEERCRED`), one 4-byte-prefixed bounded JSON request and receipt per connection, max 8 concurrent clients and 3-second IO limits; parser rejects extra fields and duplicate keys. This wires the existing authenticated denial-only `controlgate.Boundary` without an executor or new privileges.
- Added `lunactl control ipc-selftest <policy>` and [selftest implementation](../../internal/localipc/selftest_linux.go). In-CLI client/server are same-process goroutines over a real Unix socket; [Go integration tests](../../internal/localipc/unix_linux_test.go) also launch **a separate subprocess client** and cover UID denial, insecure/symlink socket directory, frame oversize, malformed JSON, valid preview/replay/tamper/apply refusal. Documentation and current Code Map updated every increment.
- **Honest status:** not production IPC/service identity; allowed Linux UID could host malicious same-UID process. No durable replay, external key broker, root privileged daemon, netfilter modifications, Cloud connection, actual scanner or NetBird integration. Cross-process tests are not multi-user/UID isolation proof; require isolated VM and GCC-enabled race CI before a production service.

## 2026-10-14 — Two independent OS processes and bounded restart-persistent replay journal

- Added [`cmd/luna-gate`](../../cmd/luna-gate/main.go) and [`cmd/luna-agent`](../../cmd/luna-agent/main.go) as separately compiled Linux binaries launched explicitly for a short nonroot developer test lease (not installed daemons). The same-UID shared secret used only for test is generated by [devkeys](../../internal/devkeys/keys_linux.go), kept in 0700/0600 scratch paths, with no credentials in Git.
- Extended [Control Gate](../../internal/controlgate/control.go) with [Linux locked journal](../../internal/controlgate/journal_linux.go): strict bounded JSONL nonce + request ID records; exclusive nonblocking flock, append + fsync, restart hydration of unexpired tickets, corruption/permissions/full-store fail closed. **Not** multi-user isolation, power-loss audited storage or privileged authorization.
- [End-to-end test script](../../scripts/test_p0_gate_processes.sh) demonstrated true separate Gate/client PIDs, signed preview, denied replay and apply, restart replay denial and refusal to start after corrupting the journal. **Full Go tests repeated three times, vet, four Linux binaries x amd64/arm64 crossbuild and 99 indexed doc links PASS**; race detector still unavailable (no gcc). No live host nftables, Cloud, NetBird or pre-existing Sentinel modified. [Full verification boundaries](../architecture/p0-two-process-gate-and-journal.md).

## 2026-10-14 — Ed25519 request credentials, group Unix IPC and direct HOMESV tests

- Added Ed25519 `SignEd25519/NewEd25519Boundary` to [controlgate](../../internal/controlgate/control.go), a pinned public verifier with no access to the Agent private key; distinct domain from old HMAC (no fallback, reject mixed). Retained locked fsynced replay journal for new Ed request path and denial-only APPLY.
- Added `internal/identitykeys` public/private files in separate 0700 directories, `localipc.ListenGroup` 0710/0660 pre-provisioned Gate group socket policy and client Gate UID pinning. [Current engineering contract](../architecture/p0-ed25519-agent-gate-identity.md) and [E2E test](../../scripts/test_p0_ed25519_identities.sh) verify separate processes under **same** nonroot HOMESV UID, signature negative cases, restart replay and corrupted journal refusal.
- **HOMESV verification:** full Go tests three times PASS, Go vet PASS, Ed25519 and legacy HMAC E2E PASS, 4 binary cross-builds x amd64/arm64 PASS, source catalog 107 indexed docs and links PASS. No real different-UID Linux test, service account provisioning, trusted owner login, kernel rule changes or Cloud/NetBird deployment. Multi-UID positive/negative isolation testing is a future disposable-VM gate. Ed25519 tests are **source proof**, not production identity verification.

## 2026-10-14 — HOMESV LXD lab preparation, distinct-UID test BLOCKED

- Reconnected to authorized HOMESV (Light Remote). Read-only checks showed Ubuntu 24.04.5, Linux 6.8, around 7.1 GiB RAM available and 105 GiB disk free. LXD client launchers exist, but the LXD snap is **not installed** and `lxc` cannot access a ready daemon. Host physical/VPN interfaces were observed read-only; no network or firewall changes made.
- Attempted to install LXD snap via governed execution and was blocked by the remote execution safety layer. Explicitly **did not retry the privileged command through PTY or another tool**. No LXD daemon, container, storage pool, profile, new host user/group or networking was created.
- Prepared [no-NIC lab scripts](../../scripts/lab/lxd/README.md) for a dedicated resource-bounded instance and cross-UID Ed25519 Agent/Gate positive/negative matrix. [Full acceptance contract](../architecture/p0-lxd-cross-uid-lab.md). All five scripts passed shell `bash -n`; generated documentation catalog validated **114 indexed sources and links**, and existing Go unit tests `go test -count=1 ./...` passed. No actual guest execution, LXD CLI semantics, distinct-UID behavior or kernel acceptance was verified. Running the otherwise read-only LXD host preflight script was also blocked by the remote safety check; the earlier individual read-only LXD availability checks are the source of the blocker assessment.

## 2026-10-14 — LXD installed, no-NIC instance and distinct-UID E2E PASS\n\n- Connected to authorized HOMESV through Light Remote. Direct operation with approved `lxd`, `sudo-on-demand` and `package-manager` capabilities installed official LXD snap **5.21.8** and initialized LXD. The minimal initialization unexpectedly created `lxdbr0`; verified it was newly created, removed its default-profile NIC then deleted **that newly-created bridge**. No existing physical/VPN interface, Docker or NetBird settings were intentionally modified.\n- Imported the Ubuntu 24.04 official `ubuntu:24.04` image to local alias `light-luna-ubuntu2404`, initialized storage pool `default` (**dir** backend, no enforced per-container disk quota), and created dedicated unprivileged `light-luna-p0-lab` with **only a root disk**, 2 vCPU, 1536MiB RAM, 256 processes, no NIC, no nesting. Verified expanded instance config and network inventory (no LXD managed bridges after cleanup).\n- Fixed [test harness](../../scripts/lab/lxd/run-tests.sh) to use repo-local temporary compiled binaries because strict LXD snap cannot access host `/tmp`; added opt-in `LUNA_LXD_USE_SUDO=1` for existing authorized sudo without putting the developer into the root-equivalent `lxd` group.\n- **Actual cross-UID guest test PASS:** separate `lunaagent` and `lunagate` UIDs, successful Ed25519 owner-demo preview only, signed APPLY rejected, expected-Gate-UID mismatch rejected, outsider without IPC group denied, different UID inside IPC group **even with copied synthetic key** denied by `SO_PEERCRED`, replay after normal Gate restart rejected, corrupted journal denied Gate startup. Test stdout: `LXD_DISTINCT_UID_POSITIVE_NEGATIVE_RESTART_PASS` and `LXD_CROSS_UID_TEST_RUN_COMPLETED`.\n- **Post-run cleanup/observability:** removed the deliberately copied synthetic Ed25519 private key from the wrong-UID test directory inside the LXD guest. Confirmed guest has only loopback (127.0.0.1/::1); resident RAM about 47.56MiB after tests. The LXD-managed `table inet lxd` exists but is empty (no chains or rules), and `lxdbr0` does not exist. The lab remains RUNNING for inspection; it is a one-shot test instance with intentionally corrupted replay journal, not a ready service.
- Original same-UID code tests and explicit owner login are separate. This tests process UID isolation **inside LXD container sharing the host kernel**, not a separate-kernel nftables, eBPF/XDP, rollback, malicious-root or production account-broker security gate. Container is one-shot (test users and intentionally corrupted journal remain), not a releasable node.\n\n## 2026-10-14 — HOMESV separate-kernel LXD VM nft syntax + synthetic watchdog PASS twice

- Verified HOMESV `/dev/kvm` and Intel VT-x. Cached official `ubuntu:24.04 --vm` image alias `light-luna-ubuntu2404-vm` (remote fingerprint `d4ff18198fb2ed8c88d408421e610c143cf1f5c1c0178e57178fb586880c3e56`). Created LXD VM profile `light-luna-p0-kernel-vm` (2 vCPU, 2048MiB, disk-only), created and booted real KVM **VIRTUAL-MACHINE** `light-luna-p0-nftvm` with **no NIC**. Guest Ubuntu 24.04 kernel `6.8.0-142-generic` differs from HOMESV `6.8.0-146-generic`. Kept earlier unprivileged different-UID container untouched.
- Added [VM runner](../../scripts/lab/vm/run-tests.sh), [guest test](../../scripts/lab/vm/guest-nft-test.sh), [fresh VM creator](../../scripts/lab/vm/create-vm.sh), explicit [cleanup](../../scripts/lab/vm/delete-vm.sh) and [runbook](../../scripts/lab/vm/README.md). Both successful runs checked exact VM profile, KVM separate kernel and no NIC before transferring **unhooked** Light Luna Go nft preview, then `nft -c` PASS, malformed-syntax negative PASS, guest netns baseline ICMP PASS, **independently authored synthetic** drop-hook with nft counter PASS, and timed separate-process watchdog revert restoring packet PASS.
- The synthetic guest test killed a stand-in controller, not a production executor. VM-only nft tables and temporary namespaces were cleaned; the guest ruleset remains empty. Host LXD-managed nft table `inet lxd` observed empty, with no `lxdbr0` managed bridge. Host firewall never intentionally modified. LXD `dir` storage has no proven disk quota; VM RAM ~404MiB after first run.
- **Source validation:** full `go test -count=1 ./...` PASS, `go vet ./...` PASS, `bash -n` all VM scripts PASS, documentation catalog and links **120 indexed sources PASS**, `git diff --check` PASS. Both real VM test runs passed.
- **Limitations:** no real Light Luna hooked nft compiler, executor/transactions, owner-presence/rescue route, reboot/power-loss/partition rollback or Docker/UFW/NetBird coexistence. Full VM/kernel test evidence and accepted scope [here](../architecture/p0-lxd-kernel-nft-rollback.md). G-STANDALONE remains blocked.

## Standing limitations

No production server, WebUI, autonomous firewall, WAF or NetBird connector. Development tests do not satisfy P0 or G-STANDALONE. Future handoffs are historical, Code Map/Git/test evidence take precedence.
