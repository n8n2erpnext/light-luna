# Light Luna P0 — isolated LXD cross-UID lab

**Status:** **LXD INSTALLED; NIC-LESS CONTAINER RUNNING; distinct-UID E2E PASS**, 2026-10-14. The previous preparation-only checkpoint is superseded by the verified results below.
**Target:** HOMESV / Ubuntu 24.04, Light Luna `feat/p0-standalone-foundation`.
**Authority:** [LXD VM/lab verification contract](../../../docs/architecture/p0-lxd-cross-uid-lab.md) and [documentation safety rules](../../../docs/project-book/LIBRARY_RULES.md).

This is a **networkless unprivileged LXD container** for *Linux filesystem/Unix socket/UID identity* tests. LXD containers share the host kernel: **do not run nft, BPF or kernel-level firewall tests in this container**. Those require a separately authorized disposable virtual machine with kernel isolation.

## Actual installed lab — verification update (2026-10-14)

LXD snap **5.21.8** was installed with the authorized HOMESV Light Remote `lxd`, `package-manager` and `sudo-on-demand` capabilities. `lxd init --minimal` created pool `default` using **dir** driver and an unexpected `lxdbr0` managed bridge; the default-profile `eth0` was removed and that new managed bridge deleted. Current physical/VPN interfaces were not reconfigured by our lab scripts.

A fresh unprivileged `light-luna-p0-lab` Ubuntu 24.04 container is **RUNNING** (2 CPUs, 1536MiB, 256 process cap; expanded config: **root disk only, no NIC**). The live E2E script `scripts/lab/lxd/run-tests.sh` completed with `LXD_DISTINCT_UID_POSITIVE_NEGATIVE_RESTART_PASS` and `LXD_CROSS_UID_TEST_RUN_COMPLETED`: separate Agent/Gate identities, wrong Gate UID denied, outsiders denied, malicious same-group wrong UID with **copied synthetic private key** denied, restart replay denied, corrupted journal startup refused.

Host invocation without joining root-equivalent LXD group:
```sh
export PATH=/home/ubuntu/.local/opt/go1.27.2/bin:$PATH
LUNA_LXD_USE_SUDO=1 bash scripts/lab/lxd/check-host.sh
# The guest-test is ONE-SHOT; run it only on a *new fresh instance*.
```

**After the test:** `light-luna-p0-lab` remains RUNNING for future inspection, with a deliberately corrupted synthetic journal and guest-only test users; the copied rogue synthetic test key was removed. The LXD-owned nftables table `inet lxd` is **empty** after the temporary auto-created bridge was removed. This is *not* a certification of zero host firewall changes. Use a fresh dedicated instance for another one-shot run. The LXD `dir` storage backend offers no enforced per-container quota.

**Caution:** `run-tests.sh` now keeps temporary static binaries inside the working tree because the strict LXD snap cannot read host `/tmp`, then removes them automatically. The test leaves fake identities and an intentionally tainted journal in this **networkless lab**; do not treat it as a reusable production service. A future fresh lab must be created using a new container instance, not by rerunning the same one-shot script.

## Previous blocker (resolved on 2026-10-14)

HOMESV has `/usr/sbin/lxc` and `/usr/sbin/lxd` launchers, but `snap list lxd` reports **no LXD snap installed**. Ordinary `lxc version` reports a missing LXD system group/snap; there are no live LXD instances. The attempted LXD snap installation via managed remote command was **blocked by the tool's safety control**. We did **not** switch to PTY to bypass that denial. No LXD host services, networks, storage pools or containers were created during this session.

Only an authorized owner/admin should install/init LXD, using an independent trusted local console. **Do not add untrusted identities to the host's `lxd` group:** LXD daemon control is effectively root-equivalent.

## Host preparation (owner/admin, manual)

1. Initialize LXD in a **networkless** configuration, with **no LXD-managed bridge, NAT, NIC, DNS or DHCP**; keep `eno1`, existing VPN and host firewall untouched. Prefer a dedicated local storage pool with a driver that supports enforced volume quotas. Note that a `dir` pool does **not** guarantee a per-container disk quota.
2. Check a local Ubuntu 24.04 image is available under alias `light-luna-ubuntu2404`. If acquiring a new official image, the host LXD client (not the guest) may require outbound access once to populate local cache. Do not let the guest have a NIC.
3. Run read-only readiness: `bash scripts/lab/lxd/check-host.sh`. Confirm the storage pool name and image alias. **Do not execute the create script before verifying the LXD config**.

LXD official documentation: <https://documentation.ubuntu.com/lxd/stable-5.21/tutorial/first_steps/> and <https://documentation.ubuntu.com/lxd/stable-5.21/howto/profiles/>.

## Create and test — only after LXD is available

```bash
# Change pool name to one already present in 'lxc storage list'.
export LUNA_LXD_POOL=default
export LUNA_LXD_IMAGE=light-luna-ubuntu2404
export LUNA_LXD_INSTANCE=light-luna-p0-lab

bash scripts/lab/lxd/create-lab.sh
bash scripts/lab/lxd/run-tests.sh
```

`create-lab.sh` creates a **dedicated new** profile `light-luna-p0-isolated`, 2 vCPU and 1536MiB memory limit, only a root disk. It uses `lxc init` (not `launch`) and inspects the fully expanded config for NIC devices *before* starting the instance, then verifies that guest `/sys/class/net` contains **lo only**. It marks the container with `/etc/light-luna-p0-lxd-lab`. If the instance/profile exists, it refuses to overwrite or modify it. LXD itself must already be initialized and available; scripts do not install snap, mutate host networking or apply any firewall policy.

`run-tests.sh` builds static `luna-gate` and `luna-agent` binaries on the host, pushes only those binaries, a synthetic host-baseline policy and `guest-test.sh`, then executes tests **inside the marked LXD container**.

Inside the guest, the test creates `lunaagent`, `lunagate`, `lunaevil` and `lunaoutsider` **guest-only** service identities plus a temporary bootstrap identity. The bootstrap generates a **synthetic random Ed25519 keypair**, places private and public material into separate owner-private directories, and relinquishes them. A dedicated `lunaipc` group permits socket access but never replaces UID checks or signatures. Test suite exercises:
- Correct Agent UID, membership and signing key → `ADVISORY_ONLY` for a synthetic preview.
- Gate rejects correctly signed `APPLY_POLICY`.
- Wrong expected Gate UID → Agent refuses the server.
- User outside IPC group → filesystem/socket permission denies connection.
- Different UID *with IPC group membership and a deliberately copied test private key* → Gate must refuse on `SO_PEERCRED` UID; the copied key exists **only inside this disposable lab**.
- Previously consumed ID and nonce → still `REJECTED` after Gate restart.
- Corrupt replay journal → Gate refuses to start.

The guest lab is **one-shot**; rerun with a fresh dedicated instance instead of attempting to repair or overwrite existing users/journal. It does not test unattended services, human-owner login, malicious root inside the guest, restart after power loss or cryptographically signed Gate receipts.

## Cleanup — explicit opt-in

```bash
LUNA_LXD_DELETE_ACK=DELETE_LIGHT_LUNA_LAB bash scripts/lab/lxd/delete-lab.sh
```

This deletes **only** the marked dedicated instance. It deliberately does **not** remove storage pools, images or host-owned profiles. The instance must be running so the script can verify its lab marker.

## Safety, capacity and truth

HOMESV preflight: approximately 7.7GiB total RAM (7.1GiB available) and 105GiB free storage at observation time. LXD image disk use and any storage-specific quota are **not measured**. All Guest UID/GID acceptance results remain **NOT RUN** until a real LXD instance exists. `bash -n` only verifies shell syntax, not LXD API semantics, user namespace isolation, group permission behavior or process security. Never label these prepared scripts as an E2E PASS.
