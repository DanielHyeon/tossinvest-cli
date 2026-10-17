#!/usr/bin/env python3
"""a112 7.5 — 편집 뒤 번들(FLM/BTM)을 ast.json 에서 채운다(render_5222_bundles.main 재사용 — 좌표 · 호출 · return 은 AST 만).

편집 전 번들: `analysis/measurements/lot-7.5/pre-edit/`(HEAD f03b6ec1 — Lane.Health 는 편집 뒤 HEAD 사본에서 렌더, review 기록).
변이 원장: `analysis/measurements/lot-7.5/mutation-7.5.tsv`(R01~R11).
사용: python3 render_75_bundles.py   (저장소 루트에서)
"""
from __future__ import annotations

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import render_5222_bundles as base  # noqa: E402

LEDGER = "`analysis/measurements/lot-7.5/mutation-7.5.tsv`"
OT = "`a112_lane_operability_test.go`"
LT = "`a112_lane_latency_testseam_test.go`"
FT = "`a112_lane_fanout_test.go`"
PT = "`a112_lane_coordinator_projection_test.go`"
SAME = ("분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로", "no — 이 로트가 바꾸지 않음", "yes")
SAFE = ["High-risk 인접(레인 런타임 동시성) — 주문 · 원장 · 활성화 쓰기 없음. 레인끼리 상태 공유 0(Lane 구조), goroutine 하나가 레인 하나, "
        "관측은 자기 색인 칸에만. 생산 레인은 전부 DORMANT(서명 매니페스트 0)."]
E = "internal/app/engine/"

FUNCS = [
    {"bundle": "internal-app-engine--strategylaneruntime.evaluate", "file": E + "strategy_lane_runtime.go", "func": "strategyLaneRuntime.evaluate",
     "title": "strategyLaneRuntime.evaluate",
     "revision": "편집 전 6 분기 → 8: 레인 순회(B3)가 레인마다 goroutine 하나를 띄우고(go 문 1 · defer 2 — join.Done · recover) join 한 뒤, "
                 "B6 · B7(레인 goroutine 의 panic 을 시장 주기 goroutine 에서 다시 던짐)을 더했다. 관측은 레인 순서 색인으로 모음. 나머지 분기 불변.",
     "scen": {
         "B1": ("nil 런타임", "진입 0(편집 전 번들 기록 그대로)", "no — 이 로트가 바꾸지 않음", "n/a"),
         "B2": ("복구 실패 → 오류", "진입 0(편집 전 번들 기록 그대로)", "no — 이 로트가 바꾸지 않음", "n/a"),
         "B3": ("**(편집)** 레인마다 goroutine 하나 · join(시장 주기 안) — 멈춘 레인이 이웃을 세우지 않음, 시장 지연 = 최댓값",
                f"{LT} `TestAHungLaneDoesNotDelayItsPeersInTheSameWave`(경과 = 마감 시한 정확히 1 회, 이웃은 가상 시각 0 에 반환)",
                f"yes — 편집 전 컴파일 실패(`red-7.5.log` — seam 부재), 변이 R01(순차 되돌림) · R04(join 제거) CAUGHT({LEDGER})", "yes"),
         "B4": ("레인 입력 찾기", f"{FT} `TestEightLanesShareOneAuthorityWaveAndEachProposalReachesOneLane`", *SAME[1:]),
         "B5": ("자기 제안 소유 — 제안 하나 → 레인 하나", f"{FT} `TestEightLanesShareOneAuthorityWaveAndEachProposalReachesOneLane`",
                f"no — 분기 불변, 변이 R08(모든 레인에 줌) CAUGHT({LEDGER})", "yes"),
         "B6": ("**(새)** 레인 goroutine 들의 panic 값 순회", f"{LT} `TestAPanicOutsideALaneStepStillReachesTheMarketCycle`",
                f"yes — 편집 전 컴파일 실패, 변이 R02 · R03 CAUGHT({LEDGER})", "yes"),
         "B7": ("**(새)** panic 이 있으면 시장 주기 goroutine 에서 다시 던짐(순차 때와 같은 회복 경로)",
                f"{LT} `TestAPanicOutsideALaneStepStillReachesTheMarketCycle`", "yes — 변이 R02(삼킴) CAUGHT", "yes"),
         "B8": ("잠금 기록 실패 → 오류", "`TestALaneLatchThatCannotBeRecordedIsCountedNotEscalated`", *SAME[1:]),
     },
     "invariants": ["join 은 시장 주기 안이다(분리된 레인 goroutine 금지 — Manager 조건 ①). 각 레인은 RunBounded 마감 시한으로 끊기므로 join ≤ 마감 시한 1 회.",
                    "버려진 사이클의 step goroutine 은 step 이 돌아올 때까지 산다(strategyworker.invokeBounded — 결과 채널 버퍼 1 이라 막히지 않고 끝남). "
                    "생산 step 은 순수 메모리 평가라 곧 끝나고, 버림은 비정상이라 레인이 즉시 잠기므로 레인당 최대 하나(조건 ② — review 「7.5」).",
                    "레인 goroutine 의 panic 은 삼키지 않는다 — join 뒤 다시 던져 invokeStrategyCycle 이 받는다(다른 goroutine 의 panic 은 그 경로가 못 잡는다)."],
     "state": "레인 상태는 각 레인의 Offer · RunBounded 만 바꾼다(레인당 goroutine 하나). 관측 맵 · 물결 번호는 record 가 쓰기 잠금으로.",
     "safety": SAFE},
    {"bundle": "internal-app-engine--strategylaneruntime.runlane", "file": E + "strategy_lane_runtime.go", "func": "strategyLaneRuntime.runLane",
     "title": "strategyLaneRuntime.runLane",
     "revision": "분기 불변(2). `RunBounded` 에 넘기는 값이 `strategyFamilyLaneStep(lane, promotion)` → `runtime.laneStepFor(lane, promotion)` 한 곳 — "
                 "생산 정의(`strategy_lane_step.go`, `!tossos_testseams`)는 strategyFamilyLaneStep 한 줄이고 seam 은 태그 빌드(`strategy_lane_step_testseam.go`)에만 있다. "
                 "첫 구현은 무태그 함수 필드였고 5.1.2.1 핀이 잡았다(「핀이 잡은 자기 이탈」 — review).",
     "scen": {
         "B1": ("투입 거절(DISABLED · FULL) → 건강만 싣고 반환", f"{PT} `TestEightLatchedLanesAreProjectedInProductionOrderWithTheirFirstFailure`", *SAME[1:]),
         "B2": ("유계 사이클 오류 → 실패 문장", "`a112_lane_latency_testseam_test.go` `TestAHungLaneDoesNotDelayItsPeersInTheSameWave`(멈춘 레인의 마감 시한 오류)",
                "no — 분기 불변(편집 전 진입 0 → 이 로트의 시험이 처음 진입)", "yes"),
     },
     "invariants": ["레인 안에서 도는 일의 자리 · 생산 정의 본문 · seam 빌드 태그 · 정의 수는 `TestOnlyThePackageLevelStepEverRunsInsideALane` 이 못 박는다 — "
                    f"변이 R12(생산 정의가 훅을 봄) · R13(seam 태그 확장) CAUGHT({LEDGER})."],
     "state": "레인 상태는 Offer · RunBounded 만 바꾼다(편집 전과 같음).", "safety": SAFE},
    {"bundle": "internal-app-engine--strategylaneprojection", "file": E + "strategy_lane_projection.go", "func": "strategyLaneProjection",
     "title": "strategyLaneProjection",
     "revision": "분기 불변(5). 레인 상태를 접근자별 잠금 아홉 번 대신 `lane.Status()` 한 번(한 잠금)으로 읽는다 — 찢긴 행 제거(D2).",
     "scen": {
         "B1": ("미관측 레인 → 관측 사실 없이 반환", f"{PT} `TestAProcessWithoutLanesProjectsTheEightUnobservedDefaults` 계열", *SAME[1:]),
         "B2": ("투입이 들어간 물결 → 시작", f"{PT} `TestTheCycleGenerationIsTheMarketWaveInWhichTheLaneWasLastObserved`", *SAME[1:]),
         "B3": ("연 사이클 → 결과 · 비정상", f"{PT} `TestTheCycleGenerationIsTheMarketWaveInWhichTheLaneWasLastObserved`", *SAME[1:]),
         "B4": ("결과 있음", f"{PT} `TestTheCycleGenerationIsTheMarketWaveInWhichTheLaneWasLastObserved`", *SAME[1:]),
         "B5": ("REFUSED 결과만 거절 코드", f"{PT} `TestTheLaneDesiredAndEffectiveAreTheActivationTheWaveRanWith`", *SAME[1:]),
     },
     "invariants": ["한 행은 한 잠금으로 읽힌다 — AST 핀 `TestTheLaneProjectionReadsALaneRowUnderOneLock`(Status 1 회 · 상태 접근자 0, 변이 R05 CAUGHT); "
                    "행 불변식 동시성 시험 `TestAProjectedLaneRowIsNeverTorn`(-race, 비결정이라 결정적 핀은 AST — Manager 승인)."],
     "state": "상태 변경 없음 — 읽기 전용.", "safety": ["High-risk 아님 — 읽기 전용 투영."]},
    {"bundle": "internal-strategyworker--lane.health", "file": "internal/strategyworker/lane.go", "func": "Lane.Health", "title": "Lane.Health",
     "revision": "편집 전 2 분기 → 0: 판정 본문을 `healthLocked` 로 옮기고 Health 는 잠금 + 위임. Status 가 같은 잠금 안에서 같은 판정을 쓴다(판정 한 벌).",
     "scen": {},
     "happy": ("잠금을 잡고 healthLocked 의 판정을 돌려줌(LATCHED · DEGRADED · HEALTHY — 판정은 healthLocked 로 이동, 값 불변)",
               "`TestTheLaneStatusIsTheRowItsAccessorsRead` · `TestAnOrdinaryFailureCountsAndLatchesExactlyAtTheThreshold` · `TestEveryProductionLaneIsBornHealthyAndUnlatched`"),
     "invariants": ["Health 와 Status 의 건강 판정은 같은 함수다 — 변이 R06(Status 판정 분리) CAUGHT."],
     "state": "상태 변경 없음 — 읽기.", "safety": ["High-risk 아님 — 레인 건강 읽기(값 불변)."]},
]


def _pin_bundle() -> dict:
    """5.1.2.1 핀(레인 안에서 도는 일)의 경량 번들 — 분기는 이 시험 자신의 판정 갈래라 AST id 를 그대로 쓴다."""
    file, func = "internal/app/engine/a112_lane_runtime_test.go", "TestOnlyThePackageLevelStepEverRunsInsideALane"
    ast, _ = base.ast_for(file, func)
    ids = [b["id"] for b in ast.get("branches") or []]
    return {"bundle": "internal-app-engine--testonlythepackagelevelstepeverrunsinsidealane", "file": file, "func": func, "title": func + " (시험)",
            "revision": "자리 철자를 `runtime.laneStepFor(lane, promotion)` 로 바꾸고, laneStepFor 정의 전부를 세는 대조(정의 둘 · 생산 본문 한 줄 · 태그)를 "
                        "`assertLaneStepForIsTheProductionStepOutsideTestSeams` 로 더했다 — 핀 강화(Manager 판정 (A)).",
            "scen": {i: ("RunBounded 자리 세기 · 인자 수 가드 · 자리 목록 대조", "이 시험 자신", "no — 시험 코드",
                         "yes — 무태그 seam 첫 구현에서 실패 관측(`green-7.5-first.log`)") for i in ids},
            "invariants": ["자리만 세면 메서드 본문이 바뀌어도 목록은 그대로다 — 정의 본문 · 빌드 태그까지 대조한다(변이 R12 · R13 CAUGHT)."],
            "state": "시험 코드.", "safety": ["시험 코드 — 생산 경로 없음."]}


if __name__ == "__main__":
    base.main(FUNCS + [_pin_bundle()], tag="a112 7.5", pre="lot-7.5", ledger=LEDGER)
