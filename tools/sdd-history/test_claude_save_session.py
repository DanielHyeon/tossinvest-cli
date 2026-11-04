#!/usr/bin/env python3
"""Claude 세션 저장 훅의 호출 설정과 단일 실행·간격 제한·요약 보존 동작을 검증함."""

from __future__ import annotations

import fcntl
import json
import os
import shutil
import subprocess
import tempfile
import time
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "save-session.sh"
SETTINGS = ROOT / ".claude" / "settings.json"
MEMO_HEADING = "## 수동 메모 (세션 종료 전 직접 기록)\n"


class SaveSessionScriptTest(unittest.TestCase):
    def setUp(self):
        # 실제 저장소·홈을 건드리지 않도록 임시 프로젝트에 스크립트 사본을 두고 실행함
        self.tmp = Path(tempfile.mkdtemp())
        self.addCleanup(shutil.rmtree, self.tmp)
        self.project = self.tmp / "project"
        self.project.mkdir()
        shutil.copy(SCRIPT, self.project / "save-session.sh")
        (self.project / ".gitignore").write_text(".ai-context/\nsave-session.sh\n", encoding="utf-8")
        subprocess.run(["git", "init", "-q"], cwd=self.project, check=True)
        self.context = self.project / ".ai-context"
        self.context.mkdir()
        self.summary = self.context / "session-summary.md"
        self.env = {**os.environ, "HOME": str(self.tmp / "home")}

    def run_script(self, *args: str, check: bool = True, env: dict | None = None) -> subprocess.CompletedProcess:
        # 훅은 작업 폴더가 어디든 불릴 수 있으므로 프로젝트 밖에서 실행함
        return subprocess.run(
            ["bash", str(self.project / "save-session.sh"), *args],
            cwd=self.tmp,
            env=env or self.env,
            capture_output=True,
            text=True,
            check=check,
        )

    def age_summary(self, seconds: int) -> None:
        then = time.time() - seconds
        os.utime(self.summary, (then, then))

    def saved_text(self) -> str:
        return self.summary.read_text(encoding="utf-8")

    def section(self, start: str, end: str) -> str:
        return self.saved_text().split(start, 1)[1].split(end, 1)[0]

    def write_memo(self, memo: str) -> None:
        self.run_script()
        self.summary.write_text(self.saved_text().replace(MEMO_HEADING, MEMO_HEADING + memo + "\n", 1), encoding="utf-8")

    def test_hook_skips_fresh_summary_but_manual_run_saves(self):
        self.summary.write_text("old\n", encoding="utf-8")
        # 나이는 정수초 차이이므로 300초에 바짝 붙이면 초 경계를 넘을 때 무작위로 실패함
        self.age_summary(290)
        self.run_script("--hook")
        self.assertEqual(self.saved_text(), "old\n")
        self.run_script()
        self.assertIn("이전 세션 컨텍스트", self.saved_text())

    def test_hook_refreshes_stale_summary_and_keeps_backup(self):
        self.summary.write_text("old\n", encoding="utf-8")
        self.age_summary(301)
        self.run_script("--hook")
        self.assertIn("이전 세션 컨텍스트", self.saved_text())
        backups = list(self.context.glob("session-summary_*.md"))
        self.assertEqual(len(backups), 1)
        self.assertEqual(backups[0].read_text(encoding="utf-8"), "old\n")

    def test_hook_skips_at_once_while_another_save_runs(self):
        self.summary.write_text("old\n", encoding="utf-8")
        self.age_summary(3600)
        with open(self.context / ".save-session.lock", "w", encoding="utf-8") as lock:
            fcntl.flock(lock, fcntl.LOCK_EX)
            started = time.monotonic()
            self.run_script("--hook")
            elapsed = time.monotonic() - started
        self.assertEqual(self.saved_text(), "old\n")
        # 기다리지 않아야 비동기 실행이 줄지어 쌓이지 않음
        self.assertLess(elapsed, 2)

    def test_manual_run_waits_for_running_save_then_saves(self):
        self.summary.write_text("old\n", encoding="utf-8")
        with open(self.context / ".save-session.lock", "w", encoding="utf-8") as lock:
            fcntl.flock(lock, fcntl.LOCK_EX)
            process = subprocess.Popen(
                ["bash", str(self.project / "save-session.sh")],
                cwd=self.tmp,
                env=self.env,
                stdout=subprocess.DEVNULL,
            )
            self.addCleanup(process.kill)
            time.sleep(1.5)
            # 잠금을 잡고 있는 동안은 끝나지도, 요약을 쓰지도 않아야 함
            self.assertIsNone(process.poll())
            self.assertEqual(self.saved_text(), "old\n")
        self.assertEqual(process.wait(timeout=30), 0)
        self.assertIn("이전 세션 컨텍스트", self.saved_text())

    def test_lock_timeout_fails_loudly_and_leaves_summary(self):
        self.summary.write_text("old\n", encoding="utf-8")
        env = {**self.env, "SAVE_SESSION_LOCK_WAIT_SEC": "1"}
        with open(self.context / ".save-session.lock", "w", encoding="utf-8") as lock:
            fcntl.flock(lock, fcntl.LOCK_EX)
            result = self.run_script(check=False, env=env)
        self.assertEqual(result.returncode, 1)
        self.assertIn("저장 실패", result.stderr)
        self.assertEqual(self.saved_text(), "old\n")
        # 자동 실행은 표준 오류가 버려지므로 실패 기록이 파일에 남아야 함
        self.assertIn("저장 실패", (self.context / "save-failures.log").read_text(encoding="utf-8"))

    def test_default_lock_wait_is_the_specified_value(self):
        self.assertIn('LOCK_WAIT_SEC="${SAVE_SESSION_LOCK_WAIT_SEC:-120}"', SCRIPT.read_text(encoding="utf-8"))

    def test_saving_does_not_rewrite_the_git_index(self):
        # git status 가 index 를 갱신하면 index.lock 을 잡아 동시에 도는 git add·commit 을 실패시킴
        tracked = self.project / "tracked.txt"
        tracked.write_text("x\n", encoding="utf-8")
        identity = ["-c", "user.email=t@example.invalid", "-c", "user.name=t"]
        subprocess.run(["git", "add", "tracked.txt"], cwd=self.project, check=True)
        subprocess.run(["git", *identity, "commit", "-qm", "seed"], cwd=self.project, check=True)
        then = time.time() - 7200
        os.utime(tracked, (then, then))
        index = self.project / ".git" / "index"
        before = index.read_bytes()
        self.run_script()
        self.assertEqual(index.read_bytes(), before)

    def test_write_failure_keeps_previous_summary_and_memo(self):
        self.write_memo("MY_MEMO_NOTE " + "x" * 4000)
        before = self.saved_text()
        # 파일 크기 제한으로 디스크 가득 참과 같은 쓰기 실패를 만듦
        command = f'ulimit -f 1; exec bash "{self.project / "save-session.sh"}"'
        result = subprocess.run(["bash", "-c", command], cwd=self.tmp, env=self.env, capture_output=True, text=True)
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(self.saved_text(), before)

    def test_unwritable_directory_fails_and_keeps_summary(self):
        self.run_script()
        before = self.saved_text()
        os.chmod(self.context, 0o555)
        self.addCleanup(os.chmod, self.context, 0o755)
        result = self.run_script(check=False)
        self.assertEqual(result.returncode, 1)
        self.assertEqual(self.saved_text(), before)

    def test_no_temporary_file_is_left_after_saving(self):
        self.run_script()
        self.run_script()
        self.assertEqual(list(self.context.glob(".session-summary.*")), [])

    def test_git_status_is_capped_and_reports_omitted_count(self):
        for index in range(300):
            (self.project / f"untracked-{index:03d}.txt").write_text("x\n", encoding="utf-8")
        self.run_script()
        status = self.section("변경된 파일)\n```\n", "\n```").splitlines()
        # .gitignore 까지 미추적 301줄이므로 상한 200줄 + 생략 안내 1줄이 남음
        self.assertEqual(len(status), 201)
        self.assertEqual(status[-1], "... 외 101개 생략")
        self.assertTrue(all(line.startswith("?? ") for line in status[:200]))

    def test_archive_is_not_walked_for_recent_files(self):
        archive = self.project / "openspec" / "changes" / "archive"
        archive.mkdir(parents=True)
        (archive / "archived-evidence.md").write_text("x\n", encoding="utf-8")
        (self.project / "recent-work.md").write_text("x\n", encoding="utf-8")
        self.run_script()
        recent = self.section("최근 30분 내 수정된 파일", "수동 메모")
        self.assertIn("recent-work.md", recent)
        # -print 가 빠지면 파일이 아닌 archive 디렉터리 줄이 찍히므로 경로 조각까지 함께 확인함
        self.assertNotIn("openspec/changes/archive", recent)
        # 걸러 내기만 하면 수십만 파일을 그대로 돌므로 prune 자체를 고정함
        self.assertIn('-path "$PROJECT_DIR/openspec/changes/archive" -prune -o', SCRIPT.read_text(encoding="utf-8"))

    def test_manual_memo_survives_repeated_saves(self):
        self.run_script()
        self.assertIn("여기에 현재 하고 있던 작업", self.saved_text())
        self.write_memo("MY_MEMO_NOTE")
        self.run_script()
        self.run_script()
        text = self.saved_text()
        self.assertEqual(text.count("MY_MEMO_NOTE"), 1)
        # 안내문도 메모 본문으로 옮겨지므로 저장을 거듭해도 늘어나지 않아야 함
        self.assertEqual(text.count("여기에 현재 하고 있던 작업"), 1)

    def test_memo_line_resembling_a_heading_does_not_cut_the_memo(self):
        self.write_memo("FIRST\n## Claude Code 최근 대화 정리 메모\nLAST")
        self.run_script()
        memo = self.section(MEMO_HEADING, "## Claude Code 최근 대화 — ")
        self.assertIn("FIRST", memo)
        self.assertIn("LAST", memo)


class ClaudeHookSettingsTest(unittest.TestCase):
    def setUp(self):
        hooks = json.loads(SETTINGS.read_text(encoding="utf-8"))["hooks"]
        self.post_groups = hooks["PostToolUse"]
        self.stop_groups = hooks.get("Stop", [])

    def only_hook(self, groups: list, needle: str) -> tuple[dict, dict]:
        found = [
            (group, hook)
            for group in groups
            for hook in group["hooks"]
            if needle in hook.get("command", "")
        ]
        self.assertEqual(len(found), 1, needle)
        return found[0]

    def test_save_session_hook_finds_script_from_any_cwd_without_blocking(self):
        group, hook = self.only_hook(self.post_groups, "save-session.sh")
        # 실행 비트가 없는 체크아웃에서도 돌도록 bash 로 직접 실행함
        self.assertEqual(hook["command"], 'bash "$CLAUDE_PROJECT_DIR/save-session.sh" --hook > /dev/null 2>&1')
        self.assertIs(hook.get("async"), True)
        self.assertEqual(group["matcher"], "Edit|Write|Bash")

    def test_turn_end_saves_last_state_that_interval_limit_skipped(self):
        _, hook = self.only_hook(self.stop_groups, "save-session.sh")
        # 턴 종료 저장은 간격 제한을 받지 않아야 마지막 작업이 요약에 남음
        self.assertEqual(hook["command"], 'bash "$CLAUDE_PROJECT_DIR/save-session.sh" > /dev/null 2>&1')
        self.assertIs(hook.get("async"), True)

    def test_agent_save_hook_finds_script_from_any_cwd(self):
        _, hook = self.only_hook(self.post_groups, "agent_save_hook.py")
        self.assertEqual(hook["command"], 'python3 "$CLAUDE_PROJECT_DIR/tools/sdd-history/agent_save_hook.py"')


if __name__ == "__main__":
    unittest.main()
