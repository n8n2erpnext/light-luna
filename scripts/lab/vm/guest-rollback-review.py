#!/usr/bin/env python3
"""Read-only VM synthetic rollback receipt / journal review, NOT recovery.

Reports ambiguity after crashes/reboots; never changes kernel, journal,
receipt, service, or device state. No output status authorizes enforcement.
"""
import importlib.util
import json
import os
from pathlib import Path
import re
import sys

ROOT = Path("/opt/luna-p0-kernel-tests")
PHASES = ["STAGED", "VALIDATED", "COMMITTED_SYNTHETIC", "ROLLED_BACK"]
RECEIPT_STATUSES = {
    "NO_ROLLBACK_RECEIPT",
    "PREPARED_NO_KERNEL_ABSENCE_PROOF",
    "KERNEL_ABSENCE_OBSERVED_JOURNAL_NOT_TERMINAL_AT_RECEIPT",
    "JOURNAL_TERMINAL_RECEIPT_PRESENT_DONE_NOT_EVALUATED",
    "INVALID_RECEIPT_FAIL_CLOSED",
}
SCHEMA = "lightluna.vm-rollback-review/v1alpha1"


def classify(journal, ready, current_boot, inventory, receipt, done=None, *,
             archived=False, archive_review=None):
    """Pure fail-closed classification; never confers kernel or journal authority."""
    result = {
        "schema": SCHEMA, "mode": "READ_ONLY_VM_SYNTHETIC_ROLLBACK_REVIEW",
        "status": "UNVERIFIED_REVIEW_REQUIRED", "currentBootId": current_boot,
        "recordBootId": None, "journalState": None, "receiptStatus": None,
        "receiptEpoch": None, "receiptWriter": None, "archived": bool(archived),
        "manualReviewRequired": True, "kernelWrites": False,
        "autoReconciliationAllowed": False, "ownerApproved": False,
        "signedOwnerReceipt": False, "successAcknowledged": False,
    }
    if (not isinstance(inventory, dict)
        or inventory.get("kernelInventoryComplete") is not True
        or type(inventory.get("rootNftEntries")) is not int
        or type(inventory.get("nonRootNetnsProcesses")) is not int
        or type(inventory.get("namedNetnsPresent")) is not bool
        or inventory["rootNftEntries"] < 0
        or inventory["nonRootNetnsProcesses"] < 0):
        result["status"] = "INCOMPLETE_OR_INVALID_KERNEL_INVENTORY"
        return result
    if (inventory["rootNftEntries"] or inventory["nonRootNetnsProcesses"]
        or inventory["namedNetnsPresent"]):
        result["status"] = "KERNEL_STATE_PRESENT_OR_UNKNOWN"
        return result
    if not isinstance(journal, dict) or not isinstance(ready, dict):
        result["status"] = "MISSING_OR_UNREADABLE_JOURNAL_READY"
        return result
    if (set(journal) != {"stageId", "bootId", "state", "history"}
        or not isinstance(journal.get("history"), list)
        or not journal["history"]
        or journal["history"] != PHASES[:len(journal["history"])]
        or journal["state"] != journal["history"][-1]):
        result["status"] = "MALFORMED_JOURNAL"
        return result
    stage = journal["stageId"]
    boot = journal["bootId"]
    result["recordBootId"] = boot
    result["journalState"] = journal["state"]
    if (type(stage) is not str or not re.fullmatch(r"vm-[0-9a-f]{24}", stage)
        or type(boot) is not str or not re.fullmatch(
            r"[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}", boot)
        or ready.get("stageId") != stage or ready.get("bootId") != boot
        or ready.get("armed") is not True
        or type(ready.get("netnsInode")) is not str
        or not re.fullmatch(r"net:\[[0-9]+\]", ready["netnsInode"])):
        result["status"] = "JOURNAL_READY_IDENTITY_MISMATCH"
        return result
    if not isinstance(receipt, dict) or receipt.get("status") not in RECEIPT_STATUSES:
        result["status"] = "RECEIPT_UNVERIFIABLE"
        return result
    result["receiptStatus"] = receipt["status"]
    epoch = receipt.get("epoch")
    result["receiptEpoch"] = epoch
    if receipt["status"] == "INVALID_RECEIPT_FAIL_CLOSED":
        result["status"] = "RECEIPT_CORRUPTED_FAIL_CLOSED"
        return result
    if receipt["status"] == "NO_ROLLBACK_RECEIPT":
        if done is not None:
            result["status"] = "DONE_WITHOUT_RECEIPT_CHAIN_INVALID"
        else:
            result["status"] = "NO_ROLLBACK_OBSERVATIONS_REVIEW_REQUIRED"
        return result
    if (type(epoch) is not str or not re.fullmatch(r"[0-9a-f]{24}", epoch)
        or receipt.get("writer") not in ("primary","fallback")
        or receipt.get("ownerApproved") is not False
        or receipt.get("automaticRecoveryAuthorized") is not False):
        result["status"] = "INVALID_RECEIPT_METADATA"
        return result
    result["receiptWriter"] = receipt["writer"]
    if type(current_boot) is not str or boot != current_boot:
        result["status"] = "CROSS_BOOT_RECEIPT_REVIEW_REQUIRED"
        return result
    if archived:
        if (not isinstance(archive_review, dict)
            or archive_review.get("status") != "JOURNAL_TERMINAL_DONE_MISSING_POST_CLEANUP_REVIEW"
            or archive_review.get("rollbackObservationEpoch") != epoch
            or archive_review.get("bootId") != boot
            or archive_review.get("stageId") != stage
            or archive_review.get("successDoneObserved") is not False
            or archive_review.get("reviewTiming") != "AFTER_TEST_EXIT_CLEANUP_NOT_PRE_CLEANUP_EVIDENCE"):
            result["status"] = "ARCHIVED_REVIEW_EVIDENCE_MISSING_OR_MISMATCH"
        else:
            result["status"] = "ARCHIVED_TERMINAL_DONE_MISSING_AT_PRIOR_REVIEW"
        return result
    phases = receipt["status"]
    if phases == "PREPARED_NO_KERNEL_ABSENCE_PROOF":
        result["status"] = "PREPARED_KERNEL_OUTCOME_UNKNOWN"
        return result
    if phases == "KERNEL_ABSENCE_OBSERVED_JOURNAL_NOT_TERMINAL_AT_RECEIPT":
        result["status"] = ("JOURNAL_TERMINAL_OBSERVATION_MISSING" if journal["state"]=="ROLLED_BACK"
                            else "KERNEL_ABSENT_JOURNAL_INCOMPLETE")
        return result
    if phases != "JOURNAL_TERMINAL_RECEIPT_PRESENT_DONE_NOT_EVALUATED":
        result["status"] = "UNRECOGNIZED_RECEIPT_PHASE"
        return result
    if journal["state"] != "ROLLED_BACK":
        result["status"] = "TERMINAL_RECEIPT_JOURNAL_CONTRADICTION"
        return result
    if done is None:
        result["status"] = "TERMINAL_JOURNAL_DONE_MISSING"
        return result
    if (not isinstance(done, dict)
        or done.get("stageId") != stage or done.get("bootId") != boot
        or done.get("rollbackObservationEpoch") != epoch
        or done.get("rollbackFenced") is not True):
        result["status"] = "DONE_METADATA_MISMATCH_REVIEW_REQUIRED"
        return result
    if receipt["writer"] == "primary":
        good = done.get("status") == "AUTO_ROLLBACK_CONFIRMED"
    else:
        good = (done.get("status") == "SYNTHETIC_FALLBACK_ROLLBACK_CONFIRMED"
                and done.get("netnsInode") == ready["netnsInode"]
                and done.get("syntheticGuestOnly") is True
                and done.get("ownerApproved") is False)
    result["status"] = ("SYNTHETIC_OBSERVATION_CHAIN_CONSISTENT_NOT_AUTHORITY"
                        if good else "DONE_STATUS_OR_WRITER_CONTRADICTION")
    # Even a consistent lab DONE does not authorize host kernel actions or
    # confer cryptographic authenticity / power-loss durability.
    return result


def load_guest(filename, label):
    path = ROOT / filename
    if path.is_symlink() or not path.is_file():
        raise ValueError("VM-only source module missing")
    spec = importlib.util.spec_from_file_location(label, path)
    obj = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(obj)
    return obj


def main(argv):
    if len(argv) not in (0, 2) or (argv and argv[0] != "--archived-terminal"):
        raise ValueError("usage: [--archived-terminal unique-tag]")
    if argv and not re.fullmatch(r"[a-z0-9-]{1,32}", argv[1]):
        raise ValueError("invalid archive tag")
    inspector = load_guest("guest-boot-inspect.py","vm_read_only_kernel_inspector")
    inspector.assert_guest()
    journalmod = inspector.load_stage_module()
    receipts = load_guest("guest-rollback-receipts.py","vm_read_only_observations")
    current_boot = journalmod.boot_id()
    kernel = inspector.inspect_kernel()
    tag = argv[1] if argv else None
    if tag:
        jp = ROOT / ("stage-journal.receipt-missing-fallback_after_terminal."+tag+".json")
        rp = ROOT / ("auto-watchdog-ready.receipt-missing."+tag+".json")
        ep = ROOT / ("terminal-review."+tag+".json")
        if any(p.is_symlink() for p in (jp,rp,ep)):
            raise ValueError("symlinked archive source")
        journal = journalmod.safe_load(jp)
        ready = journalmod.safe_load(rp)
        prior = journalmod.safe_load(ep)
        done = None   # Past absence inferred ONLY from prior review record.
    else:
        jp = journalmod.JOURNAL
        rp = ROOT / "auto-watchdog-ready.json"
        journal = journalmod.safe_load(jp) if jp.exists() else None
        ready = journalmod.safe_load(rp) if rp.exists() else None
        prior = None
        done = None
        for p in (ROOT/"auto-watchdog-done.json", ROOT/"fallback-done.json"):
            if p.is_symlink():
                raise ValueError("symlinked DONE")
            if p.exists():
                if done is not None: raise ValueError("conflicting dual DONE")
                done = journalmod.safe_load(p)
    if ready is None:
        observation = {"status":"MISSING_READY"}
    else:
        try:
            observation = receipts.inspect(journal["stageId"],journal["bootId"],
                                            ready["netnsInode"],ready)
        except (ValueError,OSError,KeyError,TypeError):
            observation = {"status":"INVALID_RECEIPT_FAIL_CLOSED"}
    report = classify(journal,ready,current_boot,kernel,observation,done,
                      archived=bool(tag),archive_review=prior)
    print(json.dumps(report,sort_keys=True,indent=2))
    return 0 if report["status"]=="SYNTHETIC_OBSERVATION_CHAIN_CONSISTENT_NOT_AUTHORITY" else 3


if __name__ == "__main__":
    try:
        sys.exit(main(sys.argv[1:]))
    except (OSError,ValueError,KeyError,TypeError,json.JSONDecodeError) as err:
        print(json.dumps({"schema":SCHEMA,"status":"INVENTORY_OR_EVIDENCE_ERROR_FAIL_CLOSED",
            "errorClass":type(err).__name__,"kernelWrites":False,
            "successAcknowledged":False,"autoReconciliationAllowed":False},sort_keys=True))
        sys.exit(3)
