#!/usr/bin/env python3
"""Explicit LAB-only archival of stale synthetic guest journal after inspection.

Never marks a transaction ROLLED_BACK; never touches nft/netns. This is NOT a
production boot reconciler or a trust assertion about arbitrary kernel state.
"""
import importlib.util
import json
import os
from pathlib import Path
import re
import sys
import tempfile

ROOT = Path("/opt/luna-p0-kernel-tests")
ACK = "ARCHIVE_STALE_SYNTHETIC_BOOT_EVIDENCE"


def load(name, label):
    file = ROOT / name
    if file.is_symlink() or not file.is_file() or file.stat().st_size > 4096:
        raise ValueError("invalid " + label)
    def pairs(seq):
        obj = {}
        for key, value in seq:
            if key in obj:
                raise ValueError("duplicate JSON field in " + label)
            obj[key] = value
        return obj
    with file.open(encoding="utf-8") as f:
        return json.load(f, object_pairs_hook=pairs)


def atomic_receipt(path, data):
    fd, temp = tempfile.mkstemp(prefix=".manual-inspection-", dir=ROOT)
    try:
        os.fchmod(fd, 0o600)
        with os.fdopen(fd,"w",encoding="utf-8") as handle:
            json.dump(data,handle,sort_keys=True,separators=(",",":"))
            handle.write("\n")
            handle.flush()
            os.fsync(handle.fileno())
        # O_EXCL to avoid changing or overwriting earlier evidence.
        with open(temp,"rb") as source:
            dst = os.open(path, os.O_WRONLY|os.O_CREAT|os.O_EXCL|os.O_NOFOLLOW, 0o600)
            with os.fdopen(dst,"wb") as target:
                target.write(source.read())
                target.flush()
                os.fsync(target.fileno())
    finally:
        if os.path.exists(temp): os.unlink(temp)


def main():
    if os.environ.get("LUNA_P0_ARCHIVE_ACK") != ACK:
        raise ValueError("explicit archival acknowledgment required")
    module_path = ROOT / "guest-boot-inspect.py"
    if module_path.is_symlink() or not module_path.is_file():
        raise ValueError("read-only inspector not present")
    spec = importlib.util.spec_from_file_location("vm_boot_inspect_reviewer",module_path)
    reviewer = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(reviewer)
    reviewer.assert_guest()
    mod = reviewer.load_stage_module()
    plan = mod.manifest()
    record = mod.safe_load(mod.JOURNAL)
    observed = reviewer.inspect_kernel()
    report = reviewer.classify(record,plan,mod.boot_id(),observed)
    if report["status"] != "CROSS_BOOT_NO_KERNEL_STATE_OBSERVED_REVIEW_REQUIRED":
        raise ValueError("inspection does not establish required stale-boot fixture state")
    before = load("boot-fault-pre.json", "preboot evidence")
    if (before.get("schema")!="lightluna.vm-boot-fault-evidence/v1alpha1"
        or before.get("preBootId")!=record["bootId"]
        or before.get("stageId")!=record["stageId"]
        or before.get("journalState")!="COMMITTED_SYNTHETIC"
        or record.get("state")!="COMMITTED_SYNTHETIC"
        or before.get("icmpDropCounterPositive") is not True
        or before.get("watchdogArmed") is not False
        or record["bootId"]==mod.boot_id()):
        raise ValueError("stale boot fixture evidence mismatch")
    old_boot = record["bootId"]
    if not re.fullmatch(r"[a-f0-9-]{36}",old_boot):
        raise ValueError("invalid prior boot id")
    archive = ROOT / ("stage-journal.cross-boot-unreconciled-" + old_boot + ".json")
    receipt = ROOT / ("manual-cross-boot-review-" + old_boot + ".json")
    if archive.exists() or archive.is_symlink() or receipt.exists() or receipt.is_symlink():
        raise ValueError("pre-existing evidence; refusing overwrite")
    audit = {
        "schema":"lightluna.vm-manual-archive/v1alpha1",
        "sourceBootId":old_boot,
        "inspectedBootId":mod.boot_id(),
        "stageId":record["stageId"],
        "sourceJournalState":"COMMITTED_SYNTHETIC",
        "outcome":"ARCHIVED_STALE_SYNTHETIC_JOURNAL_WITHOUT_ROLLBACK_CLAIM",
        "scope":"EXACT_NO_NIC_KVM_ONLY",
        "observedKernel":observed,
        "noKernelWrites":True,
        "autoReconciliationAllowed":False,
        "notEvidenceOfProductionRecovery":True,
    }
    # Archive before claiming archival in a receipt. A crash after rename but
    # before receipt leaves preserved evidence for manual review, not a false
    # success receipt for a journal which never moved.
    os.rename(mod.JOURNAL,archive)
    directory = os.open(ROOT,os.O_RDONLY|os.O_DIRECTORY)
    try: os.fsync(directory)
    finally: os.close(directory)
    if archive.read_bytes() == b"" or mod.JOURNAL.exists():
        raise ValueError("journal archival postcondition failed")
    atomic_receipt(receipt,audit)
    print("VM_MANUAL_STALE_JOURNAL_ARCHIVED_WITHOUT_ROLLBACK_CLAIM_PASS")


if __name__ == "__main__":
    try: main()
    except (OSError,ValueError,KeyError,TypeError,IndexError,json.JSONDecodeError) as exc:
        print("VM_MANUAL_ARCHIVE_DENIED:",type(exc).__name__,file=sys.stderr)
        sys.exit(3)
