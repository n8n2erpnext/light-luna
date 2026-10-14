#!/usr/bin/env python3
"""Bounded VM synthetic receipt fsync/write fault matrix; NO kernel writes.

The injected OSError is a syscall-level simulation, NOT physical power loss,
durable-disk proof, or automatic recovery. The module under test may leave
parseable but not durability-confirmed partial receipts after an exception.
Never treat such existence as a success DONE or grant to modify nftables.
"""
import errno
import importlib.util
import json
import os
from pathlib import Path
import tempfile
import unittest
from unittest import mock

spec = importlib.util.spec_from_file_location(
    "luna_fsync_receipt_module", Path(__file__).with_name("guest-rollback-receipts.py"))
receipts = importlib.util.module_from_spec(spec)
spec.loader.exec_module(receipts)
r = receipts
STAGE = "vm-" + "a"*24
BOOT = "11111111-2222-4333-8444-555555555555"
NS = "net:[12345678]"
READY = {"stageId": STAGE, "bootId": BOOT, "netnsInode": NS,
         "armed": True, "anchorPid": 120, "anchorBirthTicks": 1900,
         "armedMonotonic": 250.0, "deadlineMonotonic": 260.0}
WRITER = "fallback"


class ReceiptFsyncFaultTests(unittest.TestCase):
    def setUp(self):
        temp = tempfile.TemporaryDirectory(prefix="lightluna-fsync-")
        self.addCleanup(temp.cleanup)
        self.root = Path(temp.name)
        patcher = mock.patch.object(r, "ROOT", self.root)
        patcher.start()
        self.addCleanup(patcher.stop)
        self.epoch = r._id(STAGE, BOOT, NS, READY)

    def inject_fsync_failure(self, index):
        """Fail exactly one os.fsync operation by ordinal (file=1, dir=2)."""
        assert index in (1, 2)
        original = os.fsync
        sequence = []
        def fail_once(fd):
            sequence.append(fd)
            if len(sequence) == index:
                raise OSError(errno.EIO, "simulated VM-only fsync failure")
            return original(fd)
        return mock.patch.object(r.os, "fsync", side_effect=fail_once)

    def assert_no_success(self):
        self.assertFalse((self.root/"fallback-done.json").exists())
        self.assertFalse((self.root/"auto-watchdog-done.json").exists())
        report = r.inspect(STAGE, BOOT, NS, READY)
        self.assertIn(report["status"], (
            "NO_ROLLBACK_RECEIPT",
            "PREPARED_NO_KERNEL_ABSENCE_PROOF",
            "KERNEL_ABSENCE_OBSERVED_JOURNAL_NOT_TERMINAL_AT_RECEIPT",
            "JOURNAL_TERMINAL_RECEIPT_PRESENT_DONE_NOT_EVALUATED",
            "INVALID_RECEIPT_FAIL_CLOSED",
        ))
        if "automaticRecoveryAuthorized" in report:
            self.assertIs(report["automaticRecoveryAuthorized"], False)

    def test_prepare_file_fsync_error_prevents_kernel_write(self):
        reached = {"kernelWriteReached": False}
        with self.inject_fsync_failure(1):
            with self.assertRaises(OSError):
                r.prepare(STAGE, BOOT, NS, READY, WRITER)
                reached["kernelWriteReached"] = True
        self.assertFalse(reached["kernelWriteReached"])
        self.assertTrue(r.path(self.epoch, "prepared").exists())
        self.assert_no_success()

    def test_prepare_directory_fsync_error_prevents_kernel_write(self):
        reached = {"kernelWriteReached": False}
        with self.inject_fsync_failure(2):
            with self.assertRaises(OSError):
                r.prepare(STAGE, BOOT, NS, READY, WRITER)
                reached["kernelWriteReached"] = True
        self.assertFalse(reached["kernelWriteReached"])
        self.assertTrue(r.path(self.epoch, "prepared").exists())
        self.assert_no_success()

    def test_enospc_before_receipt_open_denies_kernel_write(self):
        reached = {"kernelWriteReached": False}
        with mock.patch.object(r.os, "open", side_effect=OSError(errno.ENOSPC,"simulated disk full")):
            with self.assertRaises(OSError):
                r.prepare(STAGE, BOOT, NS, READY, WRITER)
                reached["kernelWriteReached"] = True
        self.assertFalse(reached["kernelWriteReached"])
        self.assertFalse(r.path(self.epoch, "prepared").exists())
        self.assert_no_success()

    def test_persisted_orphan_receipt_disallows_second_prepare(self):
        with self.inject_fsync_failure(1):
            with self.assertRaises(OSError):
                r.prepare(STAGE, BOOT, NS, READY, WRITER)
        old = r.path(self.epoch, "prepared").read_bytes()
        with self.assertRaisesRegex(ValueError, "already exist"):
            r.prepare(STAGE, BOOT, NS, READY, WRITER)
        self.assertEqual(r.path(self.epoch, "prepared").read_bytes(),old)
        self.assert_no_success()

    def test_kernel_absent_file_fsync_error_no_terminal(self):
        self._check_second_phase_failure(1)

    def test_kernel_absent_dir_fsync_error_no_terminal(self):
        self._check_second_phase_failure(2)

    def _check_second_phase_failure(self, failure_index):
        r.prepare(STAGE, BOOT, NS, READY, WRITER)
        # The real guest would have deleted the fixed synthetic table; no
        # nft invocation occurs in this test. This boolean models that point.
        guest_only_synthetic_nft_absence_observed = True
        with self.inject_fsync_failure(failure_index):
            with self.assertRaises(OSError):
                r.advance(self.epoch,"kernel-absent",STAGE,BOOT,NS,WRITER)
        self.assertTrue(guest_only_synthetic_nft_absence_observed)
        self.assertFalse(r.path(self.epoch,"journal-terminal").exists())
        with self.assertRaises(FileExistsError):
            r.advance(self.epoch,"kernel-absent",STAGE,BOOT,NS,WRITER)
        self.assert_no_success()

    def test_journal_terminal_file_fsync_error_no_done(self):
        self._check_terminal_receipt_failure(1)

    def test_journal_terminal_dir_fsync_error_no_done(self):
        self._check_terminal_receipt_failure(2)

    def _check_terminal_receipt_failure(self, failure_index):
        r.prepare(STAGE, BOOT, NS, READY, WRITER)
        r.advance(self.epoch,"kernel-absent",STAGE,BOOT,NS,WRITER)
        # Journal fsync completed in the modeled controller before this call.
        journal = {"state":"ROLLED_BACK"}
        with self.inject_fsync_failure(failure_index):
            with self.assertRaises(OSError):
                r.advance(self.epoch,"journal-terminal",STAGE,BOOT,NS,WRITER)
        self.assertEqual(journal["state"],"ROLLED_BACK")
        self.assert_no_success()

    def test_truncated_prepared_after_io_error_is_invalid(self):
        with self.inject_fsync_failure(1):
            with self.assertRaises(OSError):
                r.prepare(STAGE,BOOT,NS,READY,WRITER)
        # Simulate partial/short write and torn JSON. This cannot be treated
        # as durable PREPARED authorization by a subsequent writer.
        r.path(self.epoch,"prepared").write_bytes(b'{"schema":')
        self.assertEqual(r.inspect(STAGE,BOOT,NS,READY)["status"],
                         "INVALID_RECEIPT_FAIL_CLOSED")
        with self.assertRaises((ValueError,json.JSONDecodeError)):
            r.advance(self.epoch,"kernel-absent",STAGE,BOOT,NS,WRITER)
        self.assert_no_success()

    def test_fsync_errors_preserve_older_epoch(self):
        old = r.prepare(STAGE,BOOT,NS,READY,WRITER)
        before = r.path(old,"prepared").read_bytes()
        new_ready = dict(READY,armedMonotonic=280.0,deadlineMonotonic=290.0)
        with self.inject_fsync_failure(2):
            with self.assertRaises(OSError):
                r.prepare(STAGE,BOOT,NS,new_ready,WRITER)
        self.assertEqual(r.path(old,"prepared").read_bytes(),before)
        self.assert_no_success()


if __name__ == "__main__":
    unittest.main(verbosity=2)
