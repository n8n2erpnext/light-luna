#!/usr/bin/env python3
"""Offline VM-only manifest identity tests: TTL must be bound to stageId.

Reads only temporary synthetic JSON; never enters a namespace, calls nft,
touches systemd, or installs privileges. This is not kernel rollback proof.
"""
import hashlib
import importlib.util
import json
import os
import tempfile
import unittest
from pathlib import Path
from unittest import mock

MODULE = Path(__file__).with_name("guest-stage-journal.py")
spec = importlib.util.spec_from_file_location("stage_journal_ttl_under_test", MODULE)
journal = importlib.util.module_from_spec(spec)
spec.loader.exec_module(journal)

HASH = "a" * 64


def stage_id(ttl: int) -> str:
    content = (
        journal.SCHEMA + "\x00" + HASH + "\x00" + journal.TABLE
        + "\x00" + str(ttl)
    ).encode()
    return "vm-" + hashlib.sha256(content).hexdigest()[:24]


def manifest(ttl: int) -> dict:
    return {
        "schema": journal.SCHEMA,
        "mode": journal.MODE,
        "policyName": "host-baseline",
        "policyRevision": 1,
        "sourcePreviewHash": HASH,
        "stageId": stage_id(ttl),
        "table": journal.TABLE,
        "ttlSeconds": ttl,
        "phases": list(journal.PHASES),
        "effect": journal.EFFECT,
        "kernelWrites": False,
        "ownerApproved": False,
        "scope": "KVM_GUEST_PRIVATE_NETNS_ONLY",
    }


class StageTTLIdentityTests(unittest.TestCase):
    def setUp(self):
        work = tempfile.TemporaryDirectory(prefix="luna-stage-ttl-")
        self.addCleanup(work.cleanup)
        self.path = Path(work.name) / "stage-intent.json"
        self.patcher = mock.patch.object(journal, "MANIFEST", self.path)
        self.patcher.start()
        self.addCleanup(self.patcher.stop)

    def write(self, data):
        self.path.write_text(json.dumps(data), encoding="utf-8")

    def test_same_manifest_is_repeatable(self):
        data = manifest(8)
        self.write(data)
        self.assertEqual(journal.manifest()["stageId"], data["stageId"])
        self.assertEqual(stage_id(8), stage_id(8))

    def test_different_ttl_has_different_stage_identity(self):
        self.assertNotEqual(stage_id(1), stage_id(10))
        self.write(manifest(1))
        self.assertEqual(journal.manifest()["ttlSeconds"], 1)
        self.write(manifest(10))
        self.assertEqual(journal.manifest()["ttlSeconds"], 10)

    def test_changed_ttl_with_old_stage_id_is_denied(self):
        data = manifest(8)
        data["ttlSeconds"] = 9
        self.write(data)
        with self.assertRaisesRegex(ValueError, "stage id mismatch"):
            journal.manifest()

    def test_legacy_unbound_stage_id_is_denied(self):
        data = manifest(8)
        legacy = (journal.SCHEMA + "\x00" + HASH + "\x00" + journal.TABLE).encode()
        data["stageId"] = "vm-" + hashlib.sha256(legacy).hexdigest()[:24]
        self.write(data)
        with self.assertRaisesRegex(ValueError, "stage id mismatch"):
            journal.manifest()

    def test_forged_effect_claim_rejected(self):
        for forged in ("kernel rules applied", "owner authorized", None):
            with self.subTest(effect=forged):
                data = manifest(8)
                data["effect"] = forged
                self.write(data)
                with self.assertRaisesRegex(ValueError, "unsafe manifest"):
                    journal.manifest()

    def test_revision_must_be_exact_integer(self):
        for wrong in (True, 1.0):
            with self.subTest(revision=wrong):
                data = manifest(8)
                data["policyRevision"] = wrong
                self.write(data)
                with self.assertRaisesRegex(ValueError, "unsafe manifest"):
                    journal.manifest()

    def test_invalid_ttl_rejected_even_with_recomputed_id(self):
        for ttl in (0, 11):
            with self.subTest(ttl=ttl):
                self.write(manifest(ttl))
                with self.assertRaisesRegex(ValueError, "unsafe manifest"):
                    journal.manifest()



class StageManifestFileReadTests(unittest.TestCase):
    """Offline tests for stale path checks and alternate inode names."""

    def setUp(self):
        work = tempfile.TemporaryDirectory(prefix="luna-stage-io-")
        self.addCleanup(work.cleanup)
        self.root = Path(work.name)
        self.path = self.root / "stage-intent.json"
        self.payload = json.dumps(manifest(8))

    def test_legitimate_single_link_regular_file_reads(self):
        self.path.write_text(self.payload, encoding="utf-8")
        self.assertEqual(journal.safe_load(self.path), manifest(8))

    def test_hardlinked_manifest_rejected(self):
        source = self.root / "other-name.json"
        source.write_text(self.payload, encoding="utf-8")
        os.link(source, self.path)
        with self.assertRaises((ValueError, OSError)):
            journal.safe_load(self.path)

    def test_symlink_denied_even_if_precheck_lies(self):
        source = self.root / "other-name.json"
        source.write_text(self.payload, encoding="utf-8")
        self.path.symlink_to(source)
        # Simulate stale is_symlink() information: the final open must still
        # refuse to traverse a symlink, even when the precheck was inaccurate.
        with mock.patch.object(Path, "is_symlink", return_value=False):
            with self.assertRaises((ValueError, OSError)):
                journal.safe_load(self.path)

    def test_file_growth_denied_even_if_precheck_size_stale(self):
        self.path.write_text(self.payload, encoding="utf-8")
        original_stat = self.path.stat()
        self.path.write_text(self.payload + " " * 4097, encoding="utf-8")
        # Original stat would claim the old safe size. The opened FD must be
        # checked independently and reading must remain bounded to 4096 bytes.
        with mock.patch.object(Path, "stat", return_value=original_stat):
            with self.assertRaises((ValueError, OSError)):
                journal.safe_load(self.path)

    def test_duplicate_keys_still_rejected(self):
        self.path.write_text('{"a":1,"a":2}', encoding="utf-8")
        with self.assertRaisesRegex(ValueError, "duplicate JSON key"):
            journal.safe_load(self.path)


if __name__ == "__main__":
    unittest.main(verbosity=2)
