#!/usr/bin/env bash
# Explicit operator recovery ONLY after interrupted P0 synthetic VM test.
# Marks terminal rollback after verifying no live test namespace or nft table.
# Never applies or deletes firewall rules. Do NOT automate this in test runner.
set -euo pipefail
[[ "$(id -u)" == 0 ]] || exit 2
[[ -f /etc/light-luna-p0-kernel-vm ]] || exit 2
[[ "$(systemd-detect-virt)" == kvm ]] || exit 2
[[ "$(ls -1 /sys/class/net)" == lo ]] || exit 2
[[ "$(ip netns list)" == "" ]] || { echo "ACTIVE_NETNS_REFUSE";exit 3; }
[[ "$(nft list tables)" == "" ]] || { echo "UNEXPECTED_ROOT_NFT_TABLE_REFUSE";exit 3; }
# Catch surviving unnamed network namespaces held by lingering processes.
python3 - <<'PY'
from pathlib import Path
import os
baseline=os.readlink("/proc/1/ns/net")
for d in Path("/proc").iterdir():
    if not d.name.isdigit(): continue
    try: ns=os.readlink(str(d/"ns/net"))
    except (OSError,FileNotFoundError): continue
    if ns != baseline:
        raise SystemExit("LIVE_DIFFERENT_NETNS_REFUSE pid="+d.name+" "+ns)
print("VM_NAMESPACE_PROCESSES_CLEAN_PASS")
PY
work=/opt/luna-p0-kernel-tests
current_state="$(python3 - "$work/stage-journal.json" <<'PY'
import json,sys
with open(sys.argv[1]) as f: j=json.load(f)
assert set(j)=={"stageId","state","history"}
assert isinstance(j["stageId"],str)
state=j["state"]
prefix=["STAGED","VALIDATED","COMMITTED_SYNTHETIC"]
assert state in prefix
assert j["history"]==prefix[:prefix.index(state)+1]
print(state)
PY
)"
if [[ "$current_state" == COMMITTED_SYNTHETIC ]];then
  # We verified no namespace process/table still exists. Mark that a
  # post-interruption rollback already took effect; no table is recreated.
  python3 "$work/guest-stage-journal.py" transition ROLLED_BACK
  echo "VM_INTERRUPTED_COMMITTED_STAGE_RECONCILED_PASS"
else
  # No policy hook was committed. Preserve the aborted journal *as evidence*,
  # never pretend rollback ran; refuse to overwrite an old evidence file.
  evidence="$work/stage-journal.aborted-evidence.json"
  [[ ! -e "$evidence" ]] || { echo "PREVIOUS_ABORT_EVIDENCE_EXISTS_REFUSE";exit 3; }
  mv -n -- "$work/stage-journal.json" "$evidence"
  [[ -f "$evidence" && ! -f "$work/stage-journal.json" ]] || exit 3
  echo "VM_PRECOMMIT_STAGE_ABORT_ARCHIVED_PASS"
fi
echo "VM_INTERRUPTED_STAGE_JOURNAL_RECONCILED_AFTER_CLEANUP_PASS"
