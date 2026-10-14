#!/usr/bin/env bash
# Explicit destructive cleanup of the exact owner-approved P0 VM only.
set -euo pipefail
if [[ "${LUNA_LXD_USE_SUDO:-0}" == 1 ]];then lxc(){ sudo -n /snap/bin/lxc "$@"; };fi
[[ "${LUNA_LXD_VM_DELETE_ACK:-}" == DELETE_LIGHT_LUNA_NFT_VM ]] || {
 echo "Set LUNA_LXD_VM_DELETE_ACK=DELETE_LIGHT_LUNA_NFT_VM before deleting the lab VM" >&2;exit 2;
}
vm=light-luna-p0-nftvm
lxc list --format csv | grep -Eq "^$vm,[A-Z]+,,,VIRTUAL-MACHINE,"
lxc config show "$vm" --expanded | grep -qx -- "- light-luna-p0-kernel-vm"
lxc exec "$vm" -- test -f /etc/light-luna-p0-kernel-vm
lxc stop "$vm"
lxc delete "$vm"
echo "Exact marked VM removed; image/storage/profile retained."
