import tempfile
import unittest
import json
from pathlib import Path

import check_agent_config_sync as sync


class AgentConfigTests(unittest.TestCase):
    def test_repository_contract_is_synchronized(self):
        self.assertEqual(sync.check(), [])

    def test_drift_is_detected(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            for relative in sync.REQUIRED_PATHS:
                path = root / relative
                path.parent.mkdir(parents=True, exist_ok=True)
                path.write_text("", encoding="utf-8")
            source = (sync.ROOT / sync.SOURCE).read_text(encoding="utf-8")
            source_path = root / sync.SOURCE
            source_path.parent.mkdir(parents=True, exist_ok=True)
            source_path.write_text(source, encoding="utf-8")
            mirror_path = root / sync.MIRROR
            mirror_path.parent.mkdir(parents=True, exist_ok=True)
            mirror_path.write_text(sync.mirror_text(source).replace("RED → GREEN", "RED → BROKEN", 1), encoding="utf-8")
            (root / "CLAUDE.md").write_text(".claude/CLAUDE.md", encoding="utf-8")
            (root / "AGENTS.md").write_text(".claude/CLAUDE.md", encoding="utf-8")
            errors = sync.check(root)
            self.assertTrue(any("drift" in error for error in errors))

    def test_broad_mutating_or_mcp_permissions_are_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            for relative in sync.REQUIRED_PATHS:
                path = root / relative
                path.parent.mkdir(parents=True, exist_ok=True)
                path.write_text("", encoding="utf-8")
            source = (sync.ROOT / sync.SOURCE).read_text(encoding="utf-8")
            source_path = root / sync.SOURCE
            source_path.parent.mkdir(parents=True, exist_ok=True)
            source_path.write_text(source, encoding="utf-8")
            mirror_path = root / sync.MIRROR
            mirror_path.parent.mkdir(parents=True, exist_ok=True)
            mirror_path.write_text(sync.mirror_text(source), encoding="utf-8")
            (root / "CLAUDE.md").write_text(".claude/CLAUDE.md", encoding="utf-8")
            (root / "AGENTS.md").write_text(".claude/CLAUDE.md", encoding="utf-8")
            settings = root / ".claude" / "settings.json"
            settings.parent.mkdir(parents=True, exist_ok=True)
            settings.write_text(
                json.dumps({"permissions": {"allow": ["Bash(openspec *)"]}}),
                encoding="utf-8",
            )
            errors = sync.check(root)
            self.assertTrue(any("broad agent auto-approval" in error for error in errors))


if __name__ == "__main__":
    unittest.main()
