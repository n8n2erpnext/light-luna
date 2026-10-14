#!/usr/bin/env bash
# Offline, no-network LXD VM kernel/netns nftables verification runner.
# Light Luna policy preview goes through nft --check ONLY.
# The independent synthetic hook exists solely in VM netns.
set -euo pipefail
[[ "$(id -u)" != 0 ]] || { echo "Run as nonroot authorized LXD operator" >&2;exit 2; }
if [[ "${LUNA_LXD_USE_SUDO:-0}" == "1" ]];then lxc(){ sudo -n /snap/bin/lxc "$@"; };fi
instance="${LUNA_LXD_VM:-light-luna-p0-nftvm}"
profile="light-luna-p0-kernel-vm"
[[ "$instance" == "light-luna-p0-nftvm" ]] || { echo "Refusing unexpected VM" >&2;exit 2; }
lxc list --format csv | grep -q "^$instance,RUNNING,,,VIRTUAL-MACHINE,"
lxc config show "$instance" --expanded | grep -qx -- "- $profile"
if lxc config show "$instance" --expanded | grep -Eq '^    type: nic$';then
 echo "NIC is present - fail closed" >&2;exit 3
fi
vm_kernel="$(lxc exec "$instance" -- uname -r)"
[[ -n "$vm_kernel" && "$vm_kernel" != "$(uname -r)" ]] || {
 echo "Separate guest kernel not verified" >&2;exit 3;
}
[[ "$(lxc exec "$instance" -- ls -1 /sys/class/net)" == "lo" ]] || {
 echo "Guest NIC detected" >&2;exit 3;
}
[[ "$(lxc exec "$instance" -- systemd-detect-virt)" == "kvm" ]] || {
 echo "Expected KVM guest" >&2;exit 3;
}
echo "SEPARATE_KERNEL_NO_NIC_KVM_PASS: host=$(uname -r) guest=$vm_kernel"
# Marker only after verified profile, type and NIC/kernel status.
lxc exec "$instance" -- sh -c 'test -e /etc/light-luna-p0-kernel-vm || install -m 0644 /dev/null /etc/light-luna-p0-kernel-vm'
tmp="$(mktemp -d "$PWD/lxd-vm-artifacts.XXXXXXXX")"
trap 'rm -rf -- "$tmp"' EXIT
export GOTOOLCHAIN=local GOPROXY=off CGO_ENABLED=0
go build -trimpath -o "$tmp/lunactl" ./cmd/lunactl
"$tmp/lunactl" firewall policies nft-preview testdata/host-baseline.json > "$tmp/nft-plan.json"
python3 - "$tmp/nft-plan.json" "$tmp/nft-preview.nft" <<'PY'
import json,sys
with open(sys.argv[1],encoding="utf-8") as f: p=json.load(f)
assert p["status"]=="UNHOOKED_TEXT_PREVIEW_ONLY",p
assert p["kernelWrites"] is False
assert p["syntaxChecked"] is False
assert p["kernelCompatibilityChecked"] is False
script=p["unhookedScript"]
assert "table inet luna_p0_preview" in script
assert " hook " not in "\n".join(s for s in script.splitlines() if not s.startswith("#"))
with open(sys.argv[2],"w",encoding="utf-8") as f:f.write(script)
print("P0_NFT_UNHOOKED_COMPILER_NO_EXECUTOR_PASS")
PY
lxc exec "$instance" -- install -d -m 0700 /opt/luna-p0-kernel-tests
lxc file push "$tmp/nft-preview.nft" "$instance/opt/luna-p0-kernel-tests/nft-preview.nft" --mode=0600 >/dev/null 2>&1
lxc file push scripts/lab/vm/guest-rollback.sh "$instance/opt/luna-p0-kernel-tests/guest-rollback.sh" --mode=0700 >/dev/null 2>&1
lxc file push scripts/lab/vm/guest-nft-test.sh "$instance/opt/luna-p0-kernel-tests/guest-nft-test.sh" --mode=0700 >/dev/null 2>&1
lxc exec "$instance" -- bash /opt/luna-p0-kernel-tests/guest-nft-test.sh
# New typed VM-only staging intent. This is a manifest, NOT a system call.
"$tmp/lunactl" vm stage-intent testdata/host-baseline.json 8 > "$tmp/stage-intent.json"
for name in guest-stage-test.sh guest-stage-journal.py guest-netprobe.py guest-auto-watchdog.py;do
 lxc file push "scripts/lab/vm/$name" "$instance/opt/luna-p0-kernel-tests/$name" --mode=0700 >/dev/null 2>&1
done
lxc file push "$tmp/stage-intent.json" "$instance/opt/luna-p0-kernel-tests/stage-intent.json" --mode=0600 >/dev/null 2>&1
lxc exec "$instance" -- bash /opt/luna-p0-kernel-tests/guest-stage-test.sh
# This is a NEW stage transaction: watchdog MUST arm on its own before
# synthetic nft commit, and MUST roll back without another controller call.
"$tmp/lunactl" vm stage-intent testdata/host-baseline.json 10 > "$tmp/stage-intent-auto.json"
lxc file push "$tmp/stage-intent-auto.json" "$instance/opt/luna-p0-kernel-tests/stage-intent.json" --mode=0600 >/dev/null 2>&1
lxc exec "$instance" -- env LUNA_VM_AUTO_TTL=1 bash /opt/luna-p0-kernel-tests/guest-stage-test.sh
echo "VM_KERNEL_STAGE_AND_AUTONOMOUS_TTL_TESTS_EXECUTED"
