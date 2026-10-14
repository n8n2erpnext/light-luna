# P0 autonomous TTL watchdog — independent KVM guest lab

**Scope:** HOMESV `light-luna-p0-nftvm` (LXD KVM VM, separate Linux kernel, NIC-less). Development-only synthetic nft table, not product kernel enforcement.
**Date:** 2026-10-14. Branch `feat/p0-standalone-foundation`.
**Status:** **3/3 consecutive full KVM VM suites PASS after the final keeper/cleanup fix**. The automatically triggered synthetic rollback was observed in all three runs, including restored IPv4/IPv6 TCP/UDP. No production or G-STANDALONE claim.

## Boot identity hardening — source + negative tests (2026-10-14)

The VM-only `guest-stage-journal.py` records the Linux kernel `/proc/sys/kernel/random/boot_id` on `STAGED` and refuses later transitions from a different boot or from legacy records without a boot ID. `guest-auto-watchdog.py` checks that ID before arming and before namespace-targeted rollback, and records it in READY/DONE receipts. `guest-stage-test.sh` checks receipts; `reconcile-interrupted-stage.sh` refuses cross-boot journal reuse. This prevents stale PID/netns/monotonic records being accepted as if they belonged to the new boot.

[Offline regression tests](../../scripts/lab/vm/test_boot_journal.py) cover same-boot transitions, simulated reboot denial, missing boot ID, duplicate JSON keys and invalid boot ID. **Not a real VM reboot, a power-loss test, a boot-time reconciler or a product rollback guarantee.** A stale-boot transaction still needs independent guest-kernel inspection and explicit operator reconciliation; no automatic journal state transition is authorized.

The regression after boot-ID hardening included one complete KVM runner with exit 0, retained the `VM_AUTO_TTL_EXPIRY_NO_TRIGGER_ROLLBACK_PASS` marker, and restored synthetic IPv4/IPv6 TCP/UDP. A harmless shell diagnostic (`wait_for: No record of process`) appeared on stderr in the old ICMP guest test; it did not fail the suite and still merits a separate harness review. This latest run is not an additional three-run statistical claim.

## Actual forced KVM reboot and read-only inspection (2026-10-14)

**Lab verification only — not production boot recovery.** On HOMESV's explicitly selected `light-luna-p0-nftvm` (KVM, independent kernel, no NIC), the acknowledged [fault runner](../../scripts/lab/vm/run-reboot-fault.sh) and [guest preparation](../../scripts/lab/vm/guest-prepare-reboot-fault.sh) installed a **fixed synthetic nft table only in a temporary guest network namespace** and proved an ICMP drop counter. Journal state was `COMMITTED_SYNTHETIC`, intentionally **without** arming a watchdog. The host then ran `lxc stop --force` and `lxc start` on the exact VM, not HOMESV or a customer device.

Observed Linux boot ID changed across the restart; the exact ephemeral UUIDs remain only in the private guest's evidence files, not in the public repository. The [read-only inspector](../../scripts/lab/vm/guest-boot-inspect.py) reported `CROSS_BOOT_NO_KERNEL_STATE_OBSERVED_REVIEW_REQUIRED`: guest root nft inventory empty, no named netns or foreign process netns observed. Journal retained `COMMITTED_SYNTHETIC` with old boot ID, and a stale `ROLLED_BACK` transition was denied. Runner exited **0** with `VM_FORCED_REBOOT_STALE_JOURNAL_BLOCKED_PASS`. This confirms fail-closed **detection** of a reboot-interrupted transaction, **not** a functioning automatic boot-time reconciler. Namespace scope checking is not exhaustive against malicious root/fd-held namespaces.

After independent observation, the explicitly acknowledged [manual archival helper](../../scripts/lab/vm/guest-archive-stale-boot.py) preserved the stale journal unchanged with a separate review receipt, **never wrote `ROLLED_BACK` or nft state**. The absent-ack test was denied and left the stale journal intact. The archive is evidence, not a recovery success attestation. No action is automatically performed by the inspector.

The [14 offline classification tests](../../scripts/lab/vm/test_boot_inspect.py) and earlier five boot-ID tests PASS. **One complete post-reboot KVM regression PASS**, including ICMP, IPv4/IPv6 TCP/UDP drops, explicit service rollback and autonomous TTL rollback in the later independent synthetic test. Go tests (`-count=3`) and vet PASS. **Still unproven:** power loss during journal fsync or product commit; independent watchdog restart after power loss; actual privileged executor; signed owner approval; real firewall rule, network admin reachability, Docker/UFW coexistence or recovery guarantees. **P0 OPEN; G-STANDALONE BLOCKED.**

## KVM guest service death fault injection — keeper/watchdog, three scenarios (2026-10-14)

**Status: 3/3 real SIGKILL fault cases PASS for detection and evidence archival, NOT for self-recovery.** Starting from development `feat/p0-standalone-foundation@c6090fec8dd8422445efcc4f50f873ea7aacaabb`, [isolated host runner](../../scripts/lab/vm/run-service-faults.sh) verified HOMESV's exact `light-luna-p0-nftvm` remains a KVM **virtual machine with no NIC** and a different Linux kernel. The temporary systemd keeper/watchdog units were created by the pre-existing [guest synthetic stage test](../../scripts/lab/vm/guest-stage-test.sh). Only their **named MainPID processes** were sent SIGKILL using `systemctl kill --kill-who=main`; no production process, host firewall, Cloud or NetBird was touched. The regular unhooked product nft compiler output was **never** applied.

- **`keeper_precommit`**: independent watchdog READY was observed after journal `VALIDATED`; the keeper was killed before synthetic commit. After the monotonic deadline, no autonomous DONE receipt appeared, and the state stayed `VALIDATED`. PASS means failure detection, not successful rollback (there was no committed nft hook).
- **`keeper_postcommit`**: committed guest-only synthetic TCP/UDP/IPv6 drops with positive nft counters were observed; killing keeper broke the pinned PID/netns identity. After deadline, the fixed `inet luna_p0_stage` synthetic table was **still present** in the guest's named test namespace. Journal remained `COMMITTED_SYNTHETIC`; there was **no** DONE receipt. PASS means the unsafe recovery was not falsely acknowledged. **It also reveals absence of a working backup rollback path.**
- **`watchdog_postcommit`**: similarly committed and verified synthetic packet drops; killing the watchdog's service MainPID left the same synthetic table active after expiry. No DONE, journal stayed `COMMITTED_SYNTHETIC`. Again this is a **detected failure**, not a recovery pass.

[Guest read-only evidence observer](../../scripts/lab/vm/guest-fault-evidence.py) waited beyond the armed deadline, verified a specific failed unit, matching stage/boot identities, absent DONE receipt and incomplete journal, and fsynced a **`FAILURE_DETECTED_NOT_RECOVERED`** record. The **original guest-stage-test EXIT trap** then removed only its synthetic table and guest namespace fixtures. A separate explicitly acknowledged [manual archive tool](../../scripts/lab/vm/guest-archive-fault.py) confirmed no remaining root nft/named/foreign-process netns and archived the incomplete journal plus fault evidence **without writing `ROLLED_BACK`**. No-ack archival was denied in each case. The host run exited 0 with `VM_THREE_SERVICE_DEATH_FAULTS_DETECTED_AND_ARCHIVED_PASS`, no stderr. The three on-guest archives are evidence, not signed/audited product receipts. A separate missing-ACK negative test on `guest-stage-test.sh` was denied **before journal modification**. The regular VM suite was then rerun with default `LUNA_VM_FAULT=none` and exited 0: unhooked `nft -c`, synthetic ICMP/TCP/UDP/IPv6, explicitly triggered rollback and autonomous TTL rollback all PASS; the final guest had terminal `ROLLED_BACK`, no root nft tables or named netns. Go test x3 and vet PASS; 19 earlier offline boot tests PASS. Two `wait_for: No record of process` stderr notices (old ICMP and stage fixture) remained nonfatal and must be investigated as test-harness reliability issues. These do not establish a percentile or a production service restart SLA.

**Remaining blockers:** no independent watchdog replacement, no recovery if its keeper dies, no owner-authenticated executor, no production service identity, no adversarial process/FD-held netns coverage, no real product enforcement, no recovery after disk/power loss, no race/concurrency proof, and no SLA. A precommit fault is not a rollback benchmark; a successful EXIT trap is not persistent autonomous rescue. **P0 OPEN, G-STANDALONE BLOCKED.**

## Dual-armed arbitration and after-delete crash point — no-NIC KVM verified (2026-10-14)

**Status: development-only VM test PASS for two bounded observations, not transaction recovery.** Source parent `feat/p0-standalone-foundation@f5f97ada0fb4725706e91d6e7d481e40256ceeb6`. [Explicitly acknowledged VM runner](../../scripts/lab/vm/run-race-crash-tests.sh) exercises the exact `light-luna-p0-nftvm` (KVM, distinct kernel, loopback-only) using the existing fixed synthetic guest-netns IPv4/IPv6 TCP/UDP/ICMPv6 table. It never applies the product policy compiler's preview or touches HOMESV networking.

**Case A — primary/fallback both READY, primary remains healthy (`primary_race`).** Both independent PID1-managed processes were armed before the synthetic table was installed. At TTL the primary watchdog removed the table, journal reached `ROLLED_BACK` and primary emitted `AUTO_ROLLBACK_CONFIRMED`; at deadline + 1s the fallback observed the already terminal journal, recorded `PRIMARY_ALREADY_TERMINAL_NO_TAKEOVER` as **rejection**, emitted no fallback DONE and did not attempt a second deletion. IPv4/IPv6 TCP/UDP were restored. This is *one controlled dual-armed arbitration observation*, **not** a simultaneous last-millisecond race test or multi-process lock proof.

**Case B — fallback exits after nft deletion but before journal/DONE (`fallback_after_delete`).** The primary watchdog was SIGKILLed after an observed synthetic drop; the independently pinned fallback verified the strict fixed nft shape, deleted the guest synthetic table and checked absence, then deliberately called `os._exit(72)` at the exact **VM-only failpoint** before `guest-stage-journal.py` advances state. After the monotonic deadline, [guest fault observer](../../scripts/lab/vm/guest-fault-evidence.py) saw that the fallback service MainPID had exited, the synthetic table was absent, **journal still `COMMITTED_SYNTHETIC`**, and neither primary nor fallback had a success DONE. It wrote fsynced `FAILURE_DETECTED_NOT_RECOVERED` evidence with `kernelAbsentJournalIncomplete=true` — not `AUTO_ROLLBACK_CONFIRMED`. The stage-test EXIT handler subsequently cleaned only the temporary namespaces. [Manual archive](../../scripts/lab/vm/guest-archive-fault.py) rejected missing ACK, and on explicit ACK preserved the original incomplete journal and fault record unchanged; the fallback READY was separately saved under a unique tag. The runner exited zero with `VM_RACE_AND_AFTER_DELETE_CRASH_CONSISTENCY_DETECTED_PASS`.

**Measured safety boundary:** fail-closed *observation* of the commit/kernel divergence, not repaired atomicity. After both cases the VM was checked: only `lo`, root nft inventory empty, no named guest test netns, active journal archived, and unique test evidence retained in the VM. The existing single-controller JSON journal has no ownership fencing, multi-writer lock or atomic kernel+disk commit. **Post-fault default KVM regression** also completed with exit code 0 and no stderr: original unhooked `nft -c`, independent explicit revert, primary TTL autonomy and synthetic IPv4/IPv6 TCP/UDP accepted/dropped/restored as expected. This regression does **not** repair the separate observed after-delete journal inconsistency. **Open gates:** crash during nft write/delete, crash after journal fsync before DONE, primary-vs-fallback simultaneous delete race, fallback reboot/power-loss, fd-held/foreign netns and hostile-root cases. No production self-healing, signed authority, installer or remote reachability claim. **P0 OPEN; G-STANDALONE BLOCKED.**

## Independent VM-only fallback after keeper/watchdog SIGKILL — lab pass (2026-10-14)

**Verification, not a production recovery system.** Starting from development baseline `feat/p0-standalone-foundation@73518264e2e9de2e2926fc482fc4421fe452701e`, the [independent fallback prototype](../../scripts/lab/vm/guest-fallback-supervisor.py) is launched as its own PID1-managed systemd transient service **before** synthetic nft commit. It executes `nsenter -n` once while the verified keeper is still alive, then **holds the guest test netns itself**; subsequent rollback does not need the keeper PID to remain available. The fallback matches primary watchdog's boot ID, stage ID, TTL deadline, keeper process start ticks and target netns inode and produces an fsynced READY file before any commit.

After deadline + 1 second, it checks the current boot and **its own** netns inode, exact stage journal and whether the named primary watchdog or keeper process has died. Only when `COMMITTED_SYNTHETIC` remains outstanding does it inspect the **exact nftables JSON** from the guest-only netns: one fixed `inet luna_p0_stage` table, one input chain with filter/input/-50/accept, and exactly three matching synthetic IP+TCP:18080, IP+UDP:18081 and IPv6+ICMPv6 echo-request counter+DROP rules. Foreign chain, changed port/source, added/duplicate rule or reordered expression is denied. It can delete **only** that fixed test table, confirm absence, atomically progress the single-controller journal to `ROLLED_BACK` and write its own `SYNTHETIC_FALLBACK_ROLLBACK_CONFIRMED` receipt. It never accepts a table name from an external request. This is a **VM-only privileged prototype** and not owner-approved Light Luna policy execution.

The [explicit-ack host runner](../../scripts/lab/vm/run-fallback-faults.sh) exercised **two actual SIGKILL cases** in the same exact no-NIC KVM VM: `keeper_postcommit` and `watchdog_postcommit`. In each, packet drop counters were positive before the fault, the primary DONE was absent, and the independent fallback returned positive evidence **before the existing test EXIT cleanup**, restoring synthetic IPv4/IPv6 TCP/UDP and terminal journal. Two runs of both cases PASS, including the second run using a strictly verified nft JSON rule shape. The runner retains per-mode/verification-tag records without overwriting earlier test evidence. After the last run, guest had only `lo`, empty root nft/named netns and `ROLLED_BACK` journal. **A full independent normal KVM regression also completed with exit code 0**, including nft syntax, IPv4/IPv6 packet probes and primary autonomous TTL rollback without enabling fallback. One nonfatal legacy ICMP shell `wait_for: No record of process` stderr diagnostic remains an open harness reliability item. A missing fallback-ACK negative call was also denied **before journal mutation**, leaving the last terminal `ROLLED_BACK` state unchanged.

[Offline parser negative tests](../../scripts/lab/vm/test_fallback_supervisor.py) cover 12 synthetic rule-shape cases (including changed port, source and duplicate rules). A real KVM nft JSON probe was used to determine expected shape, and its temporary table/netns were removed immediately afterward. **Not yet proven:** fallback service death, failure during nft delete↔fsynced journal/DONE, malicious root/FD-held namespace, concurrency, actual power loss/reboot, packaging or health-managed service restarts; primary completion racing fallback also needs explicit live testing. Existing service-death detection tests without opt-in fallback are historical independent evidence, not invalidated. **P0 OPEN, G-STANDALONE BLOCKED.**

## Why the first watchdog attempt was not accepted

Earlier source code proved only that a **separately triggered service** can undo a synthetic nft rule. It did **not** prove a service autonomously fires when TTL expires. Attempts using detached `setsid`, `systemd-run` with an internal `sleep`, and transient systemd timers intermittently failed to complete. Their failures were recorded and the gate remained open.

We identified two independent VM-test harness concerns:

1. `ip netns` names created from `lxc exec` may not be visible in the VM's systemd PID1 **mount namespace**. A rollback service cannot rely on a named netns being visible. The systemd process must target the network namespace of a **verified process PID**.
2. A test shell's short-lived background PID could disappear when the simulated controller terminated. The guest suite now uses a **PID1-managed systemd namespace keeper** started via `nsenter -t <initial-anchor-pid> -n sleep 240`. It validates `/proc/<keeper-pid>/ns/net` matches the expected namespace inode before terminating the original controller-owned anchor. [Keeper smoke fixture](../../scripts/lab/vm/guest-keeper-smoke.sh) explicitly demonstrated the keeper survives that original anchor being terminated. Both legacy ICMP and new stage tests use the keeper. Cleanup is scoped to the **top-level Bash PID** to prevent child/subshell exit traps stopping the keeper early.

## VM-only prototype runtime boundary

| Module | Implemented VM proof | Production nonclaim |
|---|---|---|
| `vmstage.Build` and `lunactl vm stage-intent` | Inert, hash-identified source manifest with bounded TTL | Not a policy permit or owner approval |
| [Guest single-controller journal](../../scripts/lab/vm/guest-stage-journal.py) | Rejects invalid schema/TTL/stage ID, repeated and out-of-order transitions; fsync+atomic rename | No multi-client locking, signed receipt, boot epoch or hostile-root protection |
| [Independent keeper](../../scripts/lab/vm/guest-keeper-smoke.sh) | VM systemd service owns stable netns PID independent of controller-owned shell | Not a production service/account installer |
| [Autonomous watchdog](../../scripts/lab/vm/guest-auto-watchdog.py) | Separate VM PID1 service; validates root/KVM/no NIC, manifest, journal state, process start ticks+netns inode; emits fsynced READY before any rule is installed | No production privileged executor, durable rescue, hardware clock/host reboot proof |
| [Synthetic nft stage test](../../scripts/lab/vm/guest-stage-test.sh) | Requires READY and ≥5 seconds remaining before committing fixed VM-netns table; drops synthetic TCP/UDP/IPv6, kills controller placeholder, **does not trigger rollback** in auto mode | Never applies Light Luna's unhooked nft compiler output |
| [VM host runner](../../scripts/lab/vm/run-tests.sh) | Runs original unhooked grammar test, explicit separate-service revert and new autonomous TTL test with TTL 10s | No Docker/UFW/NetBird coexistence, outside reachability or actual customer hosts |

## State and deadline sequence

```text
HOMESV (nonprivileged CLI)                   KVM guest without NIC
        │                                        │
 typed inert manifest --------------------> exact schema/hash/TTL validation
                                               STAGED → VALIDATED
                                               │
                                independent systemd keeper holds target netns
                                               │
                           watchdog (PID1) checks expected PID birth + net inode
                                               │
                           verifies synthetic table does not yet exist
                                               │
                 atomic READY with monotonic arm/deadline/TTL and stage ID
                                               │
                    test refuses commit unless >=5s before expiry
                                               │
             ONLY FIXED GUEST-NETNS SYNTHETIC nft hook applied
                                               │
                            journal COMMITTED_SYNTHETIC
                                               │
                            positive drop counters and probes
                                               │
                            simulated controller terminated
                                               │
                           watchdog wakes at monotonic deadline
                                               │
                     verify keeper PID birth+netns inode again
                                               │
                     nft delete ONLY `inet luna_p0_stage`
                                               │
                         verify table no longer exists
                                               │
                       journal terminal ROLLED_BACK
                                               │
                    atomic DONE receipt with elapsed/lag evidence
                                               │
                   verify IPv4/IPv6, TCP/UDP probes restored
```

The deadline is **relative to arming**, not commit. This is intentionally conservative: commit occurs only if there are at least five seconds remaining. The lab requires `DONE.status=AUTO_ROLLBACK_CONFIRMED`, journal terminal `ROLLED_BACK`, no nft table after cleanup, and measured monotonic completion lag ≤3.5s. These numeric limits are **test acceptance settings**, not an SLA or guaranteed production latency.

## Safety behavior / negative cases

- When PID, process birth time or network namespace inode changes before rollback, **refuse unsafe namespace entry** rather than possibly modifying an unrelated namespace. This needs a separate rescue solution in the eventual product, because refusing the wrong target is not the same as restoring a valid target.
- If a transaction is not actually committed at expiry, no success receipt is recorded; a synthetic table seen unexpectedly is still removed in a fail-closed way.
- If journal is corrupt, missing or a table disappears before verification, do not issue `AUTO_ROLLBACK_CONFIRMED`.
- The separate service never accepts a table name from the product or user; it can touch only the fixed `luna_p0_stage` table in the verified VM target namespace.
- No LXD NIC, LXD bridge, host nft apply, Cloud, NetBird, live WAF or eBPF/XDP involvement.
- Reconciliations after interrupted negative test runs are **manual and evidence-based**; no silent journal reset.

## Evidence and follow-up

The first complete VM run returned `VM_AUTO_WATCHDOG_PRECOMMIT_READY_PASS`, `VM_AUTO_TTL_EXPIRY_NO_TRIGGER_ROLLBACK_PASS`, `VM_STAGE_AUTONOMOUS_TTL_IPV6_TCP_UDP_RESTORED_PASS` and `VM_KERNEL_STAGE_AND_AUTONOMOUS_TTL_TESTS_EXECUTED`. An independent receipt reported monotonic completion `0.079s` after its scheduled deadline in that one observation. **This is one measured result, not a latency percentile or reliability guarantee.**

**Final acceptance evidence:** Three consecutive post-fix full VM suites returned exit 0 and the marker `FULL_LAB_THREE_POSTFIX_CONSECUTIVE_PASS`; each auto phase emitted `VM_AUTO_TTL_EXPIRY_NO_TRIGGER_ROLLBACK_PASS` and restored IPv4/IPv6 TCP/UDP. Final observed journal was `ROLLED_BACK`, no guest nft tables or named netns. Failures before the keeper/cleanup fix were not counted as passing runs. One readable final per-run receipt remains in the guest; no aggregate latency percentile is claimed. **Still not tested:** actual process/executor death rather than placeholder, VM restart during pending transaction, boot reconciliation, power-loss/snapshot rollback, host firewall coexistence, multiple clients, hostile VM root, signed owner auth, load/stress/race, release installation or ARM compatibility. Keep G-STANDALONE blocked.
