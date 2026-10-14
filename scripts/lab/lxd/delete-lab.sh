#!/usr/bin/env bash
# Explicit destructive cleanup, only for lab instance/profile, never default.
set -euo pipefail
# Trusted operators can use preauthorized sudo without joining root-equivalent lxd group.
if [[ "${LUNA_LXD_USE_SUDO:-0}" == "1" ]]; then
  lxc() { sudo -n /snap/bin/lxc "$@"; }
fi
: "${LUNA_LXD_INSTANCE:=light-luna-p0-lab}"
: "${LUNA_LXD_PROFILE:=light-luna-p0-isolated}"
[[ "$LUNA_LXD_INSTANCE" =~ ^[a-zA-Z0-9][a-zA-Z0-9_-]{1,55}$ ]] || exit 2
if [[ "${LUNA_LXD_DELETE_ACK:-}" != "DELETE_LIGHT_LUNA_LAB" ]];then
 echo "Set LUNA_LXD_DELETE_ACK=DELETE_LIGHT_LUNA_LAB to delete this dedicated lab" >&2
 exit 2
fi
# Marker must exist. Never delete an arbitrary matching-name instance.
lxc exec "$LUNA_LXD_INSTANCE" -- test -f /etc/light-luna-p0-lxd-lab || {
 echo "Lab marker missing; refusing destructive action" >&2; exit 3;
}
lxc stop "$LUNA_LXD_INSTANCE"
lxc delete "$LUNA_LXD_INSTANCE"
# Do not remove user-owned unrelated profiles or storage pools/images.
echo "Deleted marked lab instance $LUNA_LXD_INSTANCE; profile and storage retained."
