# P0 VM typed staging and explicitly triggered independent rollback — verified

Status: **Historical explicit-service milestone PASS x2; the later [autonomous TTL watchdog](p0-vm-autonomous-ttl-watchdog.md) PASSED three consecutive full KVM suites. Both remain development-only synthetic VM labs, NOT product executors.**
Date: 2026-10-14
Branch: `feat/p0-standalone-foundation`; obtain actual HEAD with `git rev-parse HEAD`.
Scope: HOMESV host / LXD KVM `light-luna-p0-nftvm` with independent guest kernel and **no NIC**.
Previous lab evidence: [kernel/nft grammar and synthetic packet rollback](p0-lxd-kernel-nft-rollback.md). This is a new stage-state verification increment, not proof of production policy enforcement.
Primary sources: [VM-only typed Go intent](../../internal/vmstage/plan.go), [Go tests](../../internal/vmstage/plan_test.go), [CLI](../../cmd/lunactl/main.go), [host runner](../../scripts/lab/vm/run-tests.sh), [VM-stage journal](../../scripts/lab/vm/guest-stage-journal.py), [VM-stage guest test](../../scripts/lab/vm/guest-stage-test.sh), [VM watchdog payload](../../scripts/lab/vm/guest-rollback.sh), [TCP/UDP fixture](../../scripts/lab/vm/guest-netprobe.py), [old regression](../../scripts/lab/vm/guest-nft-test.sh). [Security obligations](../07-SECURITY-SAFETY.md).

## Core P0 guest manifest/journal opened-inode read bounds — offline RED→GREEN (2026-10-15)

**IMPLEMENTED SYNTHETIC VM LAB FILE-READER HARDENING ONLY.** The guest-only [`safe_load`](../../scripts/lab/vm/guest-stage-journal.py) previously called `Path.is_symlink`, `Path.is_file` and `Path.stat` **before** `Path.open`, then performed an unbounded `json.load`. Path/inode or file length could change after the precheck, and a hardlink to an alternate file was accepted. This permitted a stale-size read larger than its 4096-byte contract, or accepting a symlink if the earlier symlink precheck was stale.

The reader now opens **one descriptor** using `os.open(O_RDONLY | O_CLOEXEC | O_NOFOLLOW | O_NONBLOCK)`, checks `os.fstat` on that descriptor for regular-file type, **exactly one hardlink**, positive byte size and maximum **4096 bytes**, then reads at most 4097 bytes and rejects any growth beyond the limit before strict UTF-8 JSON decoding with existing duplicate-key rejection. This covers the fixed guest manifest and stage journal read paths. No file is rewritten, migrated, elevated or executed by this correction.

Five [offline focused reader regressions](../../scripts/lab/vm/test_stage_manifest_ttl.py) test normal single-link input, denial of a hardlink, denial of a symlink when the old precheck is deliberately mocked stale, denial of a file grown beyond the byte cap when the old `Path.stat` result is mocked stale, and preservation of duplicate-key rejection. **Three attack cases were RED under the old reader and GREEN after the fix**, the two controls remained PASS; the manifest+reader focused suite now **12/12 PASS**, and the complete VM Python **offline** test suite now **103/103 PASS**.

**Limits:** `O_NOFOLLOW` protects the **final path component**, not all ancestor directories. The guest-only root directory still needs independent trusted ownership/isolation; no defense against malicious guest root, active multi-process manipulation of the enclosing directory, kernel power-loss, or authenticated receipt provenance is established. No live VM/KVM netfilter or rollback operation was run. **LXD socket permissions previously denied a read-only VM inventory and the live guest/journal state remains UNKNOWN; no bypass attempted.** P0 OPEN / G-STANDALONE BLOCKED.

## Core P0 guest manifest field integrity — offline RED→GREEN (2026-10-15)

**IMPLEMENTED VM LAB PARSER HARDENING ONLY, NOT OWNER AUTH OR A LIVE FIREWALL TEST.** After the TTL-bound stage ID increment, a separate read-only inspection of the [Python guest manifest verifier](../../scripts/lab/vm/guest-stage-journal.py) revealed two acceptance gaps: Python treats `True == 1` and `1.0 == 1`, so it accepted a JSON boolean/float as `policyRevision: 1`; the JSON `effect` field was present but had no value check, so an attacker-controlled string such as `kernel rules applied` or `owner authorized` (or JSON null) was accepted as manifest metadata. These are **misleading fixture declarations**, not evidence that any kernel writes or authorization occurred.

[Existing offline manifest tests](../../scripts/lab/vm/test_stage_manifest_ttl.py) now include two additional negative methods covering **five mutation cases**: wrong `effect` values (three) and non-integer revisions (`true`, `1.0`). All five produced **RED** on the previous parser. [Guest verifier](../../scripts/lab/vm/guest-stage-journal.py) now requires `type(policyRevision) is int` in addition to value 1, and requires `effect` to match exactly the inert Go builder text `none: manifest only; synthetic VM-only hook tested separately`. Tests are **GREEN 7/7**, and the complete Python VM offline suite is **98/98 PASS**. An unprivileged, offline-built real `lunactl vm stage-intent` manifest was accepted by the guest parser, then mutations of revision to `true`/`1.0` or effect to a false kernel claim were refused.

**VM live state UNKNOWN:** Read-only `lxc list light-luna-p0-nftvm` attempted on HOMESV through the official session failed with **LXD Unix socket permission denied**. No alternative privilege, sudo, PTY or network access was used; the current guest manifest, journal/receipts, live nft state and clean-archive status could **not** be inspected. No VM state change, cleanup, new NIC-less KVM test or live rollback validation was made. Preserve any preexisting VM journal/receipts and continue only with authorized read-only access plus independent archive/cleanup review.

**Boundary:** this parser checks an inert, fixed synthetic VM fixture's declarative fields, not policy-content authenticity, signed owner permission, power-loss durability, epoch/replay uniqueness, host firewall coexistence or guaranteed rescue. **P0 OPEN; G-STANDALONE BLOCKED; NetBird/Cloud DEFERRED.**

## Core P0 synthetic stage identity binds TTL — offline source verification (2026-10-15)

**IMPLEMENTED DEVELOPMENT MANIFEST CONSISTENCY ONLY; NOT A PRODUCTION ROLLBACK OR SIGNED LEASE.** The earlier [Go VM-stage builder](../../internal/vmstage/plan.go) derived `stageId` from `schema\0policyPreviewHash\0isolatedTable`, but **not `ttlSeconds`**. Two distinct synthetic stage requests with the same policy/table but different TTL values therefore had an identical stage identifier (`vm-8420f04e562c64dacbc27d03` was observed for 1s and 10s in the original fixture). Because the [guest journal](../../scripts/lab/vm/guest-stage-journal.py) uses `stageId` for manifest/journal binding, a TTL substitution could alias an existing synthetic transaction.

The builder now computes the truncated SHA-256 over `schema\0policyPreviewHash\0isolatedTable\0decimalTTLSeconds`, maintaining determinism for identical policy and TTL while separating differing deadlines. The guest's independent Python verifier uses the same binding and refuses a changed TTL with the old stage ID. [Go regression](../../internal/vmstage/plan_test.go) demonstrated **RED** on the original 1-vs-10 collision, then GREEN ×3, and checks stable same-input IDs, constant source hash for TTL changes, and revision binding. A new [offline Python guest regression](../../scripts/lab/vm/test_stage_manifest_ttl.py) adds five tests for stable manifest, different TTL IDs, stale-ID substitution denial, old unbound formula denial, and TTL bounds. Total Python VM **offline** suite now **96 tests PASS**. An independently built real `lunactl vm stage-intent` generated 1s and 10s manifests, each accepted by the guest Python parser; substituting TTL without recomputing its ID was rejected. **No live guest KVM/packet/rollback test was run for this change.**

**Compatibility boundary:** A previously generated **old-formula** VM manifest is intentionally rejected by the updated parser. This is fail-closed and must not trigger blind journal migration or clearing of ambiguous VM lab evidence. Any later staged integration requires a clean/archived **isolated guest** fixture and separate compatibility review. A deterministic stage ID is NOT unique per attempt or boot, not owner-authenticated, not a power-loss-proof lease, and cannot authorize APPLY. No production nft, Host networking, NetBird/Cloud, real watchdog or executor is touched. **P0 OPEN / G-STANDALONE BLOCKED.**

## Subsequent autonomous TTL milestone (2026-10-14)

[Latest independent-KVM watchdog and measured results](p0-vm-autonomous-ttl-watchdog.md) supersede the earlier unverified-timer outcome below: after adding an independent systemd keeper process and monotonic service arming-before-commit handshake, **three full post-fix VM suites PASS** with autonomous TTL rollback. The original transient timer failures below remain historically accurate and are retained for debugging. Boot/power-loss rescue and production owner authorization remain **NOT tested**.

## New code reachable today

```bash
go run ./cmd/lunactl vm stage-intent testdata/host-baseline.json 8
```

`lunactl.run` parses policy using strict `policy.Decode`, then `vmstage.Build` runs the existing inert `nftpreview.Compile`, checks it is **UNHOOKED** with `KernelWrites=false`, derives a deterministic identifier from the policy content hash, test table name and declared TTL (the original pre-2026-10-15 implementation omitted the TTL), and outputs a bounded typed JSON **manifest**. Manifest contains policy identity/hash/revision, TTL **1–10 seconds**, exact four phase labels, `kernelWrites=false`, `ownerApproved=false`, and `scope=KVM_GUEST_PRIVATE_NETNS_ONLY`. `vmstage.Build` contains **no nft, shell, Unix socket, privileged executor, signer or kernel code**. Manifest is only a hint to a **synthetic VM fixture**, not a permit to enforce. TTL/invalid-policy negative tests exist under `internal/vmstage`.

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
