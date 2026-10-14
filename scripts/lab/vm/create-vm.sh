#!/usr/bin/env bash
# Explicitly launched fresh kernel-isolated, NO-NIC LXD VM for source-only P0 tests.
# Does not install LXD, create NAT/bridges, change firewall or reconfigure NetBird.
set -euo pipefail
if [[ "${LUNA_LXD_USE_SUDO:-0}" == 1 ]]; then lxc(){ sudo -n /snap/bin/lxc "$@"; }; fi
[[ "$(id -u)" -ne 0 ]] || { echo "Run as trusted nonroot LXD operator"; exit 2; }
instance="light-luna-p0-nftvm"
profile="light-luna-p0-kernel-vm"
image="light-luna-ubuntu2404-vm"
pool="${LUNA_LXD_POOL:-default}"
[[ "$pool" =~ ^[a-zA-Z0-9][a-zA-Z0-9_-]{1,55}$ ]] || exit 2
lxc info >/dev/null
lxc storage show "$pool" >/dev/null
lxc image info "local:$image" | grep -q '^Type: virtual-machine'
if lxc info "$instance" >/dev/null 2>&1;then echo "Instance already exists; refusing overwrite" >&2;exit 3;fi
if lxc profile show "$profile" >/dev/null 2>&1;then echo "Profile exists; refusing overwrite" >&2;exit 3;fi
lxc profile create "$profile"
lxc profile set "$profile" limits.cpu 2
lxc profile set "$profile" limits.memory 2048MiB
lxc profile set "$profile" security.secureboot false
lxc profile device add "$profile" root disk path=/ "pool=$pool"
lxc init "local:$image" "$instance" --vm --profile "$profile"
if lxc config show "$instance" --expanded | grep -Eq '^    type: nic$';then
 echo "Unexpected NIC; VM not started" >&2;exit 3
fi
echo "VM_NO_NIC_CONFIG_VERIFIED"
lxc start "$instance"
ok=0
for _ in $(seq 1 100);do
 if lxc exec "$instance" -- /bin/true >/dev/null 2>&1;then ok=1;break;fi
 sleep 1
done
[[ "$ok" == 1 ]] || { echo "VM agent not reachable" >&2;exit 3; }
[[ "$(lxc exec "$instance" -- ls -1 /sys/class/net)" == lo ]] || { echo "VM NIC unexpectedly present" >&2; exit 3; }
[[ "$(lxc exec "$instance" -- systemd-detect-virt)" == kvm ]] || { echo "VM isolation not KVM" >&2; exit 3; }
[[ "$(lxc exec "$instance" -- uname -r)" != "$(uname -r)" ]] || { echo "Cannot prove separate kernel" >&2; exit 3; }
echo "VM_SEPARATE_KERNEL_NO_NIC_READY"
