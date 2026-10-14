# Light Luna P0 kernel-separated LXD VM

**Status (2026-10-14): DEPLOYED and kernel/namespace nft tests PASS on HOMESV.** LXD 5.21.8, KVM, local Ubuntu 24.04 VM image release 20260926. Repo branch `feat/p0-standalone-foundation`, see latest HEAD and [evidence document](../../../docs/architecture/p0-lxd-kernel-nft-rollback.md).

## Journal-terminal-before-DONE failure / immutable observation receipts (2026-10-14)

[Architecture and correct acceptance limits](../../../docs/architecture/p0-vm-autonomous-ttl-watchdog.md). Both PID1 rollback services now require `guest-rollback-receipts.py` delivered by every KVM host runner, and write non-overwriting, file+directory fsynced **prepared → kernel-absent → journal-terminal** observations under the existing rollback flock. No observation receipt is a signature, an owner grant, an atomic nft+journal commit or permission to auto-reconcile a missing DONE.

On the exact disposable no-NIC VM, after a clean terminal synthetic stage, test a fallback `os._exit(73)` **after terminal journal, before DONE**, and explicitly preserve missing-DONE evidence:

```bash
LUNA_LXD_USE_SUDO=1 LUNA_VM_TERMINAL_FAULT_ACK=TEST_ONLY_GUEST_TERMINAL_WITHOUT_DONE LUNA_VM_EVIDENCE_TAG=unique-tag bash scripts/lab/vm/run-terminal-receipt-fault.sh
```

Verified lab marker: `VM_GUEST_TERMINAL_WITHOUT_DONE_FAILURE_INJECTION_PASS`; status is `TERMINAL_JOURNAL_DONE_MISSING_REVIEW_REQUIRED`, never a forged DONE. Prior harness observer failures require explicit **post-cleanup** review; `guest-preserve-terminal-review.py` distinguishes such evidence from observations made before the test EXIT trap. Once an operator archives a fault, the active journal is absent and normal `run-tests.sh` must explicitly establish a new synthetic transaction before further fault runners. Do not delete/archive evidence as part of an automatic retry.

## Cooperative rollback writer lock — isolated VM only (2026-10-14)

[Canonical architecture and bounded test evidence](../../../docs/architecture/p0-vm-autonomous-ttl-watchdog.md). The primary watchdog and independently netns-pinned fallback now **both require** `guest-rollback-fence.py`, delivered by the main KVM runner and all synthetic fault runners. Its `fcntl.flock` is held across current journal/boot revalidation, fixed guest-netns nft rollback, journal fsync, and success receipt; no process may claim success based on a stale pre-lock read. The lock file is a persistent VM-local `0600` inode; do not remove it between transactions. Without the fence, primary/fallback fail closed.

A dual-armed test with **same monotonic deadline** and explicit VM-only ACK:

```bash
LUNA_LXD_USE_SUDO=1 LUNA_VM_RACE_CRASH_ACK=TEST_ONLY_NIC_LESS_KVM_RACE_CRASH LUNA_VM_EVIDENCE_TAG=unique-tag LUNA_P0_RACE_ONLY=1 bash scripts/lab/vm/run-race-crash-tests.sh
```

A further full race/crash execution omitting `LUNA_P0_RACE_ONLY=1` exercises the deterministic `os._exit(72)` after nft deletion while lock is held; journal remains unfinished and must be **separately inspected/archived**, never marked `ROLLED_BACK`. New unique archive tag support prevents overwriting earlier failed-transaction evidence. [Offline process/unit tests](test_rollback_fence.py) require **no VM/kernel access**. These scripts **can only** mutate the fixed synthetic table inside the dedicated no-NIC KVM; no actual Light Luna product policy is installed or enforced.

## VM-only dual-arm arbitration and crash-consistency detection (2026-10-14)

[Canonical evidence and remaining gaps](../../../docs/architecture/p0-vm-autonomous-ttl-watchdog.md). The `primary_race` fault mode dual-arms primary and independent fallback while keeping the primary alive. On the tested timeline only the primary rolls back; fallback records `PRIMARY_ALREADY_TERMINAL_NO_TAKEOVER` (REJECTED, **not** DONE). The `fallback_after_delete` mode SIGKILLs the primary, lets fallback remove the **fixed** guest synthetic nft table, then forces `os._exit(72)` **before** terminal journal persistence; this deliberately produces **absent kernel table + outstanding `COMMITTED_SYNTHETIC` journal**, with neither service claiming success. The VM fault observer and explicit manual archive preserve the inconsistency without rewriting it into rollback.

Run only on exact prechecked no-NIC lab VM and with a fresh unique archival tag:

```bash
LUNA_LXD_USE_SUDO=1 LUNA_VM_RACE_CRASH_ACK=TEST_ONLY_NIC_LESS_KVM_RACE_CRASH LUNA_VM_EVIDENCE_TAG=unique-tag bash scripts/lab/vm/run-race-crash-tests.sh
```

The runner was executed with exit 0, both assertions PASS and archives retained. After archive, the active journal is **absent** by design until the next ordinary `run-tests.sh` creates a fresh stage. This is a **crash-inconsistency detector, not automatic power-loss recovery**. It does not test simultaneous deletion, disk tear, or a product firewall.

## Independent namespace-pinning fallback (2026-10-14 — KVM synthetic PASS only)

[Full evidence, limits and source](../../../docs/architecture/p0-vm-autonomous-ttl-watchdog.md). `guest-fallback-supervisor.py` runs as a **second** VM PID1-managed process, entering the exact guest test namespace before commit and keeping it pinned independently of the keeper. Opt-in `LUNA_VM_FALLBACK=1` and explicit fallback ACK apply only to postcommit keeper/watchdog SIGKILL fault modes; the normal stage test still defaults to no fallback. It validates boot/stage/namespace/deadline and exact nft JSON test-rule shape before deleting only `inet luna_p0_stage`, then checks absence and writes a separate **synthetic-only** DONE receipt.

To reproduce after ensuring a clean terminal VM journal and choosing a new evidence tag (do not overwrite archives):

```bash
LUNA_LXD_USE_SUDO=1 LUNA_P0_FALLBACK_TEST_ACK=TEST_GUEST_ONLY_SYNTHETIC_FALLBACK LUNA_VM_EVIDENCE_TAG=your-unique-lab-tag bash scripts/lab/vm/run-fallback-faults.sh
```

Two executions of both `keeper_postcommit` and `watchdog_postcommit` fault cases PASS, including the stricter JSON rule validator. Post-fault traffic recovery is verified **before test cleanup**; a later clean VM inventory confirms empty root nft and named netns. Twelve offline whitelist positive/negative tests PASS. **No real Light Luna kernel enforcement, production deployment, signed owner authentication, persistent daemon or independent power-loss rescue** has been built or demonstrated.

## Three real isolated service-death fault injections (2026-10-14)

[Architecture, evidence, nonclaims](../../../docs/architecture/p0-vm-autonomous-ttl-watchdog.md). `guest-stage-test.sh` has explicit `LUNA_VM_FAULT` modes `keeper_precommit`, `keeper_postcommit` and `watchdog_postcommit`. They deliberately SIGKILL only their own temporary systemd MainPID **inside the no-NIC KVM guest**. `guest-fault-evidence.py` observes the failed unit after the monotonic deadline; for both postcommit cases the fixed synthetic nft table remains until the test EXIT cleanup. **No recovery or automatic rollback success is claimed.** After scoped guest cleanup, `guest-archive-fault.py` requires a separate exact acknowledgment and archives the still-incomplete stage journal unchanged along with a failure record; no-ack is rejected.

Lab runner (disruptive to temporary guest-only services; the named VM, profile and no-NIC checks are mandatory):

```bash
LUNA_LXD_USE_SUDO=1 LUNA_P0_SERVICE_FAULT_ACK=KILL_ONLY_ISOLATED_VM_LAB_SERVICES bash scripts/lab/vm/run-service-faults.sh
```

A no-ACK attempt against `guest-stage-test.sh` was denied before changing journal state. After the three fault cases, the **normal nonfault full KVM suite was rerun and PASS**, with terminal `ROLLED_BACK` and empty observed guest nft/named netns. Two nonfatal `wait_for: No record of process` shell diagnostics were observed and remain a fixture-quality issue. One full 3-mode execution **PASS** with marker `VM_THREE_SERVICE_DEATH_FAULTS_DETECTED_AND_ARCHIVED_PASS`. The runner preserves fixed per-mode archival files and refuses overwriting them; do not erase history merely to rerun. Synthetic service failure **exposes missing failover** and must not be reported as live firewall crash recovery. Normal `run-tests.sh` is independent and uses the default `LUNA_VM_FAULT=none` path.

## Forced reboot of committed guest-only synthetic stage (2026-10-14)

Verified isolated fault-injection: `guest-boot-inspect.py` is **read-only**, checks exact no-NIC KVM guest, journal boot ID and bounded root nft/named/process netns inventory; it never clears journals or grants approval. `test_boot_inspect.py` runs 14 offline tests.

The explicitly acknowledged `run-reboot-fault.sh` invokes `guest-prepare-reboot-fault.sh` to demonstrate a fixed synthetic ICMP drop only in a **private network namespace inside the no-NIC VM**, then performs `lxc stop --force` and `lxc start` of **exact** `light-luna-p0-nftvm`. Run this **only** when authorized to reset the disposable lab VM; it disrupts all guest processes, not HOMESV:

```bash
LUNA_LXD_USE_SUDO=1 LUNA_P0_VM_REBOOT_ACK=REBOOT_ISOLATED_LUNA_P0_VM bash scripts/lab/vm/run-reboot-fault.sh
```

The verified result is `VM_FORCED_REBOOT_STALE_JOURNAL_BLOCKED_PASS`, with a different boot ID, old journal `COMMITTED_SYNTHETIC` preserved, empty observed guest nft/netns and denied cross-boot transition. The runner deliberately **does not** clean that journal. Following separate explicit operator inspection, an acknowledged guest-only `guest-archive-stale-boot.py` can archive this exact stale synthetic journal and write a report **without marking rollback**. Missing acknowledgment is denied. The prior run already did this once and retained archived evidence on the guest. Never automate archival as part of the forced-reboot runner.

The full KVM regression after manual evidence preservation **PASS** and leaves no staged nft/netns. This is kernel-reset/stale-transaction detection, **not** power-loss safe production recovery or a reboot-time enforcement watchdog.

## Boot-ID journal hardening (2026-10-14)

Linux `boot_id` is now pinned in the synthetic VM stage journal and watchdog READY/DONE receipts. Cross-boot state transitions and cross-boot manual interrupted-stage reconciliation reject old state. The unprivileged `python3 scripts/lab/vm/test_boot_journal.py` test suite has five passing negative/positive tests. A complete KVM regression with the new code returned exit 0 and previously established explicit/automatic TTL markers. **This does not reboot the guest and does not demonstrate crash-safe boot recovery.** Preserve archived legacy terminal records; do not silently clear incomplete journals.

## Autonomous TTL lab update (2026-10-14)

[Latest design, failures and acceptance](../../../docs/architecture/p0-vm-autonomous-ttl-watchdog.md): `run-tests.sh` now performs **three separate VM-only phases**: old nft syntax+ICMP explicit service revert, staged TCP/UDP/IPv6 explicit service revert, and *independently armed autonomous monotonic TTL revert* of the fixed synthetic table. Both original suites now use systemd keeper units for named network namespaces created from `lxc exec`; the auto service writes atomic READY before commit, then acts on TTL **without a second controller trigger**. All three phases **PASS across three consecutive full post-fix runs**. Verification requires a terminal `ROLLED_BACK` journal, validated DONE receipt, restored probes, and no guest nft/netns. No Light Luna-generated kernel hook is applied.

The earlier note below saying autonomous timer was not verified documents the *previous* flaky transient timer implementation, not the current monotonic independent-service test. **Still no real product executor, reboot/power-loss rescue, owner authentication or G-STANDALONE.**

## Live instance

- Name: `light-luna-p0-nftvm`, type **VIRTUAL-MACHINE** (not LXD container).
- CPU: 2 vCPU; memory cap: 2048MiB. LXD `dir` pool `default`: per-VM root-disk quota **not proven**.
- Image alias: `light-luna-ubuntu2404-vm` (official `ubuntu:24.04 --vm` fetched once via host-side LXD; no VM NIC).
- Guest kernel: `6.8.0-142-generic`; HOMESV kernel: `6.8.0-146-generic` as observed. VM is isolated by QEMU/KVM, with **only loopback** and no LXD-managed NIC, bridge or NAT.
- Host does **not** have `lxdbr0`. LXD-managed nftables table `inet lxd` was observed empty with no chains/rules after the lab.
- Host `/dev/kvm` was present and Intel VT-x enabled; the VM was deliberately created with a profile containing **only a root disk**.
- `lxc info` after tests showed VM RAM around 404MiB; host free disk around 101GiB. `dir` does not enforce an instance disk quota.

## Re-run the safe suite

```sh
cd /home/ubuntu/n8n2erpnext/light-luna
export PATH=/home/ubuntu/.local/opt/go1.27.2/bin:$PATH
LUNA_LXD_USE_SUDO=1 bash scripts/lab/vm/run-tests.sh
```

`run-tests.sh` insists on the **exact** VM name and profile, type `VIRTUAL-MACHINE`, independent guest kernel, `kvm` detection and loopback-only NIC inventory before copying anything. It builds the **current** Go `lunactl` and produces an unhooked policy preview from `testdata/host-baseline.json`. The compiler's flags must say `kernelWrites=false`, `syntaxChecked=false`, `kernelCompatibilityChecked=false`, and the text must have no `hook` statement. The text is pushed to the VM and passed only to **`nft -c`** in a fresh network namespace. It is **never applied**.

The guest test script creates **two throwaway network namespaces within the separate VM kernel** and joins them with a synthetic veth pair. It confirms an ICMP packet can pass, uses `nft -c` to validate the Light Luna preview, rejects intentionally malformed nft syntax, then installs a **separately authored synthetic drop rule** inside the *VM destination namespace only*. It verifies ping is blocked and a drop counter increments. It kills a stand-in controller, then **explicitly triggers an independent transient systemd rollback service** that deletes only the synthetic table and restores traffic. **Automatic TTL expiration is not verified**; earlier systemd timers timed out in this VM and must not be counted as PASS. Guest cleanup removes both namespaces and rule fixtures.

Expected PASS markers:

```text
SEPARATE_KERNEL_NO_NIC_KVM_PASS
P0_NFT_UNHOOKED_COMPILER_NO_EXECUTOR_PASS
VM_PACKET_BASELINE_ACCEPT_PASS
LIGHT_LUNA_UNHOOKED_NFT_KERNEL_SYNTAX_PASS
INVALID_NFT_NEGATIVE_PASS
VM_PACKET_DROP_COUNTER_PASS
SYNTHETIC_CONTROLLER_TERMINATED
VM_SYSTEMD_ROLLBACK_EXPLICIT_TRIGGER_PASS
VM_SEPARATE_SERVICE_ROLLBACK_PACKET_RESTORED_PASS
VM_ISOLATED_NFT_SYNTAX_FLOW_ROLLBACK_PASS
VM_STAGE_TTL_AND_STAGE_ID_NEGATIVE_PASS
VM_STAGE_OUT_OF_ORDER_DENIED_PASS
VM_STAGE_DUPLICATE_COMMIT_DENIED_PASS
VM_STAGE_ANCHOR_NETNS_VERIFIED
VM_STAGE_TCP_UDP_IPV6_DROP_COUNTER_PASS
VM_STAGE_SYSTEMD_ROLLBACK_EXPLICIT_TRIGGER_PASS
VM_STAGE_SEPARATE_SERVICE_ROLLBACK_IPV6_TCP_UDP_RESTORED_PASS
VM_STAGE_SYNTHETIC_TRANSACTION_TEST_PASS
VM_KERNEL_AND_STAGE_TESTS_EXECUTED
```

## Fresh lab setup and cleanup

The existing VM was created by authorized LXD commands with the same profile parameters. To create a *new* instance on another lab host (provided LXD, pool and image already exist), use `LUNA_LXD_USE_SUDO=1 bash scripts/lab/vm/create-vm.sh`. It refuses an existing profile or instance. Provisioning of the initial LXD daemon and official image is a separate owner-authorized step; the script does **not** install packages or create any network.

For destructive cleanup, `LUNA_LXD_USE_SUDO=1 LUNA_LXD_VM_DELETE_ACK=DELETE_LIGHT_LUNA_NFT_VM bash scripts/lab/vm/delete-vm.sh` removes only the marked exact VM and retains the profile/image/pool. Do not call it while evidence is needed.

## Limits of the proof

1. The kernel in VM is separate from HOMESV, but this does NOT prove future production executor, safety guarantees against a compromised hypervisor, or LXD/QEMU isolation certifications.
2. The Light Luna compiler emits **UNHOOKED text** only; acceptance is **real kernel syntax** via `nft -c`. It does NOT test Light Luna enforcement, coexistence with Docker/UFW, routing, NAT, VPN, remote admin reachability or true policy default drop.
3. The **synthetic-only** VM hook and separate systemd rollback service demonstrate packet drop, counters and explicitly triggered revert **inside the VM guest namespace**. Timed automatic rollback has been flaky and is **NOT PASSED**. They are NOT the production privileged executor/transaction manager, nor durable rescue after crash, power loss or reboot.
4. Never apply synthetic drop rules, nft tests or future XDP programs on HOMESV. Keep kernel tests in a separate virtual machine. No Cloud, NetBird or original Sentinel integrations are touched.

[Current P0 architecture and verified scope](../../../docs/architecture/p0-lxd-kernel-nft-rollback.md)
