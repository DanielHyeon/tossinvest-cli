import json
import tempfile
import unittest
from pathlib import Path

import generate_master_tracker as tracker


class TrackerTests(unittest.TestCase):
    def fixture(self, root: Path) -> None:
        portfolio = root / "docs" / "pm" / "portfolio"
        for path in tracker.PORTFOLIO.rglob("*.yaml"):
            target = portfolio / path.relative_to(tracker.PORTFOLIO)
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_text(path.read_text(encoding="utf-8"), encoding="utf-8")
        story_path = portfolio / "stories" / "STORY-TOS-001.yaml"
        story = json.loads(story_path.read_text(encoding="utf-8"))
        story["status"] = "in_progress"
        story_path.write_text(json.dumps(story), encoding="utf-8")
        changes = root / "openspec" / "changes"
        for change in (
            "adopt-stockos-full-sdd",
            "verify-execution-capability",
            "add-net-rr-measurement",
            "console-adoption-controls",
            "console-click-approval",
            "verify-us-market",
            "apply-us-measurement-fixes",
            "verify-clears-leftovers",
        ):
            path = changes / change
            path.mkdir(parents=True)
            (path / "proposal.md").write_text("# proposal", encoding="utf-8")

    def test_current_repository_is_valid(self):
        self.assertEqual(tracker.validate(), [])

    def test_orphan_change_is_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            self.fixture(root)
            path = root / "openspec" / "changes" / "orphan-change"
            path.mkdir(parents=True)
            (path / "proposal.md").write_text("# proposal", encoding="utf-8")
            errors = tracker.validate(root)
            self.assertTrue(any("orphan-change" in error for error in errors))

    def test_reverse_link_duplicate_change_and_stale_allowlist_are_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            self.fixture(root)
            portfolio = root / "docs" / "pm" / "portfolio"
            epic_path = portfolio / "epics" / "EPIC-TOS-001.yaml"
            epic = json.loads(epic_path.read_text(encoding="utf-8"))
            epic["initiative"] = "INIT-MISSING"
            epic_path.write_text(json.dumps(epic), encoding="utf-8")

            story_path = portfolio / "stories" / "STORY-TOS-001.yaml"
            duplicate = json.loads(story_path.read_text(encoding="utf-8"))
            duplicate["id"] = "STORY-TOS-002"
            (portfolio / "stories" / "STORY-TOS-002.yaml").write_text(
                json.dumps(duplicate),
                encoding="utf-8",
            )
            registry_path = portfolio / "_registry.yaml"
            registry = json.loads(registry_path.read_text(encoding="utf-8"))
            registry["stories"].append("STORY-TOS-002")
            registry["bootstrap_change_allowlist"].append("stale-change")
            registry_path.write_text(json.dumps(registry), encoding="utf-8")
            feature_path = portfolio / "features" / "FEAT-TOS-001.yaml"
            feature = json.loads(feature_path.read_text(encoding="utf-8"))
            feature["stories"].append("STORY-TOS-002")
            feature_path.write_text(json.dumps(feature), encoding="utf-8")

            errors = tracker.validate(root)
            self.assertTrue(any("initiative reverse link" in error for error in errors))
            self.assertTrue(any("expected one story" in error for error in errors))
            self.assertTrue(any("stale bootstrap" in error for error in errors))

    def test_stale_generated_tracker_is_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            self.fixture(root)
            errors = tracker.generate(root, check=True)
            self.assertTrue(any("generated tracker stale" in error for error in errors))

    def test_completed_story_must_point_to_archive(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            self.fixture(root)
            story_path = (
                root
                / "docs"
                / "pm"
                / "portfolio"
                / "stories"
                / "STORY-TOS-001.yaml"
            )
            story = json.loads(story_path.read_text(encoding="utf-8"))
            story["status"] = "done"
            story_path.write_text(json.dumps(story), encoding="utf-8")
            active = root / "openspec" / "changes" / "adopt-stockos-full-sdd"
            archived = (
                root
                / "openspec"
                / "changes"
                / "archive"
                / "2026-07-27-adopt-stockos-full-sdd"
            )
            archived.parent.mkdir(parents=True)
            active.rename(archived)
            self.assertEqual(tracker.validate(root), [])

    def test_duplicate_ids_filename_mismatch_and_forward_orphans_are_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            self.fixture(root)
            portfolio = root / "docs" / "pm" / "portfolio"
            story = json.loads(
                (portfolio / "stories" / "STORY-TOS-001.yaml").read_text(
                    encoding="utf-8"
                )
            )
            (portfolio / "stories" / "WRONG-NAME.yaml").write_text(
                json.dumps(story),
                encoding="utf-8",
            )
            feature_path = portfolio / "features" / "FEAT-TOS-001.yaml"
            feature = json.loads(feature_path.read_text(encoding="utf-8"))
            feature["stories"].append("STORY-MISSING")
            feature_path.write_text(json.dumps(feature), encoding="utf-8")
            errors = tracker.validate(root)
            self.assertTrue(any("filename/id mismatch" in error for error in errors))
            self.assertTrue(any("duplicate id" in error for error in errors))
            self.assertTrue(any("invalid story forward link" in error for error in errors))


if __name__ == "__main__":
    unittest.main()
