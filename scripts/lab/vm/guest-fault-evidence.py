#!/usr/bin/env python3
"""Read-only detection and fsynced evidence of killed guest-only lab services.

No policy executor, auto recovery, production firewall or forged success receipt.
The caller is responsible for explicit VM-only namespace cleanup after exit.
"""
import importlib.util
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import time

ROOT = Path("/opt/luna-p0-kernel-tests")
FAULT = ROOT / "stage-fault-evidence.json"
READY = ROOT / "auto-watchdog-ready.json"
DONE = ROOT / "auto-watchdog-done.json"
MOD = ROOT / "guest-stage-journal.py"
MODES = {"keeper_precommit", "keeper_postcommit", "watchdog_postcommit"}


def run(*args):
    return subprocess.run(args, capture_output=True, text=True, timeout=5, check=False)


def load_module():
    if MOD.is_symlink() or not MOD.is_file():
        raise ValueError("missing or symlinked stage validator")
    spec = importlib.util.spec_from_file_location("fault_journal", MOD)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def inspect(mode, watchdog, keeper, pid=None, inode=None, netns=None, *, root=ROOT):
    """Pure-ish observation: no mutations. Read-only guest kernel queries only."""
    if os.geteuid() != 0 or not Path("/etc/light-luna-p0-kernel-vm").is_file():
        raise ValueError("guest VM marker/root required")
    if sorted(os.listdir("/sys/class/net")) != ["lo"]:
        raise ValueError("guest NIC detected")
    if run("systemd-detect-virt").stdout.strip() != "kvm":
        raise ValueError("not KVM")
    if mode not in MODES or not watchdog.startswith("luna-p0-auto-wd-") or not keeper.startswith("luna-p0-stage-keeper-"):
        raise ValueError("invalid fault mode/unit identity")
    if FAULT.exists() or FAULT.is_symlink() or DONE.exists() or DONE.is_symlink():
        raise ValueError("preexisting fault or success evidence")
    mod = load_module()
    manifest = mod.manifest()
    journal = mod.safe_load(mod.JOURNAL)
    ready = mod.safe_load(READY)
    boot = mod.boot_id()
    if (journal.get("bootId") != boot or journal.get("stageId") != manifest["stageId"]
        or ready.get("bootId") != boot or ready.get("stageId") != manifest["stageId"]
        or ready.get("armed") is not True):
        raise ValueError("stage/ready/boot identity mismatch")
    if type(ready.get("deadlineMonotonic")) not in (float, int):
        raise ValueError("unverified monotonic deadline")
    timeout = ready["deadlineMonotonic"] + 1.25 - time.monotonic()
    if timeout > 12:
        raise ValueError("deadline is not bounded")
    if timeout > 0:
        time.sleep(timeout)
    if DONE.exists() or DONE.is_symlink():
        raise ValueError("watchdog unexpectedly recorded success after fault")
    killed = keeper if mode.startswith("keeper_") else watchdog
    main_pid = run("systemctl", "show", killed, "-p", "MainPID", "--value")
    if main_pid.returncode != 0 or main_pid.stdout.strip() != "0":
        raise ValueError("killed service is not stopped")
    expected = "VALIDATED" if mode == "keeper_precommit" else "COMMITTED_SYNTHETIC"
    if journal.get("state") != expected:
        raise ValueError("unexpected pre-read state")
    latest = mod.safe_load(mod.JOURNAL)
    if latest != journal:
        raise ValueError("journal mutated during fault observation")
    # Only the guest's specifically named synthetic private namespace may be inspected.
    table_observed = False
    if mode != "keeper_precommit":
        if (not netns or not netns.startswith("luna_stage_dst_")
            or not pid or not inode or inode != ready.get("netnsInode")):
            raise ValueError("fault target is not the expected synthetic namespace")
        result = run("ip", "netns", "exec", netns, "nft", "list", "table", "inet", "luna_p0_stage")
        table_observed = result.returncode == 0
        if not table_observed:
            raise ValueError("synthetic table unexpectedly absent; cannot claim un-recovered filter")
    else:
        if pid or netns or inode:
            raise ValueError("precommit fault must not require target for kernel write")
    record = {
        "schema": "lightluna.vm-service-fault-evidence/v1alpha1",
        "mode": mode,
        "bootId": boot,
        "stageId": manifest["stageId"],
        "expectedJournal": expected,
        "deadlineElapsed": True,
        "watchdogDonePresent": False,
        "serviceMainPIDIsZero": True,
        "syntheticRuleStillPresent": table_observed,
        "status": "FAILURE_DETECTED_NOT_RECOVERED",
        "kernelWrites": False,
        "ownerApproved": False,
        "automaticRollbackConfirmed": False,
    }
    return record


def persist(record):
    fd, temp = tempfile.mkstemp(prefix=".fault-evidence-", dir=ROOT)
    try:
        os.fchmod(fd, 0o600)
        with os.fdopen(fd, "w", encoding="utf-8") as output:
            json.dump(record, output, sort_keys=True, separators=(",", ":"))
            output.write("\n")
            output.flush()
            os.fsync(output.fileno())
        if FAULT.exists() or FAULT.is_symlink():
            raise ValueError("previous fault record exists")
        os.replace(temp, FAULT)
        directory = os.open(ROOT, os.O_RDONLY | os.O_DIRECTORY)
        try: os.fsync(directory)
        finally: os.close(directory)
    finally:
        if os.path.exists(temp):
            os.unlink(temp)


def main():
    if len(sys.argv) not in (5, 8) or sys.argv[1] != "observe":
        raise ValueError("usage: observe MODE WATCHDOG_UNIT KEEPER_UNIT [PID INODE NETNS]")
    mode, watchdog, keeper = sys.argv[2:5]
    if (mode == "keeper_precommit" and len(sys.argv) != 5
        or mode != "keeper_precommit" and len(sys.argv) != 8):
        raise ValueError("mismatched arguments for stage")
    args = sys.argv[5:]
    result = inspect(mode, watchdog, keeper, *args)
    persist(result)
    print("VM_FAULT_FAILURE_EVIDENCE_FSYNCED_NO_FALSE_SUCCESS_PASS", flush=True)


if __name__ == "__main__":
    try: main()
    except Exception as exc:
        print("VM_FAULT_EVIDENCE_DENIED", type(exc).__name__, str(exc)[:160], file=sys.stderr)
        sys.exit(3)
