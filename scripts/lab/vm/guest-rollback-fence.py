#!/usr/bin/env python3
"""VM-only advisory single-writer fence for a fixed synthetic nft rollback.

fcntl.flock serializes the primary PID1 watchdog and the independently
namespace-pinned fallback across *observation -> nft deletion -> journal ->
DONE*. The lock is NOT a durable transaction commit, signed owner authority,
cross-boot recovery, or protection against non-cooperating privileged writers.
"""
from contextlib import contextmanager
import fcntl
import os
from pathlib import Path
import re
import stat
import time

ROOT = Path("/opt/luna-p0-kernel-tests")
LOCK = ROOT / "synthetic-rollback-fence.lock"

@contextmanager
def guard(mod, stage_id, boot_id, *, max_wait=3.0):
    if not isinstance(stage_id, str) or not re.fullmatch(r"vm-[a-f0-9]{24}", stage_id):
        raise ValueError("invalid synthetic stage identity")
    if not isinstance(boot_id, str) or not re.fullmatch(
        r"[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}", boot_id
    ):
        raise ValueError("invalid boot identity")
    if not (0.01 <= max_wait <= 5):
        raise ValueError("unbounded rollback lock deadline")
    if ROOT.is_symlink() or not ROOT.is_dir():
        raise ValueError("unsafe lab root")
    flags = os.O_CREAT | os.O_RDWR | os.O_CLOEXEC | os.O_NOFOLLOW
    fd = os.open(LOCK, flags, 0o600)
    acquired = False
    try:
        info = os.fstat(fd)
        pathinfo = os.stat(LOCK, follow_symlinks=False)
        if (not stat.S_ISREG(info.st_mode) or info.st_uid != os.geteuid()
            or stat.S_IMODE(info.st_mode) != 0o600 or info.st_nlink != 1
            or info.st_ino != pathinfo.st_ino or info.st_dev != pathinfo.st_dev):
            raise ValueError("unsafe or replaced rollback lock")
        deadline = time.monotonic() + max_wait
        while True:
            try:
                fcntl.flock(fd, fcntl.LOCK_EX | fcntl.LOCK_NB)
                acquired = True
                break
            except BlockingIOError:
                if time.monotonic() >= deadline:
                    raise ValueError("rollback writer lock contention timed out")
                time.sleep(0.015)
        # Revalidate all transaction truth *after* obtaining exclusive access.
        if mod.boot_id() != boot_id:
            raise ValueError("rollback fence crossed a Linux boot")
        latest = mod.safe_load(mod.JOURNAL)
        if (set(latest) != {"stageId", "bootId", "state", "history"}
            or latest["stageId"] != stage_id or latest["bootId"] != boot_id):
            raise ValueError("rollback transaction identity mismatched under fence")
        history = latest["history"]
        if (type(history) is not list or not history
            or history != mod.PHASES[:len(history)]
            or latest["state"] != history[-1]):
            raise ValueError("invalid journal history under fence")
        yield latest
    finally:
        if acquired:
            fcntl.flock(fd, fcntl.LOCK_UN)
        os.close(fd)
