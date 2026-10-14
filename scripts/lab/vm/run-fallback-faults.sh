#!/usr/bin/env bash
# Explicitly approved P0 fallback prototype in the exact NIC-less KVM lab VM.
set -euo pipefail
[[ "$(id -u)" != 0 && "${LUNA_LXD_USE_SUDO:-}" == 1 ]] || exit 2
[[ "${LUNA_P0_FALLBACK_TEST_ACK:-}" == "TEST_GUEST_ONLY_SYNTHETIC_FALLBACK" ]] || { echo ACK_REQUIRED >&2; exit 2; }
lxc(){ sudo -n /snap/bin/lxc "$@"; }
vm=light-luna-p0-nftvm
lxc list --format csv | grep -q "^$vm,RUNNING,,,VIRTUAL-MACHINE,"
lxc config show "$vm" --expanded | grep -qx -- '- light-luna-p0-kernel-vm'
if lxc config show "$vm" --expanded | grep -Eq '^    type: nic$';then exit 3;fi
[[ "$(lxc exec "$vm" -- systemd-detect-virt)" == kvm ]]
[[ "$(lxc exec "$vm" -- ls /sys/class/net)" == lo ]]
[[ "$(lxc exec "$vm" -- uname -r)" != "$(uname -r)" ]]
lxc exec "$vm" -- test -e /etc/light-luna-p0-kernel-vm
lxc exec "$vm" -- python3 -c '
import json
from pathlib import Path
p=Path("/opt/luna-p0-kernel-tests")
m=json.loads((p/"stage-intent.json").read_text())
j=json.loads((p/"stage-journal.json").read_text())
assert m["ttlSeconds"]==10 and m["kernelWrites"] is False and m["ownerApproved"] is False
assert j["state"]=="ROLLED_BACK" and j["bootId"]==Path("/proc/sys/kernel/random/boot_id").read_text().strip()
'
for file in guest-fallback-supervisor.py guest-stage-test.sh guest-auto-watchdog.py guest-rollback-fence.py;do
 lxc file push "scripts/lab/vm/$file" "$vm/opt/luna-p0-kernel-tests/$file" --mode=0700 >/dev/null
done
for mode in keeper_postcommit watchdog_postcommit;do
 echo "BEGIN_VM_FALLBACK:$mode"
 lxc exec "$vm" -- env LUNA_VM_AUTO_TTL=1 LUNA_VM_FAULT="$mode" \
   LUNA_VM_FAULT_ACK=KILL_VM_SYNTHETIC_KEEPER_OR_WATCHDOG \
   LUNA_VM_FALLBACK=1 LUNA_VM_FALLBACK_ACK=ENABLE_GUEST_SYNTHETIC_FALLBACK \
   bash /opt/luna-p0-kernel-tests/guest-stage-test.sh
 lxc exec "$vm" -- env LUNA_VM_EVIDENCE_TAG="${LUNA_VM_EVIDENCE_TAG:-initial}" python3 - "$mode" <<'PY'
import json,sys,subprocess,os,re
from pathlib import Path
mode=sys.argv[1];root=Path("/opt/luna-p0-kernel-tests")
tag=os.environ.get("LUNA_VM_EVIDENCE_TAG","initial")
assert re.fullmatch(r"[a-z0-9-]{1,36}",tag)
state=json.loads((root/"stage-journal.json").read_text())
done=json.loads((root/"fallback-done.json").read_text())
assert state["state"]=="ROLLED_BACK" and done["status"]=="SYNTHETIC_FALLBACK_ROLLBACK_CONFIRMED"
assert not (root/"auto-watchdog-done.json").exists()
assert subprocess.check_output(["ip","netns","list"],text=True).strip()==""
assert subprocess.check_output(["nft","list","tables"],text=True).strip()==""
for base in ("stage-journal.json","fallback-ready.json","fallback-done.json"):
    source=root/base
    archived=root/(base.removesuffix(".json")+"."+mode+"."+tag+".json")
    assert not archived.exists(), "prior lab evidence exists"
    data=source.read_bytes()
    fd=archived.open("xb");fd.write(data);fd.flush()
    import os
    os.fsync(fd.fileno());fd.close()
    assert archived.read_bytes()==data
(root/"fallback-ready.json").unlink()
(root/"fallback-done.json").unlink()
print("VM_FALLBACK_PER_MODE_EVIDENCE_ARCHIVED_PASS",mode)
PY
done
echo "VM_TWO_GUEST_FALLBACK_FAILOVER_TESTS_PASS"
