#!/usr/bin/env bash
# Builds locally, pushes only fixtures/binaries into the marked offline LXD lab.
set -euo pipefail
# Trusted operators can use preauthorized sudo without joining root-equivalent lxd group.
if [[ "${LUNA_LXD_USE_SUDO:-0}" == "1" ]]; then
  lxc() { sudo -n /snap/bin/lxc "$@"; }
fi
if [[ "$(id -u)" -eq 0 ]]; then echo "No root on HOMESV" >&2; exit 2; fi
: "${LUNA_LXD_INSTANCE:=light-luna-p0-lab}"
: "${LUNA_LXD_PROFILE:=light-luna-p0-isolated}"
[[ "$LUNA_LXD_INSTANCE" =~ ^[a-zA-Z0-9][a-zA-Z0-9_-]{1,55}$ ]] || exit 2
test -f scripts/lab/lxd/guest-test.sh
test -f testdata/host-baseline.json
lxc info "$LUNA_LXD_INSTANCE" >/dev/null
profileInfo="$(lxc config show "$LUNA_LXD_INSTANCE" --expanded)"
if printf '%s\n' "$profileInfo" | grep -Eq '^[[:space:]]*type:[[:space:]]*nic[[:space:]]*$'; then
  echo "DENY: container has network device" >&2
  exit 3
fi
lxc exec "$LUNA_LXD_INSTANCE" -- test -f /etc/light-luna-p0-lxd-lab
# LXD is a snap: its private mount namespace cannot read the host /tmp.
# Keep throwaway binaries under this accessible working tree; trap removes them.
tmp="$(mktemp -d "$PWD/lxd-artifacts.XXXXXXXX")"
trap 'rm -rf -- "$tmp"' EXIT
export GOTOOLCHAIN=local GOPROXY=off CGO_ENABLED=0
GOOS=linux GOARCH=amd64 go build -trimpath -o "$tmp/luna-gate" ./cmd/luna-gate
GOOS=linux GOARCH=amd64 go build -trimpath -o "$tmp/luna-agent" ./cmd/luna-agent
lxc exec "$LUNA_LXD_INSTANCE" -- install -d -m 0755 /opt/light-luna-p0/bin
lxc file push "$tmp/luna-gate" "$LUNA_LXD_INSTANCE/opt/light-luna-p0/bin/luna-gate"
lxc file push "$tmp/luna-agent" "$LUNA_LXD_INSTANCE/opt/light-luna-p0/bin/luna-agent"
lxc file push testdata/host-baseline.json "$LUNA_LXD_INSTANCE/opt/light-luna-p0/host-baseline.json"
lxc file push scripts/lab/lxd/guest-test.sh "$LUNA_LXD_INSTANCE/root/luna-guest-test.sh"
lxc exec "$LUNA_LXD_INSTANCE" -- chmod 0755 /opt/light-luna-p0/bin/luna-agent /opt/light-luna-p0/bin/luna-gate
lxc exec "$LUNA_LXD_INSTANCE" -- chmod 0644 /opt/light-luna-p0/host-baseline.json
lxc exec "$LUNA_LXD_INSTANCE" -- chmod 0700 /root/luna-guest-test.sh
# Test updates only throwaway credentials/accounts INSIDE this exact container.
lxc exec "$LUNA_LXD_INSTANCE" -- /bin/bash /root/luna-guest-test.sh
echo "LXD_CROSS_UID_TEST_RUN_COMPLETED"
