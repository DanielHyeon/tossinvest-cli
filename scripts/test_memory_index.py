import tempfile
import unittest
import json
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path
from unittest import mock

import memory_index


class MemoryIndexTests(unittest.TestCase):
    def episode(
        self,
        root: Path,
        body: str = "A verified lesson.",
        record_id: str = "EP-TOS-001",
    ) -> Path:
        path = root / "episodes" / f"{record_id}.md"
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(
            """---
id: {record_id}
change_id: add-example
status: episodic
created_at: 2026-07-27
summary: verified retry lesson
---
""".format(record_id=record_id)
            + body,
            encoding="utf-8",
        )
        return path

    def test_retain_search_and_explicit_promotion(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            result = memory_index.retain(self.episode(root), root)
            self.assertTrue(result["ok"])
            self.assertEqual(len(memory_index.search("retry", root)), 1)
            denied = memory_index.promote("EP-TOS-001", "", root)
            self.assertFalse(denied["ok"])
            promoted = memory_index.promote("EP-TOS-001", "gate green", root)
            self.assertTrue(promoted["ok"])
            record = memory_index.load_ledger(root)[0]
            self.assertEqual(record["status"], "canonical")
            self.assertTrue((root / record["source_path"]).is_file())
            self.assertIn(
                "status: canonical",
                (root / record["source_path"]).read_text(encoding="utf-8"),
            )
            self.assertFalse((root / "episodes" / "EP-TOS-001.md").exists())

    def test_sensitive_marker_is_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            errors = memory_index.validate_episode(self.episode(root, "password: leaked"))
            self.assertTrue(any("forbidden" in error for error in errors))
            errors = memory_index.validate_episode(self.episode(root, "token=x"))
            self.assertTrue(any("assignment" in error for error in errors))

    def test_secret_shapes_and_external_sources_are_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp) / "memory"
            external = Path(tmp) / "outside.md"
            external.write_text("Bearer short-token", encoding="utf-8")
            result = memory_index.retain(external, root)
            self.assertFalse(result["ok"])
            errors = memory_index.validate_episode(
                self.episode(root, "contact user@example.com")
            )
            self.assertTrue(any("sensitive shape" in error for error in errors))

    def test_malicious_ledger_path_is_not_read(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp) / "memory"
            root.mkdir(parents=True)
            outside = Path(tmp) / "outside.md"
            outside.write_text("DO-NOT-INDEX", encoding="utf-8")
            record = {
                "id": "evil",
                "status": "episodic",
                "change_id": "evil",
                "summary": "evil",
                "source_path": str(outside),
            }
            (root / "ledger.jsonl").write_text(
                json.dumps(record) + "\n",
                encoding="utf-8",
            )
            self.assertEqual(memory_index.load_ledger(root), [])
            self.assertEqual(memory_index.rebuild(root), [])
            self.assertNotIn(
                "DO-NOT-INDEX",
                (root / "index.jsonl").read_text(encoding="utf-8"),
            )

    def test_promotion_evidence_rejects_short_secret_assignments(self):
        for evidence in (
            "password=x",
            "api_key=x",
            "token=x",
            "cookie=x",
            "authorization=x",
        ):
            with self.subTest(evidence=evidence), tempfile.TemporaryDirectory() as tmp:
                root = Path(tmp)
                self.assertTrue(memory_index.retain(self.episode(root), root)["ok"])
                result = memory_index.promote("EP-TOS-001", evidence, root)
                self.assertFalse(result["ok"])
                self.assertEqual(memory_index.load_ledger(root)[0]["status"], "episodic")

    def test_concurrent_retain_keeps_every_record(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            paths = [
                self.episode(root, record_id=f"EP-TOS-{number:03d}")
                for number in range(8)
            ]
            with ThreadPoolExecutor(max_workers=8) as pool:
                results = list(pool.map(lambda path: memory_index.retain(path, root), paths))
            self.assertTrue(all(result["ok"] for result in results))
            self.assertEqual(len(memory_index.load_ledger(root)), 8)

    def test_canonical_id_cannot_be_demoted_or_replaced(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            self.assertTrue(memory_index.retain(self.episode(root), root)["ok"])
            self.assertTrue(
                memory_index.promote("EP-TOS-001", "gate green", root)["ok"]
            )
            replacement = self.episode(root, body="replacement")
            result = memory_index.retain(replacement, root)
            self.assertFalse(result["ok"])
            record = memory_index.load_ledger(root)[0]
            self.assertEqual(record["status"], "canonical")
            self.assertIn(
                "A verified lesson.",
                (root / record["source_path"]).read_text(encoding="utf-8"),
            )

    def test_promotion_rejects_symlinked_canonical_directory(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp) / "memory"
            outside = Path(tmp) / "outside"
            outside.mkdir()
            self.assertTrue(memory_index.retain(self.episode(root), root)["ok"])
            (root / "canonical").symlink_to(outside, target_is_directory=True)
            result = memory_index.promote("EP-TOS-001", "gate green", root)
            self.assertFalse(result["ok"])
            self.assertEqual(list(outside.iterdir()), [])

    def test_check_rejects_untracked_invalid_memory_source(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            self.episode(root, "password=short")
            self.assertTrue(memory_index.check_memory(root))

    def test_check_rejects_valid_orphan_and_unsafe_ledger_record(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            self.episode(root)
            (root / "ledger.jsonl").write_text(
                json.dumps(
                    {
                        "id": "evil",
                        "status": "episodic",
                        "source_path": "/tmp/outside.md",
                    }
                )
                + "\n",
                encoding="utf-8",
            )
            errors = memory_index.check_memory(root)
            self.assertTrue(any("absolute" in error for error in errors))
            self.assertTrue(any("not linked" in error for error in errors))

    def test_promotion_rebuild_failure_restores_episodic_ledger(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            source = self.episode(root)
            self.assertTrue(memory_index.retain(source, root)["ok"])
            with mock.patch("memory_index._rebuild", side_effect=OSError("disk full")):
                result = memory_index.promote("EP-TOS-001", "gate green", root)
            self.assertFalse(result["ok"])
            record = memory_index.load_ledger(root)[0]
            self.assertEqual(record["status"], "episodic")
            self.assertEqual(record["source_path"], "episodes/EP-TOS-001.md")
            self.assertTrue(source.exists())
            self.assertFalse((root / "canonical" / source.name).exists())
            self.assertEqual(memory_index.check_memory(root), [])

    def test_promotion_index_build_failure_leaves_derived_indexes_episodic(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            source = self.episode(root)
            self.assertTrue(memory_index.retain(source, root)["ok"])
            original_json = (root / "index.jsonl").read_text(encoding="utf-8")
            original_db = (root / "index.db").read_bytes()
            with mock.patch(
                "memory_index.sqlite3.connect",
                side_effect=OSError("sqlite unavailable"),
            ):
                result = memory_index.promote("EP-TOS-001", "gate green", root)
            self.assertFalse(result["ok"])
            self.assertEqual(
                (root / "index.jsonl").read_text(encoding="utf-8"),
                original_json,
            )
            self.assertEqual((root / "index.db").read_bytes(), original_db)
            self.assertEqual(list(root.glob(".index-*")), [])

    def test_promotion_second_temp_creation_failure_cleans_first_temp(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            source = self.episode(root)
            self.assertTrue(memory_index.retain(source, root)["ok"])
            with mock.patch(
                "memory_index.tempfile.mkstemp",
                side_effect=OSError("no temp database"),
            ):
                result = memory_index.promote("EP-TOS-001", "gate green", root)
            self.assertFalse(result["ok"])
            self.assertEqual(memory_index.load_ledger(root)[0]["status"], "episodic")
            self.assertEqual(list(root.glob(".index-*")), [])

    def test_promotion_first_temp_write_failure_cleans_created_file(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            source = self.episode(root)
            self.assertTrue(memory_index.retain(source, root)["ok"])
            original = memory_index.tempfile.NamedTemporaryFile

            class FailingIndexTemp:
                def __init__(self):
                    self.name = str(root / ".index-write-failure")
                    self.handle = None

                def __enter__(self):
                    self.handle = open(self.name, "w", encoding="utf-8")
                    return self

                def write(self, _value):
                    raise OSError("index write failed")

                def __exit__(self, *_args):
                    self.handle.close()

            def temporary_file(*args, **kwargs):
                if kwargs.get("prefix") == ".index-":
                    return FailingIndexTemp()
                return original(*args, **kwargs)

            with mock.patch(
                "memory_index.tempfile.NamedTemporaryFile",
                side_effect=temporary_file,
            ):
                result = memory_index.promote("EP-TOS-001", "gate green", root)
            self.assertFalse(result["ok"])
            self.assertEqual(memory_index.load_ledger(root)[0]["status"], "episodic")
            self.assertEqual(list(root.glob(".index-*")), [])


if __name__ == "__main__":
    unittest.main()
