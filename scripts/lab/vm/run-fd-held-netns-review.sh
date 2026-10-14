#!/usr/bin/env bash
# The dedicated no-NIC KVM may contain a network namespace held ONLY by an FD.
# This test never modifies nft or a host namespace: guest-local, read-only audit.
set -euo pipefail
[[ "$(id -u)" != 0 && "${LUNA_LXD_USE_SUDO:-0}" == 1 ]] || exit 2
[[ "${LUNA_VM_FD_NS_ACK:-}" == "TEST_ONLY_GUEST_FD_HELD_NETNS_INVENTORY" ]] || { echo VM_FD_NS_ACK_REQUIRED >&2;exit 2; }
lxc() { sudo -n /snap/bin/lxc "$@"; }
vm=light-luna-p0-nftvm
lxc list --format csv | grep -q "^$vm,RUNNING,,,VIRTUAL-MACHINE,"
lxc config show "$vm" --expanded | grep -qx -- '- light-luna-p0-kernel-vm'
if lxc config show "$vm" --expanded | grep -Eq '^    type: nic$';then exit 3;fi
[[ "$(lxc exec "$vm" -- systemd-detect-virt)" == kvm ]]
[[ "$(lxc exec "$vm" -- ls /sys/class/net)" == lo ]]
[[ "$(lxc exec "$vm" -- uname -r)" != "$(uname -r)" ]]
lxc exec "$vm" -- test -f /etc/light-luna-p0-kernel-vm
# Only the two read-only scanner files are copied. No policy installation.
for file in guest-boot-inspect.py guest-rollback-review.py;do
  lxc file push "scripts/lab/vm/$file" "$vm/opt/luna-p0-kernel-tests/$file" --mode=0700 >/dev/null
done
lxc exec "$vm" -- python3 - <<'PY'
import hashlib,importlib.util,json,os,subprocess
from pathlib import Path
root=Path("/opt/luna-p0-kernel-tests")
def load(name):
    p=root/name
    assert p.is_file() and not p.is_symlink()
    spec=importlib.util.spec_from_file_location("luna_fd_test_"+name.replace(".","_").replace("-","_"),p)
    m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m);return m
def run(*args):
    return subprocess.run(args,check=True,text=True,capture_output=True,timeout=8)
inspector=load("guest-boot-inspect.py")
review=load("guest-rollback-review.py")
inspector.assert_guest()
before=inspector.inspect_kernel()
assert before["foreignNetnsFdHolders"]==0 and before["nonRootNetnsProcesses"]==0
assert before["rootNftEntries"]==0 and before["namedNetnsPresent"] is False
assert sorted(os.listdir("/sys/class/net"))==["lo"]
journal_path=root/"stage-journal.json"
before_digest=hashlib.sha256(journal_path.read_bytes()).hexdigest()
journalmod=inspector.load_stage_module()
journal=journalmod.safe_load(journal_path)
assert journal["state"]=="ROLLED_BACK"
ready=journalmod.safe_load(root/"auto-watchdog-ready.json")
receiptmod=load("guest-rollback-receipts.py")
receipt=receiptmod.inspect(journal["stageId"],journal["bootId"],ready["netnsInode"],ready)
done_file=root/"auto-watchdog-done.json"
assert done_file.is_file()
done=journalmod.safe_load(done_file)
name="luna_fd_probe"
path=Path("/run/netns")/name
assert not path.exists()
fd=None
created=False
try:
    run("ip","netns","add",name)
    created=True
    fd=os.open(path,os.O_RDONLY|os.O_CLOEXEC)
    ref=os.readlink("/proc/self/fd/"+str(fd))
    f=os.fstat(fd)
    base=os.stat("/proc/1/ns/net")
    assert f.st_dev==base.st_dev and f.st_ino!=base.st_ino
    run("ip","netns","del",name)
    created=False
    assert not path.exists()
    inside=inspector.inspect_kernel()
    assert inside["rootNftEntries"]==0 and not inside["namedNetnsPresent"]
    assert inside["nonRootNetnsProcesses"]==0
    assert inside["foreignNetnsFdHolders"]>=1
    report=review.classify(journal,ready,journalmod.boot_id(),inside,receipt,done)
    assert report["status"]=="KERNEL_STATE_PRESENT_OR_UNKNOWN",report
    assert report["kernelWrites"] is False and report["successAcknowledged"] is False
    assert report["autoReconciliationAllowed"] is False
    print("VM_FD_ONLY_FOREIGN_NETNS_DETECTED_AND_REVIEW_DENIED_PASS")
finally:
    if fd is not None: os.close(fd)
    if created:
        run("ip","netns","del",name)
after=inspector.inspect_kernel()
assert after["foreignNetnsFdHolders"]==0 and after["nonRootNetnsProcesses"]==0
assert after["rootNftEntries"]==0 and after["namedNetnsPresent"] is False
assert before_digest==hashlib.sha256(journal_path.read_bytes()).hexdigest()
assert not run("nft","list","tables").stdout.strip()
assert not run("ip","netns","list").stdout.strip()
print("VM_FD_NS_CLEANUP_JOURNAL_HASH_UNCHANGED_NO_NFT_PASS")
PY
