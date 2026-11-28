#!/usr/bin/env python3
"""a168 tasks 0.3 — 대사 읽기 경로 함수의 `tools/logic-map` AST 산출물을 만듦(손으로 고른 줄이 아니라 도구 출력).

사용(저장소 루트에서):
    GOFLAGS=-trimpath go build -o <scratch>/logicmap ./tools/logic-map
    python3 -I openspec/changes/a168-verify-read-decode-hardening/analysis/ast/generate.py <scratch>/logicmap
    python3 -I openspec/changes/a168-verify-read-decode-hardening/analysis/ast/enumerate.py > .../enumerate-output.md

산출물 이름은 `<패키지>--<함수 소문자>.ast.json`. 실패하면 그 함수 이름을 찍고 rc 1 로 끝냄(빈 파일을 남기지 않음).
"""

from __future__ import annotations

import subprocess
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[4]

O, V, C = "internal/official", "internal/verifylive", "cmd/tossctl"
TARGETS = [
    # 편집 후보(대사 읽기·해독)
    (O + "/reconcile_reads.go", "decodeReconcileEnvelope", "official"),
    (O + "/reconcile_reads.go", "DecodeReconcileConditionalPage", "official"),
    (O + "/reconcile_reads.go", "DecodeReconcileOrderPage", "official"),
    (O + "/reconcile_reads.go", "Client.ReconcileConditionalOrdersPage", "official"),
    (O + "/reconcile_reads.go", "Client.ReconcileOpenOrdersPage", "official"),
    (O + "/reconcile_reads.go", "Client.ReconcileInstrument", "official"),
    (O + "/reconcile_detail_reads.go", "reconcileObject", "official"),
    (O + "/reconcile_detail_reads.go", "DecodeReconcileConditionalDetail", "official"),
    (O + "/reconcile_detail_reads.go", "Client.ReconcileConditionalOrder", "official"),
    # 공용 전송(편집하지 않음 — Q1 (a) 근거)
    (O + "/client.go", "unwrapAndDecode", "official"),
    (O + "/client.go", "Client.send", "official"),
    (O + "/client.go", "Client.doRequest", "official"),
    (O + "/client.go", "Client.getAcct", "official"),
    (O + "/client.go", "Client.getWithHeaders", "official"),
    (O + "/client.go", "Client.get", "official"),
    # 재사용 후보(search-first)
    (O + "/strict_minute_candles.go", "strictMinuteFinalAttempt", "official"),
    (O + "/strict_minute_candles.go", "strictMinuteCheckJSON", "official"),
    (O + "/strict_minute_candles.go", "strictMinuteWalk", "official"),
    (O + "/strict_quote_reads.go", "Client.strictQuoteFetch", "official"),
    # 호출 자리(가드 순서)
    (V + "/reconcile.go", "Reconcile", "verifylive"),
    (V + "/reconcile_check.go", "readReconcileSet", "verifylive"),
    (V + "/reconcile_check.go", "readReconcileConditionalGroup", "verifylive"),
    (V + "/reconcile_check.go", "readReconcilePlainGroup", "verifylive"),
    (V + "/reconcile_check.go", "checkReconcileInstrument", "verifylive"),
    (V + "/reconcile_expired.go", "readReconcileByID", "verifylive"),
    # 판정 함수(리뷰 1회차 P2 — 모드별 CLOSED 행 처리 근거)
    (V + "/reconcile_check.go", "checkReconcileAbsence", "verifylive"),
    (V + "/reconcile_check.go", "checkReconcileNoFiredTrace", "verifylive"),
    (V + "/reconcile_check.go", "checkReconcileClosedStatus", "verifylive"),
    (V + "/reconcile_expired.go", "checkReconcileExpiry", "verifylive"),
    (C + "/verify_reconcile.go", "reconcileReads.ReconcileConditionalOrdersPage", "tossctl"),
    (C + "/verify_reconcile.go", "reconcileReads.ReconcileOpenOrdersPage", "tossctl"),
    (C + "/verify_reconcile.go", "reconcileReads.ReconcileInstrument", "tossctl"),
    (C + "/verify_reconcile.go", "reconcileReads.ensureIdentity", "tossctl"),
    (C + "/verify_reconcile_expired.go", "reconcileReads.ReconcileConditionalOrder", "tossctl"),
]


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: generate.py <logic-map binary>", file=sys.stderr)
        return 2
    tool = sys.argv[1]
    failed = 0
    for src, func, pkg in TARGETS:
        proc = subprocess.run([tool, "--file", src, "--func", func], cwd=ROOT, capture_output=True, text=True)
        if proc.returncode != 0 or not proc.stdout.strip():
            print(f"FAIL {src} {func}: rc={proc.returncode} {proc.stderr.strip()}", file=sys.stderr)
            failed += 1
            continue
        (HERE / f"{pkg}--{func.lower()}.ast.json").write_text(proc.stdout)
    print(f"generated {len(TARGETS) - failed}/{len(TARGETS)}")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
