#!/usr/bin/env python3
"""Offline VM-only manifest identity tests: TTL must be bound to stageId.

Reads only temporary synthetic JSON; never enters a namespace, calls nft,
touches systemd, or installs privileges. This is not kernel rollback proof.
"""
import hashlib
import importlib.util
import json
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


if __name__ == "__main__":
    unittest.main(verbosity=2)
