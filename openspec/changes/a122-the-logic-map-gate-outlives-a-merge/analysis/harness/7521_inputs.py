#!/usr/bin/env python3
"""판정이 읽는 **입력**을 AST 로 전부 센다 (task 7.5.2.1).

7.5.2 는 "묶었다" 고 적기 전에 판정이 읽는 입력의 목록을 세지 않았다. 그래서 바이트는 묶고
`HEAD`(역사) · 디렉터리 존재 · 목록 열거 셋을 놓쳤고 재리뷰가 셋 다 재현했다. 손으로 고른 목록은
**볼 곳을 고르므로** 선택적이다 — 이 열거는 진입점에서 닿는 모든 함수의 모든 I/O 호출을 적는다.

세는 것:
- `subprocess.run([...])` — 첫 인자 목록의 앞 몇 칸(git 의 하위 명령)과, 인자에 `"HEAD"` 가 있는가
- 파일시스템 메서드 호출 (`read_bytes` · `read_text` · `exists` · `is_dir` · `is_file` · `iterdir` ·
  `glob` · `rglob` · `resolve` · `is_symlink` · `open` · `lstat` · `stat` · `write_bytes` · `write_text`)
- `os.environ` 읽기
- 같은 모듈(과 `role_check`)의 함수 호출 — 도달 집합을 넓히는 데만 쓴다

    python3 <이 파일> [check|record_landing|main ...]
"""
import ast
import sys
from pathlib import Path

ROOT = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
MODULES = [ROOT / "tools" / "logic-map" / "check_analysis.py", ROOT / "tools" / "logic-map" / "role_check.py"]
FS = {"read_bytes", "read_text", "exists", "is_dir", "is_file", "iterdir", "glob", "rglob", "resolve",
      "is_symlink", "open", "lstat", "stat", "write_bytes", "write_text", "unlink", "mkstemp"}


def functions() -> dict[str, tuple[Path, ast.FunctionDef]]:
    found = {}
    for module in MODULES:
        tree = ast.parse(module.read_text(encoding="utf-8"))
        for node in ast.walk(tree):
            if isinstance(node, ast.FunctionDef):
                found.setdefault(node.name, (module, node))
    return found


def io_sites(fn: ast.FunctionDef) -> list[tuple[int, str]]:
    sites = []
    for node in ast.walk(fn):
        if isinstance(node, ast.Call):
            callee = ast.unparse(node.func)
            if callee.endswith("subprocess.run") or callee == "subprocess.run":
                argv = node.args[0] if node.args else None
                words = []
                if isinstance(argv, ast.List):
                    for element in argv.elts[:5]:
                        words.append(element.value if isinstance(element, ast.Constant) else "…")
                head = " HEAD" if "HEAD" in ast.unparse(node) else ""
                sites.append((node.lineno, "git " + " ".join(str(word) for word in words[1:]) + head))
            elif callee.startswith("os.") and not callee.startswith("os.environ"):
                # `os.path.realpath` · `os.path.lexists` 도 파일시스템 읽기다 — 첫 판은 메서드만 셌다(7.5.2.1 GREEN 뒤).
                sites.append((node.lineno, f"fs {callee}()"))
            elif isinstance(node.func, ast.Attribute) and node.func.attr in FS:
                sites.append((node.lineno, f"fs .{node.func.attr}() on `{ast.unparse(node.func.value)[:50]}`"))
        elif isinstance(node, ast.Attribute) and ast.unparse(node) == "os.environ":
            sites.append((node.lineno, "env os.environ"))
        elif isinstance(node, ast.Constant) and isinstance(node.value, str) and "HEAD" in node.value \
                and not node.value.startswith(("HEAD 커밋", " ")) and len(node.value) < 40:
            # 상징 `HEAD` 를 **인자로** 넘기는 자리(`_committed_bytes(root, "HEAD", …)`)도 역사 읽기다 —
            # `subprocess.run` 만 세면 `_landing_record` 가 안 보인다(첫 판이 그랬다).
            sites.append((node.lineno, f"history {node.value!r}"))
    return sorted(set(sites))


def callees(fn: ast.FunctionDef, known: dict) -> set[str]:
    names = set()
    for node in ast.walk(fn):
        if isinstance(node, ast.Call):
            target = node.func.id if isinstance(node.func, ast.Name) else (
                node.func.attr if isinstance(node.func, ast.Attribute) else "")
            if target in known:
                names.add(target)
        elif isinstance(node, ast.Name) and node.id in known and isinstance(node.ctx, ast.Load):
            names.add(node.id)                    # 인자로 넘긴 함수(`_decoded` 등)
    return names


def reach(entry: str, known: dict) -> list[str]:
    seen, order, stack = set(), [], [entry]
    while stack:
        name = stack.pop()
        if name in seen:
            continue
        seen.add(name)
        order.append(name)
        stack.extend(sorted(callees(known[name][1], known) - seen, reverse=True))
    return order


if __name__ == "__main__":
    known = functions()
    for entry in sys.argv[1:] or ["check", "record_landing", "main"]:
        print(f"=== {entry} ===")
        for name in reach(entry, known):
            module, fn = known[name]
            for line, what in io_sites(fn):
                print(f"  {module.name}:{line:<5} {name:<28} {what}")
