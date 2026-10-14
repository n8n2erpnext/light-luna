#!/usr/bin/env python3
"""Luna P0 ONLY: fail-closed lab receipt progression, no kernel actions.

This does NOT provide production journal concurrency, owner authorization,
signed receipts, power-loss guarantees, or true executor transaction atomicity.
"""
import hashlib
import json
import os
import re
import sys
import tempfile
from pathlib import Path

SCHEMA = "lightluna.vm-stage-intent/v1alpha1"
MODE = "SYNTHETIC_VM_ONLY_NO_EXECUTOR"
TABLE = "luna_p0_stage"
EFFECT = "none: manifest only; synthetic VM-only hook tested separately"
PHASES = ["STAGED", "VALIDATED", "COMMITTED_SYNTHETIC", "ROLLED_BACK"]
ROOT = Path("/opt/luna-p0-kernel-tests")
MANIFEST = ROOT / "stage-intent.json"
JOURNAL = ROOT / "stage-journal.json"
EXPECTED_KEYS = {"schema","mode","policyName","policyRevision","sourcePreviewHash",
                 "stageId","table","ttlSeconds","phases","effect",
                 "kernelWrites","ownerApproved","scope"}

def boot_id():
    # A kernel boot boundary must not reuse a synthetic transaction identity.
    # Linux boot_id is regenerated on each boot and is not owner authentication.
    value = Path("/proc/sys/kernel/random/boot_id").read_text(encoding="ascii").strip()
    if not re.fullmatch(r"[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}", value) or value == "00000000-0000-0000-0000-000000000000":
        raise ValueError("missing or invalid kernel boot ID")
    return value

def unique_pairs(pairs):
    output = {}
    for key, value in pairs:
        if key in output:
            raise ValueError("duplicate JSON key")
        output[key] = value
    return output

def safe_load(path):
    if path.is_symlink() or not path.is_file() or path.stat().st_size > 4096:
        raise ValueError("missing, symlinked or over-limit staging file")
    with path.open("r", encoding="utf-8") as handle:
        return json.load(handle, object_pairs_hook=unique_pairs)

def manifest():
    m = safe_load(MANIFEST)
    if set(m) != EXPECTED_KEYS:
        raise ValueError("manifest must contain exact known keys")
    if (m["schema"] != SCHEMA or m["mode"] != MODE
        or m["table"] != TABLE or m["scope"] != "KVM_GUEST_PRIVATE_NETNS_ONLY"
        or m["kernelWrites"] is not False or m["ownerApproved"] is not False
        or m["phases"] != PHASES or m["policyName"] != "host-baseline"
        or m["effect"] != EFFECT
        or type(m["policyRevision"]) is not int or m["policyRevision"] != 1
        or type(m["ttlSeconds"]) is not int
        or not 1 <= m["ttlSeconds"] <= 10):
        raise ValueError("unexpected or unsafe manifest")
    hash_text = m["sourcePreviewHash"]
    if not isinstance(hash_text,str) or not re.fullmatch("[a-f0-9]{64}",hash_text):
        raise ValueError("missing valid source policy hash")
    expected = "vm-" + hashlib.sha256(
        (SCHEMA + "\x00" + hash_text + "\x00" + TABLE + "\x00" + str(m["ttlSeconds"])).encode()
    ).hexdigest()[:24]
    if m["stageId"] != expected:
        raise ValueError("stage id mismatch")
    return m

def persist(state):
    # This is a single-controller VM lab. No multi-process flock or verified
    # fs power-loss durability claim; fsync is for basic crash fault simulation.
    fd, temp_name = tempfile.mkstemp(prefix=".stage-journal-", dir=str(ROOT))
    try:
        os.fchmod(fd, 0o600)
        with os.fdopen(fd,"w",encoding="utf-8") as f:
            json.dump(state,f,sort_keys=True,separators=(",",":"))
            f.write("\n")
            f.flush()
            os.fsync(f.fileno())
        os.replace(temp_name, JOURNAL)
        parent = os.open(ROOT, os.O_RDONLY|os.O_DIRECTORY)
        try: os.fsync(parent)
        finally: os.close(parent)
    finally:
        if os.path.exists(temp_name):
            os.unlink(temp_name)

def transition(command, target, m):
    if command == "init":
        if JOURNAL.exists() or JOURNAL.is_symlink():
            raise ValueError("previous transaction not cleared; fail closed")
        if target != "STAGED":
            raise ValueError("invalid initial state")
        state = {"stageId": m["stageId"],"bootId":boot_id(),"state":"STAGED","history":["STAGED"]}
    else:
        state = safe_load(JOURNAL)
        if set(state) != {"stageId","bootId","state","history"} or state["stageId"] != m["stageId"]:
            raise ValueError("journal mismatch")
        if state["bootId"] != boot_id():
            raise ValueError("journal belongs to a different Linux boot; independent recovery required")
        history = state["history"]
        if not isinstance(history,list) or history != PHASES[:len(history)] or state["state"] != history[-1]:
            raise ValueError("journal history invalid")
        step = len(history)
        if step >= len(PHASES) or target != PHASES[step]:
            raise ValueError("out-of-order or repeated transition")
        state["state"] = target
        state["history"] = history + [target]
    persist(state)
    print("VM_STAGE_" + target + "_PASS")

def main():
    if os.geteuid() != 0 or not Path("/etc/light-luna-p0-kernel-vm").is_file():
        raise ValueError("VM lab identity marker/root required")
    if os.listdir("/sys/class/net") != ["lo"]:
        raise ValueError("guest contains NIC")
    if ROOT.is_symlink() or not ROOT.is_dir():
        raise ValueError("stage root unavailable")
    if len(sys.argv) != 3 or sys.argv[1] not in ("init","transition") or sys.argv[2] not in PHASES:
        raise ValueError("usage: guest-stage-journal.py init|transition STATE")
    m = manifest()
    transition(sys.argv[1],sys.argv[2],m)

if __name__=="__main__":
    try: main()
    except (OSError,ValueError,KeyError,TypeError,IndexError) as err:
        print("VM_STAGE_DENIED:",str(err),file=sys.stderr)
        sys.exit(2)
