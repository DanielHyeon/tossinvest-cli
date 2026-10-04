#!/usr/bin/env python3
"""a112 게이트 준비(2026-10-04) — 후보 base 의 창이 요구하는 함수 전수와 각 함수의 번들 신선도·귀속 영수증.

base 재고정(사람 승인 항목)의 **준비물**이다 — 재고정을 실행하지 않는다(base-commit.txt 를 읽지도 쓰지도 않음).
사용(저장소 루트에서): gateprep_window_receipt.py <candidate-base> > receipt.tsv

- 요구 집합 = 게이트와 같은 계산 `check_analysis.changed_existing_functions(root, base, "")`(대상 = 워킹트리).
- 귀속 = `git log -L :^func …:file base..HEAD` 의 커밋 제목에 처음 나오는 change 태그(aNNN); -L 이 함수를 못 찾거나 커밋 0 이면 그 파일의 이력.
- GONE = 번들은 있으나 파일이 지금 없음(삭제된 시험 등 — 게이트는 그 번들을 요구 집합 밖으로 둔다).
- 번들 = 이 change 의 analysis/function-logic/*/ast.json 중 file + (receiver.)function 이 같은 것; FRESH = 그 ast 의 source_sha256 이 지금 파일 바이트와 같음.
출력: `#` 요약 줄 + 함수별 TSV(file, function, attribution, bundle, freshness).
"""
from __future__ import annotations

import collections
import hashlib
import json
import re
import subprocess
import sys
from pathlib import Path

ROOT = Path(subprocess.run(["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True, check=True).stdout.strip())
sys.path.insert(0, str(ROOT / "tools" / "logic-map"))
import check_analysis  # noqa: E402

CHANGE = ROOT / "openspec/changes/a112-run-four-strategy-families-independently"


def bundles() -> dict[tuple[str, str], tuple[str, str]]:
    out = {}
    for ast_path in sorted((CHANGE / "analysis/function-logic").glob("*/ast.json")):
        ast = json.loads(ast_path.read_text())
        name = ast["function"] if not ast.get("receiver") else ast["receiver"].lstrip("*") + "." + ast["function"]
        out[(ast["file"], name)] = (ast_path.parent.name, ast["source_sha256"])
    return out


def attribution(base: str, path: str, function: str) -> str:
    if "." in function:
        receiver, name = function.split(".", 1)
        pattern = f"^func ([^)]*{receiver}) {name}("
    else:
        pattern = f"^func {function}("
    run = subprocess.run(["git", "-C", str(ROOT), "log", "-L", f":{pattern}:{path}", "--format=%s", "-s", f"{base}..HEAD"],
                         capture_output=True, text=True)
    subjects = [s for s in run.stdout.splitlines() if s.strip()] if run.returncode == 0 else []
    if not subjects:  # -L 이 함수를 못 찾았거나(오류) 찾고도 커밋 0(이름 · 시그니처가 창 안에서 바뀐 경우) — 파일 이력으로
        subjects = subprocess.run(["git", "-C", str(ROOT), "log", "--format=%s", f"{base}..HEAD", "--", path],
                                  capture_output=True, text=True).stdout.splitlines()
    tags = sorted({(re.findall(r"\b(a\d{3})\b", s) or ["untagged"])[0] for s in subjects if s.strip()})
    return ",".join(tags) or "none"


def main() -> int:
    base = sys.argv[1]
    required = check_analysis.changed_existing_functions(ROOT, base, "")
    have = bundles()
    rows, tally = [], collections.Counter()
    for (path, function) in sorted(required):
        bundle, sha = have.get((path, function), ("", ""))
        current = hashlib.sha256((ROOT / path).read_bytes()).hexdigest() if (ROOT / path).exists() else "GONE"
        fresh = "NO-BUNDLE" if not bundle else ("GONE" if current == "GONE" else ("FRESH" if sha == current else "STALE"))
        owner = attribution(base, path, function)
        tally[("a112" if "a112" in owner.split(",") else "other", fresh)] += 1
        rows.append((path, function, owner, bundle or "-", fresh))
    head = subprocess.run(["git", "-C", str(ROOT), "rev-parse", "--short=12", "HEAD"], capture_output=True, text=True).stdout.strip()
    print(f"# base {base} -> working tree at HEAD {head}: required {len(required)}")
    for key, count in sorted(tally.items()):
        print(f"# {key[0]} {key[1]}: {count}")
    print("file\tfunction\tattribution\tbundle\tfreshness")
    for row in rows:
        print("\t".join(row))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
