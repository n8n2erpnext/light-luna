#!/usr/bin/env bash
# Read-only host readiness check; no snap install, bridge creation or LXD init.
set -euo pipefail
# Trusted operators can use preauthorized sudo without joining root-equivalent lxd group.
if [[ "${LUNA_LXD_USE_SUDO:-0}" == "1" ]]; then
  lxc() { sudo -n /snap/bin/lxc "$@"; }
fi
echo "LIGHT_LUNA_LXD_HOST_PREFLIGHT"
printf 'Host: '; hostname
printf 'Kernel: '; uname -r
free -h
df -h /
if ! command -v lxc >/dev/null 2>&1; then
  echo "BLOCKED: LXD client is missing" >&2; exit 2
fi
if ! snap list lxd >/dev/null 2>&1; then
  echo "BLOCKED: LXD snap is not installed. Do not use this script to install it." >&2; exit 2
fi
if ! lxc info >/dev/null 2>&1; then
  echo "BLOCKED: LXD daemon is not initialized or current operator is unauthorized." >&2; exit 2
fi
echo "LXD_SERVER_READY"
lxc storage list
lxc network list
lxc profile list
lxc list
echo "This does not establish no host firewall changes; record a separate baseline."
