"""Codex 호스트에서 확립된 PostToolUse 이름과 저장 훅 matcher 의 대응을 고정하는 회귀 시험.

이름은 추측하지 않고 살균 픽스처(`fixtures/codex_post_tool_use_names.json`)에서만 읽음.
픽스처의 근거는 a119 `analysis/host-evidence.md` 에 있음. 이 시험의 통과는 matcher 가 그
이름들을 덮는다는 뜻일 뿐, 호스트가 실제로 훅을 전달했다는 관측이 아님(스펙 시나리오
"Host coverage has not been observed").
"""

import json
import re
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
HOOKS = ROOT / ".codex" / "hooks.json"
SCRIPT = ROOT / ".codex" / "hooks" / "save_session.py"
FIXTURE = Path(__file__).resolve().parent / "fixtures" / "codex_post_tool_use_names.json"
SAVER_SUFFIX = '/.codex/hooks/save_session.py"'
FIXTURE_KEYS = {"schema", "host", "post_tool_use_names"}
EVIDENCE_KINDS = {"delivery-inferred", "binary-constant"}
TOOL_NAME = re.compile(r"[A-Za-z_][A-Za-z0-9_]*")
UUID_LIKE = re.compile(r"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", re.I)


def load_fixture() -> dict:
    return json.loads(FIXTURE.read_text(encoding="utf-8"))


def saver_matcher(config: dict) -> str:
    """Codex 저장기를 부르는 그룹이 정확히 하나일 때 그 matcher 를 반환함."""
    groups = [
        group
        for group in config["hooks"]["PostToolUse"]
        if any(
            handler.get("command", "").endswith(SAVER_SUFFIX)
            for handler in group.get("hooks", [])
        )
    ]
    if len(groups) != 1:
        raise AssertionError(f"Codex saver group count is {len(groups)}, expected 1")
    return groups[0]["matcher"]


def uncovered_names(matcher: str, names: list[str]) -> list[str]:
    """matcher 가 전체 일치로 받지 못하는 호스트 이름 목록을 반환함."""
    pattern = re.compile(matcher)
    return [name for name in names if pattern.fullmatch(name) is None]


class CodexHostEventFixtureTests(unittest.TestCase):
    def test_fixture_carries_only_sanitized_names_and_evidence_kinds(self):
        fixture = load_fixture()
        self.assertEqual(set(fixture), FIXTURE_KEYS)
        self.assertEqual(fixture["schema"], 1)
        self.assertEqual(set(fixture["host"]), {"product", "version"})
        entries = fixture["post_tool_use_names"]
        self.assertTrue(entries)
        for entry in entries:
            self.assertEqual(set(entry), {"name", "evidence"})
            self.assertIsNotNone(TOOL_NAME.fullmatch(entry["name"]), entry)
            self.assertTrue(entry["evidence"])
            self.assertLessEqual(set(entry["evidence"]), EVIDENCE_KINDS)
        # 살균 확인: 경로·세션 ID 모양 문자열이 픽스처에 들어오지 않았음을 확인함.
        raw = FIXTURE.read_text(encoding="utf-8")
        self.assertNotIn("/", raw)
        self.assertIsNone(UUID_LIKE.search(raw))


class CodexHostEventCoverageTests(unittest.TestCase):
    def setUp(self):
        self.names = [entry["name"] for entry in load_fixture()["post_tool_use_names"]]
        self.matcher = saver_matcher(json.loads(HOOKS.read_text(encoding="utf-8")))

    def test_every_established_host_name_schedules_the_saver(self):
        # 정본 시나리오의 두 이름은 픽스처에서 빠질 수 없음.
        self.assertIn("Bash", self.names)
        self.assertIn("apply_patch", self.names)
        self.assertEqual(uncovered_names(self.matcher, self.names), [])

    def test_matcher_is_anchored_so_host_match_mode_cannot_widen_it(self):
        # 호스트가 부분 일치로 판정해도 전체 일치와 같은 결과가 나오도록 양끝 고정을 요구함.
        self.assertTrue(self.matcher.startswith("^"), self.matcher)
        self.assertTrue(self.matcher.endswith("$"), self.matcher)


class CodexSaverToolResultTests(unittest.TestCase):
    """PostToolUse 훅의 stdout 은 호스트가 판정 JSON 으로 읽으므로 저장기는 비워 둬야 함."""

    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.project = Path(self.temporary.name)
        for args in (
            ("init", "-q"),
            ("config", "user.email", "codex-test@example.invalid"),
            ("config", "user.name", "Codex Test"),
        ):
            subprocess.run(["git", "-C", str(self.project), *args], check=True)
        (self.project / "README.md").write_text("fixture\n", encoding="utf-8")
        subprocess.run(["git", "-C", str(self.project), "add", "README.md"], check=True)
        subprocess.run(
            ["git", "-C", str(self.project), "commit", "-qm", "test: initialize"],
            check=True,
        )

    def tearDown(self):
        self.temporary.cleanup()

    def run_saver(self, hook_input: str) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            [sys.executable, str(SCRIPT)],
            cwd=self.project,
            input=hook_input,
            capture_output=True,
            text=True,
            timeout=10,
            check=False,
        )

    def test_saver_writes_nothing_to_stdout_for_each_established_name(self):
        for entry in load_fixture()["post_tool_use_names"]:
            with self.subTest(name=entry["name"]):
                result = self.run_saver(
                    json.dumps(
                        {
                            "hook_event_name": "PostToolUse",
                            "tool_name": entry["name"],
                            "session_id": "fixture-session",
                            "cwd": str(self.project),
                        }
                    )
                )
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertEqual(result.stdout, "")
                self.assertTrue(
                    (self.project / ".codex-context" / "session-summary.md").is_file()
                )

    def test_failed_save_still_leaves_stdout_empty(self):
        # 실패 경로: 저장소가 심볼릭 링크이면 저장을 건너뛰되 경고는 stderr 로만 보냄.
        target = self.project / "elsewhere"
        target.mkdir()
        (self.project / ".codex-context").symlink_to(target, target_is_directory=True)
        result = self.run_saver(json.dumps({"cwd": str(self.project)}))
        self.assertEqual(result.returncode, 0)
        self.assertEqual(result.stdout, "")
        self.assertIn("save skipped", result.stderr)
        self.assertEqual(list(target.iterdir()), [])


if __name__ == "__main__":
    unittest.main()
