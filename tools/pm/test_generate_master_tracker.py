import json
import tempfile
import unittest
from pathlib import Path

import generate_master_tracker as tracker


class TrackerTests(unittest.TestCase):
    def write_json(self, path: Path, value: dict) -> None:
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(json.dumps(value, indent=2) + "\n", encoding="utf-8")

    def fixture(self, root: Path) -> None:
        portfolio = root / "docs" / "pm" / "portfolio"
        self.write_json(
            portfolio / "_registry.yaml",
            {
                "project": "TossOS",
                "prefix": "TOS",
                "initiatives": ["INIT-TOS-001"],
                "epics": ["EPIC-TOS-001"],
                "features": ["FEAT-TOS-001"],
                "stories": ["STORY-TOS-001"],
            },
        )
        self.write_json(
            portfolio / "initiatives" / "INIT-TOS-001.yaml",
            {
                "id": "INIT-TOS-001",
                "title": "Delivery governance",
                "intent": "active",
                "epics": ["EPIC-TOS-001"],
            },
        )
        self.write_json(
            portfolio / "epics" / "EPIC-TOS-001.yaml",
            {
                "id": "EPIC-TOS-001",
                "initiative": "INIT-TOS-001",
                "title": "SDD",
                "intent": "active",
                "features": ["FEAT-TOS-001"],
            },
        )
        self.write_json(
            portfolio / "features" / "FEAT-TOS-001.yaml",
            {
                "id": "FEAT-TOS-001",
                "epic": "EPIC-TOS-001",
                "title": "SDD controls",
                "intent": "active",
                "stories": ["STORY-TOS-001"],
            },
        )
        self.write_json(
            portfolio / "stories" / "STORY-TOS-001.yaml",
            {
                "id": "STORY-TOS-001",
                "feature": "FEAT-TOS-001",
                "title": "Example delivery",
                "intent": "active",
                "openspec": {
                    "change_id": "example-change",
                    "path": "openspec/changes/example-change",
                },
                "acceptance": ["Example acceptance"],
            },
        )
        change = root / "openspec" / "changes" / "example-change"
        change.mkdir(parents=True)
        (change / "proposal.md").write_text("# proposal\n", encoding="utf-8")

    def story(self, root: Path) -> tuple[Path, dict]:
        path = (
            root
            / "docs"
            / "pm"
            / "portfolio"
            / "stories"
            / "STORY-TOS-001.yaml"
        )
        return path, json.loads(path.read_text(encoding="utf-8"))

    def replace_with_numbered_story(
        self,
        root: Path,
        *,
        story_id: str = "STORY-TOS-a040",
        change_id: str = "a040-example-change",
    ) -> None:
        portfolio = root / "docs" / "pm" / "portfolio"
        old_path, story = self.story(root)
        old_path.unlink()
        story["id"] = story_id
        story["openspec"] = {
            "change_id": change_id,
            "path": f"openspec/changes/{change_id}",
        }
        self.write_json(portfolio / "stories" / f"{story_id}.yaml", story)

        registry_path = portfolio / "_registry.yaml"
        registry = json.loads(registry_path.read_text(encoding="utf-8"))
        registry["stories"] = [story_id]
        self.write_json(registry_path, registry)

        feature_path = portfolio / "features" / "FEAT-TOS-001.yaml"
        feature = json.loads(feature_path.read_text(encoding="utf-8"))
        feature["stories"] = [story_id]
        self.write_json(feature_path, feature)

        old_change = root / "openspec" / "changes" / "example-change"
        new_change = root / "openspec" / "changes" / change_id
        old_change.rename(new_change)

    def test_current_repository_is_valid(self):
        self.assertEqual(tracker.validate(), [])

    def test_numbered_story_and_change_with_same_number_are_valid(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            self.fixture(root)
            self.replace_with_numbered_story(root)
            self.assertEqual(tracker.validate(root), [])

    def test_numbered_story_change_mismatch_is_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            self.fixture(root)
            self.replace_with_numbered_story(
                root,
                story_id="STORY-TOS-a040",
                change_id="a041-example-change",
            )
            errors = tracker.validate(root)
            self.assertTrue(any("number mismatch" in error for error in errors))

    def test_numbered_change_requires_kebab_intent(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            self.fixture(root)
            self.replace_with_numbered_story(
                root,
                story_id="STORY-TOS-a040",
                change_id="a040_bad_slug",
            )
            errors = tracker.validate(root)
            self.assertTrue(any("invalid numbered change id" in error for error in errors))

    def test_duplicate_numbered_change_number_is_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            self.fixture(root)
            self.replace_with_numbered_story(root)
            duplicate = root / "openspec" / "changes" / "a040-other-change"
            duplicate.mkdir(parents=True)
            (duplicate / "proposal.md").write_text("# proposal\n", encoding="utf-8")
            errors = tracker.validate(root)
            self.assertTrue(any("duplicate numbered change a040" in error for error in errors))

    def test_numeric_legacy_story_is_rejected_at_or_after_cutover(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            self.fixture(root)
            self.replace_with_numbered_story(
                root,
                story_id="STORY-TOS-040",
                change_id="new-legacy-style-change",
            )
            errors = tracker.validate(root)
            self.assertTrue(any("legacy numeric Story id" in error for error in errors))

    def test_numbered_story_below_cutover_is_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            self.fixture(root)
            self.replace_with_numbered_story(
                root,
                story_id="STORY-TOS-a001",
                change_id="a001-reused-number",
            )
            errors = tracker.validate(root)
            self.assertTrue(any("below cutoff a040" in error for error in errors))

    def test_bootstrap_allowlist_is_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            self.fixture(root)
            path = root / "docs" / "pm" / "portfolio" / "_registry.yaml"
            registry = json.loads(path.read_text(encoding="utf-8"))
            registry["bootstrap_change_allowlist"] = ["orphan-change"]
            self.write_json(path, registry)
            errors = tracker.validate(root)
            self.assertTrue(any("bootstrap" in error for error in errors))

    def test_orphan_active_change_is_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            self.fixture(root)
            path = root / "openspec" / "changes" / "orphan-change"
            path.mkdir(parents=True)
            (path / "proposal.md").write_text("# proposal\n", encoding="utf-8")
            errors = tracker.validate(root)
            self.assertTrue(any("orphan-change" in error for error in errors))

    def test_duplicate_story_mapping_is_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            self.fixture(root)
            portfolio = root / "docs" / "pm" / "portfolio"
            _, duplicate = self.story(root)
            duplicate["id"] = "STORY-TOS-002"
            self.write_json(
                portfolio / "stories" / "STORY-TOS-002.yaml",
                duplicate,
            )
            registry_path = portfolio / "_registry.yaml"
            registry = json.loads(registry_path.read_text(encoding="utf-8"))
            registry["stories"].append("STORY-TOS-002")
            self.write_json(registry_path, registry)
            feature_path = portfolio / "features" / "FEAT-TOS-001.yaml"
            feature = json.loads(feature_path.read_text(encoding="utf-8"))
            feature["stories"].append("STORY-TOS-002")
            self.write_json(feature_path, feature)
            errors = tracker.validate(root)
            self.assertTrue(any("expected one story" in error for error in errors))

    def test_invalid_openspec_path_is_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            self.fixture(root)
            path, story = self.story(root)
            story["openspec"]["path"] = "openspec/changes/not-the-change"
            self.write_json(path, story)
            errors = tracker.validate(root)
            self.assertTrue(any("invalid OpenSpec path" in error for error in errors))

    def test_manual_story_status_is_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            self.fixture(root)
            path, story = self.story(root)
            story["status"] = "done"
            self.write_json(path, story)
            errors = tracker.validate(root)
            self.assertTrue(any("manual status" in error for error in errors))

    def test_story_status_is_derived_from_tasks_and_archive(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            self.fixture(root)
            rendered = tracker.render(root)["01-active-change-map.md"]
            self.assertIn("| `example-change` | STORY-TOS-001 | designed |", rendered)

            tasks = root / "openspec" / "changes" / "example-change" / "tasks.md"
            tasks.write_text("- [ ] implement\n", encoding="utf-8")
            rendered = tracker.render(root)["01-active-change-map.md"]
            self.assertIn("| `example-change` | STORY-TOS-001 | in_progress |", rendered)

            tasks.write_text("- [x] implement\n", encoding="utf-8")
            rendered = tracker.render(root)["01-active-change-map.md"]
            self.assertIn("| `example-change` | STORY-TOS-001 | implemented |", rendered)

            active = root / "openspec" / "changes" / "example-change"
            archived = (
                root
                / "openspec"
                / "changes"
                / "archive"
                / "2026-07-31-example-change"
            )
            archived.parent.mkdir(parents=True)
            active.rename(archived)
            story_path, story = self.story(root)
            story["openspec"]["path"] = (
                "openspec/changes/archive/2026-07-31-example-change"
            )
            self.write_json(story_path, story)
            self.assertEqual(tracker.validate(root), [])
            rendered = tracker.render(root)["01-active-change-map.md"]
            self.assertIn("| `example-change` | STORY-TOS-001 | archived |", rendered)

    def test_reverse_link_and_forward_orphan_are_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            self.fixture(root)
            portfolio = root / "docs" / "pm" / "portfolio"
            epic_path = portfolio / "epics" / "EPIC-TOS-001.yaml"
            epic = json.loads(epic_path.read_text(encoding="utf-8"))
            epic["initiative"] = "INIT-MISSING"
            self.write_json(epic_path, epic)
            feature_path = portfolio / "features" / "FEAT-TOS-001.yaml"
            feature = json.loads(feature_path.read_text(encoding="utf-8"))
            feature["stories"].append("STORY-MISSING")
            self.write_json(feature_path, feature)
            errors = tracker.validate(root)
            self.assertTrue(any("initiative reverse link" in error for error in errors))
            self.assertTrue(any("invalid story forward link" in error for error in errors))

    def test_stale_generated_tracker_is_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            self.fixture(root)
            errors = tracker.generate(root, check=True)
            self.assertTrue(any("generated tracker stale" in error for error in errors))


if __name__ == "__main__":
    unittest.main()
