#!/usr/bin/env bash
# Crash after fsynced VM synthetic rollback journal, before fallback DONE.
# Guard EXACT disposable no-NIC KVM, never host/production.
set -euo pipefail
[[ "$(id -u)" != 0 && "${LUNA_LXD_USE_SUDO:-0}" == 1 ]] || exit 2
[[ "${LUNA_VM_TERMINAL_FAULT_ACK:-}" == "TEST_ONLY_GUEST_TERMINAL_WITHOUT_DONE" ]] || { echo ACK_REQUIRED >&2; exit 2; }
lxc(){ sudo -n /snap/bin/lxc "$@"; }
vm=light-luna-p0-nftvm
lxc list --format csv | grep -q "^$vm,RUNNING,,,VIRTUAL-MACHINE,"
lxc config show "$vm" --expanded | grep -qx -- '- light-luna-p0-kernel-vm'
if lxc config show "$vm" --expanded | grep -Eq '^    type: nic$';then exit 3;fi
[[ "$(lxc exec "$vm" -- systemd-detect-virt)" == kvm ]]
[[ "$(lxc exec "$vm" -- ls /sys/class/net)" == lo ]]
[[ "$(lxc exec "$vm" -- uname -r)" != "$(uname -r)" ]]
lxc exec "$vm" -- test -f /etc/light-luna-p0-kernel-vm
tag="${LUNA_VM_EVIDENCE_TAG:-terminal-first}"
[[ "$tag" =~ ^[a-z0-9-]{1,32}$ ]] || exit 2
lxc exec "$vm" -- python3 - <<'PY'
import json,subprocess
from pathlib import Path
root=Path("/opt/luna-p0-kernel-tests")
m=json.loads((root/"stage-intent.json").read_text())
j=json.loads((root/"stage-journal.json").read_text())
assert m["mode"]=="SYNTHETIC_VM_ONLY_NO_EXECUTOR" and m["ttlSeconds"]==10
assert m["kernelWrites"] is False and m["ownerApproved"] is False
assert j["state"]=="ROLLED_BACK" and j["bootId"]==Path("/proc/sys/kernel/random/boot_id").read_text().strip()
assert not subprocess.check_output(["nft","list","tables"],text=True).strip()
assert not subprocess.check_output(["ip","netns","list"],text=True).strip()
PY
for file in guest-stage-test.sh guest-auto-watchdog.py guest-fallback-supervisor.py guest-rollback-fence.py guest-rollback-receipts.py guest-fault-evidence.py guest-archive-fault.py guest-boot-inspect.py;do
  lxc file push "scripts/lab/vm/$file" "$vm/opt/luna-p0-kernel-tests/$file" --mode=0700 >/dev/null
done
lxc exec "$vm" -- env LUNA_VM_AUTO_TTL=1 LUNA_VM_FAULT=fallback_after_terminal \
   LUNA_VM_FAULT_ACK=KILL_VM_SYNTHETIC_KEEPER_OR_WATCHDOG \
   LUNA_VM_FALLBACK=1 LUNA_VM_FALLBACK_ACK=ENABLE_GUEST_SYNTHETIC_FALLBACK \
   LUNA_VM_FALLBACK_CRASH_ACK=VM_ONLY_CRASH_AFTER_JOURNAL_TERMINAL \
   bash /opt/luna-p0-kernel-tests/guest-stage-test.sh
lxc exec "$vm" -- python3 - <<'PY'
import importlib.util,json,os,subprocess
from pathlib import Path
root=Path("/opt/luna-p0-kernel-tests")
j=json.loads((root/"stage-journal.json").read_text())
m=json.loads((root/"stage-intent.json").read_text())
fault=json.loads((root/"stage-fault-evidence.json").read_text())
ready=json.loads((root/"auto-watchdog-ready.json").read_text())
assert j["state"]=="ROLLED_BACK" and j["history"][-1]=="ROLLED_BACK"
assert fault["mode"]=="fallback_after_terminal"
assert fault["terminalJournalWithoutDone"] is True
assert fault["status"]=="TERMINAL_JOURNAL_DONE_MISSING_REVIEW_REQUIRED"
assert not (root/"fallback-done.json").exists() and not (root/"auto-watchdog-done.json").exists()
assert not subprocess.check_output(["nft","list","tables"],text=True).strip()
assert not subprocess.check_output(["ip","netns","list"],text=True).strip()
path=root/"guest-rollback-receipts.py"
spec=importlib.util.spec_from_file_location("vm_terminal_inspect",path)
module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
r=module.inspect(m["stageId"],j["bootId"],ready["netnsInode"],ready)
assert r["status"]=="JOURNAL_TERMINAL_RECEIPT_PRESENT_DONE_NOT_EVALUATED",r
assert r["writer"]=="fallback" and r["phases"]==list(module.KINDS)
assert r["automaticRecoveryAuthorized"] is False
print("VM_TERMINAL_JOURNAL_WITHOUT_DONE_OBSERVED_NO_FALSE_RECEIPT_PASS")
PY
if lxc exec "$vm" -- python3 /opt/luna-p0-kernel-tests/guest-archive-fault.py >/dev/null 2>&1;then
 echo ARCHIVE_WITHOUT_ACK_ACCEPTED >&2;exit 3
fi
lxc exec "$vm" -- env LUNA_VM_FAULT_ARCHIVE_ACK=ARCHIVE_ONLY_UNRECOVERED_VM_FAULT LUNA_VM_FAULT_ARCHIVE_TAG="$tag" \
 python3 /opt/luna-p0-kernel-tests/guest-archive-fault.py
lxc exec "$vm" -- env LUNA_VM_EVIDENCE_TAG="$tag" python3 - <<'PY'
import json,os
from pathlib import Path
root=Path("/opt/luna-p0-kernel-tests"); tag=os.environ["LUNA_VM_EVIDENCE_TAG"]
for name in ("fallback-ready.json","auto-watchdog-ready.json"):
    src=root/name
    dest=root/(name.removesuffix(".json")+".terminal-without-done."+tag+".json")
    assert src.is_file() and not src.is_symlink() and not dest.exists()
    with dest.open("xb") as out:
        out.write(src.read_bytes());out.flush();os.fsync(out.fileno())
    src.unlink()
dir_fd=os.open(root,os.O_RDONLY|os.O_DIRECTORY)
try:os.fsync(dir_fd)
finally:os.close(dir_fd)
state=root/("stage-journal.receipt-missing-fallback_after_terminal."+tag+".json")
assert not (root/"stage-journal.json").exists()
assert json.loads(state.read_text())["state"]=="ROLLED_BACK"
assert not (root/"fallback-done.json").exists()
print("VM_TERMINAL_WITHOUT_DONE_EVIDENCE_ARCHIVED_REVIEW_REQUIRED_PASS")
PY
echo VM_GUEST_TERMINAL_WITHOUT_DONE_FAILURE_INJECTION_PASS
