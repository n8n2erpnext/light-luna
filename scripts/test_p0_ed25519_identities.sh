#!/usr/bin/env bash
# Dev test only: same UID in an isolated scratch directory, intentionally NOT
# evidence of distinct Linux service-user isolation or privilege separation.
set -euo pipefail
[[ "$(id -u)" -ne 0 ]] || { echo "nonroot only" >&2; exit 2; }
export GOTOOLCHAIN=local GOPROXY=off CGO_ENABLED=0
tmp="$(mktemp -d)"
chmod 700 "$tmp"
pid=""
cleanup(){
 if [[ -n "$pid" ]]; then kill "$pid" 2>/dev/null || true; wait "$pid" 2>/dev/null || true; fi
 rm -rf -- "$tmp"
}
trap cleanup EXIT INT TERM
go build -trimpath -o "$tmp/luna-gate" ./cmd/luna-gate
go build -trimpath -o "$tmp/luna-agent" ./cmd/luna-agent
"$tmp/luna-gate" init-ed25519 "$tmp/agent" "$tmp/gate"
[[ -f "$tmp/agent/agent-private.ed25519" ]]
[[ -f "$tmp/gate/agent-public.ed25519" ]]
[[ ! -f "$tmp/gate/agent-private.ed25519" ]]
[[ ! -f "$tmp/agent/agent-public.ed25519" ]]
mkdir -m 710 "$tmp/shared"
chmod 710 "$tmp/shared"
p="$PWD/testdata/host-baseline.json"
uid="$(id -u)"
gid="$(id -g)"
start(){
 "$tmp/luna-gate" serve-ed25519 "$p" "$tmp/gate" "$tmp/shared" "$uid" "$gid" 20 > "$tmp/gatelog" 2>&1 &
 pid=$!
 for _ in $(seq 1 80); do
  [[ -S "$tmp/shared/gate.sock" ]] && return 0
  kill -0 "$pid" 2>/dev/null || { cat "$tmp/gatelog" >&2; exit 1; }
  sleep .05
 done
 cat "$tmp/gatelog" >&2
 echo "Gate not ready" >&2
 exit 1
}
stop(){
 kill "$pid" 2>/dev/null || true
 wait "$pid" 2>/dev/null || true
 pid=""
}
check(){
 local output="$1" expected="$2"
 printf '%s\n' "$output" | python3 -c 'import sys,json
x=json.load(sys.stdin)
assert x["decision"]==sys.argv[1],x
assert x["kernelWrites"] is False,x
assert x["executorStatus"]=="NOT_AVAILABLE",x
assert x["replayProtection"]=="FSYNC_JOURNAL_DEMO_ONLY",x' "$expected"
}
start
r1=synthetic-ed-process-request-001
n1=synthetic-ed-process-nonce-001
result="$("$tmp/luna-agent" preview-ed25519 "$p" "$tmp/agent" "$tmp/shared/gate.sock" "$uid" "$r1" "$n1")"
check "$result" ADVISORY_ONLY
check "$("$tmp/luna-agent" preview-ed25519 "$p" "$tmp/agent" "$tmp/shared/gate.sock" "$uid" "$r1" "$n1")" REJECTED
check "$("$tmp/luna-agent" apply-ed25519 "$p" "$tmp/agent" "$tmp/shared/gate.sock" "$uid" synthetic-ed-apply-request-001 synthetic-ed-apply-nonce-001)" REJECTED
# Agent refuses an unexpected Gate UID before sending the signed request.
if "$tmp/luna-agent" preview-ed25519 "$p" "$tmp/agent" "$tmp/shared/gate.sock" "$((uid+1))" > "$tmp/badUID" 2>&1;then
 echo "Client failed to reject wrong Gate UID" >&2;exit 1
fi
stop
start
check "$("$tmp/luna-agent" preview-ed25519 "$p" "$tmp/agent" "$tmp/shared/gate.sock" "$uid" "$r1" "$n1")" REJECTED
check "$("$tmp/luna-agent" preview-ed25519 "$p" "$tmp/agent" "$tmp/shared/gate.sock" "$uid" synthetic-ed-fresh-request-001 synthetic-ed-fresh-nonce-001)" ADVISORY_ONLY
stop
printf 'corrupted-tail\n' >> "$tmp/gate/replay-ed25519.jsonl"
if "$tmp/luna-gate" serve-ed25519 "$p" "$tmp/gate" "$tmp/shared" "$uid" "$gid" 1 > "$tmp/invalid" 2>&1;then
 echo "Corrupt replay journal accepted" >&2;exit 1
fi
echo "P0_ED25519_TWO_BINARIES_RESTART_FAILCLOSED_PASS"
