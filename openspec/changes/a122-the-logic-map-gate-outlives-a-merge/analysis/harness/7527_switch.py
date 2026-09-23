"""task 7.5.27 — 워킹트리를 다시 쓰는 문이 **실물 게이트 판정**에 무엇을 하는지 잰다.

공유 저장소를 건드리지 않는다: 격리 worktree 에 `.gitattributes` 만 놓고(worktree 마다 따로다),
필터 드라이버 설정은 **환경 변수**(`GIT_CONFIG_COUNT`)로만 준다. 7.5.24 에서 `git config` 를 격리
worktree 에서 했다가 공유 `.git/config` 에 킬 스위치 절반을 남긴 뒤 바꾼 방법이다.

    git worktree add --detach <PROBE> HEAD
    python3 7527_switch.py <PROBE> [<before-sha>]
    git worktree remove --force <PROBE>
"""
import importlib.util
import os
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

ROOT = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
PROBE = Path(sys.argv[1])
BEFORE = sys.argv[2] if len(sys.argv) > 2 else "6c2ce774"
CHANGE = "a112-run-four-strategy-families-independently"
BASE = (ROOT / "openspec" / "changes" / CHANGE / "base-commit.txt").read_text(encoding="utf-8").strip()


def load(tag: str, revision: str | None):
    where = Path(tempfile.mkdtemp()) / "lm"
    shutil.copytree(ROOT / "tools" / "logic-map", where,
                    ignore=shutil.ignore_patterns("__pycache__", "test_*.py"))
    if revision:
        blob = subprocess.run(["git", "show", f"{revision}:tools/logic-map/check_analysis.py"],
                              cwd=ROOT, capture_output=True, check=True).stdout
        (where / "check_analysis.py").write_bytes(blob)
    sys.path.insert(0, str(where))
    spec = importlib.util.spec_from_file_location(f"ca_{tag}", where / "check_analysis.py")
    module = importlib.util.module_from_spec(spec)
    sys.modules[f"ca_{tag}"] = module
    spec.loader.exec_module(module)
    sys.path.remove(str(where))
    return module


old, new = load("before", BEFORE), load("after", None)
hide = Path(tempfile.mkdtemp()) / "hide.sh"
hide.write_text(f'#!/bin/sh\ngit -C "{PROBE}" cat-file -p {BASE}:"$1" 2>/dev/null || cat\n', encoding="utf-8")
hide.chmod(0o755)
attributes = PROBE / ".gitattributes"


def judge(label: str) -> None:
    print(f"\n=== {label} ===")
    for tag, module in ((f"편집 전 {BEFORE[:8]}", old), ("편집 후", new)):
        facts: dict = {}
        try:
            errors = module.check(CHANGE, PROBE, facts)
        except Exception as error:  # noqa: BLE001 — 무엇이 올라오는지 그대로 찍는 것이 측정이다
            print(f"  {tag}: 예외 {type(error).__name__}: {error}")
            continue
        first = errors[0][:92] if errors else "(판정 줄 0 — evidence complete)"
        print(f"  {tag}: 판정 줄 {len(errors):2d} · required {facts.get('required_count')} · {first}")


try:
    judge("평소")
    attributes.write_text("*.go filter=hide\n", encoding="utf-8")
    os.environ.update({"GIT_CONFIG_COUNT": "1", "GIT_CONFIG_KEY_0": "filter.hide.clean",
                       "GIT_CONFIG_VALUE_0": f"{hide} %f"})
    subprocess.run(["find", ".", "-name", "*.go", "-not", "-path", "./.git/*",
                    "-exec", "touch", "{}", "+"], cwd=PROBE, check=True)
    judge("clean 필터 — 추적 안 된 `.gitattributes` 한 줄 + 환경 변수 설정 하나")
finally:
    for key in ("GIT_CONFIG_COUNT", "GIT_CONFIG_KEY_0", "GIT_CONFIG_VALUE_0"):
        os.environ.pop(key, None)
    attributes.unlink(missing_ok=True)
