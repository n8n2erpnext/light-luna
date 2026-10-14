#!/usr/bin/env bash
# Exact disposable no-NIC KVM only: primary-vs-fallback arbitration and
# crash after a fixed synthetic nft table deletion, before journal fsync.
set -euo pipefail
[[ "$(id -u)" != 0 && "${LUNA_LXD_USE_SUDO:-}" == 1 ]] || exit 2
[[ "${LUNA_VM_RACE_CRASH_ACK:-}" == "TEST_ONLY_NIC_LESS_KVM_RACE_CRASH" ]] || { echo ACK_REQUIRED >&2;exit 2; }
lxc(){ sudo -n /snap/bin/lxc "$@"; }
vm=light-luna-p0-nftvm
lxc list --format csv | grep -q "^$vm,RUNNING,,,VIRTUAL-MACHINE,"
lxc config show "$vm" --expanded | grep -qx -- '- light-luna-p0-kernel-vm'
if lxc config show "$vm" --expanded | grep -Eq '^    type: nic$';then exit 3;fi
[[ "$(lxc exec "$vm" -- systemd-detect-virt)" == kvm ]]
[[ "$(lxc exec "$vm" -- ls /sys/class/net)" == lo ]]
[[ "$(lxc exec "$vm" -- uname -r)" != "$(uname -r)" ]]
lxc exec "$vm" -- test -f /etc/light-luna-p0-kernel-vm
lxc exec "$vm" -- python3 -c '
import json,subprocess
from pathlib import Path
p=Path("/opt/luna-p0-kernel-tests")
m=json.loads((p/"stage-intent.json").read_text())
j=json.loads((p/"stage-journal.json").read_text())
assert m["mode"]=="SYNTHETIC_VM_ONLY_NO_EXECUTOR" and m["ttlSeconds"]==10
assert m["kernelWrites"] is False and m["ownerApproved"] is False
assert j["state"]=="ROLLED_BACK" and j["bootId"]==Path("/proc/sys/kernel/random/boot_id").read_text().strip()
assert not subprocess.check_output(["nft","list","tables"],text=True).strip()
assert not subprocess.check_output(["ip","netns","list"],text=True).strip()
'
for file in guest-stage-test.sh guest-fallback-supervisor.py guest-fault-evidence.py guest-archive-fault.py guest-boot-inspect.py guest-auto-watchdog.py guest-rollback-fence.py;do
 lxc file push "scripts/lab/vm/$file" "$vm/opt/luna-p0-kernel-tests/$file" --mode=0700 >/dev/null
done
tag="${LUNA_VM_EVIDENCE_TAG:-race-crash-initial}"
[[ "$tag" =~ ^[a-z0-9-]{1,32}$ ]] || exit 2
# A healthy primary must complete first; fallback is ALIVE but must not
# double-delete or create a second completion receipt.
lxc exec "$vm" -- env LUNA_VM_AUTO_TTL=1 LUNA_VM_FAULT=primary_race \
  LUNA_VM_FAULT_ACK=KILL_VM_SYNTHETIC_KEEPER_OR_WATCHDOG \
  LUNA_VM_FALLBACK=1 LUNA_VM_FALLBACK_ACK=ENABLE_GUEST_SYNTHETIC_FALLBACK \
  LUNA_VM_FALLBACK_RACE_ACK=VM_ONLY_PRIMARY_FALLBACK_SAME_DEADLINE \
  bash /opt/luna-p0-kernel-tests/guest-stage-test.sh
lxc exec "$vm" -- env LUNA_VM_EVIDENCE_TAG="$tag" python3 - <<'PY'
import json,os,subprocess
from pathlib import Path
p=Path("/opt/luna-p0-kernel-tests")
r=json.loads((p/"fallback-rejected.json").read_text())
state=json.loads((p/"stage-journal.json").read_text())
primary=json.loads((p/"auto-watchdog-done.json").read_text())
assert r["status"] in ("PRIMARY_ALREADY_TERMINAL_NO_TAKEOVER","PRIMARY_AND_KEEPER_STILL_ACTIVE_REFUSED_TAKEOVER")
assert primary["status"]=="AUTO_ROLLBACK_CONFIRMED"
assert state["state"]=="ROLLED_BACK"
assert not (p/"fallback-done.json").exists()
assert not subprocess.check_output(["nft","list","tables"],text=True).strip()
assert not subprocess.check_output(["ip","netns","list"],text=True).strip()
tag=os.environ["LUNA_VM_EVIDENCE_TAG"]
for name in ("fallback-ready.json","fallback-rejected.json","auto-watchdog-done.json"):
    src=p/name;dst=p/(name.removesuffix(".json")+".primary-race."+tag+".json")
    assert not dst.exists()
    with dst.open("xb") as f:
        f.write(src.read_bytes());f.flush();os.fsync(f.fileno())
for name in ("fallback-ready.json","fallback-rejected.json"):
    (p/name).unlink()
dir_fd=os.open(p,os.O_RDONLY|os.O_DIRECTORY)
os.fsync(dir_fd);os.close(dir_fd)
print("VM_PRIMARY_AND_FALLBACK_DUAL_ARM_NO_DOUBLE_SUCCESS_PASS")
PY
if [[ "${LUNA_P0_RACE_ONLY:-0}" == 1 ]];then
  echo "VM_CONCURRENT_DEADLINE_ADVISORY_FENCE_RACE_ONLY_PASS"
  exit 0
fi
# Inject a hard failure AFTER the fallback sees nft table deleted but BEFORE
# the lab journal transitions or a DONE receipt can be persisted.
lxc exec "$vm" -- env LUNA_VM_AUTO_TTL=1 LUNA_VM_FAULT=fallback_after_delete \
  LUNA_VM_FAULT_ACK=KILL_VM_SYNTHETIC_KEEPER_OR_WATCHDOG \
  LUNA_VM_FALLBACK=1 LUNA_VM_FALLBACK_ACK=ENABLE_GUEST_SYNTHETIC_FALLBACK \
  LUNA_VM_FALLBACK_CRASH_ACK=VM_ONLY_CRASH_AFTER_SYNTHETIC_DELETE \
  bash /opt/luna-p0-kernel-tests/guest-stage-test.sh
lxc exec "$vm" -- python3 - <<'PY'
import json,subprocess
from pathlib import Path
p=Path("/opt/luna-p0-kernel-tests")
j=json.loads((p/"stage-journal.json").read_text())
f=json.loads((p/"stage-fault-evidence.json").read_text())
assert j["state"]=="COMMITTED_SYNTHETIC"
assert f["mode"]=="fallback_after_delete"
assert f["syntheticRuleStillPresent"] is False
assert f["kernelAbsentJournalIncomplete"] is True
assert not (p/"fallback-done.json").exists()
assert not (p/"auto-watchdog-done.json").exists()
assert not subprocess.check_output(["nft","list","tables"],text=True).strip()
assert not subprocess.check_output(["ip","netns","list"],text=True).strip()
print("VM_AFTER_NFT_DELETE_BEFORE_JOURNAL_CRASH_STALE_COMMIT_NO_DONE_PASS")
PY
if lxc exec "$vm" -- python3 /opt/luna-p0-kernel-tests/guest-archive-fault.py >/dev/null 2>&1;then
 echo ARCHIVE_MISSING_ACK_WRONGLY_ACCEPTED >&2;exit 3
fi
lxc exec "$vm" -- env LUNA_VM_FAULT_ARCHIVE_ACK=ARCHIVE_ONLY_UNRECOVERED_VM_FAULT \
  LUNA_VM_FAULT_ARCHIVE_TAG="$tag" \
  python3 /opt/luna-p0-kernel-tests/guest-archive-fault.py
# Archive the non-success READY alongside the error record; never turn it
# into DONE. No active rule can survive post-test guest-only cleanup.
lxc exec "$vm" -- env LUNA_VM_EVIDENCE_TAG="$tag" python3 - <<'PY'
import os
from pathlib import Path
p=Path("/opt/luna-p0-kernel-tests")
src=p/"fallback-ready.json"
dst=p/("fallback-ready.after-delete."+os.environ["LUNA_VM_EVIDENCE_TAG"]+".json")
assert src.is_file() and not src.is_symlink() and not dst.exists()
with dst.open("xb") as f:
    f.write(src.read_bytes());f.flush();os.fsync(f.fileno())
src.unlink()
dir_fd=os.open(p,os.O_RDONLY|os.O_DIRECTORY);os.fsync(dir_fd);os.close(dir_fd)
assert not (p/"stage-journal.json").exists()
print("VM_INCOMPLETE_TRANSACTION_EVIDENCE_ARCHIVED_NO_FALSE_ROLLBACK_PASS")
PY
echo VM_RACE_AND_AFTER_DELETE_CRASH_CONSISTENCY_DETECTED_PASS
