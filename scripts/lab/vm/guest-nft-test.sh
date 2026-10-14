#!/usr/bin/env bash
# Light Luna P0 independent-kernel nftables laboratory.
# This script MUST only execute as root INSIDE the marked, NIC-less VM.
# It never runs nftables on the HOMESV host. It tests unhooked Light Luna
# source preview separately from an explicitly synthetic VM-only hook.
set -euo pipefail
[[ "$(id -u)" == "0" ]] || { echo "GUEST_ROOT_REQUIRED" >&2; exit 2; }
[[ -f /etc/light-luna-p0-kernel-vm ]] || { echo "VM_MARKER_MISSING" >&2; exit 2; }
[[ "$(ls -1 /sys/class/net)" == "lo" ]] || { echo "VM_NIC_PRESENT_ABORT" >&2; exit 2; }
[[ "$(systemd-detect-virt)" == "kvm" ]] || { echo "NON_KVM_ABORT" >&2; exit 2; }
for x in ip nft ping setsid python3; do command -v "$x" >/dev/null || { echo "MISSING_TOOL:$x"; exit 2; }; done
src="luna_src_$$"
dst="luna_dst_$$"
work=/opt/luna-p0-kernel-tests
preview="$work/nft-preview.nft"
[[ -s "$preview" ]] || { echo "NO_LIGHT_LUNA_SOURCE_PREVIEW"; exit 2; }
# Idempotent lab reruns may remove only these fixed synthetic result files.
rm -f -- "$work/watchdog-done" "$work/watchdog.log" "$work/invalid.log" "$work/bad.nft" "$work/synthetic-drop.nft"
watchdog_pid=""
controller_pid=""
cleanup(){
  [[ -z "$controller_pid" ]] || kill "$controller_pid" 2>/dev/null || true
  [[ -z "$watchdog_pid" ]] || kill "$watchdog_pid" 2>/dev/null || true
  ip netns exec "$dst" nft delete table inet luna_p0_vmtest 2>/dev/null || true
  ip netns delete "$src" 2>/dev/null || true
  ip netns delete "$dst" 2>/dev/null || true
}
trap cleanup EXIT INT TERM
echo "VM_KERNEL=$(uname -r)"
ip netns add "$src"
ip netns add "$dst"
ip link add luna_src_link type veth peer name luna_dst_link
ip link set luna_src_link netns "$src"
ip link set luna_dst_link netns "$dst"
ip -n "$src" link set lo up
ip -n "$dst" link set lo up
ip -n "$src" link set luna_src_link up
ip -n "$dst" link set luna_dst_link up
ip -n "$src" addr add 198.18.0.2/30 dev luna_src_link
ip -n "$dst" addr add 198.18.0.1/30 dev luna_dst_link
ip netns exec "$src" ping -n -c 1 -W 2 198.18.0.1 >/dev/null
echo "VM_PACKET_BASELINE_ACCEPT_PASS"
# This checks exactly the unhooked text emitted by the Light Luna Go compiler.
# No generated policy is ever installed in a packet hook.
ip netns exec "$dst" nft -c -f "$preview"
if ip netns exec "$dst" nft list tables | grep -q 'luna_p0_preview';then
  echo "NF_CHECK_CAUSED_EFFECT_FAIL" >&2;exit 3
fi
echo "LIGHT_LUNA_UNHOOKED_NFT_KERNEL_SYNTAX_PASS"
printf '%s\n' 'table inet } invalid nft' > "$work/bad.nft"
if ip netns exec "$dst" nft -c -f "$work/bad.nft" > "$work/invalid.log" 2>&1;then
  echo "INVALID_NFT_SYNTAX_ACCEPTED_FAIL" >&2;exit 3
fi
echo "INVALID_NFT_NEGATIVE_PASS"
# This synthetic hook is not emitted by Light Luna and runs only inside
# the VM's fresh, private destination network namespace.
cat > "$work/synthetic-drop.nft" <<'NFT'
table inet luna_p0_vmtest {
  chain input {
    type filter hook input priority 0; policy accept;
    ip saddr 198.18.0.2 icmp type echo-request counter drop
  }
}
NFT
ip netns exec "$dst" nft -c -f "$work/synthetic-drop.nft"
ip netns exec "$dst" nft -f "$work/synthetic-drop.nft"
if ip netns exec "$src" ping -n -c 1 -W 1 198.18.0.1 >/dev/null 2>&1;then
  echo "SYNTHETIC_PACKET_DROP_FAILED" >&2;exit 3
fi
ip netns exec "$dst" nft list chain inet luna_p0_vmtest input | grep -Eq 'counter packets [1-9][0-9]*'
echo "VM_PACKET_DROP_COUNTER_PASS"
# Start rollback in a separate session/process. Kill a stand-in controller;
# watchdog must independently delete ONLY this synthetic VM-namespace table.
setsid bash -c '
  sleep 3
  ip netns exec "$1" nft delete table inet luna_p0_vmtest
  : > "$2"
' bash "$dst" "$work/watchdog-done" > "$work/watchdog.log" 2>&1 < /dev/null &
watchdog_pid=$!
(sleep 30) &
controller_pid=$!
kill "$controller_pid" 2>/dev/null || true
wait "$controller_pid" 2>/dev/null || true
controller_pid=""
echo "SYNTHETIC_CONTROLLER_TERMINATED"
for _ in $(seq 1 55);do
  [[ -f "$work/watchdog-done" ]] && break
  sleep .15
done
[[ -f "$work/watchdog-done" ]] || { echo "WATCHDOG_ROLLBACK_TIMED_OUT" >&2;cat "$work/watchdog.log" >&2;exit 3; }
watchdog_pid=""
ip netns exec "$src" ping -n -c 1 -W 2 198.18.0.1 >/dev/null
if ip netns exec "$dst" nft list tables | grep -q luna_p0_vmtest;then
  echo "ROLLBACK_TABLE_REMAINS" >&2;exit 3
fi
echo "VM_WATCHDOG_ROLLBACK_PACKET_RESTORED_PASS"
echo "VM_ISOLATED_NFT_SYNTAX_FLOW_ROLLBACK_PASS"
