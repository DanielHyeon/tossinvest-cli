#!/usr/bin/env python3
"""Python 함수의 분기·반환·예외를 **기계로** 열거한다.

이 저장소의 FLM 산출물은 `tools/logic-map/extract_go_ast.go` 가 **Go** 함수에서
뽑는다. Python 에는 그 도구가 없다 — a120 이 `analysis/python-function-logic/` 로
같은 공백을 메운 선례가 있다. 손으로 읽은 증거는 **볼 곳을 고르므로** 선택적이고,
이 열거는 선택적이지 않다는 것이 이 파일이 존재하는 이유다.

**이것은 저장소 도구가 아니라 이 change 의 일회용 열거기다.** 커버리지·안전을
주장하지 않는다. 오직 AST 에 실제로 있는 노드만 세어 적는다.

    python3 <이 파일> tools/logic-map/check_analysis.py resolve_referenced_change
"""
import ast
import hashlib
import json
import subprocess
import sys
from pathlib import Path

# 루트는 세어서 올라가지 않고 **유도**한다 — 이 파일이 사는 깊이는 change 경로가
# 바뀌면 같이 바뀐다.
ROOT = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())

BRANCHING = (ast.If, ast.For, ast.While, ast.Try, ast.ExceptHandler, ast.IfExp,
             ast.BoolOp, ast.Match, ast.match_case, ast.comprehension, ast.Assert)


def find(tree: ast.AST, name: str) -> ast.AST:
    for node in ast.walk(tree):
        if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)) and node.name == name:
            return node
    raise SystemExit(f"function not found: {name}")


def read_source(path: Path, revision: str) -> bytes:
    """`worktree` 는 디스크를, 그 밖은 그 revision 의 blob 을 읽는다.

    편집 전후를 **같은 열거기로** 뽑아야 두 표를 대조할 수 있다. 편집 전을
    디스크에서 읽으려면 편집을 되돌려야 하므로 blob 에서 읽는다.
    """
    if revision == "worktree":
        return (ROOT / path).read_bytes()
    return subprocess.run(
        ["git", "show", f"{revision}:{path.as_posix()}"],
        cwd=ROOT, capture_output=True, check=True,
    ).stdout


def enumerate_function(path: Path, name: str, revision: str = "worktree") -> dict:
    source = read_source(path, revision)
    fn = find(ast.parse(source.decode("utf-8")), name)
    # `ast.comprehension` 처럼 lineno 가 없는 노드가 있다. 줄 없는 분기는 표에서
    # 쓸 수 없으므로 **가장 가까운 조상의 줄**을 물려받는다 — 지어내지 않고 물려받는다.
    line_of: dict[int, int] = {}
    for parent in ast.walk(fn):
        for child in ast.iter_child_nodes(parent):
            line_of[id(child)] = getattr(child, "lineno", None) or line_of.get(id(parent)) \
                or getattr(parent, "lineno", 0)
    line_of.setdefault(id(fn), fn.lineno)

    branches, returns, raises, calls = [], [], [], []
    for node in ast.walk(fn):
        line = getattr(node, "lineno", None) or line_of.get(id(node), 0)
        if isinstance(node, BRANCHING):
            branches.append({
                "id": "",
                "kind": type(node).__name__,
                "line": line,
                "inherited_line": not hasattr(node, "lineno"),
                "source": ast.unparse(node).splitlines()[0][:120],
            })
        elif isinstance(node, ast.Return):
            returns.append({"line": line,
                            "value": ast.unparse(node.value) if node.value else None})
        elif isinstance(node, ast.Raise):
            raises.append({"line": line, "source": ast.unparse(node)[:160]})
        elif isinstance(node, ast.Call):
            calls.append({"line": line, "callee": ast.unparse(node.func)})
    for bucket in (branches, returns, raises, calls):
        bucket.sort(key=lambda item: (item["line"], str(item)))
    for position, branch in enumerate(branches, start=1):
        branch["id"] = f"B{position}"
    return {
        "language": "python",
        "file": str(path),
        "revision": revision,
        "head_commit": subprocess.run(
            ["git", "rev-parse", "HEAD"], cwd=ROOT, capture_output=True, text=True, check=True
        ).stdout.strip(),
        "source_sha256": hashlib.sha256(source).hexdigest(),
        "function": name,
        "signature": f"{name}({', '.join(argument.arg for argument in fn.args.args)})",
        "start": {"line": fn.lineno, "column": fn.col_offset + 1},
        "end": {"line": fn.end_lineno, "column": (fn.end_col_offset or 0) + 1},
        "branches": branches,
        "returns": returns,
        "raises": raises,
        "calls": calls,
    }


if __name__ == "__main__":
    print(json.dumps(
        enumerate_function(Path(sys.argv[1]), sys.argv[2],
                           sys.argv[3] if len(sys.argv) > 3 else "worktree"),
        ensure_ascii=False, indent=2))
