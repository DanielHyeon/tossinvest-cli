#!/usr/bin/env python3
"""저장 훅 시험이 실제 동작을 잡는지 사본 변이로 확인함. 실제 파일은 건드리지 않음.

실행: python3 openspec/changes/a180-claude-session-save-hook-repair/analysis/harness/mutate_save_session.py
무변이 대조군이 실패하거나, 변이가 적용되지 않거나, 살아남은 변이가 있으면 종료 코드 1을 반환함.
"""

from __future__ import annotations

import io
import json
import shutil
import sys
import tempfile
import unittest
from pathlib import Path


def repository_root() -> Path:
    # archive 로 옮겨도 동작하도록 절대경로 대신 저장소 표지 파일을 찾아 올라감
    for parent in Path(__file__).resolve().parents:
        if (parent / "save-session.sh").exists() and (parent / "Makefile").exists():
            return parent
    raise SystemExit("repository root not found")


ROOT = repository_root()
sys.path.insert(0, str(ROOT / "tools" / "sdd-history"))
import test_claude_save_session as subject  # noqa: E402

SCRIPT_MUTANTS = {
    "drop -print": lambda s: s.replace("-mmin -30 -print", "-mmin -30"),
    "drop archive prune": lambda s: s.replace('-path "$PROJECT_DIR/openspec/changes/archive" -prune -o \\\n', ""),
    "archive filtered, not pruned": lambda s: s.replace(
        '-path "$PROJECT_DIR/openspec/changes/archive" -prune -o \\\n', '-not -path "$PROJECT_DIR/openspec/changes/archive*" \\\n'
    ),
    "no interval limit": lambda s: s.replace('-lt "$HOOK_MIN_INTERVAL_SEC"', "-lt 0"),
    "interval 1s": lambda s: s.replace("HOOK_MIN_INTERVAL_SEC=300", "HOOK_MIN_INTERVAL_SEC=1"),
    "interval 3000s": lambda s: s.replace("HOOK_MIN_INTERVAL_SEC=300", "HOOK_MIN_INTERVAL_SEC=3000"),
    "hook waits on lock": lambda s: s.replace("flock -n 9 || exit 0", "flock -w 3 9 || exit 0"),
    "hook writes despite lock": lambda s: s.replace("flock -n 9 || exit 0", "flock -n 9 || true"),
    "manual does not wait": lambda s: s.replace('flock -w "$LOCK_WAIT_SEC" 9 ||', "flock -n 9 ||"),
    "default wait changed": lambda s: s.replace("SAVE_SESSION_LOCK_WAIT_SEC:-120", "SAVE_SESSION_LOCK_WAIT_SEC:-1"),
    "lock timeout ignored": lambda s: s.replace('flock -w "$LOCK_WAIT_SEC" 9 || fail', 'flock -w "$LOCK_WAIT_SEC" 9 || true fail'),
    "lock timeout exits 0": lambda s: s.replace('flock -w "$LOCK_WAIT_SEC" 9 || fail "다른 저장이 끝나지 않아 저장 실패"', 'flock -w "$LOCK_WAIT_SEC" 9 || exit 0'),
    "failure not logged": lambda s: s.replace('>> "$OUTPUT_DIR/save-failures.log"', "> /dev/null"),
    "memo not carried": lambda s: s.replace("MANUAL_MEMO=$(awk", "UNUSED=$(awk"),
    "memo boundary loose": lambda s: s.replace("/^## Claude Code 최근 대화 — 사용자 요청 \\+ Claude 완료 작업 \\(자동 수집\\)$/", "/^## Claude Code 최근 대화/"),
    "guidance always added": lambda s: s.replace('if [ -z "$MANUAL_MEMO" ]; then', 'if true; then'),
    "no omission note": lambda s: s.replace('[ "$GIT_STATUS_TOTAL" -gt "$GIT_STATUS_MAX_LINES" ]', "false"),
    "no head limit": lambda s: s.replace('| head -n "$GIT_STATUS_MAX_LINES")', ")"),
    "status takes index lock": lambda s: s.replace("export GIT_OPTIONAL_LOCKS=0", ":"),
    "diff refreshes index": lambda s: s.replace("-c diff.autoRefreshIndex=false ", ""),
    "non-atomic write": lambda s: s.replace('if ! cat > "$TMP_FILE" << HEREDOC', 'if ! cat > "$OUTPUT_FILE" << HEREDOC').replace('mv -f "$TMP_FILE" "$OUTPUT_FILE" || fail "요약을 교체하지 못해 저장 실패"', ":"),
    "write failure ignored": lambda s: s.replace('then\n  fail "요약을 쓰지 못해 저장 실패"\nfi', 'then\n  :\nfi'),
    "temp file left": lambda s: s.replace("trap 'rm -f \"$TMP_FILE\"' EXIT", ":").replace('mv -f "$TMP_FILE" "$OUTPUT_FILE"', 'cp -f "$TMP_FILE" "$OUTPUT_FILE"'),
    "backup emptied": lambda s: s.replace('if cp "$OUTPUT_FILE" "$OUTPUT_DIR/session-summary_${PREV_TS}.md"; then', 'if : > "$OUTPUT_DIR/session-summary_${PREV_TS}.md"; then'),
}


def edit_hooks(settings: dict, needle: str, event: str, change) -> None:
    for group in settings["hooks"].get(event, []):
        for hook in group["hooks"]:
            if needle in hook.get("command", ""):
                change(hook)


def narrow_matcher(settings: dict) -> None:
    for group in settings["hooks"]["PostToolUse"]:
        if any("save-session.sh" in hook.get("command", "") for hook in group["hooks"]):
            group["matcher"] = "Edit|Write"


def swallow_failure(settings: dict) -> None:
    edit_hooks(settings, "save-session.sh", "Stop", lambda h: h.update(command=h["command"] + "; exit 0"))


def relative_save(settings: dict) -> None:
    edit_hooks(settings, "save-session.sh", "PostToolUse", lambda h: h.update(command="./save-session.sh --hook > /dev/null 2>&1"))


def sync_save(settings: dict) -> None:
    edit_hooks(settings, "save-session.sh", "PostToolUse", lambda h: h.pop("async"))


def relative_agent(settings: dict) -> None:
    edit_hooks(settings, "agent_save_hook.py", "PostToolUse", lambda h: h.update(command="python3 tools/sdd-history/agent_save_hook.py"))


def no_stop(settings: dict) -> None:
    settings["hooks"].pop("Stop")


def stop_throttled(settings: dict) -> None:
    edit_hooks(settings, "save-session.sh", "Stop", lambda h: h.update(command=h["command"].replace('.sh" >', '.sh" --hook >')))


SETTINGS_MUTANTS = {
    "save hook relative": relative_save,
    "save hook sync": sync_save,
    "agent hook relative": relative_agent,
    "no Stop hook": no_stop,
    "Stop hook throttled": stop_throttled,
    "matcher drops Bash": narrow_matcher,
    "Stop command extended": swallow_failure,
}


def run_suite() -> list[str]:
    suite = unittest.TestSuite()
    for case in (subject.SaveSessionScriptTest, subject.ClaudeHookSettingsTest):
        suite.addTests(unittest.defaultTestLoader.loadTestsFromTestCase(case))
    result = unittest.TextTestRunner(stream=io.StringIO()).run(suite)
    assert result.testsRun > 0
    return [item[0].id().split(".")[-1] for item in result.failures + result.errors]


def main() -> int:
    source = subject.SCRIPT.read_text(encoding="utf-8")
    settings_text = subject.SETTINGS.read_text(encoding="utf-8")
    work = Path(tempfile.mkdtemp(prefix="a180-mutate"))
    script = work / "save-session.sh"
    settings = work / "settings.json"
    subject.SCRIPT, subject.SETTINGS = script, settings
    failed = False
    try:
        script.write_text(source, encoding="utf-8")
        settings.write_text(settings_text, encoding="utf-8")
        control = run_suite()
        print(f"control: failing={control}")
        failed |= bool(control)
        for name, mutate in SCRIPT_MUTANTS.items():
            mutated = mutate(source)
            if mutated == source:
                print(f"{name}: NOT APPLIED")
                failed = True
                continue
            script.write_text(mutated, encoding="utf-8")
            caught = run_suite()
            print(f"{name}: {'CAUGHT' if caught else 'SURVIVED'} {caught}")
            failed |= not caught
        script.write_text(source, encoding="utf-8")
        for name, mutate in SETTINGS_MUTANTS.items():
            data = json.loads(settings_text)
            mutate(data)
            mutated = json.dumps(data, ensure_ascii=False, indent=2)
            if json.loads(mutated) == json.loads(settings_text):
                print(f"{name}: NOT APPLIED")
                failed = True
                continue
            settings.write_text(mutated, encoding="utf-8")
            caught = run_suite()
            print(f"{name}: {'CAUGHT' if caught else 'SURVIVED'} {caught}")
            failed |= not caught
    finally:
        shutil.rmtree(work)
    return 1 if failed else 0


if __name__ == "__main__":
    raise SystemExit(main())
