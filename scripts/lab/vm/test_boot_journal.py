#!/usr/bin/env python3
"""Offline, unprivileged P0 VM journal boot-identity regression tests.

No LXD, nft, kernel mutation, device access or privileges.
"""
import importlib.util
import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path
from unittest import mock

MODULE = Path(__file__).with_name("guest-stage-journal.py")
spec = importlib.util.spec_from_file_location("luna_stage_journal_under_test", MODULE)
journal = importlib.util.module_from_spec(spec)
spec.loader.exec_module(journal)

BOOT_A = "11111111-2222-4333-8444-555555555555"
BOOT_B = "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee"

class BootJournalTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory(prefix="luna-boot-identity-")
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.patch_root = mock.patch.object(journal, "ROOT", self.root)
        self.patch_journal = mock.patch.object(journal, "JOURNAL", self.root / "stage-journal.json")
        self.patch_root.start()
        self.patch_journal.start()
        self.addCleanup(self.patch_root.stop)
        self.addCleanup(self.patch_journal.stop)
        self.manifest = {"stageId": "vm-test-boot-identity"}

    def state(self):
        return json.loads(journal.JOURNAL.read_text(encoding="utf-8"))

    def test_same_boot_progresses_then_reboot_denies_transition(self):
        with mock.patch.object(journal, "boot_id", return_value=BOOT_A):
            journal.transition("init", "STAGED", self.manifest)
            journal.transition("transition", "VALIDATED", self.manifest)
            self.assertEqual(self.state()["bootId"], BOOT_A)
        with mock.patch.object(journal, "boot_id", return_value=BOOT_B):
            with self.assertRaisesRegex(ValueError, "different Linux boot"):
                journal.transition("transition", "COMMITTED_SYNTHETIC", self.manifest)
        self.assertEqual(self.state()["state"], "VALIDATED")
        self.assertEqual(len(self.state()["history"]), 2)

    def test_old_journal_without_boot_id_denied(self):
        journal.JOURNAL.write_text(
            json.dumps({"stageId": self.manifest["stageId"], "state": "STAGED", "history": ["STAGED"]}),
            encoding="utf-8",
        )
        with mock.patch.object(journal, "boot_id", return_value=BOOT_A):
            with self.assertRaisesRegex(ValueError, "journal mismatch"):
                journal.transition("transition", "VALIDATED", self.manifest)
        self.assertNotIn("bootId", self.state())

    def test_corrupt_or_duplicate_journal_denied(self):
        journal.JOURNAL.write_text('{"stageId":"x","state":"STAGED","state":"VALIDATED"}', encoding="utf-8")
        with mock.patch.object(journal, "boot_id", return_value=BOOT_A):
            with self.assertRaisesRegex(ValueError, "duplicate JSON key"):
                journal.transition("transition", "VALIDATED", self.manifest)

    def test_journal_is_bound_at_init_and_cannot_be_reinitialized(self):
        with mock.patch.object(journal, "boot_id", return_value=BOOT_A):
            journal.transition("init", "STAGED", self.manifest)
            with self.assertRaisesRegex(ValueError, "previous transaction"):
                journal.transition("init", "STAGED", self.manifest)
        self.assertEqual(self.state()["bootId"], BOOT_A)

    def test_init_does_not_overwrite_journal_created_after_precheck(self):
        # A second writer publishes evidence after the initial path check.
        # Initial commit must be create-only, preserving existing contents.
        real_mkstemp = journal.tempfile.mkstemp
        unexpected = '{"foreign":"previous-or-concurrent-evidence"}'
        def concurrent_writer(*args, **kwargs):
            fd, path = real_mkstemp(*args, **kwargs)
            journal.JOURNAL.write_text(unexpected, encoding="utf-8")
            return fd, path
        with mock.patch.object(journal.tempfile, "mkstemp", side_effect=concurrent_writer):
            with mock.patch.object(journal, "boot_id", return_value=BOOT_A):
                with self.assertRaises((ValueError, OSError)):
                    journal.transition("init", "STAGED", self.manifest)
        self.assertEqual(journal.JOURNAL.read_text(encoding="utf-8"), unexpected)
        self.assertEqual(sorted(p.name for p in self.root.iterdir()), ["stage-journal.json"])

    def test_init_does_not_replace_symlink_created_after_precheck(self):
        target = self.root / "protected.json"
        target.write_text("preserve-me", encoding="utf-8")
        real_mkstemp = journal.tempfile.mkstemp
        def concurrent_symlink(*args, **kwargs):
            fd, path = real_mkstemp(*args, **kwargs)
            journal.JOURNAL.symlink_to(target)
            return fd, path
        with mock.patch.object(journal.tempfile, "mkstemp", side_effect=concurrent_symlink):
            with mock.patch.object(journal, "boot_id", return_value=BOOT_A):
                with self.assertRaises((ValueError, OSError)):
                    journal.transition("init", "STAGED", self.manifest)
        self.assertTrue(journal.JOURNAL.is_symlink())
        self.assertEqual(target.read_text(encoding="utf-8"), "preserve-me")
        self.assertEqual(sorted(p.name for p in self.root.iterdir()), ["protected.json", "stage-journal.json"])

    def test_competing_transition_cannot_publish_from_stale_state(self):
        # Model two cooperating journal writers: the outer writer has checked
        # STAGED, but has not yet published VALIDATED. An inner writer must not
        # reread that same STAGED state and also report VALIDATED as success.
        with mock.patch.object(journal, "boot_id", return_value=BOOT_A):
            journal.transition("init", "STAGED", self.manifest)
            real_persist = journal.persist
            entered = False
            competing = []
            def pause_before_publish(state, **kwargs):
                nonlocal entered
                if not entered:
                    entered = True
                    try:
                        journal.transition("transition", "VALIDATED", self.manifest)
                    except (BlockingIOError, ValueError, OSError):
                        competing.append("DENIED")
                    else:
                        competing.append("ACCEPTED")
                return real_persist(state, **kwargs)
            with mock.patch.object(journal, "persist", side_effect=pause_before_publish):
                journal.transition("transition", "VALIDATED", self.manifest)
        self.assertEqual(competing, ["DENIED"], "two writers accepted the same STAGED transition")
        self.assertEqual(self.state()["history"], ["STAGED", "VALIDATED"])

    def test_separate_python_process_respects_directory_writer_lock(self):
        # A separate interpreter independently opens the same directory.
        # This tests real Linux kernel advisory-lock contention, not merely
        # an in-process mock; no guest VM or service is started.
        code = """
import importlib.util
import sys
from pathlib import Path
spec = importlib.util.spec_from_file_location("isolated_guest_writer", sys.argv[1])
mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mod)
mod.ROOT = Path(sys.argv[2])
try:
    with mod.stage_writer_guard():
        sys.exit(7)   # concurrent lock unexpectedly granted
except ValueError as err:
    if "another VM stage journal writer" not in str(err):
        raise
    sys.exit(0)
"""
        with journal.stage_writer_guard():
            process = subprocess.run(
                [sys.executable, "-c", code, str(MODULE), str(self.root)],
                check=False, capture_output=True, text=True, timeout=5,
            )
            self.assertEqual(process.returncode, 0, process.stderr)
        # Once the owner releases, an unrelated writer can make progress.
        with journal.stage_writer_guard():
            pass

    def test_unavailable_kernel_boot_id_fails_closed(self):
        with mock.patch.object(Path, "read_text", return_value="not-a-boot-id"):
            with self.assertRaisesRegex(ValueError, "invalid kernel boot ID"):
                journal.boot_id()

if __name__ == "__main__":
    unittest.main(verbosity=2)
