#!/usr/bin/env python3
"""Immutable, fsynced VM-only *observation receipts* around synthetic nft rollback.

This is NOT a signed WAL, kernel/disk atomic commit, production executor,
automatic recovery, or proof that any nft action on another network is safe.
Call methods only while holding the cooperative guest-rollback-fence flock.
"""
import hashlib
import json
import os
from pathlib import Path
import re
import stat

ROOT = Path("/opt/luna-p0-kernel-tests")
SCHEMA = "lightluna.vm-rollback-observation/v1alpha1"
WRITERS = {"primary", "fallback"}
KINDS = ("prepared", "kernel-absent", "journal-terminal")


def _id(stage, boot, netns, ready):
    if not re.fullmatch(r"vm-[a-f0-9]{24}", stage):
        raise ValueError("invalid stage id")
    if not re.fullmatch(r"[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}", boot):
        raise ValueError("invalid boot id")
    if not re.fullmatch(r"net:\[[0-9]+\]", netns):
        raise ValueError("invalid target netns")
    if (not isinstance(ready, dict) or ready.get("armed") is not True
        or ready.get("stageId") != stage or ready.get("bootId") != boot
        or ready.get("netnsInode") != netns
        or type(ready.get("armedMonotonic")) not in (int, float)
        or type(ready.get("deadlineMonotonic")) not in (int, float)
        or type(ready.get("anchorPid")) is not int
        or type(ready.get("anchorBirthTicks")) is not int):
        raise ValueError("primary READY not bound to stage/boot/netns")
    if ready["deadlineMonotonic"] <= ready["armedMonotonic"] or ready["anchorPid"] < 2:
        raise ValueError("invalid monotonic deadline / keeper")
    snapshot = {k: ready[k] for k in ("stageId", "bootId", "netnsInode",
        "anchorPid", "anchorBirthTicks", "armedMonotonic", "deadlineMonotonic")}
    return hashlib.sha256(
        json.dumps(snapshot, sort_keys=True, separators=(",", ":")).encode()
    ).hexdigest()[:24]


def path(epoch, kind):
    if not re.fullmatch(r"[0-9a-f]{24}", epoch) or kind not in KINDS:
        raise ValueError("invalid receipt path")
    return ROOT / ("rollback-" + epoch + "-" + kind + ".json")


def _safe_load(p):
    # Inspect the OPENED receipt inode rather than relying on pre-open Path
    # checks. O_NOFOLLOW protects the final pathname component; O_NONBLOCK
    # prevents FIFO open hangs. Trusted ancestor paths remain a separate lab
    # requirement. No observation is inferred from a refused receipt.
    flags = os.O_RDONLY | os.O_CLOEXEC | os.O_NOFOLLOW | os.O_NONBLOCK
    with os.fdopen(os.open(p, flags), "rb") as source:
        inode = os.fstat(source.fileno())
        if (not stat.S_ISREG(inode.st_mode) or inode.st_nlink != 1
            or not 1 <= inode.st_size <= 4096):
            raise ValueError("unsafe, hardlinked or over-limit receipt")
        raw = source.read(4097)
        if len(raw) > 4096:
            raise ValueError("receipt exceeded read-time byte limit")
    def pairs(seq):
        d = {}
        for k, v in seq:
            if k in d: raise ValueError("duplicate receipt key")
            d[k] = v
        return d
    return json.loads(raw.decode("utf-8"), object_pairs_hook=pairs)


def _write_exclusive(p, record):
    if ROOT.is_symlink() or not ROOT.is_dir() or p.parent != ROOT:
        raise ValueError("unsafe receipt root")
    encoded = (json.dumps(record, sort_keys=True, separators=(",", ":"))+"\n").encode()
    if len(encoded) > 4096:
        raise ValueError("receipt oversized")
    # No overwrite, no replacement. fsync file and directory; this is a
    # bounded process-crash experiment, not a proven power-loss barrier.
    fd = os.open(p, os.O_WRONLY | os.O_CREAT | os.O_EXCL | os.O_NOFOLLOW, 0o600)
    with os.fdopen(fd, "wb") as output:
        output.write(encoded)
        output.flush()
        os.fsync(output.fileno())
    directory = os.open(ROOT, os.O_RDONLY | os.O_DIRECTORY)
    try: os.fsync(directory)
    finally: os.close(directory)
    return record


def _digest(record):
    return hashlib.sha256(json.dumps(record, sort_keys=True,
        separators=(",", ":")).encode()).hexdigest()


def _record(kind, epoch, stage, boot, netns, writer, preceding):
    if writer not in WRITERS: raise ValueError("unexpected rollback writer")
    return {"schema":SCHEMA, "phase":kind, "epoch":epoch,
        "stageId":stage, "bootId":boot, "netnsInode":netns,
        "writer":writer, "previousSha256":preceding,
        "guestSyntheticOnly":True, "ownerApproved":False,
        "automaticRecoveryAuthorized":False}


def prepare(stage, boot, netns, ready, writer):
    epoch = _id(stage, boot, netns, ready)
    # Fail closed rather than retrying an incomplete old kernel mutation.
    if any(path(epoch, k).exists() or path(epoch, k).is_symlink() for k in KINDS):
        raise ValueError("rollback receipts already exist; manual investigation required")
    record = _record("prepared", epoch, stage, boot, netns, writer, None)
    _write_exclusive(path(epoch, "prepared"), record)
    return epoch


def advance(epoch, kind, stage, boot, netns, writer):
    if kind not in ("kernel-absent", "journal-terminal"):
        raise ValueError("invalid advance phase")
    prior = "prepared" if kind == "kernel-absent" else "kernel-absent"
    prev = _safe_load(path(epoch, prior))
    expect = _record(prior, epoch, stage, boot, netns, writer,
        None if prior == "prepared" else _digest(_safe_load(path(epoch, "prepared"))))
    if prev != expect:
        raise ValueError("prior receipt identity/hash chain invalid")
    record = _record(kind, epoch, stage, boot, netns, writer, _digest(prev))
    return _write_exclusive(path(epoch, kind), record)


def inspect(stage, boot, netns, ready):
    """Read-only lab report; missing DONE never becomes success by inference."""
    epoch = _id(stage, boot, netns, ready)
    found = []
    writer = None
    previous = None
    for kind in KINDS:
        p = path(epoch, kind)
        if not p.exists() and not p.is_symlink():
            break
        try:
            item = _safe_load(p)
            if writer is None: writer = item.get("writer")
            if item != _record(kind, epoch, stage, boot, netns, writer, previous):
                return {"status":"INVALID_RECEIPT_FAIL_CLOSED", "epoch":epoch}
            previous = _digest(item)
        except (ValueError, OSError, json.JSONDecodeError):
            return {"status":"INVALID_RECEIPT_FAIL_CLOSED", "epoch":epoch}
        found.append(kind)
    if any(path(epoch, k).exists() or path(epoch, k).is_symlink()
           for k in KINDS[len(found):]) and len(found) < len(KINDS):
        return {"status":"INVALID_RECEIPT_FAIL_CLOSED", "epoch":epoch}
    states = {0:"NO_ROLLBACK_RECEIPT",1:"PREPARED_NO_KERNEL_ABSENCE_PROOF",
              2:"KERNEL_ABSENCE_OBSERVED_JOURNAL_NOT_TERMINAL_AT_RECEIPT",
              3:"JOURNAL_TERMINAL_RECEIPT_PRESENT_DONE_NOT_EVALUATED"}
    return {"status":states[len(found)],"epoch":epoch, "writer":writer,
            "phases":found,"ownerApproved":False,
            "automaticRecoveryAuthorized":False}
