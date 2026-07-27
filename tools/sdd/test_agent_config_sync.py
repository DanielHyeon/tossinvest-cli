import json
import tempfile
import unittest
from pathlib import Path

import check_agent_config_sync as sync


class AgentConfigTests(unittest.TestCase):
    def make_root(self, root: Path) -> tuple[Path, Path]:
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
        routing = f"{sync.SOURCE}\n{sync.WORKFLOW}\n"
        (root / "CLAUDE.md").write_text(routing, encoding="utf-8")
        (root / "AGENTS.md").write_text(routing, encoding="utf-8")
        settings = root / ".claude" / "settings.json"
        settings.parent.mkdir(parents=True, exist_ok=True)
        settings.write_text('{"permissions":{"allow":[]}}', encoding="utf-8")
        return source_path, mirror_path

    def test_repository_contract_is_synchronized(self):
        self.assertEqual(sync.check(), [])

    def test_bootstrap_drift_is_detected(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            _, mirror_path = self.make_root(root)
            mirror_path.write_text(
                mirror_path.read_text(encoding="utf-8").replace(
                    "사람 승인 없는 LIVE 주문", "승인 계약 누락", 1
                ),
                encoding="utf-8",
            )
            errors = sync.check(root)
            self.assertTrue(any("drift" in error for error in errors))

    def test_workflow_pointer_is_required_in_both_entry_files(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            _, mirror_path = self.make_root(root)
            mirror_path.write_text(
                mirror_path.read_text(encoding="utf-8").replace(
                    sync.WORKFLOW, "missing-workflow"
                ),
                encoding="utf-8",
            )
            errors = sync.check(root)
            self.assertTrue(
                any(
                    sync.MIRROR in error and "does not route" in error
                    for error in errors
                )
            )

    def test_required_bootstrap_contract_is_enforced(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            source_path, mirror_path = self.make_root(root)
            source = source_path.read_text(encoding="utf-8").replace(
                "운영 토글 flip", "운영 변경", 1
            )
            source_path.write_text(source, encoding="utf-8")
            mirror_path.write_text(sync.mirror_text(source), encoding="utf-8")
            errors = sync.check(root)
            self.assertTrue(
                any("required bootstrap contract missing" in error for error in errors)
            )

    def test_broad_mutating_or_mcp_permissions_are_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            self.make_root(root)
            settings = root / ".claude" / "settings.json"
            settings.write_text(
                json.dumps({"permissions": {"allow": ["Bash(openspec *)"]}}),
                encoding="utf-8",
            )
            errors = sync.check(root)
            self.assertTrue(
                any("broad agent auto-approval" in error for error in errors)
            )


if __name__ == "__main__":
    unittest.main()
