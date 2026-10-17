#!/usr/bin/env python3
"""a112 8.4 — BTM 행 처분 census(gate-8.1-8.3-2026-10-04/btm-disposition-census.tsv 의 머리 규칙을 코드로 옮김, 손 분류 없음).

규칙(옛 census 머리 그대로):
- 마지막 열(GREEN observed / Measured disposition / Observed)을 분류: GAP = 'not entered' · '진입 0' · '시험 0개' · '아니오' · 'no' · '미실행'
  로 시작/포함하고 'arm entered N' 이 없음.
- kind = 번들 대상이 시험 함수(Test… / …Fixture / …fixture)인가 생산 함수인가.
- exists = 대상 함수가 현재 트리에 아직 있는가(ast.json 의 file 이 있고 그 안에 함수 이름이 있는가).

쓰기: btm_census.py <저장소 루트> <출력 tsv>
"""
from __future__ import annotations

import json
import re
import sys
from collections import Counter
from pathlib import Path

GAP_WORDS = ("not entered", "진입 0", "시험 0개", "아니오", "미실행")


def is_gap(cell: str) -> bool:
    text = cell.strip().strip("*").strip()
    if re.search(r"arm entered \d", text):
        return False
    lowered = text.lower()
    if lowered.startswith("no") and (len(lowered) == 2 or not lowered[2].isalpha()):
        return True
    return any(word in text for word in GAP_WORDS)


def rows(btm: Path) -> list[list[str]]:
    out = []
    for line in btm.read_text(encoding="utf-8").splitlines():
        if not line.startswith("| B") and not re.match(r"^\| *B\d", line):
            continue
        cells = [c.strip() for c in line.strip().strip("|").split("|")]
        if cells and re.match(r"^B\d+", cells[0]):
            out.append(cells)
    return out


def main() -> int:
    root, target = Path(sys.argv[1]), Path(sys.argv[2])
    bundles = root / "openspec/changes/a112-run-four-strategy-families-independently/analysis/function-logic"
    lines, totals = [], Counter()
    for bundle in sorted(p for p in bundles.iterdir() if p.is_dir()):
        btm, ast_path = bundle / "branch-test-map.md", bundle / "ast.json"
        if not btm.exists() or not ast_path.exists():
            continue
        ast = json.loads(ast_path.read_text(encoding="utf-8"))
        function = ast["function"]
        kind = "test" if function.startswith("Test") or function.endswith(("Fixture", "fixture")) else "prod"
        source = root / ast["file"]
        exists = "yes" if source.exists() and re.search(r"\bfunc (\([^)]*\) )?" + re.escape(function) + r"\b", source.read_text(encoding="utf-8")) else "NO"
        table = rows(btm)
        gaps = [r for r in table if is_gap(r[-1])]
        totals[(kind, exists, "gap")] += len(gaps)
        totals[(kind, exists, "rows")] += len(table)
        sample = gaps[0][-1][:60] if gaps else ""
        lines.append(f"{bundle.name}\t{kind}\t{exists}\t{len(gaps)}\t{len(table)}\t{sample}")
    with target.open("w", encoding="utf-8") as handle:
        handle.write("# a112 8.4 — BTM 행 처분 census(harness/btm_census.py — 규칙은 그 파일 머리, 손 분류 없음)\n")
        handle.write("bundle\tkind\texists\tgap_rows\ttotal_rows\tsample_disposition\n")
        handle.write("\n".join(lines) + "\n\n")
        summary = ", ".join(f"{k}={v}" for k, v in sorted(totals.items()))
        handle.write(f"# 합계 (kind, exists, gap|rows): {summary}\n")
    print(f"bundles={len(lines)} " + ", ".join(f"{k}={v}" for k, v in sorted(totals.items())))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
