#!/usr/bin/env python3
"""Luna P0 ONLY: fail-closed lab receipt progression, no kernel actions.

This does NOT provide production journal concurrency, owner authorization,
signed receipts, power-loss guarantees, or true executor transaction atomicity.
"""
import fcntl
import hashlib
import json
import os
import re
import stat
import sys
import tempfile
from contextlib import contextmanager
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
    # Inspect the opened inode, NOT a potentially stale path precheck. Final
    # symlinks are refused by O_NOFOLLOW; O_NONBLOCK avoids FIFO open hangs.
    # This guest-only fixture does not defend against hostile VM root or
    # replacement of trusted ancestor directories.
    flags = os.O_RDONLY | os.O_CLOEXEC | os.O_NOFOLLOW | os.O_NONBLOCK
    with os.fdopen(os.open(path, flags), "rb") as handle:
        opened = os.fstat(handle.fileno())
        if (not stat.S_ISREG(opened.st_mode) or opened.st_nlink != 1
            or opened.st_size < 1 or opened.st_size > 4096):
            raise ValueError("nonregular, hardlinked or over-limit staging file")
        # A file can grow after fstat: bound the actual bytes read as well.
        raw = handle.read(4097)
        if len(raw) > 4096:
            raise ValueError("staging file grew beyond byte limit")
    return json.loads(raw.decode("utf-8"), object_pairs_hook=unique_pairs)

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

def persist(state, create_only=False):
    # Synthetic VM-only persistence. Cooperating transition() callers hold
    # a Linux directory flock, but persist() alone provides no writer lock.
    # Fsync here is basic crash simulation, not physical power-loss proof.
    # Initial STAGED is atomic no-clobber; later transitions use replace
    # only while the cooperating writer holds its full-scope directory lock.
    fd, temp_name = tempfile.mkstemp(prefix=".stage-journal-", dir=str(ROOT))
    try:
        os.fchmod(fd, 0o600)
        with os.fdopen(fd,"w",encoding="utf-8") as f:
            json.dump(state,f,sort_keys=True,separators=(",",":"))
            f.write("\n")
            f.flush()
            os.fsync(f.fileno())
        if create_only:
            # link() fails with EEXIST if another journal or symlink appears.
            # fsynced temp contents become the new entry without overwrite;
            # the finally block unlinks the now-secondary temp name.
            os.link(temp_name, JOURNAL)
        else:
            os.replace(temp_name, JOURNAL)
        parent = os.open(ROOT, os.O_RDONLY|os.O_DIRECTORY)
        try: os.fsync(parent)
        finally: os.close(parent)
    finally:
        if os.path.exists(temp_name):
            os.unlink(temp_name)

@contextmanager
def stage_writer_guard():
    # Advisory lock on the existing lab DIRECTORY inode, no new lockfile and
    # no journal mutation before acquisition. Opens are independent, so even
    # two callers within one process contend. Only cooperating writers are
    # fenced; hostile guest root or a replaced ancestor is outside P0 proof.
    flags = os.O_RDONLY | os.O_DIRECTORY | os.O_CLOEXEC | os.O_NOFOLLOW
    fd = os.open(ROOT, flags)
    try:
        if not stat.S_ISDIR(os.fstat(fd).st_mode):
            raise ValueError("stage root is not a directory")
        try:
            fcntl.flock(fd, fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError as err:
            raise ValueError("another VM stage journal writer is active") from err
        try:
            yield
        finally:
            fcntl.flock(fd, fcntl.LOCK_UN)
    finally:
        os.close(fd)

def transition(command, target, m):
    # One cooperative writer holds the same directory inode from the first
    # read through validation, publication and parent-directory fsync.
    with stage_writer_guard():
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
        persist(state, create_only=(command == "init"))
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
