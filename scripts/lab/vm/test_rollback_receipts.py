#!/usr/bin/env python3
"""Offline immutable VM rollback observations: identity, ordering, crash gaps."""
import importlib.util
import json
import tempfile
from pathlib import Path
import unittest
from unittest import mock

source=Path(__file__).with_name("guest-rollback-receipts.py")
spec=importlib.util.spec_from_file_location("vm_tx_obs_test",source)
r=importlib.util.module_from_spec(spec)
spec.loader.exec_module(r)
STAGE="vm-"+"a"*24
BOOT="11111111-2222-4333-8444-555555555555"
NETNS="net:[12345678]"
READY={"stageId":STAGE,"bootId":BOOT,"netnsInode":NETNS,
       "armed":True,"anchorPid":120,"anchorBirthTicks":1900,
       "armedMonotonic":250.0,"deadlineMonotonic":260.0}

class ReceiptTests(unittest.TestCase):
    def setUp(self):
        td=tempfile.TemporaryDirectory(prefix="luna-txn-observed-")
        self.addCleanup(td.cleanup)
        self.root=Path(td.name)
        p=mock.patch.object(r,"ROOT",self.root)
        p.start();self.addCleanup(p.stop)

    def prepare(self):
        return r.prepare(STAGE,BOOT,NETNS,READY,"fallback")

    def test_ordered_all_three_records_persisted(self):
        e=self.prepare()
        self.assertEqual(r.inspect(STAGE,BOOT,NETNS,READY)["status"],"PREPARED_NO_KERNEL_ABSENCE_PROOF")
        r.advance(e,"kernel-absent",STAGE,BOOT,NETNS,"fallback")
        self.assertEqual(r.inspect(STAGE,BOOT,NETNS,READY)["status"],
                         "KERNEL_ABSENCE_OBSERVED_JOURNAL_NOT_TERMINAL_AT_RECEIPT")
        r.advance(e,"journal-terminal",STAGE,BOOT,NETNS,"fallback")
        self.assertEqual(r.inspect(STAGE,BOOT,NETNS,READY)["status"],
                         "JOURNAL_TERMINAL_RECEIPT_PRESENT_DONE_NOT_EVALUATED")
        for k in r.KINDS:
            self.assertEqual(r.path(e,k).stat().st_mode & 0o777,0o600)

    def test_absent_is_not_success(self):
        self.assertEqual(r.inspect(STAGE,BOOT,NETNS,READY)["status"],"NO_ROLLBACK_RECEIPT")

    def test_no_double_writer_on_same_epoch(self):
        self.prepare()
        with self.assertRaisesRegex(ValueError,"already exist"):
            r.prepare(STAGE,BOOT,NETNS,READY,"primary")

    def test_commit_without_kernel_observation_denied(self):
        e=self.prepare()
        with self.assertRaises((FileNotFoundError,ValueError)):
            r.advance(e,"journal-terminal",STAGE,BOOT,NETNS,"fallback")
        self.assertEqual(r.inspect(STAGE,BOOT,NETNS,READY)["status"],"PREPARED_NO_KERNEL_ABSENCE_PROOF")

    def test_wrong_writer_cannot_advance(self):
        e=self.prepare()
        with self.assertRaisesRegex(ValueError,"identity/hash chain"):
            r.advance(e,"kernel-absent",STAGE,BOOT,NETNS,"primary")

    def test_boot_change_denies_identity(self):
        with self.assertRaisesRegex(ValueError,"READY not bound"):
            r.prepare(STAGE,"aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee",NETNS,READY,"fallback")

    def test_bad_netns_denies_identity(self):
        with self.assertRaisesRegex(ValueError,"invalid target netns"):
            r.prepare(STAGE,BOOT,"net:[evil]",READY,"fallback")

    def test_corrupted_prepared_denies_progression(self):
        e=self.prepare()
        f=r.path(e,"prepared")
        v=json.loads(f.read_text())
        v["writer"]="primary"
        f.write_text(json.dumps(v))
        with self.assertRaisesRegex(ValueError,"identity/hash chain"):
            r.advance(e,"kernel-absent",STAGE,BOOT,NETNS,"fallback")

    def test_duplicate_key_fails_closed(self):
        e=self.prepare()
        p=r.path(e,"prepared")
        p.write_text('{"schema":"x","schema":"y"}')
        self.assertEqual(r.inspect(STAGE,BOOT,NETNS,READY)["status"],"INVALID_RECEIPT_FAIL_CLOSED")

    def test_out_of_order_receipt_detected(self):
        e=r._id(STAGE,BOOT,NETNS,READY)
        r._write_exclusive(r.path(e,"journal-terminal"),{"invalid":True})
        self.assertEqual(r.inspect(STAGE,BOOT,NETNS,READY)["status"],"INVALID_RECEIPT_FAIL_CLOSED")

    def test_receipt_immutable_no_overwrite(self):
        e=self.prepare()
        p=r.path(e,"prepared")
        before=p.read_bytes()
        with self.assertRaises(FileExistsError):
            r._write_exclusive(p,{"tampered":True})
        self.assertEqual(p.read_bytes(),before)

    def test_different_epoch_does_not_replace_old_evidence(self):
        e=self.prepare()
        new=dict(READY,armedMonotonic=290.0,deadlineMonotonic=300.0)
        e2=r.prepare(STAGE,BOOT,NETNS,new,"fallback")
        self.assertNotEqual(e,e2)
        self.assertTrue(r.path(e,"prepared").exists())

if __name__=="__main__":
    unittest.main(verbosity=2)
