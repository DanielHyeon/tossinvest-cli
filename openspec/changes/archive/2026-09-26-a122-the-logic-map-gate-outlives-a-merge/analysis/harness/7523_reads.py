"""task 7.5.2.3 VERIFY — 끝의 재확인이 **얼마나 더 읽는가**. 편집 전(기준 blob) · 편집 후를 같은 계측기로.

`7521_main_ab.py` 의 `ast_reads` 는 `ast.json` 만 센다 — 7.5.2.2 도 증거는 끝에서 다시 읽었으므로 그 수는
양쪽이 같다(9,072 → 9,072). 이 로트가 더 읽는 것은 **그 밖의 전부**(번들 산문 · 워킹트리 소스 · `review.md` ·
`base-commit.txt` · 시험 색인)이므로, 파일을 여는 **모든** 자리를 세고 시간도 같이 잰다.

    python3 7523_reads.py [<change-id>] [<before-sha>]
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
CHANGE = sys.argv[1] if len(sys.argv) > 1 else "a112-run-four-strategy-families-independently"
BEFORE = sys.argv[2] if len(sys.argv) > 2 else "1d12520c"


def build(tag, revision):
    where = SP / f"7523_{tag}"
    if where.exists():
        shutil.rmtree(where)
    shutil.copytree(ROOT / "tools" / "logic-map", where,
                    ignore=shutil.ignore_patterns("__pycache__", "test_*.py"))
    if revision:
        (where / "check_analysis.py").write_bytes(subprocess.run(
            ["git", "show", f"{revision}:tools/logic-map/check_analysis.py"],
            cwd=ROOT, capture_output=True, check=True).stdout)
    sys.path.insert(0, str(where))
    spec = importlib.util.spec_from_file_location(f"ca7523_{tag}", where / "check_analysis.py")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    sys.path.pop(0)
    return module


def bucket(path: str) -> str:
    if "/analysis/function-logic/" in path:
        return "번들"
    if path.endswith("_test.go"):
        return "*_test.go"
    if path.endswith(".go"):
        return "Go 소스"
    return "그 밖"


for tag, revision in (("before", BEFORE), ("after", None)):
    module = build(tag, revision)
    opens = collections.Counter()
    spawns = 0
    real_bytes, real_text, real_run = Path.read_bytes, Path.read_text, module.subprocess.run
    # 읽는 길이 판본마다 다르다 — 편집 전은 `Path.read_*` 와 `_read_regular`, 편집 후는 깔때기 하나다.
    inner = "_opened_bytes" if hasattr(module, "_opened_bytes") else "_read_regular"
    real_inner = getattr(module, inner, None)

    def counted(path, *args, **kwargs):
        opens[bucket(str(path))] += 1
        return real_inner(path)

    def read_bytes(path):
        opens[bucket(str(path))] += 1
        return real_bytes(path)

    def read_text(path, *args, **kwargs):
        opens[bucket(str(path))] += 1
        return real_text(path, *args, **kwargs)

    def run(*args, **kwargs):
        global spawns
        spawns += 1
        return real_run(*args, **kwargs)

    Path.read_bytes, Path.read_text, module.subprocess.run = read_bytes, read_text, run
    if real_inner is not None:
        setattr(module, inner, counted)
    start = time.monotonic()
    try:
        errors = module.check(CHANGE, ROOT, {})
    finally:
        Path.read_bytes, Path.read_text, module.subprocess.run = real_bytes, real_text, real_run
        if real_inner is not None:
            setattr(module, inner, real_inner)
    total = sum(opens.values())
    parts = " · ".join(f"{name} {count}" for name, count in opens.most_common())
    print(f"{tag:6s} 판정 줄 {len(errors):3d} · 파일 열기 {total:5d} ({parts}) · "
          f"프로세스 {spawns} · {time.monotonic() - start:.2f}s")
