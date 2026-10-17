#!/usr/bin/env python3
"""task 7.5.14 보수 — 구조 시험의 계측기(`_disk_calls` · `_violations`)가 **미래의 우회 모양**을 잡는가.

독립 적대 리뷰가 사본에 한 줄씩 심어 통과(= 미래 우회)한 26 종을 보고했다. 그 스크립트는 저장소에 없어서, 보고된 모양을 이 파일이
**다시 만든다**(리뷰어의 `probe_bypass.py` 가 아니다). 모양마다 `check_analysis.py` 소스 끝에 새 함수 하나(필요하면 모듈 줄 하나)를
붙이고, 워킹트리의 시험 클래스가 가진 계측기로 세어 위반이 새 함수를 대는지 본다. 파일은 안 쓴다 — 소스 글자만 센다.

    python3 7514_probe_bypass.py
"""
import sys
from pathlib import Path

REPO = next(parent for parent in Path(__file__).resolve().parents if (parent / "tools" / "logic-map").is_dir())
sys.path.insert(0, str(REPO / "tools" / "logic-map"))
import test_check_analysis  # noqa: E402

CENSUS = test_check_analysis.TheRecheckReadsWhatTheVerdictRead
# (이름, 모듈 줄, 함수 몸통). 몸통은 `def _probe(p, q):` 안에 들어간다.
SHAPES = [
    ("os.path.getsize", "", "os.path.getsize(p)"),
    ("os.path.getmtime", "", "os.path.getmtime(p)"),
    ("os.path.getatime", "", "os.path.getatime(p)"),
    ("os.path.getctime", "", "os.path.getctime(p)"),
    ("Path.is_fifo", "", "p.is_fifo()"),
    ("Path.is_socket", "", "p.is_socket()"),
    ("Path.is_block_device", "", "p.is_block_device()"),
    ("Path.is_char_device", "", "p.is_char_device()"),
    ("Path.is_mount", "", "p.is_mount()"),
    ("Path.owner", "", "p.owner()"),
    ("Path.group", "", "p.group()"),
    ("os.statvfs", "", "os.statvfs(p)"),
    ("os.listxattr", "", "os.listxattr(p)"),
    ("os.getxattr", "", "os.getxattr(p, 'user.x')"),
    ("io.FileIO", "import io", "io.FileIO(p)"),
    ("glob.iglob", "import glob", "glob.iglob(p)"),
    ("os.fwalk", "", "os.fwalk(p)"),
    ("shutil.copyfile", "import shutil", "shutil.copyfile(p, q)"),
    ("linecache.getline", "import linecache", "linecache.getline(p, 1)"),
    ("from os import stat as _st", "from os import stat as _st", "_st(p)"),
    ("from os.path import getsize", "from os.path import getsize", "getsize(p)"),
    ("_o = open", "_o = open", "_o(p)"),
    ("_s = os.stat", "_s = os.stat", "_s(p)"),
    ("default reader=open", "", None),
    ("builtins.open", "import builtins", "builtins.open(p)"),
    ("getattr(os, 'stat')", "", "getattr(os, 'stat')(p)"),
    ("eval('open')", "", "eval('open')(p)"),
    ("__builtins__['open']", "", "__builtins__['open'](p)"),
    ("map(open, …)", "", "list(map(open, [p]))"),
    ("subprocess cat", "", "subprocess.run(['cat', str(p)], capture_output=True)"),
    ("subprocess git hash-object", "", "subprocess.run(['git', 'hash-object', str(p)], capture_output=True)"),
]


def main() -> None:
    source = Path(test_check_analysis.check_analysis.__file__).read_text(encoding="utf-8")
    assert CENSUS._violations(CENSUS._disk_calls(source)) == [], "무변이 소스가 이미 위반이다 — 대조군이 아니다"
    caught = limits = 0
    print(f"| 모양 | 결과 |\n|---|---|")
    for name, module_line, body in SHAPES:
        probe = f"\n{module_line}\n" if module_line else "\n"
        if body is None:
            probe += "def _probe(p, q, reader=open):\n    return reader(p)\n"
        else:
            probe += f"def _probe(p, q):\n    return {body}\n"
        violations = CENSUS._violations(CENSUS._disk_calls(source + probe))
        hit = any(line.startswith("_probe:") for line in violations)
        caught += hit
        limits += not hit
        print(f"| `{name}` | {'잡힘' if hit else '**한계**'} |")
    print(f"\n모양 {len(SHAPES)} · 잡힘 {caught} · 한계 {limits}")


if __name__ == "__main__":
    main()
