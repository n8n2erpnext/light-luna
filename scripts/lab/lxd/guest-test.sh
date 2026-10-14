#!/usr/bin/env bash
# Must only be run as guest root INSIDE the marked, NIC-less throwaway LXD lab.
# System users/groups/files below are strictly GUEST-local, not HOMESV.
set -euo pipefail
[[ "$(id -u)" -eq 0 ]] || { echo "Guest root required" >&2; exit 2; }
[[ -f /etc/light-luna-p0-lxd-lab ]] || { echo "Missing LXD lab marker; refusing host execution" >&2; exit 2; }
[[ "$(ls -1 /sys/class/net | tr '\n' ' ')" == "lo " ]] || {
  echo "Expected only loopback NIC inside LXD; refusing test" >&2; exit 2;
}
for cmd in groupadd useradd usermod runuser setpriv python3; do
  command -v "$cmd" >/dev/null || { echo "Missing offline image prerequisite $cmd" >&2; exit 2; }
done
GATE_BIN=/opt/light-luna-p0/bin/luna-gate
AGENT_BIN=/opt/light-luna-p0/bin/luna-agent
POLICY=/opt/light-luna-p0/host-baseline.json
test -x "$GATE_BIN" && test -x "$AGENT_BIN" && test -r "$POLICY"
for name in lunagate lunaagent lunaevil lunaoutsider lunabootstrap;do
  if id "$name" >/dev/null 2>&1; then
    echo "Test user already exists; dedicated lab script refuses rerun" >&2;exit 3;
  fi
done
if getent group lunaipc >/dev/null; then echo "Lab group already exists"; exit 3;fi
groupadd --system lunaipc
for name in lunagate lunaagent lunaevil lunaoutsider lunabootstrap;do
  useradd --system --no-create-home --shell /usr/sbin/nologin "$name"
done
usermod -aG lunaipc lunaagent
usermod -aG lunaipc lunaevil
GATE_UID="$(id -u lunagate)"
AGENT_UID="$(id -u lunaagent)"
EVIL_UID="$(id -u lunaevil)"
OUTSIDER_UID="$(id -u lunaoutsider)"
IPC_GID="$(getent group lunaipc | cut -d: -f3)"
for value in "$GATE_UID" "$AGENT_UID" "$EVIL_UID" "$OUTSIDER_UID";do
  [[ "$value" != 0 ]] || { echo "Never use root as a service UID";exit 3; }
done
[[ "$GATE_UID" != "$AGENT_UID" && "$GATE_UID" != "$EVIL_UID" && "$GATE_UID" != "$OUTSIDER_UID" ]] || exit 3
if id -nG lunaoutsider | tr ' ' '\n' | grep -qx lunaipc; then
 echo "Outsider was assigned IPC group unexpectedly";exit 3;
fi
install -d -m 0755 /var/lib/light-luna-p0 /run/light-luna-p0
install -d -m 0700 -o lunabootstrap -g lunabootstrap /var/tmp/light-luna-bootstrap
runuser -u lunabootstrap -- "$GATE_BIN" init-ed25519 \
 /var/tmp/light-luna-bootstrap/agent /var/tmp/light-luna-bootstrap/gate
mv /var/tmp/light-luna-bootstrap/agent /var/lib/light-luna-p0/agent
mv /var/tmp/light-luna-bootstrap/gate /var/lib/light-luna-p0/gate
chown -R lunaagent:lunaagent /var/lib/light-luna-p0/agent
chown -R lunagate:lunagate /var/lib/light-luna-p0/gate
rmdir /var/tmp/light-luna-bootstrap
AGENT_DIR=/var/lib/light-luna-p0/agent
GATE_DIR=/var/lib/light-luna-p0/gate
install -d -m 0710 -o lunagate -g lunaipc /run/light-luna-p0/shared
SOCKET=/run/light-luna-p0/shared/gate.sock
[[ -f "$AGENT_DIR/agent-private.ed25519" ]]
[[ -f "$GATE_DIR/agent-public.ed25519" ]]
[[ ! -e "$GATE_DIR/agent-private.ed25519" ]]
# The Gate UID must never be able to read the Agent's private signing key.
if setpriv --reuid "$GATE_UID" --regid "$(id -g lunagate)" --clear-groups \
 test -r "$AGENT_DIR/agent-private.ed25519"; then
 echo "FAIL: Gate can read Agent private key" >&2; exit 3;
fi
check(){
 local expected="$1" payload="$2"
 printf '%s\n' "$payload" | python3 -c '
import sys,json
r=json.load(sys.stdin)
assert r["decision"] == sys.argv[1],r
assert r["kernelWrites"] is False,r
assert r["executorStatus"] == "NOT_AVAILABLE",r
assert r["replayProtection"] == "FSYNC_JOURNAL_DEMO_ONLY",r
' "$expected"
}
gate_pid=""
cleanup(){
 if [[ -n "$gate_pid" ]];then kill "$gate_pid" 2>/dev/null || true;wait "$gate_pid" 2>/dev/null || true;fi
}
trap cleanup EXIT INT TERM
start_gate(){
 setpriv --reuid "$GATE_UID" --regid "$IPC_GID" --clear-groups \
  "$GATE_BIN" serve-ed25519 "$POLICY" "$GATE_DIR" /run/light-luna-p0/shared \
  "$AGENT_UID" "$IPC_GID" 30 > /run/light-luna-p0/gate.log 2>&1 &
 gate_pid=$!
 for _ in $(seq 1 80);do
  if [[ -S "$SOCKET" ]];then return 0;fi
  if ! kill -0 "$gate_pid" 2>/dev/null;then
   cat /run/light-luna-p0/gate.log >&2; exit 3;
  fi
  sleep 0.05
 done
 cat /run/light-luna-p0/gate.log >&2
 echo "Gate socket never appeared" >&2;exit 3
}
stop_gate(){
 kill "$gate_pid" 2>/dev/null || true
 wait "$gate_pid" 2>/dev/null || true
 gate_pid=""
}
agent(){
 runuser -u lunaagent -- "$AGENT_BIN" "$1" "$POLICY" "$AGENT_DIR" "$SOCKET" "$GATE_UID" "${@:2}"
}
start_gate
REQ=synthetic-crossuid-request-0001
NONCE=synthetic-crossuid-nonce-0001
check ADVISORY_ONLY "$(agent preview-ed25519 "$REQ" "$NONCE")"
check REJECTED "$(agent preview-ed25519 "$REQ" "$NONCE")"
check REJECTED "$(agent apply-ed25519 synthetic-crossuid-apply-0001 synthetic-crossuid-applynonce-0001)"
echo "CROSS_UID_POSITIVE_AGENT_PASS"
# Gate UID pinning must block connection to unexpected identity.
if runuser -u lunaagent -- "$AGENT_BIN" preview-ed25519 "$POLICY" "$AGENT_DIR" "$SOCKET" "$((GATE_UID+10000))" \
 > /run/light-luna-p0/bad-gate-uid.log 2>&1;then
 echo "FAIL: Agent accepted untrusted Gate UID" >&2;exit 3;
fi
echo "CROSS_UID_BAD_GATE_UID_DENIED"
# Outsider with no IPC group must fail at filesystem/socket permission boundary.
runuser -u lunaoutsider -- python3 - "$SOCKET" <<'PY'
import socket,sys,errno
c=socket.socket(socket.AF_UNIX,socket.SOCK_STREAM)
try:
    c.connect(sys.argv[1])
except OSError as e:
    assert e.errno in (errno.EACCES, errno.EPERM),repr(e)
    print("OUTSIDER_UNGROUPED_SOCKET_DENIED")
else:
    raise AssertionError("unauthorized outsider connected")
finally:
    c.close()
PY
# Compromised-key test: inside this disposable LXD container ONLY, deliberately
# copy the test Agent private key to a different UID already in IPC group.
# The Gate MUST still reject this credential at SO_PEERCRED UID boundary.
install -d -m 0700 -o lunaevil -g lunaevil /var/lib/light-luna-p0/evil
install -m 0600 -o lunaevil -g lunaevil \
 "$AGENT_DIR/agent-private.ed25519" /var/lib/light-luna-p0/evil/agent-private.ed25519
if runuser -u lunaevil -- "$AGENT_BIN" preview-ed25519 "$POLICY" \
 /var/lib/light-luna-p0/evil "$SOCKET" "$GATE_UID" \
 synthetic-rogue-request-0001 synthetic-rogue-nonce-0001 \
 > /run/light-luna-p0/rogue.log 2>&1;then
 echo "FAIL: Wrong UID with test-key copy was accepted" >&2; exit 3;
fi
echo "WRONG_UID_EVEN_WITH_TEST_KEY_DENIED"
# Delete the deliberate *guest-only* copied synthetic secret immediately.
rm -f /var/lib/light-luna-p0/evil/agent-private.ed25519
stop_gate
start_gate
check REJECTED "$(agent preview-ed25519 "$REQ" "$NONCE")"
check ADVISORY_ONLY "$(agent preview-ed25519 synthetic-fresh-after-restart-0001 synthetic-fresh-restart-nonce-0001)"
echo "JOURNAL_RESTART_REPLAY_DENIED"
stop_gate
printf 'corrupt-journal-line\n' >> "$GATE_DIR/replay-ed25519.jsonl"
if setpriv --reuid "$GATE_UID" --regid "$IPC_GID" --clear-groups \
 "$GATE_BIN" serve-ed25519 "$POLICY" "$GATE_DIR" /run/light-luna-p0/shared \
 "$AGENT_UID" "$IPC_GID" 1 > /run/light-luna-p0/corrupt.log 2>&1;then
 echo "FAIL: corrupt replay journal accepted" >&2; exit 3;
fi
echo "LXD_DISTINCT_UID_POSITIVE_NEGATIVE_RESTART_PASS"
