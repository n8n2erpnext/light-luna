#!/usr/bin/env python3
"""Explicit archival after post-cleanup discovery of VM-only terminal/no-DONE.

Unlike the live fault observer, this cannot prove absence before EXIT cleanup.
It preserves evidence without inventing a missing success receipt.
"""
import importlib.util
import json
import os
from pathlib import Path
import re
import sys

ROOT=Path("/opt/luna-p0-kernel-tests")
ACK="PRESERVE_ONLY_VM_TERMINAL_NO_DONE_REVIEW"


def import_guest(name, label):
    p=ROOT/name
    if p.is_symlink() or not p.is_file():raise ValueError("missing guest "+label)
    spec=importlib.util.spec_from_file_location(label,p)
    m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m)
    return m


def preserve(tag):
    if os.environ.get("LUNA_VM_TERMINAL_REVIEW_ACK")!=ACK:
        raise ValueError("explicit owner/operator lab archive acknowledgment missing")
    if not re.fullmatch(r"[a-z0-9-]{1,32}",tag):
        raise ValueError("unsafe immutable evidence tag")
    inspector=import_guest("guest-boot-inspect.py","guest_terminal_postcleanup_inspector")
    inspector.assert_guest()
    mod=inspector.load_stage_module()
    manifest=mod.manifest()
    state=mod.safe_load(mod.JOURNAL)
    boot=mod.boot_id()
    if state.get("state")!="ROLLED_BACK" or state.get("bootId")!=boot:
        raise ValueError("terminal current-boot stage required")
    kernel=inspector.inspect_kernel()
    classification=inspector.classify(state,manifest,boot,kernel)
    if classification["status"]!="CURRENT_BOOT_TERMINAL_NO_KERNEL_STATE_OBSERVED":
        raise ValueError("post-cleanup kernel inventory not verified")
    if any((ROOT/name).exists() or (ROOT/name).is_symlink()
           for name in ("auto-watchdog-done.json","fallback-done.json","stage-fault-evidence.json")):
        raise ValueError("unexpected successful DONE or previous fault evidence")
    receipts=import_guest("guest-rollback-receipts.py","guest_terminal_postcleanup_receipts")
    ready=mod.safe_load(ROOT/"auto-watchdog-ready.json")
    state_receipts=receipts.inspect(manifest["stageId"],boot,ready["netnsInode"],ready)
    if state_receipts["status"]!="JOURNAL_TERMINAL_RECEIPT_PRESENT_DONE_NOT_EVALUATED" or state_receipts["writer"]!="fallback":
        raise ValueError("immutable observation receipts missing or invalid")
    destinations={
        ROOT/"stage-journal.json": ROOT/("stage-journal.receipt-missing-fallback_after_terminal."+tag+".json"),
        ROOT/"auto-watchdog-ready.json": ROOT/("auto-watchdog-ready.receipt-missing."+tag+".json"),
        ROOT/"fallback-ready.json": ROOT/("fallback-ready.receipt-missing."+tag+".json"),
    }
    evidence=ROOT/("terminal-review."+tag+".json")
    if evidence.exists() or evidence.is_symlink():
        raise ValueError("previous review evidence must not be overwritten")
    for source,dest in destinations.items():
        if source.is_symlink() or not source.is_file() or dest.exists() or dest.is_symlink():
            raise ValueError("existing/missing archive source or destination")
    # Explicitly document timing limitation: this inventory was taken AFTER
    # original fault runner exited and cleaned synthetic guest namespaces.
    report={"schema":"lightluna.vm-terminal-review/v1alpha1",
        "status":"JOURNAL_TERMINAL_DONE_MISSING_POST_CLEANUP_REVIEW",
        "bootId":boot,"stageId":manifest["stageId"],
        "rollbackObservationEpoch":state_receipts["epoch"],
        "sourceJournalState":"ROLLED_BACK","successDoneObserved":False,
        "reviewTiming":"AFTER_TEST_EXIT_CLEANUP_NOT_PRE_CLEANUP_EVIDENCE",
        "kernelInventory":kernel,"ownerApproved":False,
        "autoRecoveryAuthorized":False,"nftWrites":False,
        "signedReceipt":False}
    # Preserve journal and original READY bytes with atomic same-filesystem
    # renames, no marker claiming success, no kernel mutation.
    for src,dest in destinations.items():
        os.rename(src,dest)
    fd=os.open(evidence,os.O_WRONLY|os.O_CREAT|os.O_EXCL|os.O_NOFOLLOW,0o600)
    with os.fdopen(fd,"w") as out:
        json.dump(report,out,sort_keys=True,separators=(",",":"))
        out.write("\n");out.flush();os.fsync(out.fileno())
    directory=os.open(ROOT,os.O_RDONLY|os.O_DIRECTORY)
    try:os.fsync(directory)
    finally:os.close(directory)
    print("VM_TERMINAL_WITHOUT_DONE_POST_CLEANUP_EVIDENCE_ARCHIVED_NO_FALSE_CLAIM_PASS")


if __name__=="__main__":
    try:
        if len(sys.argv)!=2:raise ValueError("usage: tag")
        preserve(sys.argv[1])
    except (ValueError,OSError,KeyError,TypeError,json.JSONDecodeError) as error:
        print("VM_TERMINAL_REVIEW_DENIED",type(error).__name__,str(error)[:180],file=sys.stderr)
        sys.exit(3)
