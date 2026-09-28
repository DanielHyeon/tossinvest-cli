#!/usr/bin/env python3
"""a125 1.1 — 분기 논증의 **실행 영수증**(옛 도구 판본 대상). 이관 기록이 없는 change 에서

  (1) `resolve_base` 의 결과(base · 문맥)는 이관 판정기를 **부르든 안 부르든** 같다 — 부르지 않는 판본은 판정기를
      `None` 을 돌려주는 대역으로 바꿔 흉내 낸다(제거 뒤 코드의 동작).
  (2) 제거될 문맥 키 `execution_baseline_adoption` 은 그 경우 `False` 이고, 옛 코드에서 그 키를 읽는 자리는 전부
      **참일 때만** 갈래가 선다(AST 로 센다) — 거짓이면 판정 · 출력 어느 쪽에도 관여하지 않는다.
  (3) `check()` 의 판정 줄과 `main()` 의 출력이 (1) 의 두 판본에서 같다.

    python3 ab_absent_record.py --tool <옛 판본 tools/logic-map 디렉터리>

제거 **뒤**의 같은 사실(기록이 있어도 판정 · 문맥 · 출력이 없을 때와 같다)은 스위트의
`TheA063ExceptionIsRetired.test_a_leftover_record_changes_neither_verdict_nor_context_nor_output` 가 못 박는다.
"""
import argparse
import ast
import io
import json
import shutil
import subprocess
import sys
import tempfile
from contextlib import redirect_stdout
from pathlib import Path
from unittest import mock

parser = argparse.ArgumentParser()
parser.add_argument("--tool", required=True)
args = parser.parse_args()
tool = Path(args.tool).resolve()
sys.path.insert(0, str(tool))
import check_analysis  # noqa: E402
import fixture_git_env  # noqa: E402

fixture_git_env.isolate()
assert Path(check_analysis.__file__).resolve().parent == tool


def git(root: Path, *argv: str) -> str:
    return subprocess.run(["git", *argv], cwd=root, check=True, capture_output=True, text=True).stdout.strip()


def fixture(raw: str) -> Path:
    root = Path(raw)
    git(root, "init", "-q")
    git(root, "config", "user.email", "a125@example.invalid")
    git(root, "config", "user.name", "a125")
    (root / "go.mod").write_text("module fixture\ngo 1.23\n")
    shutil.copytree(tool, root / "tools" / "logic-map", ignore=shutil.ignore_patterns("*.py", "__pycache__"))
    source = root / "internal" / "own.go"
    source.parent.mkdir(parents=True)
    source.write_text("package internal\nfunc Own() int { return 1 }\n")
    git(root, "add", "."); git(root, "commit", "-qm", "P")
    base = git(root, "rev-parse", "HEAD")
    change = root / "openspec" / "changes" / "ordinary"
    change.mkdir(parents=True)
    (change / "base-commit.txt").write_text(base + "\n")
    (change / "review.md").write_text("ordinary\n")
    source.write_text("package internal\nfunc Own() int { return 2 }\n")
    git(root, "add", "."); git(root, "commit", "-qm", "work")
    return root


def judge(root: Path) -> dict:
    head = check_analysis._head_commit(root)
    context: dict[str, object] = {}
    base = check_analysis.resolve_base(root / "openspec/changes/ordinary", root, context,
                                       change_id="ordinary", head=head)
    facts: dict[str, object] = {}
    verdict = check_analysis.check("ordinary", root, facts)
    output = io.StringIO()
    with mock.patch.object(sys, "argv", ["check_analysis.py", "--change", "ordinary", "--root", str(root)]), \
            redirect_stdout(output):
        code = check_analysis.main()
    return {"base": base, "base_context": context, "verdict": verdict, "facts": facts,
            "rc": code, "output": output.getvalue()}


with tempfile.TemporaryDirectory() as raw:
    root = fixture(raw)
    called = judge(root)
    with mock.patch.object(check_analysis, "validate_execution_baseline", return_value=None) as stub:
        skipped = judge(root)
    same = {key: called[key] == skipped[key] for key in called}
    print(json.dumps({"same": same, "stub_calls": stub.call_count,
                      "adoption_key": called["base_context"].get("execution_baseline_adoption")},
                     ensure_ascii=False))
    assert all(same.values()), same
    assert called["base_context"].get("execution_baseline_adoption") is False

# (2) 옛 코드에서 그 키를 읽는 자리 — 전부 참일 때만 갈래가 서는가.
tree = ast.parse((tool / "check_analysis.py").read_text(encoding="utf-8"))
reads = []
for node in ast.walk(tree):
    if isinstance(node, ast.Call) and ast.unparse(node.func).endswith(".get") and node.args \
            and isinstance(node.args[0], ast.Constant) and node.args[0].value == "execution_baseline_adoption":
        reads.append((node.lineno, ast.unparse(node)))
print(json.dumps({"reads": reads}, ensure_ascii=False))
for function in ast.walk(tree):
    if not isinstance(function, ast.FunctionDef):
        continue
    text = ast.get_source_segment((tool / "check_analysis.py").read_text(encoding="utf-8"), function) or ""
    if "execution_baseline_adoption" in text or "adopted" in text or "audited" in text:
        guards = [ast.unparse(node.test) for node in ast.walk(function) if isinstance(node, (ast.If, ast.IfExp))
                  and any(word in ast.unparse(node.test) for word in ("adopt", "audited"))]
        print(f"{function.name}: guards={guards}")
print("OK")
