# Light Luna P0 kernel-separated LXD VM

**Status (2026-10-14): DEPLOYED and kernel/namespace nft tests PASS on HOMESV.** LXD 5.21.8, KVM, local Ubuntu 24.04 VM image release 20260926. Repo branch `feat/p0-standalone-foundation`, see latest HEAD and [evidence document](../../../docs/architecture/p0-lxd-kernel-nft-rollback.md).

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
