"""task 7.5.2.3 MEASURE — 판정이 디스크에서 읽는 것의 **전수 목록**. 손으로 세지 않는다.

두 방향으로 센다. 어느 한쪽만으로는 목록이 못 닫힌다:

- **정적**: 모듈 AST 를 걸어 파일시스템 원시 호출(`read_bytes` · `read_text` · `open` · `iterdir` ·
  `rglob` · `glob` · `is_file` · `is_dir` · `exists` · `lexists` · `stat` …)이 **어느 함수 안에** 있는지
  열거한다. 함수 하나가 도달하지 않는 갈래에 숨긴 읽기도 보인다.
- **동적**: 실물 change 로 `check()` 를 한 번 돌리고 `Path` 메서드와 `_read_regular` 를 감싸 **실제로
  만진 경로**를 종류별로 센다. 그리고 지금의 끝-재확인(`_judged_state_moved`)이 다시 읽는 집합
  (`HEAD` + `Evidence`)과 뺄셈해서 **재확인이 안 보는 입력**을 찍는다.

    python3 7523_inputs.py [<change-id>]
"""
import ast
import collections
import os
import sys
from pathlib import Path

ROOT = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
SOURCE = ROOT / "tools" / "logic-map" / "check_analysis.py"
CHANGE = sys.argv[1] if len(sys.argv) > 1 else "a112-run-four-strategy-families-independently"

# 내용을 읽는 것과 이름만 재는 것을 가른다 — 재확인이 다시 읽어야 하는 것은 앞쪽이다.
CONTENT = {"read_bytes", "read_text", "open", "iterdir", "rglob", "glob", "scandir", "listdir", "walk", "read"}
STATS = {"is_file", "is_dir", "exists", "lexists", "is_symlink", "stat", "lstat", "fstat", "samefile"}


def spelled(node):
    """호출의 마지막 이름 — `path.read_bytes()` · `os.open()` · `open()` 을 같은 표로."""
    func = node.func
    if isinstance(func, ast.Attribute):
        return func.attr
    if isinstance(func, ast.Name):
        return func.id
    return ""


def static_inventory():
    tree = ast.parse(SOURCE.read_text(encoding="utf-8"))
    owner = {}
    for item in ast.walk(tree):
        if isinstance(item, (ast.FunctionDef, ast.AsyncFunctionDef)):
            for inner in ast.walk(item):
                owner.setdefault(id(inner), item.name)
    found = collections.defaultdict(list)
    for item in ast.walk(tree):
        if not isinstance(item, ast.Call):
            continue
        name = spelled(item)
        kind = "content" if name in CONTENT else "stat" if name in STATS else ""
        if not kind:
            continue
        found[(kind, owner.get(id(item), "<module>"))].append((item.lineno, name))
    return found


def dynamic_inventory(change):
    sys.path.insert(0, str(ROOT / "tools" / "logic-map"))
    import check_analysis as module

    touched = collections.defaultdict(set)
    real = {name: getattr(Path, name) for name in
            ("read_bytes", "read_text", "iterdir", "rglob", "glob", "is_file", "is_dir", "exists", "is_symlink")}
    real_regular = getattr(module, "_read_regular", None)

    def wrap(name, kind):
        original = real[name]

        def wrapped(self, *args, **kwargs):
            touched[kind].add(str(self) if not args else f"{self}::{args[0]}")
            return original(self, *args, **kwargs)
        return wrapped

    for name in ("read_bytes", "read_text", "iterdir"):
        setattr(Path, name, wrap(name, "content"))
    for name in ("rglob", "glob"):
        setattr(Path, name, wrap(name, "pattern"))
    for name in ("is_file", "is_dir", "exists", "is_symlink"):
        setattr(Path, name, wrap(name, "stat"))
    if real_regular is not None:
        def read_regular(path):
            touched["content"].add(str(path))
            return real_regular(path)
        module._read_regular = read_regular
    try:
        errors = module.check(change, ROOT, {})
    finally:
        for name, original in real.items():
            setattr(Path, name, original)
        if real_regular is not None:
            module._read_regular = real_regular

    # 지금의 끝-재확인이 다시 읽는 집합: `HEAD`(git) + `Evidence`(증거 디렉터리 목록 + 번들별 ast.json)
    analysis = ROOT / "openspec" / "changes" / change / "analysis" / "function-logic"
    rechecked = {str(analysis)} | {str(p / "ast.json") for p in (analysis.iterdir() if analysis.is_dir() else ())}
    return errors, touched, rechecked


def bucket(path):
    if "/analysis/function-logic/" in path:
        return "번들(증거)"
    if path.endswith("_test.go"):
        return "*_test.go(시험 색인)"
    if path.endswith(".go"):
        return "Go 소스(워킹트리)"
    if path.endswith("review.md"):
        return "review.md(면제 표지)"
    if path.endswith("base-commit.txt"):
        return "base-commit.txt"
    if path.endswith("function-logic-reference.txt"):
        return "function-logic-reference.txt"
    return "그 밖"


print(f"=== 정적: {SOURCE.relative_to(ROOT)} 의 파일시스템 원시 호출 (AST 열거)")
static = static_inventory()
for kind in ("content", "stat"):
    rows = sorted((owner, sites) for (this, owner), sites in static.items() if this == kind)
    total = sum(len(sites) for _, sites in rows)
    print(f"-- {kind}: 자리 {total} · 함수 {len(rows)}")
    for owner, sites in rows:
        print(f"   {owner:34s} {', '.join(f'{line}:{name}' for line, name in sorted(sites))}")

print(f"\n=== 동적: check('{CHANGE}') 한 번이 만진 경로")
errors, touched, rechecked = dynamic_inventory(CHANGE)
print(f"판정 줄 {len(errors)}")
for kind in ("content", "pattern", "stat"):
    paths = touched[kind]
    print(f"-- {kind}: {len(paths)}")
    if kind == "pattern":
        for p in sorted(paths):
            print(f"   {p}")
        continue
    counted = collections.Counter(bucket(p) for p in paths)
    for name, count in counted.most_common():
        print(f"   {name:28s} {count}")

uncovered = {p for p in touched["content"] if p not in rechecked}
print(f"\n=== 끝-재확인이 다시 읽는 것 {len(rechecked)} · 판정이 읽은 것 {len(touched['content'])} "
      f"· **재확인 밖** {len(uncovered)}")
for name, count in collections.Counter(bucket(p) for p in uncovered).most_common():
    print(f"   {name:28s} {count}")
print("-- 재확인 밖의 표본 (종류마다 하나)")
shown = set()
for p in sorted(uncovered):
    if bucket(p) in shown:
        continue
    shown.add(bucket(p))
    print(f"   {bucket(p):28s} {os.path.relpath(p, ROOT)}")
