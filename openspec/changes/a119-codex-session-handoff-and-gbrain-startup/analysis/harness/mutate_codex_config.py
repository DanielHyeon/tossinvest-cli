#!/usr/bin/env python3
"""a119 회귀 시험이 결함 모양에 실제로 실패하는지 사본 변이로 재는 하네스.

실제 저장소 파일은 읽기만 하고, 변이는 임시 디렉터리 사본에만 적용함.
무변이 대조군이 통과하지 않으면 계측기가 눈먼 것이므로 즉시 멈춤.
"""

from __future__ import annotations

import importlib.util
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


def failures(module) -> int:
    suite = unittest.defaultTestLoader.loadTestsFromModule(module)
    result = unittest.TextTestRunner(stream=open("/dev/null", "w"), verbosity=0).run(suite)
    return len(result.failures) + len(result.errors)


def run_case(label: str, hooks: str | None, config: str | None, saver_extra: str | None) -> int:
    with tempfile.TemporaryDirectory() as raw:
        work = Path(raw)
        hooks_path = work / "hooks.json"
        config_path = work / "config.toml"
        saver_path = work / "save_session.py"
        hooks_path.write_text(hooks or (ROOT / ".codex/hooks.json").read_text(), encoding="utf-8")
        config_path.write_text(config or (ROOT / ".codex/config.toml").read_text(), encoding="utf-8")
        shutil.copyfile(ROOT / ".codex/hooks/save_session.py", saver_path)
        if saver_extra:
            source = saver_path.read_text(encoding="utf-8")
            marker = "def run(payload"
            assert marker in source
            saver_path.write_text(source.replace(marker, saver_extra + marker, 1), encoding="utf-8")
        event = load(EVENT_TEST, f"event_{label}")
        event.HOOKS = hooks_path
        event.SCRIPT = saver_path
        registration = load(REGISTRATION_TEST, f"registration_{label}")
        registration.CODEX_CONFIG = config_path
        count = failures(event) + failures(registration)
    print(f"{label}: failing tests = {count}")
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
            "import atexit\natexit.register(lambda: print('{}'))\n",
        ),
    ]
    blind = []
    for label, mutated_hooks, mutated_config, saver_extra in cases:
        if mutated_hooks == hooks or mutated_config == config:
            print(f"{label}: mutation did not apply — stop")
            return 2
        if run_case(label, mutated_hooks, mutated_config, saver_extra) == 0:
            blind.append(label)
    print("SURVIVED:", blind or "none")
    return 1 if blind else 0


if __name__ == "__main__":
    raise SystemExit(main())
