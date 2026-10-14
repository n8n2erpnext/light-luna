# P0 VM typed staging and explicitly triggered independent rollback — verified

Status: **IMPLEMENTED typed inert intent and VM synthetic transaction; full combined suite PASS twice after guarded cleanup. Timed TTL watchdog NOT verified. Not a product executor.**
Date: 2026-10-14
Branch: `feat/p0-standalone-foundation`; obtain actual HEAD with `git rev-parse HEAD`.
Scope: HOMESV host / LXD KVM `light-luna-p0-nftvm` with independent guest kernel and **no NIC**.
Previous lab evidence: [kernel/nft grammar and synthetic packet rollback](p0-lxd-kernel-nft-rollback.md). This is a new stage-state verification increment, not proof of production policy enforcement.
Primary sources: [VM-only typed Go intent](../../internal/vmstage/plan.go), [Go tests](../../internal/vmstage/plan_test.go), [CLI](../../cmd/lunactl/main.go), [host runner](../../scripts/lab/vm/run-tests.sh), [VM-stage journal](../../scripts/lab/vm/guest-stage-journal.py), [VM-stage guest test](../../scripts/lab/vm/guest-stage-test.sh), [VM watchdog payload](../../scripts/lab/vm/guest-rollback.sh), [TCP/UDP fixture](../../scripts/lab/vm/guest-netprobe.py), [old regression](../../scripts/lab/vm/guest-nft-test.sh). [Security obligations](../07-SECURITY-SAFETY.md).

## New code reachable today

```bash
go run ./cmd/lunactl vm stage-intent testdata/host-baseline.json 8
```

`lunactl.run` parses policy using strict `policy.Decode`, then `vmstage.Build` runs the existing inert `nftpreview.Compile`, checks it is **UNHOOKED** with `KernelWrites=false`, derives a deterministic identifier from the policy content hash and test table name, and outputs a bounded typed JSON **manifest**. Manifest contains policy identity/hash/revision, TTL **1–10 seconds**, exact four phase labels, `kernelWrites=false`, `ownerApproved=false`, and `scope=KVM_GUEST_PRIVATE_NETNS_ONLY`. `vmstage.Build` contains **no nft, shell, Unix socket, privileged executor, signer or kernel code**. Manifest is only a hint to a **synthetic VM fixture**, not a permit to enforce. TTL/invalid-policy negative tests exist under `internal/vmstage`.

## Real VM-only transaction fixture

```text
HOMESV unprivileged CLI                 KVM VM guest / no NIC
    policy.Decode                        manifest exact-schema check, stage ID/hash/TTL
     ↓                                   only fixed /opt/luna-p0-kernel-tests/...
    vmstage.Build                         STAGED → VALIDATED → COMMITTED_SYNTHETIC → ROLLED_BACK
     ↓                                              │
    JSON inert intent ─────────────►    stage fixed synthetic nft script, nft -c
                                               │ not product compiler output
                                           create 2 ephemeral netns and veth
                                           IPv4+IPv6 / TCP+UDP echo probes
                                               │
                                            guest namespace nft -f synthetic table
                                           confirmed packet DROP and counters
                                               │
                               explicitly triggered independent systemd service (TTL auto-trigger NOT proved)
                                               │ controller placeholder killed
                                               └─ delete ONLY synthetic table; journal ROLLED_BACK
                                                 ping and TCP/UDP restored
```

- `guest-stage-journal.py` is **VM-root-only** and checks marker, no NIC, manifest schema exact fields, deterministic stage digest, false approval and bounded TTL. It denies invalid TTL/hash, duplicate transition and jumping ahead. It saves a **single-controller demo journal** via write/fsync + rename + directory sync; **no flock, MAC, owner authentication, durable reboot replay or multi-client atomicity**.
- `guest-stage-test.sh` uses only a fixed, independently authored synthetic nft table `luna_p0_stage` in a newly created **VM destination network namespace**. It never loads the Light Luna-generated unhooked preview into a hook. The two-namespace fixture has IPv4 `198.19.0.0/30`, IPv6 `fd42:198:19::/126`, echo TCP port 18080 and UDP port 18081; IPv6 uses `nodad` to avoid fixture-only address readiness races. It checks traffic ACCEPT before, DROP with three nft packet counters during synthetic commit, then ACCEPT again after rollback.
- A **VM-only transient systemd service** is explicitly triggered by the test driver **after** a placeholder controller is killed. Its [rollback payload](../../scripts/lab/vm/guest-rollback.sh) refuses execution outside the marker-checked, NIC-less KVM VM, with unknown synthetic table or marker names. For staged netns, the service checks an **anchor PID and expected `net:[inode]`** and uses `nsenter -t <pid> -n` because a systemd service cannot reliably see netns bind mounts created inside a separate `lxc exec` mount namespace. It removes only the test table and updates the VM journal. **No automatic TTL trigger is claimed.**
- **Investigation and negative evidence:** `setsid` detached sleep, `systemd-run` service-internal sleep, and `systemd-run --on-active` timer each failed to finish within the bounded test window on some VM runs. Unit journal showed timers were started but cleanup stopped them before an observed trigger. Independently invoking `systemd-run --wait --collect` for the fixed revert payload passed; the full combined suite passed **twice consecutively**, but this demonstrates **explicit separate-service rollback only**, **not automatic TTL expiration**. Do not mark P0 rescue timer complete.
- **Namespace isolation finding:** PID1 systemd cannot necessarily resolve the namespace names created by a transient `lxc exec` mount namespace. A failed revert returned `Cannot open network namespace`. The new stage fixture launches a long-lived **VM guest namespace anchor**, waits until `/proc/<pid>/ns/net` matches expected netns inode, then the separate service enters it via `nsenter`. An early PID-read race and an over-escaped inode-regex guard were fixed; their failed runs are not erased or marked PASS.
- **Interrupted transaction handling:** the [explicit reconciliation script](../../scripts/lab/vm/reconcile-interrupted-stage.sh) first checks no guest nft table, named netns **or surviving process with a differing netns inode**. For `STAGED/VALIDATED` only, it archives abort evidence; for `COMMITTED_SYNTHETIC` with verified cleanup, it records terminal `ROLLED_BACK`. It is **not run automatically** by the test harness and does not modify nftables. Replayed/duplicate/out-of-order transitions remain fail-closed.
- Guest exits cleanly, removing the VM-only nft test table, ephemeral namespaces and socket server. No guest NIC, LXD-managed bridge or HOMESV nftables table is modified by the synthetic packet tests.

## Final HOMESV verification (2026-10-14)

- Combined KVM VM suite `LUNA_LXD_USE_SUDO=1 bash scripts/lab/vm/run-tests.sh` **PASS twice consecutively** after the final PID/inode namespace fix, with `VM_KERNEL_AND_STAGE_TESTS_EXECUTED` marker; `stage-journal.json` ended `ROLLED_BACK`. Guest network namespaces and nft tables were absent after cleanup.
- `go test -count=3 ./...` and `go vet ./...` **PASS**. `lunactl` cross-build for Linux amd64 and arm64 (`CGO_ENABLED=0`) **PASS**.
- VM shell scripts passed `bash -n`; Python guest fixtures passed `py_compile`; source catalog **128 entries** passed `--check --check-links`; Project Truth JSON and `git diff --check` passed.
- `go test -race` and fault-tolerant autonomous TTL watchdog **NOT verified**. Interruption attempts during development did **not** pass; their cause, explicit recovery and safety implications are recorded below and in [Worklog](../project-book/LIGHT_LUNA_WORKLOG.md). The passing scope is **explicit independent rollback only**.

## Test assertions and limits

**Actually observed in VM:** both suites passed in **two consecutive complete combined runs** after explicitly reconciling interrupted test journals and fixing namespace anchor validation. Automatic time-based TTL revert did NOT pass. Key test markers:

```text
VM_STAGE_TTL_AND_STAGE_ID_NEGATIVE_PASS
VM_STAGE_STAGED_PASS
VM_STAGE_OUT_OF_ORDER_DENIED_PASS
VM_TCP_UDP_ACCEPT_PASS
VM_STAGE_IPV6_ACCEPT_PASS
VM_STAGE_ANCHOR_NETNS_VERIFIED
VM_STAGE_VALIDATED_PASS
VM_STAGE_DUPLICATE_TRANSITION_DENIED_PASS
VM_STAGE_COMMITTED_SYNTHETIC_PASS
VM_STAGE_DUPLICATE_COMMIT_DENIED_PASS
VM_TCP_UDP_DROP_PASS
VM_STAGE_TCP_UDP_IPV6_DROP_COUNTER_PASS
VM_STAGE_CONTROLLER_KILLED
VM_STAGE_SYSTEMD_ROLLBACK_EXPLICIT_TRIGGER_PASS
VM_STAGE_SEPARATE_SERVICE_ROLLBACK_IPV6_TCP_UDP_RESTORED_PASS
VM_STAGE_SYNTHETIC_TRANSACTION_TEST_PASS
VM_KERNEL_AND_STAGE_TESTS_EXECUTED
```

**NOT proven:** autonomous TTL expiry/revert, real Agent owner consent, signed staging intents/receipts, a production executor, reject/recovery under sudden hypervisor shutdown, independent watchdog process priority/deadline guarantee, stuck kernel calls, Docker/UFW coexistence, power-loss journal repair, NetBird/Cloud integration, packet forwarding/NAT, cgroup confinement under hostile guest root, eBPF/XDP or externally reachable firewall.

The stage manifest intentionally says `ownerApproved=false` and `kernelWrites=false`; **kernel writes happened solely within the fixed VM synthetic test fixture**. A passing synthetic rollback cannot be promoted to `G-STANDALONE` or used to bypass future owner approval.

## Next safe work

1. Isolated fault injection for transient watchdog delayed startup and systemd/reboot, bounded recovery measurement and full race checks in build CI.
2. Add typed transaction and authorization receipts to a real *denial-only* Agent↔Gate protocol, retaining per-owner local approval boundary and isolated privilege holder.
3. Prove no unintended netfilter side effects when unrelated tables or Docker-like chains coexist in VM; maintain a separate-kernel lab and one-way rollback safety.
4. Only consider a carefully restricted executor after VM rollback/rescue canary reliability, key isolation and explicit owner permission; no production host deployment or NetBird dependency yet.
