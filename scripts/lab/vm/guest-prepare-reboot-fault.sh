#!/usr/bin/env bash
# Destructive-to-synthetic-state ONLY in exact no-NIC KVM guest.
# Establishes a COMMITTED_SYNTHETIC journal and a fixed netns nft test
# rule WITHOUT autonomous watchdog. Host harness MUST reboot this guest.
set -euo pipefail
[[ "$(id -u)" == 0 && -f /etc/light-luna-p0-kernel-vm ]] || exit 2
[[ "$(systemd-detect-virt)" == kvm ]] || exit 2
[[ "$(ls -1 /sys/class/net)" == lo ]] || exit 2
[[ -z "$(ip netns list)" && -z "$(nft list tables)" ]] || { echo VM_NOT_EMPTY_REFUSE;exit 3; }
python3 - <<'PY'
from pathlib import Path
import os
base=os.readlink("/proc/1/ns/net")
for p in Path("/proc").iterdir():
    if not p.name.isdecimal():continue
    try: link=os.readlink(str(p/"ns/net"))
    except FileNotFoundError:continue
    assert link==base, "NON_DEFAULT_PROCESS_NETNS_REFUSE"
PY
work=/opt/luna-p0-kernel-tests
j="$work/guest-stage-journal.py"
manifest="$work/stage-intent.json"
journal="$work/stage-journal.json"
archive="$work/stage-journal.pre-reboot-baseline.json"
pre="$work/boot-fault-pre.json"
for f in "$j" "$manifest" "$journal";do [[ -s "$f" && ! -L "$f" ]] || exit 3;done
[[ ! -e "$archive" && ! -e "$pre" ]] || { echo OLD_REBOOT_TEST_EVIDENCE_REFUSE;exit 3; }
python3 - "$journal" "$manifest" <<'PY'
import importlib.util,json,sys
p="/opt/luna-p0-kernel-tests/guest-stage-journal.py"
spec=importlib.util.spec_from_file_location("vm_stage",p)
m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m)
state=m.safe_load(m.JOURNAL);plan=m.manifest()
assert set(state)=={"stageId","bootId","state","history"}
assert state["bootId"]==m.boot_id() and state["stageId"]==plan["stageId"]
assert state["state"]=="ROLLED_BACK" and state["history"]==m.PHASES
assert plan["kernelWrites"] is False and plan["ownerApproved"] is False
PY
# Keep the original completed transaction, not a silent journal reset.
mv -- "$journal" "$archive"
[[ -s "$archive" && ! -e "$journal" ]] || exit 3
src=luna_p0_rb_src
dst=luna_p0_rb_dst
[[ -z "$(ip netns list)" ]] || exit 3
python3 "$j" init STAGED
ip netns add "$src"
ip netns add "$dst"
ip link add lunasrc type veth peer name lunadst
ip link set lunasrc netns "$src"
ip link set lunadst netns "$dst"
ip -n "$src" link set lo up
ip -n "$dst" link set lo up
ip -n "$src" link set lunasrc up
ip -n "$dst" link set lunadst up
ip -n "$src" addr add 198.19.0.2/30 dev lunasrc
ip -n "$dst" addr add 198.19.0.1/30 dev lunadst
ip netns exec "$src" ping -n -c 1 -W 2 198.19.0.1 >/dev/null
python3 "$j" transition VALIDATED
cat > "$work/boot-fault-synthetic.nft" <<'NFT'
table inet luna_p0_stage {
 chain input {
   type filter hook input priority -50; policy accept;
   ip saddr 198.19.0.2 icmp type echo-request counter drop
 }
}
NFT
ip netns exec "$dst" nft -c -f "$work/boot-fault-synthetic.nft"
ip netns exec "$dst" nft -f "$work/boot-fault-synthetic.nft"
python3 "$j" transition COMMITTED_SYNTHETIC
if ip netns exec "$src" ping -n -c 1 -W 1 198.19.0.1 >/dev/null 2>&1;then
 echo SYNTHETIC_DROP_NOT_OBSERVED;exit 3
fi
ip netns exec "$dst" nft list chain inet luna_p0_stage input | grep 'counter packets [1-9]' >/dev/null || { echo NFT_DROP_COUNTER_NOT_POSITIVE;exit 3; }
python3 - "$work" <<'PY'
import json,os,sys,tempfile
from pathlib import Path
root=Path(sys.argv[1])
j=json.loads((root/"stage-journal.json").read_text())
assert j["state"]=="COMMITTED_SYNTHETIC"
payload={"schema":"lightluna.vm-boot-fault-evidence/v1alpha1","preBootId":j["bootId"],"stageId":j["stageId"],
         "journalState":"COMMITTED_SYNTHETIC","network":"GUEST_NO_NIC_PRIVATE_NETNS",
         "table":"inet luna_p0_stage","sourceNamespace":"luna_p0_rb_src",
         "targetNamespace":"luna_p0_rb_dst","icmpDropCounterPositive":True,
         "watchdogArmed":False,"kernelWrites":"GUEST_PRIVATE_NETNS_SYNTHETIC_ONLY"}
fd,tmp=tempfile.mkstemp(dir=root,prefix=".boot-evidence-")
with os.fdopen(fd,"w") as f:
    json.dump(payload,f,sort_keys=True);f.write("\n");f.flush();os.fsync(f.fileno())
os.replace(tmp,root/"boot-fault-pre.json")
fd=os.open(root,os.O_RDONLY|os.O_DIRECTORY);os.fsync(fd);os.close(fd)
print("VM_PRE_REBOOT_SYNTHETIC_COMMITTED_AND_DROP_VERIFIED_PASS")
PY
