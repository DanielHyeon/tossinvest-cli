#!/usr/bin/env python3
"""a125 1.1 · 3.1 (codex freeze F4) — 옛 도구와 새 도구(특례 제거)가 **기록 없는** change 에서 같은 판정 · 창 · CLI 를 내는가.

    python3 ab_old_new.py --tool <tools/logic-map 디렉터리> --out <json>     # 판 하나를 잰다
    python3 ab_old_new.py --compare <old.json> <new.json>                     # 두 판을 대조한다

모양 셋(일반 워킹트리 창 · 착지 기록 · 증거 빌림)을 **같은 커밋 시각**으로 세워 두 판에서 sha 가 같게 한다. 비교 계약:
판정 줄 · 문맥(제거가 승인된 키 `execution_baseline_adoption` 만 정규화) · rc · `main` 출력(임시 경로만 정규화)이 같아야 한다.
주장의 범위는 **현재 모집단**이다 — 기록을 가진 비-a063 change 가 생기면 옛 도구는 거절, 새 도구는 무시로 갈린다(design D2).
"""
import argparse
import io
import json
import os
import shutil
import subprocess
import sys
import tempfile
from contextlib import redirect_stdout
from pathlib import Path
from unittest import mock

parser = argparse.ArgumentParser()
parser.add_argument("--tool")
parser.add_argument("--out")
parser.add_argument("--compare", nargs=2)
args = parser.parse_args()

REMOVED_KEYS = ("execution_baseline_adoption",)

if args.compare:
    old, new = (json.loads(Path(path).read_text()) for path in args.compare)
    diffs = []
    for shape in old:
        a, b = old[shape], new[shape]
        for key in ("verdict", "rc", "output"):
            if a[key] != b[key]:
                diffs.append((shape, key, a[key], b[key]))
        ca = {k: v for k, v in a["context"].items() if k not in REMOVED_KEYS}
        if ca != b["context"]:
            diffs.append((shape, "context", ca, b["context"]))
    print(json.dumps({"shapes": sorted(old), "diffs": diffs}, ensure_ascii=False, indent=1))
    raise SystemExit(1 if diffs else 0)

tool = Path(args.tool).resolve()
sys.path.insert(0, str(tool))
import check_analysis  # noqa: E402
import fixture_git_env  # noqa: E402

fixture_git_env.isolate()
os.environ.update({"GIT_AUTHOR_DATE": "2026-09-29T00:00:00+09:00", "GIT_COMMITTER_DATE": "2026-09-29T00:00:00+09:00",
                   "GIT_AUTHOR_NAME": "a125", "GIT_AUTHOR_EMAIL": "a125@example.invalid",
                   "GIT_COMMITTER_NAME": "a125", "GIT_COMMITTER_EMAIL": "a125@example.invalid"})


def git(root: Path, *argv: str) -> str:
    return subprocess.run(["git", *argv], cwd=root, check=True, capture_output=True, text=True).stdout.strip()


def commit(root: Path, subject: str) -> str:
    git(root, "add", "-A"); git(root, "commit", "-qm", subject)
    return git(root, "rev-parse", "HEAD")


def bundle(change: Path, source: Path, root: Path) -> None:
    value = check_analysis.go_functions(source, root)[0]
    # 추출기는 절대경로를 적는다 — 임시 디렉터리 이름이 커밋에 들어가면 두 판의 sha 가 갈린다. 상대경로로 적는다.
    value.update({"file": "internal/own.go", "package": "internal", "signature": "Own(params=0, results=1)", "branches": []})
    target = change / "analysis" / "function-logic" / "internal--own"
    target.mkdir(parents=True)
    (target / "ast.json").write_text(json.dumps(value, sort_keys=True))
    (target / "function-logic-map.md").write_text(
        "# Function Logic Map: `Own`\ninternal/own.go\n## Inputs and invariants\ne\n## Branches and early returns\ne\n"
        "## Calls and live bindings\ne\n## State mutations and fallbacks\ne\n## Safety conclusion\ne\n")
    (target / "branch-test-map.md").write_text("# Branch Test Map: `Own`\n| B1 | leaf | test | yes | yes |\n")
    (target / "risk-pattern-report.md").write_text("# Risk Pattern Report\ninternal/own.go\n")


def build(root: Path, shape: str) -> str:
    git(root, "init", "-q", "-b", "main")
    (root / "go.mod").write_text("module fixture\ngo 1.23\n")
    # 추출기 Go 파일만, 같은 모드로 — 두 판의 도구 디렉터리는 파일 모드가 달라(복사본) 픽스처 sha 가 갈렸다.
    (root / "tools" / "logic-map").mkdir(parents=True)
    for name in ("extract_go_ast.go",):
        (root / "tools" / "logic-map" / name).write_bytes((tool / name).read_bytes())
    own = root / "internal" / "own.go"; own.parent.mkdir(parents=True)
    own.write_text("package internal\nfunc Own() int { return 1 }\n")
    other = root / "internal" / "other.go"
    other.write_text("package internal\nfunc Other() int { return 1 }\n")
    base = commit(root, "P")
    change = root / "openspec" / "changes" / "mine"; change.mkdir(parents=True)
    (change / "base-commit.txt").write_text(base + "\n"); (change / "review.md").write_text("mine\n")
    own.write_text("package internal\nfunc Own() int { return 2 }\n")
    commit(root, "W: this change's Go work")
    if shape == "borrowed":
        lender = root / "openspec" / "changes" / "lender"; lender.mkdir(parents=True)
        (lender / "base-commit.txt").write_text(base + "\n"); (lender / "review.md").write_text("lender\n")
        bundle(lender, own, root)
        (change / "analysis").mkdir()
        (change / "analysis" / "function-logic-reference.txt").write_text("lender\n")
    else:
        bundle(change, own, root)
    commit(root, "E: evidence")
    other.write_text("package internal\nfunc Other() int { return 2 }\n")
    commit(root, "S: a sibling lands")
    if shape == "landed":
        code, lines = check_analysis.record_landing("mine", root)
        assert code == 0, lines
        commit(root, "record the landing")
    return "mine"


rows = {}
for shape in ("working-tree", "landed", "borrowed"):
    with tempfile.TemporaryDirectory() as raw:
        root = Path(raw)
        change = build(root, shape)
        context: dict[str, object] = {}
        verdict = check_analysis.check(change, root, context)
        output = io.StringIO()
        with mock.patch.object(sys, "argv", ["check_analysis.py", "--change", change, "--root", str(root)]), \
                redirect_stdout(output):
            code = check_analysis.main()
        rows[shape] = {"verdict": [line.replace(raw, "<root>") for line in verdict],
                       "context": {k: (v.replace(raw, "<root>") if isinstance(v, str) else v) for k, v in context.items()},
                       "rc": code, "output": output.getvalue().replace(raw, "<root>")}
        print(shape, rows[shape]["rc"], len(rows[shape]["verdict"]), flush=True)
Path(args.out).write_text(json.dumps(rows, ensure_ascii=False, indent=1, default=str))
