#!/usr/bin/env python3
"""Explicitly archive incomplete VM-only keeper/watchdog fault journals.

This script NEVER transitions to ROLLED_BACK, never runs nft writes, never
claims recovery. It is not a production repair or boot-time reconcile service.
"""
import importlib.util
import json
import os
from pathlib import Path
import sys

ROOT = Path("/opt/luna-p0-kernel-tests")
ACK = "ARCHIVE_ONLY_UNRECOVERED_VM_FAULT"


def main():
    if os.environ.get("LUNA_VM_FAULT_ARCHIVE_ACK") != ACK:
        raise ValueError("explicit operator archival acknowledgment required")
    path = ROOT / "guest-boot-inspect.py"
    if path.is_symlink() or not path.is_file():
        raise ValueError("read-only inspector absent")
    spec = importlib.util.spec_from_file_location("vm_fault_reviewer",path)
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    mod.assert_guest()
    state_module = mod.load_stage_module()
    manifest = state_module.manifest()
    state = state_module.safe_load(state_module.JOURNAL)
    fault = state_module.safe_load(ROOT / "stage-fault-evidence.json")
    inventory = mod.inspect_kernel()
    status = mod.classify(state, manifest, state_module.boot_id(), inventory)
    if status["status"] != "CURRENT_BOOT_INTERRUPTED_REVIEW_REQUIRED":
        raise ValueError("unexpected post-cleanup inspection result")
    mode = fault.get("mode")
    expected = "VALIDATED" if mode == "keeper_precommit" else "COMMITTED_SYNTHETIC"
    if (mode not in ("keeper_precommit","keeper_postcommit","watchdog_postcommit")
        or state.get("state")!=expected
        or fault.get("expectedJournal")!=expected
        or fault.get("bootId")!=state_module.boot_id()
        or fault.get("stageId")!=manifest["stageId"]
        or fault.get("deadlineElapsed") is not True
        or fault.get("watchdogDonePresent") is not False
        or fault.get("serviceMainPIDIsZero") is not True
        or fault.get("status")!="FAILURE_DETECTED_NOT_RECOVERED"
        or fault.get("kernelWrites") is not False
        or fault.get("automaticRollbackConfirmed") is not False
        or fault.get("syntheticRuleStillPresent") is not (mode!="keeper_precommit")):
        raise ValueError("fault evidence does not match incomplete journal")
    if (ROOT / "auto-watchdog-done.json").exists():
        raise ValueError("unexpected autonomous success receipt")
    # Unique mode per VM run; cannot rewrite previously archived evidence.
    dest_journal = ROOT / ("stage-journal.unrecovered-" + mode + ".json")
    dest_fault = ROOT / ("stage-fault.unrecovered-" + mode + ".json")
    if any(p.exists() or p.is_symlink() for p in (dest_journal,dest_fault)):
        raise ValueError("prior archived evidence exists; no overwrite")
    # Evidence first, journal last. Crash interruption will remain detectable.
    os.rename(ROOT / "stage-fault-evidence.json", dest_fault)
    os.rename(state_module.JOURNAL,dest_journal)
    directory = os.open(ROOT, os.O_RDONLY|os.O_DIRECTORY)
    try: os.fsync(directory)
    finally: os.close(directory)
    if (state_module.JOURNAL.exists() or not dest_fault.is_file()
        or not dest_journal.is_file()
        or json.loads(dest_journal.read_text())["state"]!=expected):
        raise ValueError("archive did not preserve incomplete state")
    print("VM_FAULT_INCOMPLETE_JOURNAL_ARCHIVED_NO_ROLLBACK_CLAIM_PASS")


if __name__=="__main__":
    try: main()
    except Exception as exc:
        print("VM_FAULT_ARCHIVE_DENIED",type(exc).__name__,str(exc)[:180],file=sys.stderr)
        sys.exit(3)
