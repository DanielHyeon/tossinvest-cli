#!/usr/bin/env python3
"""a112 2.3 (b) — evaluateFresh 에 1.2 반사실 기록 갈래를 더한 편집의 편집 뒤 번들(render_5222_bundles.main 재사용).

편집 전 번들: `analysis/measurements/lot-2.3/pre-edit/`(HEAD d54f4dca). RED: `lot-2.3/red-2.3.log`. 변이: `lot-2.3/mutation-2.3.tsv`.
"""
from __future__ import annotations

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import render_5222_bundles as base  # noqa: E402

LEDGER = "`analysis/measurements/lot-2.3/mutation-2.3.tsv`"
CF = "`a112_rvol_counterfactual_test.go`"
SAME = "no — 갈래 불변"
FUNCS = [
    {"bundle": "internal-breakoutlane--evaluatefresh", "file": "internal/breakoutlane/machine.go", "func": "evaluateFresh", "title": "evaluateFresh",
     "revision": "편집 전 19 분기 → 20: 범위 뒤 봉 루프(B4) 안, 입장 갈래 B6(입장 시 break) **뒤**에 1.2 반사실 기록 B7 을 더했다(Manager 판정 2.3 (b)). "
                 "그래서 편집 전 B7~B19 는 B8~B20 이 되었다(위치 번호). 입장 갈래와 그 안의 1.2 · 2.0 · 2.5 기록 줄은 바이트 그대로다.",
     "scen": {
         "B1": ("범위 봉 순회(저항 · 범위 하단)", "`TestEvaluatorDerivesOnlyCompleteClosedBarPathToProposal`", SAME, "yes"),
         "B2": ("범위 봉 high 가 저항을 올림", "`TestEvaluatorDerivesOnlyCompleteClosedBarPathToProposal` · `TestGstackRepairFrozenVocabularyAndV1Thresholds`", SAME, "yes"),
         "B3": ("범위 봉 low 가 하단을 내림", "`TestNoAveragingDownLegAfterAFailedSetup`(범위 하단 아래 종가)", SAME, "yes"),
         "B4": ("범위 뒤 봉 순회", "`TestEvaluatorDerivesOnlyCompleteClosedBarPathToProposal`", SAME, "yes"),
         "B5": ("high 만 저항 위(첫 touch)", "`TestAdversarialFirstTouchMissingRangeAndBadBarCannotPropose`", SAME, "yes"),
         "B6": ("입장 돌파(close buffer · RVOL >= 1.5 · wick) → 반사실 1.2/2.0/2.5 기록 후 break",
                "`TestRVOLAdmissionAndCounterfactualBoundaries` · `TestTheAdmittedBreakoutPathIsUnchangedByTheCounterfactual`", "no — 갈래 · 몸 불변, 변이 CF-10 CAUGHT", "yes"),
         "B7": ("**(새)** 입장 못 한 봉: close buffer · wick 통과 · 1.2 <= RVOL → `RVOLAt1200000` 기록(기록 전용)",
                f"{CF} `TestTheOnePointTwoCounterfactualRecordsABarThatOnlyTheLowerThresholdWouldAdmit`",
                f"yes — 편집 전 1.2 정확 · 1.5 바로 아래 두 경우 실패(`red-2.3.log`); 변이 CF-01~09 CAUGHT({LEDGER})", "yes"),
         "B8": ("입장 돌파 없음 → 범위 단계 결정", "`TestRVOLAdmissionAndCounterfactualBoundaries`(1.5 바로 아래) · " + CF, SAME, "yes"),
         "B9": ("첫 touch 거절", "`TestAdversarialFirstTouchMissingRangeAndBadBarCannotPropose`", SAME, "yes"),
         "B10": ("US 시한", "`TestTimeoutExactBoundaryKRAndUS`", SAME, "yes"),
         "B11": ("돌파 뒤 봉 순회", "`TestEvaluatorDerivesOnlyCompleteClosedBarPathToProposal`", SAME, "yes"),
         "B12": ("범위 하단 아래 종가 → INVALIDATED", "`TestNoAveragingDownLegAfterAFailedSetup` · `TestTheObservedBreakoutEdgesPlusTheReservedSixAreTheGoldenSet`", SAME, "yes"),
         "B13": ("`since > timeout` — **도달 불가**(B16 이 같음에서 먼저 반환, 스위트 커버리지 0 실측) · 동작 동등 방어",
                 "`TestTheBreakoutTransitionProducersAreExactlyTheCensus`(생산 자리 수로 고정 — BK2-08)", SAME, "도달 불가 — census 만"),
         "B14": ("retest 뒤 reclaim → RECLAIMED · ARMED", "`TestEvaluatorDerivesOnlyCompleteClosedBarPathToProposal`", SAME, "yes"),
         "B15": ("거래량 확장 실패 재돌파 → INVALIDATED", "`TestGstackRepairFailedReclaimAndRiskRewardBoundary` · `TestNoAveragingDownLegAfterAFailedSetup`", SAME, "yes"),
         "B16": ("시한 정확 → TIMED_OUT", "`TestTimeoutExactBoundaryKRAndUS`", SAME, "yes"),
         "B17": ("retest 허용폭 안", "`TestGstackRepairRetestQualifiesToleranceEndpoints`", SAME, "yes"),
         "B18": ("ARMED 아님 → 진행 단계 결정", "`TestAdversarialCorrectionReplaysPreTerminalAndPreservesProposed`(RETEST_WAIT)", SAME, "yes"),
         "B19": ("호가 거절 → ARMED + typed refusal", "`TestEvaluatorInvalidationTimeoutAndQuoteVetoEdges` · `TestQuoteVetoMatchesTheGoldenFormulasAtAndBeyondEachLimit`", SAME, "yes"),
         "B20": ("사이징 거절 → ARMED + typed refusal", "`TestANonProtectiveStopNeverProposes` · `TestSizingMatchesTheExactRationalOracle`", SAME, "yes"),
     },
     "invariants": ["1.2 반사실 = 입장 문턱이 1.2 였다면 이 setup 에 돌파 봉이 있었는가(design.md 「1.2 반사실의 의미」).",
                    "기록 전용: `decisionSeal` 은 provenance 의 RVOL 플래그를 포함하지 않는다(전이 · 계보만) — 단계 · 전이 · 거절 · 수량 · 제안 · 봉인 불변(쌍둥이 비교 시험).",
                    "입장 경로(B6)와 그 안의 1.2 · 2.0 · 2.5 기록 줄은 바이트 그대로 — B7 은 B6 의 break 뒤라 입장 봉에는 닿지 않는다.",
                    "패키지 밖 소비자 0(grep: `RVOLAt1200000` · `Provenance()` 의 breakout 소비 hit 0) — breakout 미배선이라 생산 효과 0."],
     "state": "상태 변경 없음 — 반환하는 Decision 의 provenance 플래그 하나.",
     "safety": ["High-risk(돌파 판정 · 증거) — 판정 경로 무변경, 기록만 늘었다. 거절 · 수락 집합 불변(쌍둥이 비교 · 기존 스위트 · 사이징/호가 신탁 GREEN)."]},
]


if __name__ == "__main__":
    (base.FL / FUNCS[0]["bundle"]).mkdir(parents=True, exist_ok=True)
    base.main(FUNCS, tag="a112 2.3", pre="lot-2.3", ledger=LEDGER)
