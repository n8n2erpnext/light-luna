#!/usr/bin/env bash
# Exact KVM VM only; deliberately SIGKILL its temporary keeper/watchdog
# service during a synthetic private netns transaction. No product apply.
set -euo pipefail
[[ "$(id -u)" != 0 ]] || { echo "NONROOT_LXD_OPERATOR_REQUIRED" >&2;exit 2; }
[[ "${LUNA_LXD_USE_SUDO:-0}" == 1 ]] || exit 2
[[ "${LUNA_P0_SERVICE_FAULT_ACK:-}" == "KILL_ONLY_ISOLATED_VM_LAB_SERVICES" ]] || {
 echo "EXPLICIT_VM_FAULT_ACK_REQUIRED" >&2;exit 2;
}
lxc(){ sudo -n /snap/bin/lxc "$@"; }
vm=light-luna-p0-nftvm
profile=light-luna-p0-kernel-vm
lxc list --format csv | grep -q "^$vm,RUNNING,,,VIRTUAL-MACHINE,"
lxc config show "$vm" --expanded | grep -qx -- "- $profile"
if lxc config show "$vm" --expanded | grep -Eq '^    type: nic$';then exit 3;fi
[[ "$(lxc exec "$vm" -- systemd-detect-virt)" == kvm ]]
[[ "$(lxc exec "$vm" -- ls -1 /sys/class/net)" == lo ]]
[[ "$(lxc exec "$vm" -- uname -r)" != "$(uname -r)" ]]
lxc exec "$vm" -- test -f /etc/light-luna-p0-kernel-vm
lxc exec "$vm" -- python3 -c '
import json
from pathlib import Path
p=Path("/opt/luna-p0-kernel-tests/stage-intent.json")
m=json.loads(p.read_text())
assert m["ttlSeconds"]==10 and m["mode"]=="SYNTHETIC_VM_ONLY_NO_EXECUTOR"
assert m["ownerApproved"] is False and m["kernelWrites"] is False
'
for file in guest-stage-test.sh guest-fault-evidence.py guest-archive-fault.py guest-boot-inspect.py;do
 lxc file push "scripts/lab/vm/$file" "$vm/opt/luna-p0-kernel-tests/$file" --mode=0700 >/dev/null
done
for mode in keeper_precommit keeper_postcommit watchdog_postcommit;do
 echo "BEGIN_VM_FAULT:$mode"
 lxc exec "$vm" -- env LUNA_VM_AUTO_TTL=1 LUNA_VM_FAULT="$mode" \
   LUNA_VM_FAULT_ACK=KILL_VM_SYNTHETIC_KEEPER_OR_WATCHDOG \
   bash /opt/luna-p0-kernel-tests/guest-stage-test.sh
 # The stage-test EXIT trap is only VM namespace cleanup; it must NOT
 # advance the incomplete journal or produce a false DONE receipt.
 lxc exec "$vm" -- python3 -c '
import json,subprocess
from pathlib import Path
p=Path("/opt/luna-p0-kernel-tests")
assert subprocess.check_output(["ip","netns","list"],text=True).strip()==""
assert subprocess.check_output(["nft","list","tables"],text=True).strip()==""
state=json.loads((p/"stage-journal.json").read_text())
fault=json.loads((p/"stage-fault-evidence.json").read_text())
assert state["state"]==fault["expectedJournal"]
assert not (p/"auto-watchdog-done.json").exists()
print("VM_FAULT_NO_KERNEL_ARTIFACTS_NO_FALSE_TERMINAL_PASS")
'
 # The no-ack attempt must be denied and must preserve the journal.
 if lxc exec "$vm" -- python3 /opt/luna-p0-kernel-tests/guest-archive-fault.py >/dev/null 2>&1;then
  echo "ARCHIVE_WITHOUT_ACK_WRONGLY_ACCEPTED" >&2;exit 3;
 fi
 lxc exec "$vm" -- env LUNA_VM_FAULT_ARCHIVE_ACK=ARCHIVE_ONLY_UNRECOVERED_VM_FAULT \
   python3 /opt/luna-p0-kernel-tests/guest-archive-fault.py
 echo "VM_FAULT_EVIDENCE_PRESERVED:$mode"
done
echo "VM_THREE_SERVICE_DEATH_FAULTS_DETECTED_AND_ARCHIVED_PASS"
