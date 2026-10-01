#!/usr/bin/env python3
"""a112 6.3 잔여 (c) — 재검증 클로저의 drift 판정을 순수 함수로 옮긴 편집 뒤 번들(render_5222_bundles.main 재사용).

편집 전 번들: `analysis/measurements/lot-6.3/pre-edit/`(HEAD 76dc35f7). 이동 영수증: `lot-6.3/move-receipt.txt`. 변이: `lot-6.3/mutation-6.3.tsv`.
사용: python3 render_63_bundles.py   (저장소 루트에서)
"""
from __future__ import annotations

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import render_5222_bundles as base  # noqa: E402

LEDGER = "`analysis/measurements/lot-6.3/mutation-6.3.tsv`"
SAME = ("분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로", "no — 이 로트가 바꾸지 않음", "yes")
FUNCS = [
    {"bundle": "internal-app-engine--context.newpairedstrategyentryproductionassembly", "file": "internal/app/engine/strategy_entry_supervisor.go",
     "func": "Context.NewPairedStrategyEntryProductionAssembly", "title": "Context.NewPairedStrategyEntryProductionAssembly",
     "revision": "편집 전 9 분기 → 8: 재검증 클로저 안의 drift 판정 if(편집 전 B6 자리)를 순수 함수 `strategyScheduleStillMatchesAdmission`"
                 "(`strategy_schedule_revalidation.go`)로 **의미 무변경 이동** — 클로저는 수집 한 문장 + 그 함수 호출 한 문장. 조건식 철자 동일(이동 영수증). "
                 "나머지 분기는 번호만 당겨졌다.",
     "scen": {
         "B1": ("nil Context → 오류",) + SAME, "B2": ("원장 경로 있음 → 후보 · 원장 경로",) + SAME, "B3": ("원장 경로 → 근거 경로",) + SAME,
         "B4": ("레인 세우기 실패 → 오류",) + SAME, "B5": ("시계 있음 → dispatch now 배선",) + SAME, "B6": ("KR · US worker 만들기",) + SAME,
         "B7": ("감독자 생성 실패 → 오류",) + SAME, "B8": ("투영 발행 실패 → 오류",) + SAME,
     },
     "invariants": ["재검증의 drift 판정은 `strategyScheduleStillMatchesAdmission` 하나 — 클로저 모양(수집 + 그 함수 반환)과 판정 조건식 철자를 "
                    "`TestTheScheduleDriftJudgementMovedVerbatim` 이 양쪽 못 박음, 축별 거절은 `TestTheScheduleDriftJudgementRefusesEveryAxis`"
                    f"(변이 A1 · A4~A9 CAUGHT, A2 · A3 는 행동상 동등 — 철자 핀만 잡음, {LEDGER})."],
     "state": "편집 전과 같다(조립은 원장을 쓰지 않음, 투영 발행만).",
     "safety": ["High-risk 인접(주문 경로 최종 권한 재검사의 판정) — 의미 무변경 이동. 거절 · 통과 집합 불변(조건식 철자 동일)."]},
]

if __name__ == "__main__":
    base.main(FUNCS, tag="a112 6.3", pre="lot-6.3", ledger=LEDGER)
