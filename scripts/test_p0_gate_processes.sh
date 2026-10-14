#!/usr/bin/env bash
# Source-only, nonroot, local development acceptance. NEVER installs a daemon
# or calls nftables; temp test credentials are destroyed on exit.
set -euo pipefail
if [[ "$(id -u)" == 0 ]]; then
  echo "P0 test refuses root" >&2
  exit 2
fi
export GOTOOLCHAIN=local GOPROXY=off CGO_ENABLED=0
tmp="$(mktemp -d)"
chmod 700 "$tmp"
pid=""
cleanup() {
  if [[ -n "$pid" ]]; then
    kill "$pid" 2>/dev/null || true
    wait "$pid" 2>/dev/null || true
  fi
  rm -rf -- "$tmp"
}
trap cleanup EXIT INT TERM
go build -trimpath -o "$tmp/luna-gate" ./cmd/luna-gate
go build -trimpath -o "$tmp/luna-agent" ./cmd/luna-agent
"$tmp/luna-gate" init "$tmp/session"
p="$PWD/testdata/host-baseline.json"
start_gate(){
  "$tmp/luna-gate" serve "$p" "$tmp/session" 20 > "$tmp/gate.log" 2>&1 &
  pid=$!
  for _ in $(seq 1 80); do
    if [[ -S "$tmp/session/gate.sock" ]]; then
      return 0
    fi
    if ! kill -0 "$pid" 2>/dev/null; then
      cat "$tmp/gate.log" >&2
      exit 1
    fi
    sleep 0.05
  done
  echo "gate did not open socket" >&2
  cat "$tmp/gate.log" >&2
  exit 1
}
stop_gate(){
  kill "$pid" 2>/dev/null || true
  wait "$pid" 2>/dev/null || true
  pid=""
}
check(){
  local expected="$1" json="$2"
  printf '%s\n' "$json" | python3 -c '
import sys,json
x=json.load(sys.stdin)
assert x["decision"]==sys.argv[1], x
assert x["kernelWrites"] is False, x
assert x["executorStatus"]=="NOT_AVAILABLE", x
assert x["replayProtection"]=="FSYNC_JOURNAL_DEMO_ONLY", x
' "$expected"
}
start_gate
req=synthetic-separate-process-request-001
nonce=synthetic-separate-process-nonce-001
first="$("$tmp/luna-agent" preview "$p" "$tmp/session" "$req" "$nonce")"
check ADVISORY_ONLY "$first"
same="$("$tmp/luna-agent" preview "$p" "$tmp/session" "$req" "$nonce")"
check REJECTED "$same"
apply="$("$tmp/luna-agent" apply "$p" "$tmp/session" synthetic-apply-request-001 synthetic-apply-nonce-001)"
check REJECTED "$apply"
stop_gate
start_gate
after_restart="$("$tmp/luna-agent" preview "$p" "$tmp/session" "$req" "$nonce")"
check REJECTED "$after_restart"
new="$("$tmp/luna-agent" preview "$p" "$tmp/session" synthetic-fresh-request-001 synthetic-fresh-nonce-001)"
check ADVISORY_ONLY "$new"
stop_gate
# A tainted journal MUST prevent service from starting, not wipe replay state.
printf 'corrupt-json\n' >> "$tmp/session/replay.jsonl"
if "$tmp/luna-gate" serve "$p" "$tmp/session" 1 > "$tmp/corrupt.log" 2>&1; then
  echo "corrupt replay journal incorrectly allowed" >&2
  exit 1
fi
echo "P0_TWO_PROCESSES_JOURNAL_RESTART_CORRUPTION_PASS"
