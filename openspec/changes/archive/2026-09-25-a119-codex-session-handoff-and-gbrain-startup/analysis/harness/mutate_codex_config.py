#!/usr/bin/env python3
"""a119 회귀 시험이 결함 모양에 실제로 실패하는지 사본 변이로 재는 하네스.

실제 저장소 파일은 읽기만 하고, 변이는 임시 디렉터리 사본에만 적용함.
무변이 대조군이 통과하지 않으면 계측기가 눈먼 것이므로 즉시 멈춤.
"""

from __future__ import annotations

import importlib.util
import os
import shutil
import sys
import tempfile
import unittest
from pathlib import Path

# 보관(archive) 뒤에도 깊이가 바뀌므로 고정 깊이 대신 저장소 표식을 가진 조상을 찾음.
ROOT = next(p for p in Path(__file__).resolve().parents if (p / ".codex" / "hooks.json").is_file())
EVENT_TEST = ROOT / "tools/sdd-history/test_codex_host_event_coverage.py"
REGISTRATION_TEST = ROOT / "tools/sdd/test_codex_gbrain_registration.py"


def load(path: Path, name: str):
    spec = importlib.util.spec_from_file_location(name, path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def failures(module) -> tuple[int, int]:
    """(실패+오류 수, 실행한 시험 수)를 반환함. 실행 수가 0 이면 대조군 GREEN 이 공허함.

    실패 수는 subTest 마다 따로 셈(한 시험이 여러 번 셀 수 있음).
    """
    suite = unittest.defaultTestLoader.loadTestsFromModule(module)
    with open(os.devnull, "w") as sink:
        result = unittest.TextTestRunner(stream=sink, verbosity=0).run(suite)
    return len(result.failures) + len(result.errors), result.testsRun


def run_case(label: str, hooks: str | None, config: str | None, saver_edit: tuple[str, str] | None) -> int:
    with tempfile.TemporaryDirectory() as raw:
        work = Path(raw)
        hooks_path = work / "hooks.json"
        config_path = work / "config.toml"
        saver_path = work / "save_session.py"
        hooks_path.write_text(hooks or (ROOT / ".codex/hooks.json").read_text(), encoding="utf-8")
        config_path.write_text(config or (ROOT / ".codex/config.toml").read_text(), encoding="utf-8")
        shutil.copyfile(ROOT / ".codex/hooks/save_session.py", saver_path)
        if saver_edit:
            # 저장기 변이는 (원문, 대체문) 한 쌍이며 원문이 사본에 정확히 한 번 있어야 적용함.
            old, new = saver_edit
            source = saver_path.read_text(encoding="utf-8")
            if source.count(old) != 1:
                raise SystemExit(f"{label}: saver mutation did not apply — stop")
            saver_path.write_text(source.replace(old, new, 1), encoding="utf-8")
        event = load(EVENT_TEST, f"event_{label}")
        event.HOOKS = hooks_path
        event.SCRIPT = saver_path
        registration = load(REGISTRATION_TEST, f"registration_{label}")
        registration.CODEX_CONFIG = config_path
        event_failed, event_ran = failures(event)
        registration_failed, registration_ran = failures(registration)
        count = event_failed + registration_failed
        ran = event_ran + registration_ran
    if ran == 0:
        raise SystemExit(f"{label}: no test ran; harness is blind — stop")
    print(f"{label}: failing tests = {count} (ran {ran})")
    return count


def main() -> int:
    hooks = (ROOT / ".codex/hooks.json").read_text(encoding="utf-8")
    config = (ROOT / ".codex/config.toml").read_text(encoding="utf-8")
    if run_case("M0-control", None, None, None) != 0:
        print("control is not green; harness is blind — stop")
        return 2
    cases = [
        ("M1-matcher-drops-apply_patch", hooks.replace("^(Bash|apply_patch)$", "^(Bash)$"), None, None),
        ("M2-matcher-unanchored", hooks.replace("^(Bash|apply_patch)$", "Bash|apply_patch"), None, None),
        (
            "M3-duplicate-wrapper",
            None,
            config + '\n[mcp_servers.gbrain2]\ncommand = "python3"\nargs = ["tools/sdd/gbrain_project.py", "serve"]\n',
            None,
        ),
        (
            "M4-raw-gbrain",
            None,
            config.replace('command = "python3"\nargs = ["tools/sdd/gbrain_project.py", "serve"]',
                           'command = "gbrain"\nargs = ["serve"]'),
            None,
        ),
        (
            "M5-saver-prints-stdout",
            None,
            None,
            ("def run(payload", "import atexit\natexit.register(lambda: print('{}'))\ndef run(payload"),
        ),
        (
            # 락 경합 출구에서만 stdout 에 씀(2026-09-26 적대 리뷰 E9).
            "M6-saver-prints-only-on-lock-contention",
            None,
            None,
            ("        if lock is None:\n            return 0",
             "        if lock is None:\n            print('{}')\n            return 0"),
        ),
        (
            # 락 경합 출구가 예외 경고 출구로 새어 나감 — stdout 은 비지만 출구가 바뀜(2차 리뷰 E13).
            "M9-lock-contention-diverts-to-exception-exit",
            None,
            None,
            ("        if lock is None:\n            return 0",
             "        if lock is None:\n            raise OSError('diverted')"),
        ),
        (
            # 고정은 글자로 남지만 `.*` 가 모든 이름을 받음(적대 리뷰 E1).
            "M7-matcher-admits-any-name",
            hooks.replace("^(Bash|apply_patch)$", "^(Bash|apply_patch|.*)$"),
            None,
            None,
        ),
        (
            # 양끝 글자는 그대로지만 우선순위로 고정이 풀림(적대 리뷰 E2·gstack 1).
            "M8-matcher-anchor-split-by-precedence",
            hooks.replace("^(Bash|apply_patch)$", "^Bash|apply_patch$"),
            None,
            None,
        ),
    ]
    blind = []
    for label, mutated_hooks, mutated_config, saver_edit in cases:
        if mutated_hooks == hooks or mutated_config == config:
            print(f"{label}: mutation did not apply — stop")
            return 2
        if run_case(label, mutated_hooks, mutated_config, saver_edit) == 0:
            blind.append(label)
    print("SURVIVED:", blind or "none")
    return 1 if blind else 0


if __name__ == "__main__":
    raise SystemExit(main())
