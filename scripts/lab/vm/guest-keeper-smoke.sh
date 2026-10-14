#!/usr/bin/env bash
# VM-only check: an independent systemd service retains a destination netns
# even when the original lxc-exec shell-owned anchor is terminated.
set -euo pipefail
[[ "$(id -u)" == 0 ]] && [[ -f /etc/light-luna-p0-kernel-vm ]] || exit 2
[[ "$(systemd-detect-virt)" == kvm ]] && [[ "$(ls -1 /sys/class/net)" == lo ]] || exit 2
ns="luna_keeper_probe_$$"
unit="luna-p0-keeper-probe-$$"
pid=""
cleanup(){
  systemctl stop "$unit.service" >/dev/null 2>&1 || true
  [[ -z "$pid" ]] || kill "$pid" 2>/dev/null || true
  ip netns delete "$ns" 2>/dev/null || true
}
trap cleanup EXIT
ip netns add "$ns"
ip netns exec "$ns" /usr/bin/sleep 90 &
pid=$!
expected="$(ip netns exec "$ns" readlink /proc/self/ns/net)"
for _ in $(seq 1 50);do
  [[ "$(readlink "/proc/$pid/ns/net" 2>/dev/null || true)" == "$expected" ]] && break
  sleep .05
done
[[ "$(readlink "/proc/$pid/ns/net")" == "$expected" ]]
systemd-run --unit="$unit" --property=TimeoutStopSec=2 /usr/bin/nsenter -t "$pid" -n /usr/bin/sleep 120
keeper=""
for _ in $(seq 1 80);do
  keeper="$(systemctl show "$unit.service" -p MainPID --value)"
  [[ "$keeper" =~ ^[1-9][0-9]*$ ]] && [[ "$(readlink "/proc/$keeper/ns/net" 2>/dev/null || true)" == "$expected" ]] && break
  sleep .05
done
[[ "$keeper" =~ ^[1-9][0-9]*$ ]] && [[ "$(readlink "/proc/$keeper/ns/net")" == "$expected" ]] || exit 3
kill "$pid"
wait "$pid" 2>/dev/null || true
pid=""
[[ "$(readlink "/proc/$keeper/ns/net")" == "$expected" ]] || exit 3
echo VM_SYSTEMD_NAMESPACE_KEEPER_SURVIVES_CONTROLLER_ANCHOR_PASS
