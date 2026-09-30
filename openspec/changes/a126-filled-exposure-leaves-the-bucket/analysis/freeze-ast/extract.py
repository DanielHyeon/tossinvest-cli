#!/usr/bin/env python3
"""a126 freeze AST 수집 하네스 — design.md 가 인용하는 분기 열거의 유일한 출처.

각 대상 함수를 **base-commit.txt 의 커밋**에서 꺼낸 소스로 `tools/logic-map` AST 로 뽑고,
분기마다 그 줄의 소스 텍스트를 붙인 census.md 를 생성함. 손으로 고른 분기가 아니라 AST 가
열거한 전부를 적음(FLM-first — 설계 문장의 분기 주장은 이 산출물 뒤에만 씀).

실행: 저장소 루트에서 `python3 openspec/changes/a126-filled-exposure-leaves-the-bucket/analysis/freeze-ast/extract.py`
"""

from __future__ import annotations

import hashlib
import json
import os
import subprocess
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[4]  # 저장소 루트는 경로에서 유도함(절대경로 하드코딩 금지)
CHANGE_DIR = HERE.parents[1]

# (소스 파일, 함수) — design.md 가 분기를 근거로 삼는 함수 전부
TARGETS = [
    ("internal/journal/risk_bucket_owner.go", "Journal.releaseRiskBucketOwner"),
    ("internal/journal/risk_bucket_owner.go", "Journal.applyRiskBucketOwnerBindingInTx"),
    ("internal/journal/risk_bucket_owner.go", "Journal.latchReleasedOwnerLateFillInTx"),
    ("internal/journal/risk_bucket_fill.go", "Journal.applyRiskBucketFillInTx"),
    ("internal/journal/risk_bucket_fill.go", "Journal.completeRiskBucketFillActual"),
    ("internal/journal/risk_bucket_fill.go", "persistRiskBucketFillTransition"),
    ("internal/journal/risk_bucket_fill.go", "riskBucketSharedUsage"),
    ("internal/journal/risk_bucket_fill.go", "queryRiskBucketOrder"),
    ("internal/journal/risk_bucket_usage.go", "refuseStaleBucketUsage"),
    ("internal/journal/risk_bucket_usage.go", "smallestRecordedBucketLimit"),
    ("internal/journal/risk_bucket_usage.go", "latchedUsageRefusal"),
    ("internal/journal/risk_bucket_relaxation.go", "Journal.ReleaseRiskOverageLatch"),
    ("internal/riskbucket/fill.go", "ApplyFill"),
    ("internal/riskbucket/fill.go", "recomputeOverageLatches"),
    ("internal/riskbucket/fill.go", "clearResolvedUnknownLatches"),
    ("internal/riskbucket/production_snapshot_authority.go", "ReadJournalBucketUsage"),
    ("internal/riskbucket/production_snapshot_authority.go", "aggregateProductionRiskUsage"),
    ("internal/riskbucket/production_snapshot_authority.go", "loadProductionRiskEntries"),
]


def git(*args: str) -> bytes:
    return subprocess.run(["git", "-C", str(ROOT), *args], check=True, capture_output=True).stdout


def main() -> int:
    base = (CHANGE_DIR / "base-commit.txt").read_text(encoding="utf-8").strip()
    out_dir = HERE / "ast"
    out_dir.mkdir(exist_ok=True)
    env = dict(os.environ, GOFLAGS="-trimpath")
    census = [
        "# a126 freeze AST census (생성물 — 손으로 고치지 말 것)",
        "",
        f"- 기준 커밋(base-commit.txt): `{base}`",
        "- 생성기: `analysis/freeze-ast/extract.py` → `go run ./tools/logic-map --file <f> --func <fn>`",
        "- 각 분기 줄의 텍스트는 같은 base blob 에서 읽음. 열거는 AST 가 낸 전부임(선택 없음).",
        "",
    ]
    for source, function in TARGETS:
        blob = git("show", f"{base}:{source}")
        # extract 는 워킹트리 파일을 읽으므로, 워킹트리가 base blob 과 바이트 단위로 같을 때만 뽑음 — 다르면 멈춤.
        current = (ROOT / source).read_bytes()
        if current != blob:
            print(f"[freeze-ast] {source} differs from base {base[:12]} — stop (re-pin or extract from a clean tree)")
            return 1
        result = subprocess.run(["go", "run", "./tools/logic-map", "--file", source, "--func", function],
                                cwd=ROOT, env=env, check=True, capture_output=True)
        value = json.loads(result.stdout)
        if value.get("source_sha256") != hashlib.sha256(blob).hexdigest():
            print(f"[freeze-ast] {source}:{function} AST hash is not the base blob hash — stop")
            return 1
        value["revision"] = "base"
        value["base_commit"] = base
        name = f"{Path(source).parent.as_posix().replace('/', '-')}--{function.lower()}.json"
        (out_dir / name).write_text(json.dumps(value, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
        lines = blob.decode("utf-8").splitlines()
        branches = value.get("branches") or []
        census.append(f"## `{source}` · `{function}` (L{value['start']['line']}–{value['end']['line']}, 분기 {len(branches)})")
        census.append("")
        census.append(f"AST: `ast/{name}` · sha256 `{value['source_sha256'][:16]}…`")
        census.append("")
        census.append("| id | kind | line | source |")
        census.append("|---|---|---|---|")
        for branch in branches:
            line = branch["at"]["line"]
            text = lines[line - 1].strip().replace("|", "\\|")
            if len(text) > 150:
                text = text[:147] + "..."
            census.append(f"| {branch['id']} | {branch['kind']} | {line} | `{text}` |")
        census.append("")
    (HERE / "census.md").write_text("\n".join(census) + "\n", encoding="utf-8")
    print(f"[freeze-ast] {len(TARGETS)} functions extracted at {base[:12]}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
