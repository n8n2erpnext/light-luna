#!/usr/bin/env bash
# Explicit-ACK no-NIC KVM filesystem fault injection; NO nft writes.
# Runs receipt tests in temporary guest dirs, never against active journal.
set -euo pipefail
[[ "$(id -u)" != 0 && "${LUNA_LXD_USE_SUDO:-0}" == 1 ]] || exit 2
[[ "${LUNA_VM_RECEIPT_FSYNC_ACK:-}" == "TEST_ONLY_GUEST_TEMP_RECEIPT_FSYNC" ]] || {
  echo "VM_RECEIPT_FSYNC_LAB_ACK_REQUIRED" >&2
  exit 2
}
lxc() { sudo -n /snap/bin/lxc "$@"; }
vm=light-luna-p0-nftvm
lxc list --format csv | grep -q "^$vm,RUNNING,,,VIRTUAL-MACHINE,"
lxc config show "$vm" --expanded | grep -qx -- '- light-luna-p0-kernel-vm'
if lxc config show "$vm" --expanded | grep -Eq '^    type: nic$';then exit 3;fi
[[ "$(lxc exec "$vm" -- systemd-detect-virt)" == kvm ]]
[[ "$(lxc exec "$vm" -- ls /sys/class/net)" == lo ]]
[[ "$(lxc exec "$vm" -- uname -r)" != "$(uname -r)" ]]
lxc exec "$vm" -- test -f /etc/light-luna-p0-kernel-vm
# This experiment MUST start from a terminal, empty synthetic VM.
lxc exec "$vm" -- python3 - <<'PY'
import json,subprocess,os
from pathlib import Path
p=Path("/opt/luna-p0-kernel-tests")
m=json.loads((p/"stage-intent.json").read_text())
j=json.loads((p/"stage-journal.json").read_text())
assert m["mode"]=="SYNTHETIC_VM_ONLY_NO_EXECUTOR" and not m["kernelWrites"] and not m["ownerApproved"]
assert j["state"]=="ROLLED_BACK" and j["history"][-1]=="ROLLED_BACK"
assert j["bootId"]==Path("/proc/sys/kernel/random/boot_id").read_text().strip()
assert sorted(os.listdir("/sys/class/net"))==["lo"]
assert not subprocess.check_output(["nft","list","tables"],text=True).strip()
assert not subprocess.check_output(["ip","netns","list"],text=True).strip()
print("VM_FSYNC_LAB_GUEST_PRECHECK_PASS")
PY
before="$(lxc exec "$vm" -- sha256sum /opt/luna-p0-kernel-tests/stage-journal.json)"
for file in guest-rollback-receipts.py test_rollback_fsync_faults.py;do
  lxc file push "scripts/lab/vm/$file" "$vm/opt/luna-p0-kernel-tests/$file" --mode=0644 >/dev/null
done
lxc exec "$vm" -- python3 /opt/luna-p0-kernel-tests/test_rollback_fsync_faults.py
after="$(lxc exec "$vm" -- sha256sum /opt/luna-p0-kernel-tests/stage-journal.json)"
[[ "$before" == "$after" ]] || { echo "VM_ACTIVE_JOURNAL_CHANGED_DURING_FSYNC_TEST" >&2;exit 3; }
lxc exec "$vm" -- sh -c 'test -z "$(nft list tables)" && test -z "$(ip netns list)" && test "$(ls /sys/class/net)" = lo'
echo "VM_FSYNC_TEMP_DIR_ERRORS_NO_ACTIVE_JOURNAL_OR_NFT_CHANGE_PASS"
