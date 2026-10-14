#!/usr/bin/env python3
"""Offline tests of read-only post-boot evidence classifications."""
import importlib.util
import unittest
from pathlib import Path
from unittest import mock

source = Path(__file__).with_name("guest-boot-inspect.py")
spec = importlib.util.spec_from_file_location("vm_boot_inspect", source)
m = importlib.util.module_from_spec(spec)
spec.loader.exec_module(m)
A = "11111111-2222-4333-8444-555555555555"
B = "99999999-aaaa-4bbb-8ccc-dddddddddddd"
STAGE = "vm-synthetic-expected"
MANIFEST = {"stageId": STAGE}
CLEAN = {"rootNftEntries": 0, "nonRootNetnsProcesses": 0, "foreignNetnsFdHolders": 0, "namedNetnsPresent": False, "kernelInventoryComplete": True}


def journal(boot=A, state="COMMITTED_SYNTHETIC"):
    idx = m.EXPECTED_HISTORY.index(state)
    return {"stageId": STAGE, "bootId": boot, "state": state, "history": m.EXPECTED_HISTORY[:idx + 1]}


class BootInspectTests(unittest.TestCase):
    def verify(self, expected, j, boot=B, kernel=None):
        report = m.classify(j, MANIFEST, boot, CLEAN if kernel is None else kernel)
        self.assertEqual(report["status"], expected)
        self.assertIs(report["autoReconciliationAllowed"], False)
        self.assertIs(report["kernelWrites"], False)
        self.assertIs(report["ownerApproved"], False)
        return report

    def test_reboot_stale_committed_with_no_kernel_state_requires_review(self):
        r = self.verify("CROSS_BOOT_NO_KERNEL_STATE_OBSERVED_REVIEW_REQUIRED", journal())
        self.assertTrue(r["manualReviewRequired"])

    def test_reboot_stale_terminal_still_not_auto_acknowledged(self):
        self.verify("CROSS_BOOT_NO_KERNEL_STATE_OBSERVED_REVIEW_REQUIRED", journal(state="ROLLED_BACK"))

    def test_same_boot_completed(self):
        r = self.verify("CURRENT_BOOT_TERMINAL_NO_KERNEL_STATE_OBSERVED", journal(boot=B, state="ROLLED_BACK"))
        self.assertFalse(r["manualReviewRequired"])

    def test_same_boot_committed_but_empty_kernel_requires_review(self):
        self.verify("CURRENT_BOOT_INTERRUPTED_REVIEW_REQUIRED", journal(boot=B))

    def test_kernel_root_nft_entry_blocks_acceptance(self):
        self.verify("KERNEL_STATE_PRESENT_OR_UNVERIFIED", journal(), kernel={**CLEAN,"rootNftEntries":1})

    def test_named_namespace_blocks_acceptance(self):
        self.verify("KERNEL_STATE_PRESENT_OR_UNVERIFIED", journal(), kernel={**CLEAN,"namedNetnsPresent":True})

    def test_unnamed_foreign_namespace_blocks_acceptance(self):
        self.verify("KERNEL_STATE_PRESENT_OR_UNVERIFIED", journal(), kernel={**CLEAN,"nonRootNetnsProcesses":1})

    def test_fd_held_foreign_namespace_blocks_acceptance(self):
        self.verify("KERNEL_STATE_PRESENT_OR_UNVERIFIED", journal(),
                    kernel={**CLEAN,"foreignNetnsFdHolders":1})

    def test_missing_fd_inventory_denies(self):
        missing=dict(CLEAN)
        del missing["foreignNetnsFdHolders"]
        self.verify("INVALID_KERNEL_INVENTORY",journal(),kernel=missing)

    def test_invalid_fd_inventory_boolean_denies(self):
        self.verify("INVALID_KERNEL_INVENTORY",journal(),
                    kernel={**CLEAN,"foreignNetnsFdHolders":False})

    def test_incomplete_kernel_inventory_denies(self):
        self.verify("INCOMPLETE_KERNEL_INVENTORY", journal(), kernel={**CLEAN,"kernelInventoryComplete":False})

    def test_invalid_inventory_type_denies(self):
        self.verify("INVALID_KERNEL_INVENTORY", journal(), kernel={**CLEAN,"rootNftEntries":True})

    def test_corrupt_journal_chain_denies(self):
        j = journal()
        j["history"] = ["STAGED","COMMITTED_SYNTHETIC"]
        self.verify("MALFORMED_OR_UNBOUND_JOURNAL", j)

    def test_unknown_stage_id_denies(self):
        self.verify("MALFORMED_OR_UNBOUND_JOURNAL", {**journal(),"stageId":"other"})

    def test_legacy_journal_denies(self):
        j = journal()
        del j["bootId"]
        self.verify("MALFORMED_OR_UNBOUND_JOURNAL", j)

    def test_absent_journal_not_clean(self):
        self.verify("NO_JOURNAL_UNKNOWN", None)

    def test_unknown_root_nft_json_denies(self):
        with mock.patch.object(m,"command",return_value='{"nftables":[{"table":{"family":"inet","name":"foreign"}}]}'),\
             mock.patch.object(m.os,"readlink",return_value="net:[1]"),\
             mock.patch.object(m.os,"stat",return_value=mock.Mock(st_dev=4,st_ino=1)),\
             mock.patch.object(m.Path,"iterdir",return_value=[]):
            r = m.inspect_kernel()
            self.assertEqual(r["rootNftEntries"],1)

if __name__ == "__main__":
    unittest.main(verbosity=2)
