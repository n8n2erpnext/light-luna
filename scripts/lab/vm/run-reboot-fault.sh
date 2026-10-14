#!/usr/bin/env bash
# Explicitly acknowledged, destructive reboot of EXACT disposable no-NIC KVM VM.
# NEVER for a production host, existing firewall or Cloud/NetBird.
# Does not reconcile an interrupted journal; preserves evidence for review.
set -euo pipefail
[[ "$(id -u)" != 0 ]] || { echo HOST_NONROOT_REQUIRED >&2;exit 2; }
[[ "${LUNA_P0_VM_REBOOT_ACK:-}" == "REBOOT_ISOLATED_LUNA_P0_VM" ]] || { echo EXPLICIT_LAB_ACK_REQUIRED >&2;exit 2; }
[[ "${LUNA_LXD_USE_SUDO:-0}" == "1" ]] || { echo APPROVED_LXD_SUDO_CONTEXT_REQUIRED >&2;exit 2; }
lxc(){ sudo -n /snap/bin/lxc "$@"; }
vm="light-luna-p0-nftvm"
profile="light-luna-p0-kernel-vm"
lxc list --format csv | grep -q "^$vm,RUNNING,,,VIRTUAL-MACHINE,"
lxc config show "$vm" --expanded | grep -qx -- "- $profile"
if lxc config show "$vm" --expanded | grep -Eq '^    type: nic$';then echo GUEST_NIC_CONFIGURED_REFUSE >&2;exit 3;fi
[[ "$(lxc exec "$vm" -- systemd-detect-virt)" == kvm ]]
[[ "$(lxc exec "$vm" -- ls -1 /sys/class/net)" == lo ]]
[[ "$(lxc exec "$vm" -- uname -r)" != "$(uname -r)" ]]
lxc exec "$vm" -- test -f /etc/light-luna-p0-kernel-vm
lxc file push scripts/lab/vm/guest-boot-inspect.py "$vm/opt/luna-p0-kernel-tests/guest-boot-inspect.py" --mode=0700 >/dev/null
lxc file push scripts/lab/vm/guest-prepare-reboot-fault.sh "$vm/opt/luna-p0-kernel-tests/guest-prepare-reboot-fault.sh" --mode=0700 >/dev/null
# Preparation creates ONLY a test table inside a guest-only private namespace,
# with no watchdog intentionally; it verifies an ICMP drop and fsynced journal.
lxc exec "$vm" -- bash /opt/luna-p0-kernel-tests/guest-prepare-reboot-fault.sh
echo "VM_SYNTHETIC_COMMITTED_PRE_FORCE_REBOOT_PASS"
# Intentional power cycle; failure here leaves an isolated VM-only rule.
lxc stop "$vm" --force
lxc start "$vm"
ready=0
for _ in $(seq 1 45);do
 if lxc exec "$vm" -- test -f /opt/luna-p0-kernel-tests/boot-fault-pre.json >/dev/null 2>&1;then
   ready=1;break
 fi
 sleep 1
done
[[ "$ready" == 1 ]] || { echo VM_RESTART_NOT_READY_KEEP_EVIDENCE >&2;exit 3; }
# The inspector MUST refuse automatic recovery and exit 3 on stale boot.
set +e
output="$(lxc exec "$vm" -- python3 /opt/luna-p0-kernel-tests/guest-boot-inspect.py)"
rc=$?
set -e
[[ "$rc" == 3 ]] || { echo UNEXPECTED_INSPECTOR_EXIT_$rc >&2;exit 3; }
printf '%s\n' "$output" | python3 -c '
import json,sys
s=json.load(sys.stdin)
assert s["status"]=="CROSS_BOOT_NO_KERNEL_STATE_OBSERVED_REVIEW_REQUIRED",s
assert s["kernel"]["rootNftEntries"]==0 and not s["kernel"]["namedNetnsPresent"]
assert s["kernel"]["nonRootNetnsProcesses"]==0
assert s["autoReconciliationAllowed"] is False and s["kernelWrites"] is False
print("VM_READ_ONLY_STALE_BOOT_DETECTION_PASS")
'
lxc exec "$vm" -- python3 -c '
import json
from pathlib import Path
b=Path("/opt/luna-p0-kernel-tests")
old=json.loads((b/"boot-fault-pre.json").read_text())
state=json.loads((b/"stage-journal.json").read_text())
new=Path("/proc/sys/kernel/random/boot_id").read_text().strip()
assert old["preBootId"]!=new
assert state["bootId"]==old["preBootId"] and state["state"]=="COMMITTED_SYNTHETIC"
print("VM_REAL_BOOT_ID_CHANGED_AND_JOURNAL_PRESERVED_PASS")
'
# Even operator's old command to finish transaction is rejected (read-only
# negative action; the rejected transition must NOT change the journal).
if lxc exec "$vm" -- python3 /opt/luna-p0-kernel-tests/guest-stage-journal.py transition ROLLED_BACK >/dev/null 2>&1;then
 echo STALE_JOURNAL_WRONGFULLY_ADVANCED >&2;exit 3
fi
lxc exec "$vm" -- python3 -c '
import json
from pathlib import Path
p=Path("/opt/luna-p0-kernel-tests/stage-journal.json")
assert json.loads(p.read_text())["state"]=="COMMITTED_SYNTHETIC"
'
echo "VM_FORCED_REBOOT_STALE_JOURNAL_BLOCKED_PASS"
echo "VM_MANUAL_REVIEW_AND_EXPLICIT_EVIDENCE_PRESERVATION_REQUIRED"
