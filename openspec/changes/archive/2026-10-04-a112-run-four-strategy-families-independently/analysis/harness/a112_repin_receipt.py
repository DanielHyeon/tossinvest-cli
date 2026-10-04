#!/usr/bin/env python3
"""a112 base 재고정 영수증 — a092 `repin_receipt.py`(a066 계보) 에서 CHANGE · 제목 grep(`a112`) 만 바꿈. 사용법 · 판정은 아래 원문 그대로.
번들로 덮였는가)을 **전수**로 잰다(Manager 판정 2026-09-29 Q8).

사용: python3 repin_receipt.py <old-base> <landing> > receipt.tsv

- 자기 Go 커밋 = 옛 창(old-base..landing)에서 이 change 디렉터리를 만진(--full-history) **또는** 제목에 `(a092)` 이 있는
  비병합 커밋 중 `.go` 를 고친 것. 두 기준의 합집합 — 한쪽만 쓰면 Go 수리와 문서를 다른 커밋으로 쪼갠 경우를 놓친다.
- 커밋마다 게이트와 같은 계산(`check_analysis.changed_existing_functions(base=c^, target=c)`)으로 그 커밋이 바꾼
  **기존** 함수를 센다.
- 각 함수에 대해 이 change 의 번들(analysis/function-logic/*/ast.json 의 file+qualified function)을 찾고, 그 번들의
  source_sha256 이 착지 리비전의 파일 바이트와 같은지(= 착지에서 fresh) 확인한다. 착지에 그 함수가 없으면 GONE 으로 적는다.
출력: 요약 줄(# 으로 시작) + 함수별 TSV.
"""
from __future__ import annotations

import hashlib
import json
import subprocess
import sys
from pathlib import Path

ROOT = Path(subprocess.run(["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True, check=True).stdout.strip())
sys.path.insert(0, str(ROOT / "tools" / "logic-map"))
import check_analysis  # noqa: E402

CHANGE = "openspec/changes/a112-run-four-strategy-families-independently"


def git(*args: str) -> str:
    return subprocess.run(["git", "-C", str(ROOT), *args], capture_output=True, text=True, check=True).stdout


def self_go_commits(old: str, landing: str) -> list[str]:
    by_dir = set(git("log", "--no-merges", "--full-history", "--format=%H", f"{old}..{landing}", "--", f"{CHANGE}/").split())
    by_subject = set(git("log", "--no-merges", "--format=%H", "--fixed-strings", "--grep=a112", f"{old}..{landing}").split())
    ordered = git("log", "--no-merges", "--reverse", "--format=%H", f"{old}..{landing}").split()
    chosen = []
    for sha in ordered:
        if sha not in by_dir and sha not in by_subject:
            continue
        files = git("diff-tree", "--no-commit-id", "--name-only", "-r", sha).split("\n")
        if any(f.endswith(".go") for f in files):
            chosen.append(sha)
    return chosen


def bundles() -> dict[tuple[str, str], tuple[str, str]]:
    found = {}
    for ast_path in sorted((ROOT / CHANGE / "analysis" / "function-logic").glob("*/ast.json")):
        value = json.loads(ast_path.read_text())
        found[(value["file"], check_analysis.qualified(value))] = (ast_path.parent.name, value.get("source_sha256", ""))
    return found


def landing_sha(landing: str, path: str) -> str | None:
    blob = subprocess.run(["git", "-C", str(ROOT), "show", f"{landing}:{path}"], capture_output=True)
    if blob.returncode:
        return None
    return hashlib.sha256(blob.stdout).hexdigest()


def main() -> int:
    old, landing = sys.argv[1], sys.argv[2]
    commits = self_go_commits(old, landing)
    # 게이트가 요구하는 집합: 옛 base 에 **있던** 함수 중 창에서 바뀐 것(= 옛 base 에서 check_analysis 가 세는 required).
    required = set(check_analysis.changed_existing_functions(ROOT, base=old, target=landing))
    touched: dict[tuple[str, str], list[str]] = {}
    for sha in commits:
        for key in check_analysis.changed_existing_functions(ROOT, base=f"{sha}^", target=sha):
            touched.setdefault(key, []).append(sha[:8])
    have = bundles()
    rows = []
    counts: dict[str, int] = {}
    for (path, function), shas in sorted(touched.items()):
        if (path, function) not in required:
            # 옛 base 에 없던 함수(이 창에서 새로 생긴 것)이거나 창 끝에서 base 와 같아진 것 — 게이트가 요구하지 않음.
            status, bundle_name = "NOT-REQUIRED(new-since-base-or-unchanged-at-landing)", "-"
        else:
            bundle = have.get((path, function))
            at_landing = landing_sha(landing, path)
            if bundle is None:
                status, bundle_name = ("GONE(file)" if at_landing is None else "MISSING"), "-"
            else:
                bundle_name = bundle[0]
                status = "FRESH" if bundle[1] == at_landing else "STALE"
        counts[status] = counts.get(status, 0) + 1
        rows.append((path, function, ",".join(shas), bundle_name, status))
    print(f"# old base {old[:8]} -> landing {landing[:8]}; self Go commits {len(commits)}: {' '.join(c[:8] for c in commits)}")
    print(f"# required at the old base (gate window) {len(required)}; functions touched by self commits {len(touched)}; "
          + "; ".join(f"{k} {v}" for k, v in sorted(counts.items())))
    print("file\tfunction\tcommits\tbundle\tstatus")
    for row in rows:
        print("\t".join(row))
    return 0


if __name__ == "__main__":
    sys.exit(main())
