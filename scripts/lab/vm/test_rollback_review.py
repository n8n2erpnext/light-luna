#!/usr/bin/env python3
"""Offline fail-closed review matrix, no privileges, kernel or external network."""
import importlib.util
from pathlib import Path
import unittest

source=Path(__file__).with_name("guest-rollback-review.py")
spec=importlib.util.spec_from_file_location("vm_rollback_review",source)
mod=importlib.util.module_from_spec(spec);spec.loader.exec_module(mod)

STAGE="vm-"+"a"*24
BOOT="11111111-2222-4333-8444-555555555555"
NETNS="net:[12345678]"
EPOCH="a"*24
INVENTORY={"kernelInventoryComplete":True,"rootNftEntries":0,
           "namedNetnsPresent":False,"nonRootNetnsProcesses":0}
READY={"stageId":STAGE,"bootId":BOOT,"armed":True,"netnsInode":NETNS}
JOURNAL={"stageId":STAGE,"bootId":BOOT,"state":"ROLLED_BACK",
         "history":["STAGED","VALIDATED","COMMITTED_SYNTHETIC","ROLLED_BACK"]}
RECEIPT={"status":"JOURNAL_TERMINAL_RECEIPT_PRESENT_DONE_NOT_EVALUATED",
         "epoch":EPOCH,"writer":"fallback","ownerApproved":False,
         "automaticRecoveryAuthorized":False}
DONE={"status":"SYNTHETIC_FALLBACK_ROLLBACK_CONFIRMED",
      "stageId":STAGE,"bootId":BOOT,"rollbackObservationEpoch":EPOCH,
      "rollbackFenced":True,"netnsInode":NETNS,
      "syntheticGuestOnly":True,"ownerApproved":False}
ARCHIVED_REVIEW={"status":"JOURNAL_TERMINAL_DONE_MISSING_POST_CLEANUP_REVIEW",
                 "rollbackObservationEpoch":EPOCH,"bootId":BOOT,"stageId":STAGE,
                 "successDoneObserved":False,
                 "reviewTiming":"AFTER_TEST_EXIT_CLEANUP_NOT_PRE_CLEANUP_EVIDENCE"}

class ReadOnlyReviewTests(unittest.TestCase):
    def classify(self, **changes):
        data={"journal":JOURNAL,"ready":READY,"current_boot":BOOT,
              "inventory":INVENTORY,"receipt":RECEIPT,"done":DONE}
        data.update(changes)
        return mod.classify(**data)
    def status(self, expected, **changes):
        result=self.classify(**changes)
        self.assertEqual(result["status"], expected)
        self.assertIs(result["autoReconciliationAllowed"],False)
        self.assertIs(result["kernelWrites"],False)
        self.assertIs(result["ownerApproved"],False)
        self.assertIs(result["successAcknowledged"],False)
    def test_consistent_guest_done_still_not_authority(self):
        self.status("SYNTHETIC_OBSERVATION_CHAIN_CONSISTENT_NOT_AUTHORITY")
    def test_terminal_missing_done_requires_review(self):
        self.status("TERMINAL_JOURNAL_DONE_MISSING",done=None)
    def test_cross_boot_never_authorizes_terminal(self):
        self.status("CROSS_BOOT_RECEIPT_REVIEW_REQUIRED",
                    current_boot="aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee")
    def test_prepared_only_kernel_unknown_even_with_terminal_journal(self):
        self.status("PREPARED_KERNEL_OUTCOME_UNKNOWN",
                    receipt=dict(RECEIPT,status="PREPARED_NO_KERNEL_ABSENCE_PROOF"),done=None)
    def test_kernel_absent_receipt_but_committed_journal(self):
        self.status("KERNEL_ABSENT_JOURNAL_INCOMPLETE",
                    receipt=dict(RECEIPT,status="KERNEL_ABSENCE_OBSERVED_JOURNAL_NOT_TERMINAL_AT_RECEIPT"),
                    journal=dict(JOURNAL,state="COMMITTED_SYNTHETIC",
                                 history=JOURNAL["history"][:3]),done=None)
    def test_kernel_absent_receipt_but_terminal_journal(self):
        self.status("JOURNAL_TERMINAL_OBSERVATION_MISSING",
                    receipt=dict(RECEIPT,status="KERNEL_ABSENCE_OBSERVED_JOURNAL_NOT_TERMINAL_AT_RECEIPT"),done=None)
    def test_terminal_receipt_committed_journal_contradiction(self):
        self.status("TERMINAL_RECEIPT_JOURNAL_CONTRADICTION",
                    journal=dict(JOURNAL,state="COMMITTED_SYNTHETIC",
                                 history=JOURNAL["history"][:3]),done=None)
    def test_corrupted_receipt_denies(self):
        self.status("RECEIPT_CORRUPTED_FAIL_CLOSED",
                    receipt={"status":"INVALID_RECEIPT_FAIL_CLOSED"},done=None)
    def test_done_without_receipts_denies(self):
        self.status("DONE_WITHOUT_RECEIPT_CHAIN_INVALID",
                    receipt={"status":"NO_ROLLBACK_RECEIPT"})
    def test_no_receipts_no_done(self):
        self.status("NO_ROLLBACK_OBSERVATIONS_REVIEW_REQUIRED",
                    receipt={"status":"NO_ROLLBACK_RECEIPT"},done=None)
    def test_wrong_done_epoch(self):
        self.status("DONE_METADATA_MISMATCH_REVIEW_REQUIRED",
                    done=dict(DONE,rollbackObservationEpoch="b"*24))
    def test_wrong_done_writer(self):
        self.status("DONE_STATUS_OR_WRITER_CONTRADICTION",
                    done=dict(DONE,status="AUTO_ROLLBACK_CONFIRMED"))
    def test_primary_matching_done(self):
        self.status("SYNTHETIC_OBSERVATION_CHAIN_CONSISTENT_NOT_AUTHORITY",
                    receipt=dict(RECEIPT,writer="primary"),
                    done=dict(DONE,status="AUTO_ROLLBACK_CONFIRMED"))
    def test_bad_history_denied(self):
        self.status("MALFORMED_JOURNAL",journal=dict(JOURNAL,history=["STAGED","ROLLED_BACK"]))
    def test_invalid_journal_state_denied(self):
        self.status("MALFORMED_JOURNAL",journal=dict(JOURNAL,state="COMMITTED_SYNTHETIC"))
    def test_ready_stage_mismatch(self):
        self.status("JOURNAL_READY_IDENTITY_MISMATCH",ready=dict(READY,stageId="vm-"+"b"*24))
    def test_incomplete_inventory(self):
        self.status("INCOMPLETE_OR_INVALID_KERNEL_INVENTORY",
                    inventory=dict(INVENTORY,kernelInventoryComplete=False))
    def test_kernel_tables_present(self):
        self.status("KERNEL_STATE_PRESENT_OR_UNKNOWN",inventory=dict(INVENTORY,rootNftEntries=1))
    def test_foreign_process_namespace_present(self):
        self.status("KERNEL_STATE_PRESENT_OR_UNKNOWN",
                    inventory=dict(INVENTORY,nonRootNetnsProcesses=1))
    def test_missing_ready(self):
        self.status("MISSING_OR_UNREADABLE_JOURNAL_READY",ready=None)
    def test_archived_terminal_with_proven_prior_missing_done(self):
        self.status("ARCHIVED_TERMINAL_DONE_MISSING_AT_PRIOR_REVIEW",
                    archived=True,done=None,archive_review=ARCHIVED_REVIEW)
    def test_archived_without_matching_provenance(self):
        self.status("ARCHIVED_REVIEW_EVIDENCE_MISSING_OR_MISMATCH",
                    archived=True,done=None,
                    archive_review=dict(ARCHIVED_REVIEW,rollbackObservationEpoch="b"*24))
    def test_archived_crossboot_priority(self):
        self.status("CROSS_BOOT_RECEIPT_REVIEW_REQUIRED",
                    archived=True,done=None,archive_review=ARCHIVED_REVIEW,
                    current_boot="aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee")

if __name__=="__main__":
    unittest.main(verbosity=2)
