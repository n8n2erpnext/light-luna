#!/usr/bin/env bash
# P0 VM ONLY: typed inert staging intent + synthetic scoped nft lab commit.
# Does not implement a production privileged Policy Executor/rollback.
set -euo pipefail
# Cleanup belongs strictly to the top-level Bash process, never its children.
main_bash_pid="$BASHPID"
[[ "$(id -u)" == 0 ]] || { echo "GUEST_ROOT_REQUIRED" >&2; exit 2; }
[[ -f /etc/light-luna-p0-kernel-vm ]] || { echo "MISSING_VM_MARKER" >&2; exit 2; }
[[ "$(systemd-detect-virt)" == kvm ]] || { echo "VM_KERNEL_REQUIRED" >&2; exit 2; }
[[ "$(ls -1 /sys/class/net)" == lo ]] || { echo "NIC_PRESENT" >&2; exit 2; }
fault_mode="${LUNA_VM_FAULT:-none}"
case "$fault_mode" in none|keeper_precommit|keeper_postcommit|watchdog_postcommit) ;; *) echo "UNRECOGNIZED_FAULT_MODE" >&2; exit 2;; esac
if [[ "$fault_mode" != none ]];then
 [[ "${LUNA_VM_AUTO_TTL:-0}" == 1 && "${LUNA_VM_FAULT_ACK:-}" == "KILL_VM_SYNTHETIC_KEEPER_OR_WATCHDOG" ]] || {
   echo "EXPLICIT_VM_FAULT_ACK_REQUIRED" >&2;exit 2;
 }
fi
work=/opt/luna-p0-kernel-tests
journal="$work/guest-stage-journal.py"
probe="$work/guest-netprobe.py"
manifest="$work/stage-intent.json"
state="$work/stage-journal.json"
for f in "$journal" "$probe" "$manifest";do [[ -s "$f" ]] || { echo "MISSING_STAGE_FIXTURE:$f" >&2; exit 2; };done
# Re-running the test is allowed ONLY after a verified terminal rollback state.
# An incomplete transaction remains blocked until independently investigated.
if [[ -e "$state" ]];then
  python3 - "$state" <<'PY'
import json,sys
with open(sys.argv[1],encoding="utf-8") as f:s=json.load(f)
assert set(s)=={"stageId","bootId","state","history"}
from pathlib import Path
assert s["bootId"]==Path('/proc/sys/kernel/random/boot_id').read_text().strip()
assert s["state"]=="ROLLED_BACK"
assert s["history"]==["STAGED","VALIDATED","COMMITTED_SYNTHETIC","ROLLED_BACK"]
PY
  rm -f -- "$state"
fi
[[ ! -e "$work/stage-fault-evidence.json" ]] || { echo "PREVIOUS_FAULT_EVIDENCE_UNARCHIVED_REFUSE" >&2;exit 3; }
rm -f -- "$work/stage-watchdog.log" "$work/stage-watchdog-done" "$work/stage-nft.nft" "$work/stage-netprobe.log" "$work/auto-watchdog-ready.json" "$work/auto-watchdog-done.json"
src="luna_stage_src_$$"
dst="luna_stage_dst_$$"
server_pid=""
anchor_pid=""
keeper_pid=""
keeper_unit=""
watchdog_pid=""
controller_pid=""
cleanup(){
  [[ "$BASHPID" == "$main_bash_pid" ]] || return 0
 [[ -z "$controller_pid" ]] || kill "$controller_pid" 2>/dev/null || true
 [[ -z "$server_pid" ]] || kill "$server_pid" 2>/dev/null || true
 [[ -z "$anchor_pid" ]] || kill "$anchor_pid" 2>/dev/null || true
 [[ -z "$watchdog_pid" ]] || kill "$watchdog_pid" 2>/dev/null || true
 [[ -z "${unit:-}" ]] || { systemctl stop "$unit.timer" >/dev/null 2>&1 || true; systemctl stop "$unit.service" >/dev/null 2>&1 || true; }
 [[ -z "$keeper_unit" ]] || systemctl stop "$keeper_unit.service" >/dev/null 2>&1 || true
 ip netns exec "$dst" nft delete table inet luna_p0_stage >/dev/null 2>&1 || true
 ip netns delete "$src" >/dev/null 2>&1 || true
 ip netns delete "$dst" >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM
# Safety-negative: malformed TTL and mismatched stage digest must fail even
# before STAGED is recorded; restore ONLY this synthetic guest manifest.
cp -p -- "$manifest" "$work/stage-intent.backup"
python3 - "$manifest" <<'PY'
import json,sys
p=sys.argv[1];m=json.load(open(p));m["ttlSeconds"]=999
with open(p,"w") as f:json.dump(m,f)
PY
if python3 "$journal" init STAGED >/dev/null 2>&1;then echo "VM_UNSAFE_TTL_ACCEPTED";exit 3;fi
cp -p -- "$work/stage-intent.backup" "$manifest"
python3 - "$manifest" <<'PY'
import json,sys
p=sys.argv[1];m=json.load(open(p));m["stageId"]="vm-ffffffffffffffffffffffff"
with open(p,"w") as f:json.dump(m,f)
PY
if python3 "$journal" init STAGED >/dev/null 2>&1;then echo "VM_MISMATCHED_STAGE_ACCEPTED";exit 3;fi
cp -p -- "$work/stage-intent.backup" "$manifest"
rm -f -- "$work/stage-intent.backup"
echo "VM_STAGE_TTL_AND_STAGE_ID_NEGATIVE_PASS"
python3 "$journal" init STAGED
# Commit without validated state MUST be denied.
if python3 "$journal" transition COMMITTED_SYNTHETIC >/dev/null 2>&1;then
 echo "VM_OUT_OF_ORDER_COMMIT_ACCEPTED";exit 3
fi
echo "VM_STAGE_OUT_OF_ORDER_DENIED_PASS"
ip netns add "$src"
ip netns add "$dst"
ip link add stage_source type veth peer name stage_target
ip link set stage_source netns "$src"
ip link set stage_target netns "$dst"
ip -n "$src" link set lo up
ip -n "$dst" link set lo up
ip -n "$src" link set stage_source up
ip -n "$dst" link set stage_target up
ip -n "$src" addr add 198.19.0.2/30 dev stage_source
ip -n "$dst" addr add 198.19.0.1/30 dev stage_target
ip -n "$src" -6 addr add fd42:198:19::2/126 dev stage_source nodad
ip -n "$dst" -6 addr add fd42:198:19::1/126 dev stage_target nodad
ip netns exec "$src" ping -n -c 1 -W 2 198.19.0.1 >/dev/null
ip netns exec "$src" ping -6 -n -c 1 -W 2 fd42:198:19::1 >/dev/null
ip netns exec "$dst" python3 "$probe" server > "$work/stage-netprobe.log" 2>&1 &
server_pid=$!
for _ in $(seq 1 30);do
 if grep -q 'VM_NETPROBE_READY' "$work/stage-netprobe.log";then break;fi
 if ! kill -0 "$server_pid" 2>/dev/null;then cat "$work/stage-netprobe.log";exit 3;fi
 sleep .1
done
ip netns exec "$src" python3 "$probe" client accept
echo "VM_STAGE_IPV6_ACCEPT_PASS"
# This fixed synthetic template is NOT generated by the product compiler.
cat > "$work/stage-nft.nft" <<'NFT'
table inet luna_p0_stage {
 chain input {
   type filter hook input priority -50; policy accept;
   ip saddr 198.19.0.2 tcp dport 18080 counter drop
   ip saddr 198.19.0.2 udp dport 18081 counter drop
   ip6 saddr fd42:198:19::2 icmpv6 type echo-request counter drop
 }
}
NFT
ip netns exec "$dst" nft -c -f "$work/stage-nft.nft"
if ip netns exec "$dst" nft list tables | grep -q luna_p0_stage;then
 echo "VM_STAGE_CHECK_MUTATED_KERNEL";exit 3
fi
# A live anchor process gives a systemd service a mount-namespace-independent
# reference to the destination netns, which otherwise may be hidden from PID 1.
ip netns exec "$dst" sleep 90 &
anchor_pid=$!
# ip netns exec is asynchronous: don't read /proc/$! before setns/exec.
# Wait until the process has entered the expected namespace.
name_netns="$(ip netns exec "$dst" readlink /proc/self/ns/net)"
anchor_netns=""
for _ in $(seq 1 40);do
  anchor_netns="$(readlink "/proc/$anchor_pid/ns/net" 2>/dev/null || true)"
  [[ "$anchor_netns" == "$name_netns" ]] && break
  kill -0 "$anchor_pid" 2>/dev/null || { echo "VM_ANCHOR_EXITED_EARLY" >&2;exit 3; }
  sleep .05
done
[[ "$anchor_netns" == "$name_netns" ]] || { echo "VM_ANCHOR_NAMESPACE_WRONG" >&2;exit 3; }
echo "VM_STAGE_ANCHOR_NETNS_VERIFIED"
# PID1-managed keeper survives termination of the lxc-exec shell-owned
# anchor; watchdog always uses this independent keeper PID/inode.
keeper_unit="luna-p0-stage-keeper-$$"
systemd-run --unit="$keeper_unit" --property=TimeoutStopSec=2 /usr/bin/nsenter -t "$anchor_pid" -n /usr/bin/sleep 240 > "$work/stage-keeper.log" 2>&1
for _ in $(seq 1 80);do
 keeper_pid="$(systemctl show "$keeper_unit.service" -p MainPID --value)"
 [[ "$keeper_pid" =~ ^[1-9][0-9]*$ ]] && [[ "$(readlink "/proc/$keeper_pid/ns/net" 2>/dev/null || true)" == "$anchor_netns" ]] && break
 sleep .05
done
[[ "$keeper_pid" =~ ^[1-9][0-9]*$ ]] && [[ "$(readlink "/proc/$keeper_pid/ns/net" 2>/dev/null || true)" == "$anchor_netns" ]] || { echo "VM_KEEPER_NOT_READY";exit 3; }
kill "$anchor_pid" 2>/dev/null || true
wait "$anchor_pid" 2>/dev/null || true
anchor_pid=""
[[ "$(readlink "/proc/$keeper_pid/ns/net")" == "$anchor_netns" ]] || exit 3
echo "VM_STAGE_INDEPENDENT_NAMESPACE_KEEPER_READY_PASS"
python3 "$journal" transition VALIDATED
# Second VALIDATED attempts are rejected, not silently idempotent.
if python3 "$journal" transition VALIDATED >/dev/null 2>&1;then
 echo "VM_REPEATED_VALIDATION_ACCEPTED";exit 3
fi
echo "VM_STAGE_DUPLICATE_TRANSITION_DENIED_PASS"
if [[ "${LUNA_VM_AUTO_TTL:-0}" == 1 ]];then
  # Arm an independent VM PID1 watchdog and confirm READY before nft -f.
  # No second explicit rollback trigger is allowed in this mode.
  unit="luna-p0-auto-wd-$$"
  anchor_ticks="$(python3 - "$keeper_pid" <<'PY'
from pathlib import Path
import sys
s=Path("/proc/"+sys.argv[1]+"/stat").read_text().rsplit(") ",1)[1].split()
print(s[19])
PY
)"
  systemd-run --unit="$unit" --property=RuntimeMaxSec=20s /usr/bin/python3 "$work/guest-auto-watchdog.py" "$keeper_pid" "$anchor_ticks" "$anchor_netns" > "$work/stage-watchdog.log" 2>&1
  for _ in $(seq 1 70);do
    [[ -f "$work/auto-watchdog-ready.json" ]] && break
    if [[ "$(systemctl is-failed "$unit.service" || true)" == failed ]];then
      journalctl -u "$unit.service" -o cat -n 18 --no-pager >&2
      echo "VM_AUTOWATCHDOG_SERVICE_FAILED_BEFORE_READY" >&2;exit 3
    fi
    sleep .05
  done
  [[ -f "$work/auto-watchdog-ready.json" ]] || { echo "VM_AUTOWATCHDOG_NEVER_ARMED" >&2;exit 3; }
  python3 - "$work/auto-watchdog-ready.json" "$manifest" <<'PY'
import json,sys,time
r=json.load(open(sys.argv[1]));m=json.load(open(sys.argv[2]))
from pathlib import Path
assert r["armed"] is True and r["stageId"]==m["stageId"]
assert r["bootId"]==Path('/proc/sys/kernel/random/boot_id').read_text().strip()
assert r["ttlSeconds"]==m["ttlSeconds"]
assert r["deadlineMonotonic"]-time.monotonic()>=5, "TTL too near; cannot commit safely"
PY
  echo "VM_AUTO_WATCHDOG_PRECOMMIT_READY_PASS"
  if [[ "$fault_mode" == keeper_precommit ]];then
    # Fault only a specific transient PID1-managed lab keeper (never systemd).
    systemctl kill --kill-who=main --signal=SIGKILL "$keeper_unit.service"
    sleep .3
    [[ "$(systemctl show "$keeper_unit.service" -p MainPID --value)" == 0 ]] || { echo "KEEPER_DEATH_UNVERIFIED";exit 3; }
    echo "VM_FAULT_KEEPER_PRECOMMIT_KILLED"
    # Watchdog must refuse success after its pinned keeper PID is gone.
    python3 "$work/guest-fault-evidence.py" observe keeper_precommit "$unit.service" "$keeper_unit.service"
    echo "VM_FAULT_KEEPER_PRECOMMIT_NO_SUCCESS_PASS"
    exit 0
  fi
fi
ttl="$(python3 - "$manifest" <<'PY'
import json,sys
with open(sys.argv[1],encoding="utf-8") as f:m=json.load(f)
assert m["ttlSeconds"]>=4 and m["ttlSeconds"]<=10
print(m["ttlSeconds"])
PY
)"
# Start a separate-session watchdog BEFORE installing the synthetic hook.
# It has fixed table+namespace arguments and a short TTL. No host namespace.
# Systemd transient VM timer survives a killed controller and shell teardown.
if [[ "${LUNA_VM_AUTO_TTL:-0}" != 1 ]];then unit="luna-p0-stage-wd-$$";fi
# Default mode demonstrates explicitly triggered recovery after controller death.
# Auto mode already armed its independent monotonic watchdog above.
ip netns exec "$dst" nft -f "$work/stage-nft.nft"
python3 "$journal" transition COMMITTED_SYNTHETIC
if python3 "$journal" transition COMMITTED_SYNTHETIC >/dev/null 2>&1;then
 echo "VM_DUPLICATE_COMMIT_ACCEPTED";exit 3
fi
echo "VM_STAGE_DUPLICATE_COMMIT_DENIED_PASS"
ip netns exec "$src" python3 "$probe" client drop
if ip netns exec "$src" ping -6 -n -c 1 -W 1 fd42:198:19::1 >/dev/null 2>&1;then
 echo "VM_STAGE_IPV6_DROP_FAILED";exit 3
fi
for rule in 'tcp dport 18080' 'udp dport 18081' 'icmpv6 type echo-request';do
 ip netns exec "$dst" nft list chain inet luna_p0_stage input | grep -F "$rule" | grep -Eq 'counter packets [1-9][0-9]*' || {
  echo "VM_RULE_COUNTER_MISSING:$rule";exit 3;
 }
done
echo "VM_STAGE_TCP_UDP_IPV6_DROP_COUNTER_PASS"
if [[ "$fault_mode" == watchdog_postcommit || "$fault_mode" == keeper_postcommit ]];then
  if [[ "$fault_mode" == watchdog_postcommit ]];then victim="$unit.service";else victim="$keeper_unit.service";fi
  systemctl kill --kill-who=main --signal=SIGKILL "$victim"
  sleep .3
  [[ "$(systemctl show "$victim" -p MainPID --value)" == 0 ]] || { echo "FAULT_SERVICE_DEATH_UNVERIFIED";exit 3; }
  echo "VM_FAULT_SERVICE_POSTCOMMIT_KILLED"
  # Verify no false DONE, confirm synthetic table still exists after TTL.
  python3 "$work/guest-fault-evidence.py" observe "$fault_mode" "$unit.service" "$keeper_unit.service" "$keeper_pid" "$anchor_netns" "$dst"
  echo "VM_FAULT_POSTCOMMIT_UNRECOVERED_RECORDED_PASS"
  exit 0
fi
sleep 30 &
controller_pid=$!
kill "$controller_pid" 2>/dev/null || true
wait "$controller_pid" 2>/dev/null || true
controller_pid=""
echo "VM_STAGE_CONTROLLER_KILLED"
if [[ "${LUNA_VM_AUTO_TTL:-0}" == 1 ]];then
  # Crucially: no controller command to revert. Only observe watchdog output.
  for _ in $(seq 1 185);do
    [[ -f "$work/auto-watchdog-done.json" ]] && break
    sleep .10
  done
  if [[ ! -f "$work/auto-watchdog-done.json" ]];then
    journalctl -u "$unit.service" -o cat -n 30 --no-pager >&2
    echo "VM_AUTOWATCHDOG_TTL_EXPIRED_WITHOUT_ROLLBACK" >&2;exit 3
  fi
  python3 - "$work/auto-watchdog-done.json" <<'PY'
import json,sys
d=json.load(open(sys.argv[1]))
from pathlib import Path
assert d["status"]=="AUTO_ROLLBACK_CONFIRMED"
assert d["bootId"]==Path('/proc/sys/kernel/random/boot_id').read_text().strip()
assert 0<=d["lagSeconds"]<=3.5,d
print("VM_AUTO_TTL_EXPIRY_NO_TRIGGER_ROLLBACK_PASS")
PY
else
  systemd-run --wait --collect --unit="$unit" /bin/bash "$work/guest-rollback.sh" "$dst" luna_p0_stage "$work/stage-watchdog-done" "$keeper_pid" "$anchor_netns" > "$work/stage-watchdog.log" 2>&1 || { cat "$work/stage-watchdog.log" >&2;exit 3; }
  [[ -f "$work/stage-watchdog-done" ]] || { echo "VM_STAGE_SEPARATE_ROLLBACK_MISSING";cat "$work/stage-watchdog.log" >&2;exit 3; }
  echo "VM_STAGE_SYSTEMD_ROLLBACK_EXPLICIT_TRIGGER_PASS"
fi
ip netns exec "$src" python3 "$probe" client accept
ip netns exec "$src" ping -6 -n -c 1 -W 2 fd42:198:19::1 >/dev/null
if ip netns exec "$dst" nft list tables | grep -q luna_p0_stage;then echo "VM_STAGE_TABLE_REMAINS";exit 3;fi
python3 - "$state" <<'PY'
import json,sys
with open(sys.argv[1],encoding="utf-8") as f:s=json.load(f)
from pathlib import Path
assert s["bootId"]==Path('/proc/sys/kernel/random/boot_id').read_text().strip()
assert s["state"] == "ROLLED_BACK"
assert s["history"] == ["STAGED","VALIDATED","COMMITTED_SYNTHETIC","ROLLED_BACK"]
PY
if [[ "${LUNA_VM_AUTO_TTL:-0}" == 1 ]];then
  echo "VM_STAGE_AUTONOMOUS_TTL_IPV6_TCP_UDP_RESTORED_PASS"
else
  echo "VM_STAGE_SEPARATE_SERVICE_ROLLBACK_IPV6_TCP_UDP_RESTORED_PASS"
fi
echo "VM_STAGE_SYNTHETIC_TRANSACTION_TEST_PASS"
