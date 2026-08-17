import fcntl
import json
import re
import stat
import subprocess
import sys
import tempfile
import time
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
SCRIPT = ROOT / ".codex" / "hooks" / "save_session.py"
HOOKS = ROOT / ".codex" / "hooks.json"
GITIGNORE = ROOT / ".gitignore"


def response_message(role: str, text: str) -> dict:
    content_type = "input_text" if role == "user" else "output_text"
    return {
        "type": "response_item",
        "payload": {
            "type": "message",
            "role": role,
            "content": [{"type": content_type, "text": text}],
        },
    }


class CodexSessionSaveTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.project = Path(self.temporary.name)
        self.git("init", "-q")
        self.git("config", "user.email", "codex-test@example.invalid")
        self.git("config", "user.name", "Codex Test")
        (self.project / "README.md").write_text("fixture\n", encoding="utf-8")
        self.git("add", "README.md")
        self.git("commit", "-qm", "test: initialize fixture")

    def tearDown(self):
        self.temporary.cleanup()

    def git(self, *args: str) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            ["git", "-C", str(self.project), *args],
            check=True,
            capture_output=True,
            text=True,
        )

    def run_saver(
        self,
        payload: dict | None = None,
        *,
        raw_input: str | None = None,
        timeout: float = 5,
    ) -> subprocess.CompletedProcess[str]:
        hook_input = raw_input if raw_input is not None else json.dumps(payload or {})
        return subprocess.run(
            [sys.executable, str(SCRIPT)],
            cwd=self.project,
            input=hook_input,
            capture_output=True,
            text=True,
            timeout=timeout,
            check=False,
        )

    def summary(self) -> str:
        return (self.project / ".codex-context" / "session-summary.md").read_text(
            encoding="utf-8"
        )

    def test_writes_only_codex_store_and_refuses_claude_owned_transcript(self):
        claude_dir = self.project / ".claude"
        ai_context = self.project / ".ai-context"
        claude_dir.mkdir()
        ai_context.mkdir()
        claude_settings = claude_dir / "settings.json"
        claude_transcript = claude_dir / "rollout.jsonl"
        claude_summary = ai_context / "session-summary.md"
        claude_script = self.project / "save-session.sh"
        claude_settings.write_text('{"owner":"claude"}\n', encoding="utf-8")
        claude_transcript.write_text(
            json.dumps(response_message("user", "initial")) + "\n",
            encoding="utf-8",
        )
        claude_summary.write_text("claude summary\n", encoding="utf-8")
        claude_script.write_text("#!/bin/sh\n# claude owner\n", encoding="utf-8")
        self.git(
            "add",
            ".claude/settings.json",
            ".claude/rollout.jsonl",
            ".ai-context/session-summary.md",
            "save-session.sh",
        )
        self.git("commit", "-qm", "test: add other-agent fixtures")
        claude_settings.write_text('{"owner":"CLAUDE_SETTINGS_MARKER"}\n', encoding="utf-8")
        claude_transcript.write_text(
            json.dumps(response_message("user", "CLAUDE_PRIVATE_MARKER")) + "\n",
            encoding="utf-8",
        )
        claude_summary.write_text("AI_CONTEXT_PRIVATE_MARKER\n", encoding="utf-8")
        claude_script.write_text("#!/bin/sh\n# CLAUDE_SCRIPT_MARKER\n", encoding="utf-8")
        before = {
            path: path.read_bytes()
            for path in (claude_settings, claude_transcript, claude_summary, claude_script)
        }

        result = self.run_saver(
            {
                "session_id": "codex-session",
                "model": "gpt-test",
                "cwd": str(self.project),
                "transcript_path": str(claude_transcript),
            }
        )

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertTrue((self.project / ".codex-context" / "session-summary.md").is_file())
        summary = self.summary()
        self.assertNotIn("CLAUDE_PRIVATE_MARKER", summary)
        self.assertNotIn("CLAUDE_SETTINGS_MARKER", summary)
        self.assertNotIn("AI_CONTEXT_PRIVATE_MARKER", summary)
        self.assertNotIn("CLAUDE_SCRIPT_MARKER", summary)
        self.assertNotIn(".claude", summary)
        self.assertNotIn(".ai-context", summary)
        self.assertNotIn("save-session.sh", summary)
        for path, content in before.items():
            self.assertEqual(path.read_bytes(), content)

    def test_extracts_only_bounded_user_assistant_text_and_redacts_secrets(self):
        transcript = self.project / "rollout.jsonl"
        records = [response_message("user", f"old-message-{index}") for index in range(35)]
        records.extend(
            [
                response_message(
                    "assistant",
                    "safe answer api_key=topsecret-value "
                    "Bearer abcdefghijklmnopqrstuvwxyz0123456789 "
                    "sk-proj-abcdefghijklmnopqrstuvwxyz0123456789 "
                    "OPENAI_API_KEY=openai-plain-value "
                    "GITHUB_TOKEN=github-plain-value "
                    "NEO4J_PASSWORD=neo4j-plain-value "
                    "AWS_SECRET_ACCESS_KEY=aws-plain-value "
                    '"api_key": "json-topsecret-value" '
                    "'password': 'hunter2'",
                ),
                response_message("developer", "DEVELOPER_PRIVATE_MARKER"),
                {
                    "type": "response_item",
                    "payload": {"type": "reasoning", "summary": "REASONING_PRIVATE_MARKER"},
                },
                {
                    "type": "response_item",
                    "payload": {
                        "type": "function_call",
                        "name": "Bash",
                        "arguments": "TOOL_INPUT_PRIVATE_MARKER",
                    },
                },
                {
                    "type": "response_item",
                    "payload": {
                        "type": "function_call_output",
                        "output": "TOOL_OUTPUT_PRIVATE_MARKER",
                    },
                },
                response_message("user", "Z" * 10_000),
            ]
        )
        transcript.write_text(
            "\n".join(json.dumps(record) for record in records) + "\nnot-json\n",
            encoding="utf-8",
        )

        result = self.run_saver(
            {
                "session_id": "session-123",
                "model": "gpt-test",
                "cwd": str(self.project),
                "transcript_path": str(transcript),
            }
        )

        self.assertEqual(result.returncode, 0, result.stderr)
        summary = self.summary()
        self.assertIn("session-123", summary)
        self.assertIn("gpt-test", summary)
        self.assertIn("old-message-7", summary)
        self.assertNotIn("old-message-6", summary)
        self.assertIn("safe answer", summary)
        self.assertIn("[REDACTED]", summary)
        self.assertNotIn("topsecret-value", summary)
        self.assertNotIn("abcdefghijklmnopqrstuvwxyz0123456789", summary)
        self.assertIn("OPENAI_API_KEY=[REDACTED]", summary)
        self.assertIn("GITHUB_TOKEN=[REDACTED]", summary)
        self.assertIn("NEO4J_PASSWORD=[REDACTED]", summary)
        self.assertIn("AWS_SECRET_ACCESS_KEY=[REDACTED]", summary)
        self.assertNotIn("openai-plain-value", summary)
        self.assertNotIn("github-plain-value", summary)
        self.assertNotIn("neo4j-plain-value", summary)
        self.assertNotIn("aws-plain-value", summary)
        self.assertIn('"api_key": [REDACTED]', summary)
        self.assertIn("'password': [REDACTED]", summary)
        self.assertNotIn("json-topsecret-value", summary)
        self.assertNotIn("hunter2", summary)
        self.assertNotIn("DEVELOPER_PRIVATE_MARKER", summary)
        self.assertNotIn("REASONING_PRIVATE_MARKER", summary)
        self.assertNotIn("TOOL_INPUT_PRIVATE_MARKER", summary)
        self.assertNotIn("TOOL_OUTPUT_PRIVATE_MARKER", summary)
        self.assertNotIn(str(transcript), summary)
        self.assertNotIn("Z" * 3_000, summary)

    def test_malformed_stdin_fails_open_and_saves_repository_handoff(self):
        result = self.run_saver(raw_input="{not valid json")

        self.assertEqual(result.returncode, 0, result.stderr)
        summary = self.summary()
        self.assertIn("Codex Session Handoff", summary)
        self.assertIn("test: initialize fixture", summary)
        self.assertNotIn("Traceback", result.stderr)

    def test_atomic_publication_keeps_only_five_unique_backups(self):
        latest = self.project / ".codex-context" / "session-summary.md"
        for index in range(8):
            result = self.run_saver(
                {
                    "session_id": f"session-{index}",
                    "model": "gpt-test",
                    "cwd": str(self.project),
                }
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn(f"session-{index}", latest.read_text(encoding="utf-8"))

        backups = list((self.project / ".codex-context" / "backups").glob("*.md"))
        self.assertEqual(len(backups), 5)
        self.assertEqual(len({path.name for path in backups}), 5)
        context = self.project / ".codex-context"
        backup_directory = context / "backups"
        lock = context / ".save-session.lock"
        self.assertEqual(stat.S_IMODE(context.stat().st_mode), 0o700)
        self.assertEqual(stat.S_IMODE(backup_directory.stat().st_mode), 0o700)
        self.assertEqual(stat.S_IMODE(latest.stat().st_mode), 0o600)
        self.assertEqual(stat.S_IMODE(lock.stat().st_mode), 0o600)
        for backup in backups:
            self.assertEqual(stat.S_IMODE(backup.stat().st_mode), 0o600)
        self.assertEqual(
            list((self.project / ".codex-context").rglob("*.tmp")),
            [],
        )

    def test_concurrent_invocation_does_not_wait_or_overwrite(self):
        first = self.run_saver(
            {"session_id": "original", "model": "gpt-test", "cwd": str(self.project)}
        )
        self.assertEqual(first.returncode, 0, first.stderr)
        latest = self.project / ".codex-context" / "session-summary.md"
        original = latest.read_bytes()
        lock_path = self.project / ".codex-context" / ".save-session.lock"

        with lock_path.open("a+b") as lock:
            fcntl.flock(lock.fileno(), fcntl.LOCK_EX | fcntl.LOCK_NB)
            started = time.monotonic()
            second = self.run_saver(
                {"session_id": "contender", "model": "gpt-test", "cwd": str(self.project)},
                timeout=2,
            )
            elapsed = time.monotonic() - started

        self.assertEqual(second.returncode, 0, second.stderr)
        self.assertLess(elapsed, 1.5)
        self.assertEqual(latest.read_bytes(), original)


class CodexHookConfigurationTests(unittest.TestCase):
    def test_preserves_sdd_hook_and_adds_async_codex_matcher(self):
        config = json.loads(HOOKS.read_text(encoding="utf-8"))
        groups = config["hooks"]["PostToolUse"]

        sdd_handlers = [
            handler
            for group in groups
            for handler in group.get("hooks", [])
            if handler.get("command") == "python3 tools/sdd-history/agent_save_hook.py"
        ]
        self.assertEqual(len(sdd_handlers), 1)

        codex_groups = [
            group
            for group in groups
            if any(
                handler.get("command", "").endswith("/.codex/hooks/save_session.py\"")
                for handler in group.get("hooks", [])
            )
        ]
        self.assertEqual(len(codex_groups), 1)
        matcher = re.compile(codex_groups[0]["matcher"])
        self.assertIsNotNone(matcher.fullmatch("Bash"))
        self.assertIsNotNone(matcher.fullmatch("apply_patch"))
        handlers = codex_groups[0]["hooks"]
        self.assertTrue(all(handler.get("async") is True for handler in handlers))
        self.assertTrue(all(handler.get("type") == "command" for handler in handlers))

    def test_codex_generated_store_is_root_ignored(self):
        rules = {
            line.strip()
            for line in GITIGNORE.read_text(encoding="utf-8").splitlines()
            if line.strip() and not line.lstrip().startswith("#")
        }
        self.assertIn("/.codex-context/", rules)


if __name__ == "__main__":
    unittest.main()
