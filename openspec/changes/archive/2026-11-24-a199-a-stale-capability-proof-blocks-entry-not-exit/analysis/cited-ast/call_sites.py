#!/usr/bin/env python3
"""a199 1.4 — 호출자 함수의 AST(calls)에서 대상 함수를 부르는 **자리**를 셈(CodeGraph 는 함수 단위로만 답함).

사용: python3 call_sites.py   (이 폴더의 *.ast.json 을 읽음 — extract.py 를 먼저 돌릴 것)
"""

from __future__ import annotations

import json
from pathlib import Path

HERE = Path(__file__).resolve().parent

# (대상 호출 텍스트, 호출자 AST 파일) — 생산 호출자만(CodeGraph callers 원문의 비시험 행)
PAIRS = [
    ("verifyGate", "internal-app-engine--runinterlock.ast.json"),
    ("refuseStartup", "internal-app-engine--runinterlock.ast.json"),
    ("refuseStartup", "internal-app-engine--newcontext.ast.json"),
    ("logGateDecision", "internal-app-engine--runinterlock.ast.json"),
    ("runInterlock", "internal-app-engine--newcontext.ast.json"),
    ("engine.NewContext", "cmd-tossctl--assembleengine.ast.json"),
    ("NewContext", "internal-app-engine--new.ast.json"),
    ("ectx.ExitObserver", "cmd-tossctl--engineruntime.ast.json"),
    ("ectx.ReconcileDriver", "cmd-tossctl--engineruntime.ast.json"),
    ("ectx.AttestationWatch", "cmd-tossctl--engineruntime.ast.json"),
    ("judgeAttestation", "internal-app-engine--attestationwatch.prepare.ast.json"),
    ("newAttestationWatch", "internal-app-engine--context.attestationwatch.ast.json"),
    ("c.productionStrategyWorker", "internal-app-engine--context.newpairedstrategyentryproductionassembly.ast.json"),
    ("att.Verify", "internal-app-engine--verifygate.ast.json"),
    ("att.Verify", "internal-app-engine--judgeattestation.ast.json"),
]


def main() -> None:
    for target, caller in PAIRS:
        d = json.loads((HERE / caller).read_text(encoding="utf-8"))
        lines = [c["at"]["line"] for c in (d.get("calls") or []) if c.get("text") == target]
        print(f"{target:32} in {d['function']:45} 자리 {len(lines)}  줄 {lines}")


if __name__ == "__main__":
    main()
