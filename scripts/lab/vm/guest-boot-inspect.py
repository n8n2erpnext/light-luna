#!/usr/bin/env python3
"""P0 read-only post-boot synthetic VM journal/kernel inspector.

This is evidence collection, never a reboot reconciler, executor, cleanup,
approval or assertion that a real host firewall is recovered.
"""
import importlib.util
import json
import os
from pathlib import Path
import subprocess
import sys

ROOT = Path("/opt/luna-p0-kernel-tests")
STAGE_SOURCE = ROOT / "guest-stage-journal.py"
EXPECTED_HISTORY = ["STAGED", "VALIDATED", "COMMITTED_SYNTHETIC", "ROLLED_BACK"]
MAX_OUTPUT = 65536


def command(*args):
    proc = subprocess.run(args, capture_output=True, text=True, timeout=5, check=False)
    if proc.returncode != 0 or len(proc.stdout) > MAX_OUTPUT:
        raise ValueError("read-only inventory command failed or output exceeded cap: " + args[0])
    return proc.stdout


def assert_guest():
    if os.geteuid() != 0:
        raise ValueError("guest root needed for complete read-only namespace inventory")
    if not Path("/etc/light-luna-p0-kernel-vm").is_file():
        raise ValueError("missing dedicated VM identity marker")
    if ROOT.is_symlink() or not ROOT.is_dir():
        raise ValueError("unsafe lab root")
    if command("systemd-detect-virt").strip() != "kvm":
        raise ValueError("dedicated KVM VM required")
    if sorted(os.listdir("/sys/class/net")) != ["lo"]:
        raise ValueError("VM guest has an external or unexpected interface")


def inspect_kernel():
    # Root nft ruleset inventory is NOT equivalent to inspecting all netns:
    # reject unless every visible process shares PID 1's netns and named netns
    # inventory is empty. All commands are read-only and bounded.
    payload = json.loads(command("nft", "-j", "list", "ruleset"))
    if set(payload) != {"nftables"} or not isinstance(payload["nftables"], list):
        raise ValueError("unrecognized nft JSON inventory")
    root_entries = [entry for entry in payload["nftables"] if set(entry) != {"metainfo"}]
    named_netns = command("ip", "netns", "list").strip()
    baseline = os.readlink("/proc/1/ns/net")
    foreign = 0
    for child in Path("/proc").iterdir():
        if not child.name.isdecimal():
            continue
        try:
            current = os.readlink(str(child / "ns/net"))
        except FileNotFoundError:
            continue  # raced with a completed process
        if current != baseline:
            foreign += 1
    return {
        "rootNftEntries": len(root_entries),
        "namedNetnsPresent": bool(named_netns),
        "nonRootNetnsProcesses": foreign,
        "kernelInventoryComplete": True,
    }


def load_stage_module():
    if STAGE_SOURCE.is_symlink() or not STAGE_SOURCE.is_file():
        raise ValueError("missing or symlinked VM stage module")
    spec = importlib.util.spec_from_file_location("luna_stage_boot_inventory", STAGE_SOURCE)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def classify(journal, manifest, current_boot, inventory):
    """Pure classification. Explicitly incapable of authorizing recovery."""
    result = {
        "schema": "lightluna.vm-boot-inspection/v1alpha1",
        "mode": "READ_ONLY_VM_SYNTHETIC_EVIDENCE",
        "currentBootId": current_boot,
        "journalBootId": None,
        "journalState": None,
        "kernel": inventory,
        "autoReconciliationAllowed": False,
        "kernelWrites": False,
        "ownerApproved": False,
        "manualReviewRequired": True,
        "status": "UNVERIFIED",
    }
    if inventory.get("kernelInventoryComplete") is not True:
        result["status"] = "INCOMPLETE_KERNEL_INVENTORY"
        return result
    if (type(inventory.get("rootNftEntries")) is not int
        or type(inventory.get("nonRootNetnsProcesses")) is not int
        or type(inventory.get("namedNetnsPresent")) is not bool
        or inventory["rootNftEntries"] < 0 or inventory["nonRootNetnsProcesses"] < 0):
        result["status"] = "INVALID_KERNEL_INVENTORY"
        return result
    if (inventory["rootNftEntries"] > 0 or inventory["nonRootNetnsProcesses"] > 0
        or inventory["namedNetnsPresent"]):
        result["status"] = "KERNEL_STATE_PRESENT_OR_UNVERIFIED"
        return result
    if journal is None:
        result["status"] = "NO_JOURNAL_UNKNOWN"
        return result
    if (not isinstance(manifest, dict)
        or type(journal) is not dict
        or set(journal) != {"stageId", "bootId", "state", "history"}
        or journal.get("stageId") != manifest.get("stageId")
        or type(journal["bootId"]) is not str
        or type(journal["history"]) is not list
        or not journal["history"]
        or journal["history"] != EXPECTED_HISTORY[:len(journal["history"])]
        or journal.get("state") != journal["history"][-1]):
        result["status"] = "MALFORMED_OR_UNBOUND_JOURNAL"
        return result
    result["journalBootId"] = journal["bootId"]
    result["journalState"] = journal["state"]
    if journal["bootId"] != current_boot:
        result["status"] = "CROSS_BOOT_NO_KERNEL_STATE_OBSERVED_REVIEW_REQUIRED"
    elif journal["state"] == "ROLLED_BACK":
        result["status"] = "CURRENT_BOOT_TERMINAL_NO_KERNEL_STATE_OBSERVED"
        result["manualReviewRequired"] = False  # no recovery action proposed
    else:
        result["status"] = "CURRENT_BOOT_INTERRUPTED_REVIEW_REQUIRED"
    return result


def main():
    assert_guest()
    mod = load_stage_module()
    current_boot = mod.boot_id()
    inventory = inspect_kernel()
    # If the stage manifest is corrupt, do not infer safety from an apparent
    # absence of nft/netns. Only return bounded honest failure evidence.
    manifest = mod.manifest()
    try:
        journal = mod.safe_load(mod.JOURNAL) if mod.JOURNAL.exists() else None
    except (ValueError, OSError, json.JSONDecodeError) as error:
        raise ValueError("stage journal could not be verified") from error
    report = classify(journal, manifest, current_boot, inventory)
    print(json.dumps(report, sort_keys=True, indent=2))
    return 0 if report["status"] == "CURRENT_BOOT_TERMINAL_NO_KERNEL_STATE_OBSERVED" else 3


if __name__ == "__main__":
    try:
        sys.exit(main())
    except (OSError, ValueError, subprocess.TimeoutExpired, json.JSONDecodeError) as exc:
        print(json.dumps({
            "schema": "lightluna.vm-boot-inspection/v1alpha1",
            "mode": "READ_ONLY_VM_SYNTHETIC_EVIDENCE",
            "status": "INVENTORY_ERROR_FAIL_CLOSED",
            "errorClass": type(exc).__name__,
            "autoReconciliationAllowed": False,
            "kernelWrites": False,
        }, sort_keys=True))
        sys.exit(3)
