"""Codex 호스트에서 확립된 PostToolUse 이름과 저장 훅 matcher 의 대응을 고정하는 회귀 시험.

이름은 추측하지 않고 살균 픽스처(`fixtures/codex_post_tool_use_names.json`)에서만 읽음.
픽스처의 근거는 a119 `analysis/host-evidence.md` 에 있음. 이 시험의 통과는 matcher 가 그
이름들을 덮는다는 뜻일 뿐, 호스트가 실제로 훅을 전달했다는 관측이 아님(스펙 시나리오
"Host coverage has not been observed").
"""

import fcntl
import json
import re
import shutil
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
# 픽스처로 확립되지 않은 이름의 대표: host-evidence §2.2 의 모델 쪽 도구 이름과 SDD 핸들러 이름.
# 픽스처에 새로 확립되면 그 이름은 자동으로 이 음성 표본에서 빠짐.
UNESTABLISHED_PROBES = (
    "exec",
    "send_message",
    "spawn_agent",
    "mcp__codegraph__codegraph_explore",
    "Write",
    "Edit",
    "MultiEdit",
    "NotebookEdit",
)


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
        # 호스트가 부분 일치(search)로 판정해도 전체 일치와 같은 결과가 나오는지 의미로 확인함.
        # 글자로 `^`·`$` 만 보면 `^Bash|apply_patch$`(우선순위로 고정이 풀림)나 `.*` 추가가 통과함.
        pattern = re.compile(self.matcher)
        for name in self.names:
            for probe in (f"x{name}", f"{name}x"):
                with self.subTest(probe=probe):
                    self.assertIsNone(pattern.search(probe), self.matcher)

    def test_matcher_admits_no_name_the_fixture_has_not_established(self):
        # 비목표 "픽스처로 확립되지 않은 이름 추가 금지"의 고정: 미확립 이름은 부분 일치로도 안 걸려야 함.
        pattern = re.compile(self.matcher)
        for name in UNESTABLISHED_PROBES:
            if name in self.names:
                continue
            with self.subTest(name=name):
                self.assertIsNone(pattern.search(name), self.matcher)


class CodexSaverToolResultTests(unittest.TestCase):
    """저장기는 세 출구(성공·예외 경고·락 경합) 모두에서 stdout 을 비워 둬야 함.

    PostToolUse 훅의 stdout 을 호스트가 판정으로 읽을 수 있다는 것은 예방적 가정임(Codex 의
    async 훅에서 관측된 적 없음). 비어 있는 stdout 은 어느 경우에도 도구 결과를 바꾸지 않음.
    """

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
                # 이름마다 저장소를 비워서 파일 존재 확인이 앞 반복의 산출물로 통과하지 않게 함.
                shutil.rmtree(self.project / ".codex-context", ignore_errors=True)
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

    def test_lock_contention_exit_leaves_stdout_empty(self):
        # 락 경합 출구: async 훅이 겹치면 뒤의 저장기는 저장 없이 즉시 끝나고 stdout 도 비워야 함.
        context = self.project / ".codex-context"
        context.mkdir()
        with (context / ".save-session.lock").open("a+b") as lock:
            fcntl.flock(lock.fileno(), fcntl.LOCK_EX | fcntl.LOCK_NB)
            result = self.run_saver(json.dumps({"cwd": str(self.project)}))
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stdout, "")
        # 경합 출구를 실제로 탔음을 확인함: 저장 산출물이 없고, 예외 경고 출구(stderr 경고)도 아니어야 함.
        self.assertFalse((context / "session-summary.md").exists())
        self.assertEqual(result.stderr, "")


if __name__ == "__main__":
    unittest.main()
