# Light Luna P0 kernel-separated LXD VM

**Status (2026-10-14): DEPLOYED and kernel/namespace nft tests PASS on HOMESV.** LXD 5.21.8, KVM, local Ubuntu 24.04 VM image release 20260926. Repo branch `feat/p0-standalone-foundation`, see latest HEAD and [evidence document](../../../docs/architecture/p0-lxd-kernel-nft-rollback.md).

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

The guest test script creates **two throwaway network namespaces within the separate VM kernel** and joins them with a synthetic veth pair. It confirms an ICMP packet can pass, uses `nft -c` to validate the Light Luna preview, rejects intentionally malformed nft syntax, then installs a **separately authored synthetic drop rule** inside the *VM destination namespace only*. It verifies ping is blocked and a drop counter increments. It starts a timer in a separate process/session, kills a stand-in controller and verifies the timer **deletes only the synthetic table**, after which ping succeeds again. Guest cleanup removes both namespaces and rule fixtures.

Expected PASS markers:

```text
SEPARATE_KERNEL_NO_NIC_KVM_PASS
P0_NFT_UNHOOKED_COMPILER_NO_EXECUTOR_PASS
VM_PACKET_BASELINE_ACCEPT_PASS
LIGHT_LUNA_UNHOOKED_NFT_KERNEL_SYNTAX_PASS
INVALID_NFT_NEGATIVE_PASS
VM_PACKET_DROP_COUNTER_PASS
SYNTHETIC_CONTROLLER_TERMINATED
VM_WATCHDOG_ROLLBACK_PACKET_RESTORED_PASS
VM_ISOLATED_NFT_SYNTAX_FLOW_ROLLBACK_PASS
VM_KERNEL_TESTS_EXECUTED
```

## Fresh lab setup and cleanup

The existing VM was created by authorized LXD commands with the same profile parameters. To create a *new* instance on another lab host (provided LXD, pool and image already exist), use `LUNA_LXD_USE_SUDO=1 bash scripts/lab/vm/create-vm.sh`. It refuses an existing profile or instance. Provisioning of the initial LXD daemon and official image is a separate owner-authorized step; the script does **not** install packages or create any network.

For destructive cleanup, `LUNA_LXD_USE_SUDO=1 LUNA_LXD_VM_DELETE_ACK=DELETE_LIGHT_LUNA_NFT_VM bash scripts/lab/vm/delete-vm.sh` removes only the marked exact VM and retains the profile/image/pool. Do not call it while evidence is needed.

## Limits of the proof

1. The kernel in VM is separate from HOMESV, but this does NOT prove future production executor, safety guarantees against a compromised hypervisor, or LXD/QEMU isolation certifications.
2. The Light Luna compiler emits **UNHOOKED text** only; acceptance is **real kernel syntax** via `nft -c`. It does NOT test Light Luna enforcement, coexistence with Docker/UFW, routing, NAT, VPN, remote admin reachability or true policy default drop.
3. The **synthetic-only** VM hook and 3-second watchdog demonstrate packet drop, counter and timer-based revert **inside guest namespace**. They are NOT the production privileged executor/transaction manager, nor durable rescue after crash, power loss or reboot.
4. Never apply synthetic drop rules, nft tests or future XDP programs on HOMESV. Keep kernel tests in a separate virtual machine. No Cloud, NetBird or original Sentinel integrations are touched.

[Current P0 architecture and verified scope](../../../docs/architecture/p0-lxd-kernel-nft-rollback.md)
