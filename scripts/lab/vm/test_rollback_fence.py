#!/usr/bin/env python3
"""Offline, unprivileged POSIX-process locking tests for VM synthetic rollback."""
import importlib.util
import json
import multiprocessing
import os
from pathlib import Path
import tempfile
import time
import unittest
from unittest import mock

spec=importlib.util.spec_from_file_location("vm_rollback_fence",Path(__file__).with_name("guest-rollback-fence.py"))
fence=importlib.util.module_from_spec(spec)
spec.loader.exec_module(fence)
STAGE="vm-"+"a"*24
BOOT="11111111-2222-4333-8444-555555555555"
PHASES=["STAGED","VALIDATED","COMMITTED_SYNTHETIC","ROLLED_BACK"]

class FakeModule:
    PHASES=PHASES
    def __init__(self,root,boot=BOOT):
        self.JOURNAL=root/"journal.json"
        self._boot=boot
    def boot_id(self):
        return self._boot
    def safe_load(self,p):
        return json.loads(p.read_text())

def other_writer(root,deadline,result):
    fence.ROOT=root
    fence.LOCK=root/"synthetic-rollback-fence.lock"
    mod=FakeModule(root)
    try:
        with fence.guard(mod,STAGE,BOOT,max_wait=deadline):
            result.put("entered")
    except ValueError as e:
        result.put(str(e))

class FenceTests(unittest.TestCase):
    def setUp(self):
        self.dir=tempfile.TemporaryDirectory(prefix="lightluna-fence-")
        self.addCleanup(self.dir.cleanup)
        self.root=Path(self.dir.name)
        self.mod=FakeModule(self.root)
        self.mod.JOURNAL.write_text(json.dumps({
            "stageId":STAGE,"bootId":BOOT,
            "state":"COMMITTED_SYNTHETIC","history":PHASES[:3],
        }))
        a=mock.patch.object(fence,"ROOT",self.root)
        b=mock.patch.object(fence,"LOCK",self.root/"synthetic-rollback-fence.lock")
        a.start();b.start()
        self.addCleanup(a.stop);self.addCleanup(b.stop)

    def test_accepts_correct_boot_stage_and_state(self):
        with fence.guard(self.mod,STAGE,BOOT) as state:
            self.assertEqual(state["state"],"COMMITTED_SYNTHETIC")
        self.assertEqual((self.root/"synthetic-rollback-fence.lock").stat().st_mode & 0o777,0o600)

    def test_rejects_wrong_boot(self):
        self.mod._boot="aaaaaaa1-2222-4333-8444-555555555555"
        with self.assertRaisesRegex(ValueError,"crossed a Linux boot"):
            with fence.guard(self.mod,STAGE,BOOT):pass

    def test_rejects_wrong_stage(self):
        with self.assertRaisesRegex(ValueError,"identity mismatched"):
            with fence.guard(self.mod,"vm-"+"b"*24,BOOT):pass

    def test_rejects_corrupt_history(self):
        state=json.loads(self.mod.JOURNAL.read_text())
        state["history"]=["STAGED","COMMITTED_SYNTHETIC"]
        self.mod.JOURNAL.write_text(json.dumps(state))
        with self.assertRaisesRegex(ValueError,"invalid journal history"):
            with fence.guard(self.mod,STAGE,BOOT):pass

    def test_rejects_state_not_matching_final_history(self):
        state=json.loads(self.mod.JOURNAL.read_text())
        state["state"]="ROLLED_BACK"
        self.mod.JOURNAL.write_text(json.dumps(state))
        with self.assertRaisesRegex(ValueError,"invalid journal history"):
            with fence.guard(self.mod,STAGE,BOOT):pass

    def test_rejects_empty_history_with_committed_state(self):
        state=json.loads(self.mod.JOURNAL.read_text())
        state["history"]=[]
        self.mod.JOURNAL.write_text(json.dumps(state))
        with self.assertRaisesRegex(ValueError,"invalid journal history"):
            with fence.guard(self.mod,STAGE,BOOT):pass

    def test_rejects_symlink_lock(self):
        (self.root/"target").write_text("irrelevant")
        (self.root/"synthetic-rollback-fence.lock").symlink_to(self.root/"target")
        with self.assertRaises(OSError):
            with fence.guard(self.mod,STAGE,BOOT):pass
        self.assertEqual((self.root/"target").read_text(),"irrelevant")

    def test_rejects_world_writable_lock(self):
        lock=self.root/"synthetic-rollback-fence.lock"
        lock.touch(mode=0o600)
        lock.chmod(0o666)
        with self.assertRaisesRegex(ValueError,"unsafe or replaced"):
            with fence.guard(self.mod,STAGE,BOOT):pass

    def test_independent_process_fails_closed_during_lock(self):
        # Two independent processes/FDs use the same advisory flock; a
        # contender times out before touching any shared journal or nft.
        ctx=multiprocessing.get_context("fork")
        result=ctx.Queue()
        with fence.guard(self.mod,STAGE,BOOT):
            process=ctx.Process(target=other_writer,args=(self.root,0.15,result))
            process.start()
            answer=result.get(timeout=3)
            process.join(timeout=3)
        self.assertEqual(process.exitcode,0)
        self.assertIn("contention timed out",answer)

    def test_independent_process_enters_after_release(self):
        ctx=multiprocessing.get_context("fork")
        result=ctx.Queue()
        with fence.guard(self.mod,STAGE,BOOT):
            process=ctx.Process(target=other_writer,args=(self.root,1.2,result))
            process.start()
            time.sleep(0.12)
            self.assertTrue(result.empty())
        self.assertEqual(result.get(timeout=3),"entered")
        process.join(timeout=3)
        self.assertEqual(process.exitcode,0)

if __name__=="__main__":
    unittest.main(verbosity=2)
