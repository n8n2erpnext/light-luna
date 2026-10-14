# P0 Independent-Kernel nftables Syntax, Packet Flow and Rollback VM Lab

**Status: DEPLOYED + INITIAL VM ACCEPTANCE PASS (2026-10-14).** This is a kernel-isolated **development test**, not a production firewall executor, not P0/G-STANDALONE acceptance.
Project: Light Luna · HOMESV · branch `feat/p0-standalone-foundation` · current commit to be verified by `git rev-parse HEAD`.
Source of truth: [Go unhooked nft generator](../../internal/nftpreview/compile.go), [policy baseline](../../testdata/host-baseline.json), [VM host runner](../../scripts/lab/vm/run-tests.sh), [VM guest netns/nft test](../../scripts/lab/vm/guest-nft-test.sh), [VM runbook](../../scripts/lab/vm/README.md), [lab provisioning helper](../../scripts/lab/vm/create-vm.sh), [explicit cleanup](../../scripts/lab/vm/delete-vm.sh), [Code Map](../project-book/LIGHT_LUNA_CODE_MAP.md).

## Why VM rather than an LXD container

The earlier [LXD container lab](p0-lxd-cross-uid-lab.md) established independent Agent/Gate Linux UIDs and `SO_PEERCRED` checks, but LXD containers still share the HOMESV kernel. **Kernel firewall/hook tests must run inside a VM.** HOMESV offers Intel VT-x plus `/dev/kvm`; we created a real LXD **VIRTUAL-MACHINE** using QEMU/KVM, not a second container.

- Host: Ubuntu 24.04.5, Linux kernel **6.8.0-146-generic** at test; CPU Intel i5-4200M with 4 logical CPUs, 7.7GiB RAM.
- VM: `light-luna-p0-nftvm`; Ubuntu 24.04.5 LTS, **kernel 6.8.0-142-generic**; profile `light-luna-p0-kernel-vm`, 2 vCPU, 2048MiB memory, `security.secureboot=false` for lab.
- VM image: official `ubuntu:24.04 --vm`, release 20260926, remote fingerprint `d4ff18198fb2ed8c88d408421e610c143cf1f5c1c0178e57178fb586880c3e56`, cached as `light-luna-ubuntu2404-vm`.
- Guest expanded LXD devices: **only root disk**, no NIC. Inside guest `/sys/class/net` contains only `lo`; VM `systemd-detect-virt` returns `kvm`.
- Existing local LXD `default` pool driver `dir`. It **does not prove an enforced per-VM disk quota**; monitor physical disk. Guest memory after initial run ~404MiB, HOMESV free disk ~101GiB.
- Managed LXD bridge `lxdbr0` removed during previous lab setup and remains absent; existing physical/VPN interfaces have not been intentionally reconfigured. Host nftables `table inet lxd {}` observed **empty** with no chains/rules after this run. This does **not** amount to a full before/after host ruleset audit.

## Exact tested operations

```text
HOMESV nonprivileged Go compiler
  testdata/host-baseline.json
        │
        └─ lunactl firewall policies nft-preview
            ├─ declares no netfilter hooks
            ├─ kernelWrites:false / syntaxChecked:false
            └─ unhooked preview text
                         │ (file transfer only)
                         ▼
KVM VM kernel 6.8.0-142 / no NIC
 ├─ target test namespace: nft -c -f <Light Luna generated text>
 │    └─ PASS grammar/kernel parser, script NEVER installed
 └─ two independent ephemeral netns linked only by veth:
      ping baseline ACCEPT
      synthetic-only nft input hook DROP, counter increments
      separate watchdog deletes ONLY synthetic table after ~3s
      controller placeholder terminated; ping restored
      trap cleans veth namespaces and any leftover table
```

Source verification used this exact command on HOMESV:

```sh
export PATH=/home/ubuntu/.local/opt/go1.27.2/bin:$PATH
LUNA_LXD_USE_SUDO=1 bash scripts/lab/vm/run-tests.sh
```

**Evidence markers from real VM execution:**
```text
SEPARATE_KERNEL_NO_NIC_KVM_PASS: host=6.8.0-146-generic guest=6.8.0-142-generic
P0_NFT_UNHOOKED_COMPILER_NO_EXECUTOR_PASS
VM_KERNEL=6.8.0-142-generic
VM_PACKET_BASELINE_ACCEPT_PASS
LIGHT_LUNA_UNHOOKED_NFT_KERNEL_SYNTAX_PASS
INVALID_NFT_NEGATIVE_PASS
VM_PACKET_DROP_COUNTER_PASS
SYNTHETIC_CONTROLLER_TERMINATED
VM_WATCHDOG_ROLLBACK_PACKET_RESTORED_PASS
VM_ISOLATED_NFT_SYNTAX_FLOW_ROLLBACK_PASS
VM_KERNEL_TESTS_EXECUTED
```

After the test, `lxc exec light-luna-p0-nftvm -- ip netns list` returned empty; `nft list tables` was empty, and guest still contained only `lo`. Host `nft list table inet lxd` was empty. The rollback covered an independently authored **synthetic lab hook**, *never an applied Light Luna compiler rule*.

## Trust guarantees and nonclaims

**Actually verified:**
1. Ubuntu guest runs a different kernel from HOMESV with QEMU/KVM rather than sharing kernel as LXD container does.
2. Source Go `nftpreview.Compile` produced an unhooked script which passed `nft -c` parser/kernel check **in the VM test namespace**. A deliberately malformed script failed the negative check.
3. The VM kernel can drop ICMP using a synthetic nft input chain with counter evidence, and packet communication recovers after a short synthetic independent-process watchdog deletes that table.
4. VM and host network inventories had **no LXD-managed NIC/bridge**; tests did not change HOMESV's running firewall rule set intentionally.

**Not verified / not implemented:** real Light Luna hooked nft firewall compiler; gated privileged executor; host management reachability or emergency owner recovery; Docker/UFW/NetBird coexistence; all supported nft versions, high-volume stress, NAT and conntrack, full DDoS resilience; watchdog as a durable independent service through host power loss; rollback across reboot; signed owner authorization and Cloud device control; kernel-level malware defense/eBPF/XDP. A controller *placeholder* was killed, not a real privileged executor; the watchdog is a shell process in a test VM only.

**Warning:** even though this is a VM, a hypervisor or kernel vulnerability can theoretically cross boundaries. Do not test untrusted malware or dangerous kernel exploit code here. This lab is only for valid, bounded, synthetic nft/netns traffic tests.

## Next engineering gate

1. Move **advisory policy → typed nft staging plan** behind a fail-closed Gate with independent commit/revert phases. Keep the **production executor absent** until owner signs off on the full isolation and recovery acceptance.
2. Implement VM-only mock executor with unique table namespace, journal transaction id, TTL, recovery watchdog and explicit admin-rescue allowlist; test faults, process death, reconnect and repeat requests.
3. Extend kernel-namespace coverage to IPv6, TCP/UDP, noninterference and systemd/watchdog reboot in the VM, with network still disabled.
4. Run race detector in GCC-enabled CI, and verify storage pressure/rollback reliability before any host deployment.
5. Document actual evidence and update Project Truth and Code Map each increment; never treat a successful synthetic packet drop as firewall shipping readiness.
