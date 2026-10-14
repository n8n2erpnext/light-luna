#!/usr/bin/env bash
# Creates ONLY a brand-new, networkless LXD container + dedicated profile.
# Does NOT install/initialize LXD, manage host networking, or modify nftables.
set -euo pipefail
# Trusted operators can use preauthorized sudo without joining root-equivalent lxd group.
if [[ "${LUNA_LXD_USE_SUDO:-0}" == "1" ]]; then
  lxc() { sudo -n /snap/bin/lxc "$@"; }
fi
if [[ "$(id -u)" -eq 0 ]]; then echo "Run as authorized nonroot LXD operator" >&2; exit 2; fi
: "${LUNA_LXD_POOL:?Set LUNA_LXD_POOL to a previously initialized storage pool}"
: "${LUNA_LXD_IMAGE:=light-luna-ubuntu2404}"
: "${LUNA_LXD_INSTANCE:=light-luna-p0-lab}"
: "${LUNA_LXD_PROFILE:=light-luna-p0-isolated}"
for val in "$LUNA_LXD_POOL" "$LUNA_LXD_IMAGE" "$LUNA_LXD_INSTANCE" "$LUNA_LXD_PROFILE"; do
  [[ "$val" =~ ^[a-zA-Z0-9][a-zA-Z0-9_-]{1,55}$ ]] || { echo "Invalid identifier" >&2; exit 2; }
done
lxc info >/dev/null || { echo "LXD unavailable, cannot create lab" >&2; exit 2; }
lxc storage show "$LUNA_LXD_POOL" >/dev/null || { echo "Missing storage pool" >&2; exit 2; }
lxc image info "local:$LUNA_LXD_IMAGE" >/dev/null || { echo "Image alias not pre-cached locally: $LUNA_LXD_IMAGE" >&2; exit 2; }
if lxc info "$LUNA_LXD_INSTANCE" >/dev/null 2>&1; then
  echo "Instance already exists; refusing to modify it" >&2; exit 2
fi
if lxc profile show "$LUNA_LXD_PROFILE" >/dev/null 2>&1; then
  echo "Profile already exists; refusing to modify it" >&2; exit 2
fi
lxc profile create "$LUNA_LXD_PROFILE"
lxc profile set "$LUNA_LXD_PROFILE" security.privileged false
lxc profile set "$LUNA_LXD_PROFILE" security.nesting false
lxc profile set "$LUNA_LXD_PROFILE" limits.cpu 2
lxc profile set "$LUNA_LXD_PROFILE" limits.memory 1536MiB
lxc profile set "$LUNA_LXD_PROFILE" limits.processes 256
lxc profile device add "$LUNA_LXD_PROFILE" root disk "path=/" "pool=$LUNA_LXD_POOL"
# Initialize WITHOUT starting so device configuration can be audited first.
lxc init "local:$LUNA_LXD_IMAGE" "$LUNA_LXD_INSTANCE" --profile "$LUNA_LXD_PROFILE"
if lxc config show "$LUNA_LXD_INSTANCE" --expanded | grep -Eq '^[[:space:]]*type:[[:space:]]*nic[[:space:]]*$'; then
  echo "SAFETY STOP: network interface unexpectedly present; container NOT started" >&2
  exit 3
fi
echo "PROFILE_NO_NIC_VERIFIED"
lxc start "$LUNA_LXD_INSTANCE"
for _ in $(seq 1 80); do
  if lxc exec "$LUNA_LXD_INSTANCE" -- /bin/true >/dev/null 2>&1; then break; fi
  sleep 1
done
lxc exec "$LUNA_LXD_INSTANCE" -- /bin/true || { echo "Instance not ready" >&2; exit 3; }
lxc exec "$LUNA_LXD_INSTANCE" -- /bin/bash -c '
  set -eu
  test "$(ls -1 /sys/class/net | wc -l)" -eq 1
  test "$(ls -1 /sys/class/net)" = lo
  test ! -e /etc/light-luna-p0-lxd-lab
  install -m 0644 /dev/null /etc/light-luna-p0-lxd-lab
'
echo "LXD_NETLESS_INSTANCE_CREATED=$LUNA_LXD_INSTANCE"
echo "Instance remains available for further isolated tests; delete only with explicit cleanup script."
