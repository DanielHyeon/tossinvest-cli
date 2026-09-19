"""task 7.5.2 VERIFY — 착지 기록이 있는 실물(a099)에서 `check()` 한 번이 `ast.json` 을 몇 번 읽고
git 프로세스를 몇 개 띄우는가. 편집 전(기준 blob) · 편집 후(워킹트리) 사본을 **같은 계측기**로 잰다.

    python3 752_reads.py [<before-sha>]
"""
import collections
import importlib.util
import shutil
import subprocess
import sys
import time
from pathlib import Path

ROOT = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
SP = Path(__file__).resolve().parent / "_work"
BEFORE = sys.argv[1] if len(sys.argv) > 1 else "e9f905bdaac5819fd43aa5addd6b38176889b9e2"
CHANGE = "a099-a-claim-excludes-the-second-sender"


def build(tag, revision):
    where = SP / f"752_{tag}"
    if where.exists():
        shutil.rmtree(where)
    shutil.copytree(ROOT / "tools" / "logic-map", where, ignore=shutil.ignore_patterns("__pycache__", "test_*.py"))
    if revision:
        (where / "check_analysis.py").write_bytes(subprocess.run(
            ["git", "show", f"{revision}:tools/logic-map/check_analysis.py"],
            cwd=ROOT, capture_output=True, check=True).stdout)
    sys.path.insert(0, str(where))
    spec = importlib.util.spec_from_file_location(f"ca752_{tag}", where / "check_analysis.py")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    sys.path.pop(0)
    return module


for tag, revision in (("before", BEFORE), ("after", None)):
    module = build(tag, revision)
    reads = collections.Counter()
    spawns = collections.Counter()
    real_bytes, real_text, real_run = Path.read_bytes, Path.read_text, module.subprocess.run

    def read_bytes(path):
        if path.name == "ast.json":
            reads["ast.json"] += 1
        return real_bytes(path)

    def read_text(path, *args, **kwargs):
        if path.name == "ast.json":
            reads["ast.json"] += 1
        return real_text(path, *args, **kwargs)

    def run(*args, **kwargs):
        argv = args[0] if args else kwargs.get("args")
        spawns[" ".join(argv[1:3]) if isinstance(argv, list) else "?"] += 1
        return real_run(*args, **kwargs)

    Path.read_bytes, Path.read_text, module.subprocess.run = read_bytes, read_text, run
    start = time.monotonic()
    try:
        errors = module.check(CHANGE, ROOT, {})
    finally:
        Path.read_bytes, Path.read_text, module.subprocess.run = real_bytes, real_text, real_run
    bundles = len(list((ROOT / "openspec/changes/archive/2026-09-09-a099-a-claim-excludes-the-second-sender"
                        / "analysis/function-logic").glob("*/ast.json")))
    print(f"{tag:6s} errors={len(errors)} · ast.json {bundles} 개 · 읽기 {reads['ast.json']} "
          f"({reads['ast.json'] / bundles:.1f}/번들) · git {sum(spawns.values())} "
          f"(cat-file {spawns['cat-file --batch']}) · {time.monotonic() - start:.2f}s")
