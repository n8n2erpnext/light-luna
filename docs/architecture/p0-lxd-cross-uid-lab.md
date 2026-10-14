# P0 LXD Cross-UID Lab — Offline, No-NIC Identity Test Plan

**Status:** **LAB DEPLOYED / real cross-UID Linux IPC matrix PASS** (2026-10-14); **NOT** a firewall/kernel acceptance or G-STANDALONE pass. The original unexecuted-lab notes below are historical.
**Scope:** HOMESV-hosted disposable *unprivileged LXD container* with separate Linux service users for Ed25519 Gate/Agent identity, Unix peer credentials, replay durability on normal Gate restart and negative tests.
**Supersedes:** no earlier architecture. Extends [Ed25519 Agent/Gate boundary](p0-ed25519-agent-gate-identity.md) and [two-process replay journal](p0-two-process-gate-and-journal.md).
**Code:** [host preflight](../../scripts/lab/lxd/check-host.sh), [container creator](../../scripts/lab/lxd/create-lab.sh), [offline test runner](../../scripts/lab/lxd/run-tests.sh), [guest cross-UID test](../../scripts/lab/lxd/guest-test.sh), [manual deletion](../../scripts/lab/lxd/delete-lab.sh), [lab README](../../scripts/lab/lxd/README.md).
**Safety:** [Library Rules](../project-book/LIBRARY_RULES.md), [Security](../07-SECURITY-SAFETY.md). No NetBird integration before G-STANDALONE.

## Actual HOMESV execution evidence (2026-10-14)

**LXD installed:** `lxd 5.21.8` snap `5.21/stable` on HOMESV through Light Remote with approved `lxd`, `package-manager` and `sudo-on-demand` capabilities. `sudo lxd init --minimal` initialized local pool `default` (`dir` driver). Minimal init unexpectedly generated `lxdbr0` and default-profile `eth0` NIC; both were **removed before the Light Luna instance was created**. After cleanup, `lxc network list` shows only existing physical interfaces, and default profile has root disk only.

**Live isolated container:** `light-luna-p0-lab` under `light-luna-p0-isolated`; Ubuntu 24.04 x86-64 image release serial `20260926`, base image fingerprint `ee3016f85cc452b34d0ad0ff5e2638948e6c905fa5068b68b36f45ef53fd628e`. Expanded LXD config verifies `security.privileged=false`, `security.nesting=false`, 2 vCPU, 1536MiB memory, 256 process cap, **devices root disk only, no NIC**. Pool `default` has dir driver (no proven disk quota).

**Executed E2E:** `scripts/lab/lxd/run-tests.sh`, called with explicit Go toolchain PATH and `LUNA_LXD_USE_SUDO=1`, built static `luna-gate` and `luna-agent`, pushed only binaries/synthetic policy to the container, and ran the guest-only test script as guest root. Separate guest `lunagate`, `lunaagent`, `lunaevil` and `lunaoutsider` UIDs executed the full matrix. Exact PASS markers:
```text
CROSS_UID_POSITIVE_AGENT_PASS
CROSS_UID_BAD_GATE_UID_DENIED
OUTSIDER_UNGROUPED_SOCKET_DENIED
WRONG_UID_EVEN_WITH_TEST_KEY_DENIED
JOURNAL_RESTART_REPLAY_DENIED
LXD_DISTINCT_UID_POSITIVE_NEGATIVE_RESTART_PASS
LXD_CROSS_UID_TEST_RUN_COMPLETED
```
An expected signed `APPLY_POLICY` test returned `REJECTED` with `kernelWrites=false`. Journal corruption caused startup refusal. **Test scope is limited to one-shot container process/peer-UID validation.** It does not prove host firewall unchanged at all layers, kernel independence, malicious root resistance, hardware power-loss replay durability or production authorization. No NetBird/Cloud integration occurred. Earlier notes below describe **previously blocked state**, not current status.

**Post-run host observations:** `lxc list` shows the instance `RUNNING`; actual guest `lxc info` reports only loopback (127.0.0.1/::1), around 47.56MiB current RAM; `lxc network list` has physical interfaces only and no managed LXD bridge. Host nftables has an **empty** LXD-owned `table inet lxd {}` (no rules/chains). This is not an audit of all preexisting host firewall state. The deliberate test-key copy placed in the rogue guest user's directory was **deleted inside the guest** after testing; the test script now explicitly removes it automatically on subsequent fresh labs. The instance is a **one-shot** test with intentionally corrupted journal; do not deploy or trust it as a live Gate service.

**Environment issues solved:** test runner initially lacked Go on command PATH; PATH corrected. LXD snap file access is isolated from host `/tmp`, so `run-tests.sh` now builds within a short-lived working-tree `lxd-artifacts.*` directory and automatically removes it.

## HOMESV original blocker (resolved after authorized retry)

- HOMESV Ubuntu 24.04.5 / Linux 6.8.0-146 / approx. 7.7GiB RAM, 7.1GiB available / 105GiB free filesystem. The `lxc`/`lxd` Ubuntu entrypoints are installed but `snap list lxd` reports **no LXD snap**; launching LXD client complains about missing installation/group membership.
- Host has a preexisting physical network and VPN interface. These must not be bridged or reconfigured for Light Luna P0.
- `sudo -n` currently works for the authorized development identity, but the remote safety layer blocked an explicit LXD snap installation call. **No LXD installation, init, bridge, profile, storage pool, VM or container was performed.** Using PTY to evade that denial is not permitted.
- This result is an **infrastructure BLOCKER**, not a failed Light Luna code test. A manual trusted owner/admin LXD installation and networkless initialization is required before the prepared scripts can be used.

## Isolation design

| Boundary | Prepared design | Actual verification |
|---|---|---|
| Host networking | No LXD-managed bridge, LXD profile includes **root disk only**, expanded config NIC check before start | NOT RUN, LXD unavailable |
| Resources | 2 vCPU, 1536MiB RAM, max 256 processes; use existing pool; no automatic disk-quota claim | NOT RUN |
| Guest root | Used **inside container only** to create test users and initialize files | NOT RUN |
| Gate UID | `lunagate`, owns public key/journal, group socket `0710` dir + `0660` socket | NOT RUN |
| Agent UID | `lunaagent`, private Ed25519 key, member of group `lunaipc` | NOT RUN |
| Outsider | `lunaoutsider` is not in IPC group → socket denied | NOT RUN |
| Wrong UID insider | `lunaevil` in IPC group receives a **copied synthetic test key inside guest** → still denied by `SO_PEERCRED` | NOT RUN |
| Replay | fsynced journal denies previously used ID/nonce after Gate normal restart; corrupt journal denies new startup | NOT RUN |
| Enforcement | NO nftables apply, BPF/XDP, systemd service or remote/Cloud agent | Safety property of intended test scope; not runtime-proven in LXD |

The LXD container **shares the host's Linux kernel**. Therefore even an unprivileged LXD container does not furnish independent *kernel* isolation for nftables/eBPF, and must never be considered an approved nftables `nft -c`, packet-filtering, XDP or kernel fuzzing lab. Those need a **separate disposable VM with its own kernel** and independent host-network reachability tests.

## Preconditions and acceptance procedure

1. Owner/admin initializes LXD manually in *no-network* mode, verifies daemon health, ensures group access only for trusted operators. Obtain a locally cached official Ubuntu 24.04 image and a storage pool. No scripts install snap or initialize host networks.
2. Run `bash scripts/lab/lxd/check-host.sh`. Record sanitized output (OS/LXD version, storage pool backend, network inventory), without publishing private IPs, keys or user account data.
3. Run `LUNA_LXD_POOL=<preexisting-pool> bash scripts/lab/lxd/create-lab.sh`. Verify newly created profile has **only root disk**, fully expanded container config has **no NIC**, loopback-only guest state, no host services changed. If unexpected devices appear, STOP before guest boot.
4. Run `bash scripts/lab/lxd/run-tests.sh`. The runner copies the **static Go binaries** and synthetic data into a marked container and calls its guest-only test entrypoint. It never pushes host credentials or source trees.
5. Require explicit result markers, plus clean negative results. A positive `ADVISORY_ONLY` for the authorized test Agent does not mean an actual owner approved enforcement. Signed `APPLY_POLICY` must **always** return `REJECTED` and `kernelWrites=false`.
6. Report real result as PASS/FAIL/BLOCKED with LXD version, image alias/digest, source Git SHA, container and kernel versions, duration, Linux UIDs/GIDs (without secrets), guest test outcomes, and measured disk/memory. Until actual execution, all LXD-specific acceptance is **BLOCKED_NOT_RUN**.

## Tests already completed versus deferred

**Completed 2026-10-14 (HOMESV host, not LXD):** read-only LXD availability and resource preflight; prepared five scripts, passed `bash -n` on them; source review of existing Gate/Agent binaries and repo. Original separate-process same-UID Ed25519 unit/E2E tests had passed in prior commit `b631273`.

**Not completed:** LXD installation; image cache, LXD profile, LXD storage, running instance, distinct-UID matrix, VM root/kernel isolation, `nft -c`, network packet testing, race detector, interruption/power-loss/fork/snapshot replay tamper, signed receipts and actual owner login.

**Implementation caution:** the new scripts are design/test automation awaiting execution. Real usability of LXD CLI flags, image prerequisites and guest runuser/setpriv behavior must be independently tested; do not claim that `bash -n` or Go tests verify them.

## Follow-on safety tests after actual LXD availability

Once LXD distinct-UID matrix passes, create a **separate LXD VM** or equivalent full VM with its own kernel for netfilter functional validation, watchdog, independent rescue, Docker/UFW coexistence and rollback. Never attach eBPF/XDP or modify HOMESV firewall for this milestone. Update Code Map, Project Book, Worklog, Project Truth and checkpoint with actual execution evidence before advancing a project gate.
