#!/usr/bin/env bash
# Independent transient systemd timer payload, VM private namespace ONLY.
set -euo pipefail
[[ "$(id -u)" == 0 ]] || { echo REVERT_NOT_ROOT >&2;exit 2; }
[[ -f /etc/light-luna-p0-kernel-vm ]] || { echo REVERT_MISSING_VM_MARKER >&2;exit 2; }
[[ "$(systemd-detect-virt)" == kvm ]] || { echo REVERT_NOT_KVM >&2;exit 2; }
[[ "$(ls -1 /sys/class/net)" == lo ]] || { echo REVERT_NON_LOOPBACK_DEVICE >&2;exit 2; }
[[ "$#" == 3 || "$#" == 5 ]] || { echo REVERT_INVALID_ARGC >&2;exit 2; }
namespace="$1"
table="$2"
marker="$3"
[[ "$namespace" =~ ^luna_(dst|stage_dst)_[0-9]+$ ]] || { echo INVALID_NETNS >&2;exit 2; }
[[ "$table" == luna_p0_vmtest || "$table" == luna_p0_stage ]] || { echo INVALID_TABLE >&2;exit 2; }
[[ "$marker" == /opt/luna-p0-kernel-tests/watchdog-done || "$marker" == /opt/luna-p0-kernel-tests/stage-watchdog-done ]] || { echo REVERT_INVALID_MARKER >&2;exit 2; }
case "$table:$marker" in
 luna_p0_vmtest:/opt/luna-p0-kernel-tests/watchdog-done|luna_p0_stage:/opt/luna-p0-kernel-tests/stage-watchdog-done) ;;
 *) echo REVERT_TABLE_MARKER_MISMATCH >&2;exit 2 ;;
esac
if [[ "$#" == 5 ]];then
  # VM-only separate mount namespace rescue: systemd cannot resolve netns
  # bind mounts created by lxc exec. Use an independently verified PID anchor.
  pid="$4"
  expected_netns_inode="$5"
  [[ "$pid" =~ ^[1-9][0-9]*$ ]] || { echo "INVALID_ANCHOR_PID" >&2;exit 2; }
  [[ "$expected_netns_inode" =~ ^net:\[[0-9]+\]$ ]] || { echo "INVALID_ANCHOR_NETNS" >&2;exit 2; }
  observed="$(readlink "/proc/$pid/ns/net" 2>/dev/null)" || { echo REVERT_MISSING_ANCHOR_PID >&2;exit 2; }
  [[ "$observed" == "$expected_netns_inode" ]] || { echo "ANCHOR_NAMESPACE_MISMATCH" >&2;exit 2; }
  nsenter -t "$pid" -n nft delete table inet "$table"
else
  ip netns exec "$namespace" nft delete table inet "$table"
fi
if [[ "$table" == luna_p0_stage ]];then
  python3 /opt/luna-p0-kernel-tests/guest-stage-journal.py transition ROLLED_BACK
fi
: > "$marker"
echo "VM_TRANSIENT_SYSTEMD_WATCHDOG_ROLLBACK_OK"
