#!/usr/bin/env python3
"""a112 8.4 — 생산 BTM GAP 행을 모듈 전체 커버리지 union 으로 다시 잰다(gate-8.1-8.3-2026-10-04/btm-remeasure-prod.tsv 의 측정 2 와 같은 방식).

입력: 저장소 사본 루트 · 모듈 union 커버 프로필(`go test -tags tossos_testseams -covermode=set -coverpkg=<패키지들> ./...`) · 출력 tsv ·
     (선택) 이미 잰 행 목록 tsv(bundle · branch 두 열) — 거기 있는 행은 건너뛴다(옛 재측정이 처분한 128 행).
판정: ENTERED-module = 그 갈래 몸통 블록이 실행됨 · NOT-ENTERED-module = 실행 안 됨(대조 처분 필요) · NO-BLOCK = 몸통 블록 없음 ·
     STALE-ROW = 번들 행이 현재 ast 에 없음. 분기 → 블록 규칙은 branch_coverage.branch_block.

쓰기: btm_remeasure.py <사본 루트> <union.out> <출력 tsv> [<건너뛸 행 tsv>]
"""
from __future__ import annotations

import json
import re
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import branch_coverage  # noqa: E402
import btm_census  # noqa: E402


def main() -> int:
    root, profile, target = Path(sys.argv[1]), Path(sys.argv[2]), Path(sys.argv[3])
    skip: set[tuple[str, str]] = set()
    if len(sys.argv) > 4:
        for line in Path(sys.argv[4]).read_text(encoding="utf-8").splitlines():
            parts = line.split("\t")
            if len(parts) > 1 and not line.startswith(("#", "bundle")):
                skip.add((parts[0], parts[1]))
    module = branch_coverage.module_path(root)
    blocks = branch_coverage.parse_profile(profile)
    bundles = root / "openspec/changes/a112-run-four-strategy-families-independently/analysis/function-logic"
    out = ["bundle\tbranch\tmodule_union\tblock\tdisposition_cell"]
    counts: dict[str, int] = {}
    for bundle in sorted(p for p in bundles.iterdir() if p.is_dir()):
        btm, ast_path = bundle / "branch-test-map.md", bundle / "ast.json"
        if not btm.exists() or not ast_path.exists():
            continue
        ast = json.loads(ast_path.read_text(encoding="utf-8"))
        function = ast["function"]
        if function.startswith("Test") or function.endswith(("Fixture", "fixture")):
            continue
        source = root / ast["file"]
        if not source.exists() or not re.search(r"\bfunc (\([^)]*\) )?" + re.escape(function) + r"\b", source.read_text(encoding="utf-8")):
            continue
        branches = {b["id"]: b for b in ast.get("branches") or []}
        file_key = module + "/" + ast["file"]
        for row in btm_census.rows(btm):
            if not btm_census.is_gap(row[-1]):
                continue
            branch_id = re.match(r"B\d+", row[0]).group(0)
            if (bundle.name, branch_id) in skip:
                continue
            branch = branches.get(branch_id)
            if branch is None:
                verdict, span = "STALE-ROW", ""
            else:
                block = branch_coverage.branch_block(blocks.keys(), file_key, branch)
                if block is None:
                    verdict, span = "NO-BLOCK", ""
                else:
                    verdict = "ENTERED-module" if blocks[block] else "NOT-ENTERED-module"
                    span = f"{block[1]}.{block[2]}-{block[3]}.{block[4]}"
            counts[verdict] = counts.get(verdict, 0) + 1
            out.append(f"{bundle.name}\t{branch_id}\t{verdict}\t{span}\t{row[-1][:80]}")
    target.write_text("\n".join(out) + "\n\n# 합계: " + ", ".join(f"{k}={v}" for k, v in sorted(counts.items())) + "\n", encoding="utf-8")
    print(", ".join(f"{k}={v}" for k, v in sorted(counts.items())))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
