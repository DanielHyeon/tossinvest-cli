#!/usr/bin/env python3
"""a112 7.3 — 편집 뒤 번들(FLM/BTM)을 ast.json 에서 채운다(render_5222_bundles.main 재사용 — 좌표 · 호출 · return 은 AST 만, 저자는 뜻 · 시험 이름).

편집 전 번들: `analysis/measurements/lot-7.3/pre-edit/`(HEAD f0f7d668). 변이 원장: `analysis/measurements/lot-7.3/mutation-7.3.tsv`.
사용: python3 render_73_bundles.py   (저장소 루트에서)
"""
from __future__ import annotations

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import render_5222_bundles as base  # noqa: E402

LEDGER = "`analysis/measurements/lot-7.3/mutation-7.3.tsv`"
PT = "`a112_lane_coordinator_projection_test.go`"
CT = "`a112_lane_coordinator_children_test.go`"
SAME = ("분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로", "no — 이 로트가 바꾸지 않음", "yes")
SAFE = ["High-risk 아님 — 읽기 전용 투영(주문 · 원장 · 활성화 · 토글 쓰기 없음). 기존 시장 레코드의 판정은 불변이고 additive 자식만 더했다."]
E, S = "internal/app/engine/", "internal/strategyprojection/"

FUNCS = [
    {"bundle": "internal-app-engine--context.read", "file": E + "strategy_runtime_projection.go", "func": "Context.Read", "title": "Context.Read",
     "revision": "편집 전 7 분기 → 8: runtime identity 뒤 · 감독자 부재 검사 앞에 B4(레인 런타임이 있으면 여덟 레인을 지금 상태로 덧씌움)를 더했다. 나머지 일곱은 번호만 밀렸다(편집 전 B4~B7 → B5~B8).",
     "scen": {
         "B1": ("nil receiver/context → 오류", "`TestStrategyRuntimeReadWithoutAStoreStaysAnError`", *SAME[1:]),
         "B2": ("store 부재 → 오류", "`TestStrategyRuntimeReadWithoutAStoreStaysAnError`", *SAME[1:]),
         "B3": ("store 읽기 오류 → 그대로", "`TestStrategyRuntimeReadOnAFailedStoreInventsNothing`", *SAME[1:]),
         "B4": ("**(새)** 레인 런타임 있음 → `lanes = runtime.projection()`(읽기만), 없으면 저장소의 미관측 기본값",
                f"{PT} `TestEightLatchedLanesAreProjectedInProductionOrderWithTheirFirstFailure` · `TestTheCycleGenerationIsTheMarketWaveInWhichTheLaneWasLastObserved`(있음) · `TestAProcessWithoutLanesProjectsTheEightUnobservedDefaults`(없음)",
                f"yes — 편집 전 컴파일 실패(`red-7.3.log`), 변이 P03 CAUGHT({LEDGER})", "yes"),
         "B5": ("감독자 부재 → 반환(편집 전 B4)", "`TestStrategyRuntimeReadExposesThisProcessConfigAndBuildDigest`", *SAME[1:]),
         "B6": ("KR · US latch 검사(편집 전 B5)", "`TestStrategyRuntimeReadKeepsTheIdentityWhileLatchingAMarket`", *SAME[1:]),
         "B7": ("latch 안 된 시장 건너뛰기(편집 전 B6)", "`TestStrategyRuntimeReadKeepsTheIdentityWhileLatchingAMarket`", *SAME[1:]),
         "B8": ("CURRENT 시장만 덮기(편집 전 B7)", "`TestStrategyRuntimeReadKeepsTheIdentityWhileLatchingAMarket`", *SAME[1:]),
     },
     "invariants": ["레인 덧씌움은 `strategyLanesMu` 아래에서 포인터만 읽고 잠금 밖에서 projection 을 부른다(레인 접근자 · 런타임 읽기 잠금만).",
                    "읽기 전용 불변: Read 를 18 번 불러도 레인 상태 · 관측 기록 · 원장 잠금이 같다(`TestReadingTheLaneProjectionNeverChangesALane`, 변이 P01 · P02 CAUGHT)."],
     "state": "상태 변경 없음 — 저장소 사본 위에 identity · 레인 · latch overlay 를 얹어 돌려준다.",
     "safety": SAFE},
    {"bundle": "internal-app-engine--strategyprojectionfromassembly", "file": E + "strategy_runtime_projection.go", "func": "strategyProjectionFromAssembly",
     "title": "strategyProjectionFromAssembly",
     "revision": "분기 구조 불변(10). 순회가 index 를 받고, 순회 머리에 조정자 자식 대입 한 줄(`snapshot.Coordinators[index] = strategyCoordinatorProjection(...)`)을 더했다 — 시장 레코드 갈래보다 앞이라 실패 갈래(B2)에서도 조정자가 보인다.",
     "scen": {
         "B1": ("KR · US 순회 — **(편집)** 머리에서 조정자 자식을 채움(R4: 승인 범위 전부)",
                f"{PT} `TestTheCoordinatorChildShowsEveryAdmittedOwnerScope`", f"yes — 편집 전 컴파일 실패, 변이 P07 · P09 CAUGHT({LEDGER})", "yes"),
         "B2": ("worker 미승격 → 시장 실패(조정자는 이미 채워짐)", f"{PT} `TestTheCoordinatorChildShowsEveryAdmittedOwnerScope`(두 worker 미승격 — 조정자가 보임)",
                f"yes — 변이 P09(조정자 대입 제거) CAUGHT", "yes"),
         "B3": ("실패 사유 고르기", *SAME), "B4": ("활성화 부재", *SAME), "B5": ("근거 stale", *SAME), "B6": ("보호 미배선", *SAME),
         "B7": ("주문 경로와 같은 handoff 목록 순회", "`a112_owner_scope_trading_test.go` `TestAnActivatedTwoScopeMarketPromotesItsWorkerPerScope`", *SAME[1:]),
         "B8": ("조정자 순서의 첫 승인 · 유효 범위(시장 레코드 — 불변)", "`a112_owner_scope_trading_test.go` `TestAnActivatedTwoScopeMarketPromotesItsWorkerPerScope`", *SAME[1:]),
         "B9": ("승인 범위 없음 → EvidenceStale", "진입 0 — 편집 전에도 진입 0(5.2.2.2 번들 기록 그대로)", "no — 이 로트가 바꾸지 않음", "n/a"),
         "B10": ("레인 근거 digest 부재 → 후보 근거", *SAME),
     },
     "invariants": ["시장 레코드는 편집 전과 같은 값이다(첫 범위만). 범위 전부는 조정자 자식 selected[] 가 싣는다 — 같은 handoff 목록 · 같은 술어."],
     "state": "상태 변경 없음 — 조립 값에서 스냅숏을 만든다.",
     "safety": SAFE},
    {"bundle": "internal-strategyprojection--validate", "file": S + "model.go", "func": "Validate", "title": "Validate",
     "revision": "편집 전 5 분기 → 7: 시장 검사 뒤에 B6(lanes[8] 검사) · B7(coordinators[2] 검사)을 더했다. 앞 다섯은 불변.",
     "scen": {
         "B1": ("envelope 3항", "`TestValidationRejectsMissingDuplicateScopeAndInventedReadiness`", *SAME[1:]),
         "B2": ("runtime identity", "`TestValidationRejectsPartialOrNoncanonicalRuntimeIdentity`", *SAME[1:]),
         "B3": ("KR · US", "`TestDormantSnapshotContainsExactPairedHonestMarkets`", *SAME[1:]),
         "B4": ("시장 부재/교차", "`TestValidationRejectsMissingDuplicateScopeAndInventedReadiness`", *SAME[1:]),
         "B5": ("시장 레코드 판정", "`TestValidationRejectsMissingDuplicateScopeAndInventedReadiness`", *SAME[1:]),
         "B6": ("**(새)** 레인 자식: 개수 8 · 고정 순서 · 열쇠 · enum · 거절 코드 ⇔ REFUSED(판정 (A)) · 미관측 무사실 · 관측 사슬(물결 ⇔ 트리거, 투입만 시작, 연 사이클만 결과)",
                f"{CT} `TestValidateRefusesMalformedLaneAndCoordinatorChildren`(레인 26 행 — 거절 코드 짝 4 행 포함)", f"yes — 편집 전 컴파일 실패, 변이 P10 · P11 · P13 · P25 CAUGHT({LEDGER})", "yes"),
         "B7": ("**(새)** 조정자 자식: 개수 2 · KR,US · 미관측 무사실 · 사유 · 중재 코드 · 정렬 유일 gated · null 아닌 목록",
                f"{CT} `TestValidateRefusesMalformedLaneAndCoordinatorChildren`(조정자 16 행 — 선택 범위 계보 2 행 포함)", "yes — 편집 전 컴파일 실패, 변이 P20 CAUGHT", "yes"),
     },
     "invariants": ["새 두 분기는 거절 집합을 늘린다. 기본(미관측) 스냅숏은 계속 유효하다(`TestDormantAndUnavailableSnapshotsCarryEightUnobservedLanesAndTwoCoordinators`).",
                    "새로 거절되는 정상 입력: 자식이 없는 옛 모양 envelope(7.3 전 엔진이 보낸 것) — 같은 이미지가 엔진 · 콘솔을 함께 교체하므로 혼합 창은 배포 절차 밖(review 잔여)."],
     "state": "상태 변경 없음.",
     "safety": SAFE},
    {"bundle": "internal-strategyprojection--clone", "file": S + "model.go", "func": "Clone", "title": "Clone",
     "revision": "분기 불변(1). 직선 코드: 자식 둘의 깊은 복사(`cloneLanes` · `cloneCoordinators`)를 구성자에 더했다.",
     "scen": {"B1": ("시장 두 개 복사(불변)", f"`TestCloneCarriesRuntimeIdentityWithoutSharingIt` · {CT} `TestCloneDeepCopiesLaneAndCoordinatorChildren`",
                     f"yes(자식 복사 — 직선) — 편집 전 컴파일 실패, 변이 P12(얕은 복사) CAUGHT({LEDGER})", "yes")},
     "invariants": ["사본의 포인터 · 목록을 바꿔도 원본이 바뀌지 않는다."],
     "state": "상태 변경 없음 — 새 값을 만든다.",
     "safety": SAFE},
    {"bundle": "internal-app-engine--strategylaneruntime.runlane", "file": E + "strategy_lane_runtime.go", "func": "strategyLaneRuntime.runLane",
     "title": "strategyLaneRuntime.runLane",
     "revision": "분기 불변(2). 관측 구성자가 desired/effective(활성화 위임) · 입력 digest 둘을 싣고, 열린 사이클의 거절 코드(`Cycle.Refusal`)를 기록하며(판정 (A)), `lane.Offer()` 를 구성자 밖 다음 줄로 옮겼다 — 레인 호출 순서 · 횟수 불변(Offer 한 번, 투입 시 RunBounded 한 번).",
     "scen": {
         "B1": ("투입 거절(DISABLED · FULL) → 건강만 싣고 반환", f"{PT} `TestEightLatchedLanesAreProjectedInProductionOrderWithTheirFirstFailure`(DISABLED)", *SAME[1:]),
         "B2": ("유계 사이클 오류 → 실패 문장", "진입 0 — 생산 Step 은 오류를 내지 않는다(편집 전 번들 기록 그대로)", "no — 이 로트가 바꾸지 않음", "n/a"),
     },
     "invariants": ["새 필드는 이미 받은 값(활성화 · 입력)의 읽기다 — 레인 상태를 바꾸는 호출을 더하지 않았다.",
                    "desired/effective 는 관측 시점 계산(판정 Q2 부가): `TestTheLaneDesiredAndEffectiveAreTheActivationTheWaveRanWith`, 변이 P06 CAUGHT. 거절 코드 기록은 같은 시험(켜진 KR 레인 REFUSED + ARBITRATION_SEAL_MISMATCH), 변이 P22 CAUGHT."],
     "state": "레인 상태는 Offer · RunBounded 만 바꾼다(편집 전과 같음).",
     "safety": SAFE},
    {"bundle": "internal-app-engine--strategylaneruntime.evaluate", "file": E + "strategy_lane_runtime.go", "func": "strategyLaneRuntime.evaluate",
     "title": "strategyLaneRuntime.evaluate",
     "revision": "분기 불변(6). `record(observations)` → `record(market, observations)` 한 줄(시장별 물결 번호 — 판정 Q1=(B)). 순서(복구 → 돌기 → 기록 → 잠금 남기기) 불변.",
     "scen": {
         "B1": ("nil 런타임", "진입 0(편집 전 번들 기록 그대로)", "no — 이 로트가 바꾸지 않음", "n/a"),
         "B2": ("복구 실패 → 오류", "진입 0(편집 전 번들 기록 그대로)", "no — 이 로트가 바꾸지 않음", "n/a"),
         "B3": ("이 시장 네 레인 순회", f"{PT} `TestTheCycleGenerationIsTheMarketWaveInWhichTheLaneWasLastObserved`", *SAME[1:]),
         "B4": ("레인 입력 찾기", "`TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`", *SAME[1:]),
         "B5": ("자기 제안 소유", "`TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt`", *SAME[1:]),
         "B6": ("잠금 기록 실패 → 오류", "`TestALaneLatchThatCannotBeRecordedIsCountedNotEscalated`", *SAME[1:]),
     },
     "invariants": ["물결 번호를 시장 인자로 record 에 넘긴다 — 공유 계수기 변이 P05 CAUGHT."],
     "state": "편집 전과 같음(레인 · 원장 잠금).",
     "safety": SAFE},
    {"bundle": "internal-app-engine--strategylaneruntime.record", "file": E + "strategy_lane_runtime.go", "func": "strategyLaneRuntime.record",
     "title": "strategyLaneRuntime.record",
     "revision": "편집 전 2 분기 → 4: 시장 인자를 받고, 물결 맵 지연 생성(B2) · 포화 상한 아래 증가(B3)를 더했다. 관측마다 물결 번호를 찍는다(B4 몸통).",
     "scen": {
         "B1": ("nil 런타임 또는 관측 0 → 물결이 아님(번호 안 올림)", "진입 0 — evaluate 는 레인 넷을 늘 넘긴다(시장마다 넷)", "no — 편집 전과 같은 조건", "n/a"),
         "B2": ("**(새)** 물결 맵 없음 → 생성(첫 물결)", f"{PT} `TestTheCycleGenerationIsTheMarketWaveInWhichTheLaneWasLastObserved`", "yes — 편집 전 컴파일 실패", "yes"),
         "B3": ("**(새)** 포화 상한 아래면 +1", f"{PT} `TestTheCycleGenerationIsTheMarketWaveInWhichTheLaneWasLastObserved`(KR 2 · US 1)",
                f"yes — 변이 P04(증가 제거) · P05(공유 계수기) CAUGHT({LEDGER})", "yes"),
         "B4": ("관측마다 물결 번호를 찍어 덮어쓰기", f"{PT} `TestTheCycleGenerationIsTheMarketWaveInWhichTheLaneWasLastObserved`", "yes — 편집 전 컴파일 실패", "yes"),
     },
     "invariants": ["증가와 찍기가 같은 쓰기 잠금 안이다 — 두 물결이 한 번호를 나눠 갖지 않는다.", "포화 상한(^uint64(0))에서 멈춘다 — 넘쳐 0 이 되면 「미관측」으로 읽힌다."],
     "state": "런타임의 관측 맵 · 물결 맵(프로세스 메모리)만 쓴다.",
     "safety": SAFE},
    {"bundle": "internal-strategyprojection--dormantsnapshot", "file": S + "model.go", "func": "DormantSnapshot", "title": "DormantSnapshot",
     "revision": "분기 없음. 구성자에 기본 자식 둘(`defaultLanes()` · `defaultCoordinators()`)을 더했다.", "scen": {},
     "happy": ("여덟 미관측 레인(골든 순서 · OFF/OFF/UNOBSERVED) · 두 미관측 조정자", f"{CT} `TestDormantAndUnavailableSnapshotsCarryEightUnobservedLanesAndTwoCoordinators` · `TestTheDefaultLaneTableIsTheFrozenGoldenDescriptorList`"),
     "invariants": ["기본 레인 표는 골든 서술자와 같고(시험이 골든 파일을 직접 읽음) 생산 레인 목록과 같다(strategyworker 시험)."],
     "state": "상태 변경 없음.", "safety": SAFE},
    {"bundle": "internal-strategyprojection--unavailablesnapshot", "file": S + "model.go", "func": "UnavailableSnapshot", "title": "UnavailableSnapshot",
     "revision": "분기 없음. DormantSnapshot 과 같은 기본 자식 둘을 더했다.", "scen": {},
     "happy": ("엔진에 닿지 못한 스냅숏도 여덟 · 둘을 미관측으로 싣는다(건강 추론 없음)", f"{CT} `TestDormantAndUnavailableSnapshotsCarryEightUnobservedLanesAndTwoCoordinators`"),
     "invariants": ["미관측 자식은 사실을 싣지 않는다 — Validate B6 · B7 이 강제."], "state": "상태 변경 없음.", "safety": SAFE},
    {"bundle": "internal-strategyprojection--currentpair", "file": S + "projection_test.go", "func": "currentPair", "title": "currentPair (시험 도우미)",
     "revision": "시험 도우미. 옛 모양 리터럴(자식 없음) 대신 DormantSnapshot 에서 시작해 시장 레코드만 바꾼다 — envelope 계약이 자식을 요구한다.",
     "scen": {"B1": ("KR · US 시장 레코드 만들기", "`TestMarketFailureReplacesOnlyExactMarketWithoutFallback` · `TestEitherMarketFailurePreservesTheExactPeer`", "no — 시험 코드", "yes"),
              "B2": ("만든 스냅숏이 Validate 를 통과", "같은 시험들(도우미가 t.Fatal)", "no — 시험 코드", "yes")},
     "invariants": ["시장 레코드 값은 편집 전과 같다."], "state": "시험 코드.", "safety": ["시험 코드 — 생산 경로 없음."]},
    {"bundle": "internal-strategyprojectionrpc--testclientignoresadditivefieldsfromanewerengine", "file": "internal/strategyprojectionrpc/a112_additive_fields_test.go",
     "func": "TestClientIgnoresAdditiveFieldsFromANewerEngine", "title": "TestClientIgnoresAdditiveFieldsFromANewerEngine (시험)",
     "revision": "7.3 이 `coordinators` · `lanes` 를 아는 필드로 만들어 심던 이름을 아직 모르는 이름으로 바꿨다(envelope · 시장 레코드 · 레인 자식 세 층). 재는 것 불변.",
     "scen": {f"B{i}": ("arrangement 오류 가드", "이 시험 자신", "no — 시험 코드", "yes") for i in (1, 2, 3, 4, 5, 6, 7)} | {
         "B8": ("토큰 불일치 → 401", "이 시험 자신(서버 stub)", "no — 시험 코드", "yes"),
         "B9": ("구 reader 가 additive 필드에 죽음 → 실패", "이 시험 자신", "yes — 7.3 GREEN 직후 옛 주입(`coordinators` 부분 객체)이 새 계약에 거절됨을 관측(`green-7.3-first.log`)", "yes"),
         "B10": ("기존 필드 의미 · 형식 유지", "이 시험 자신", "no — 시험 코드", "yes")},
     "invariants": ["주입은 세 층 모두 이 client 가 모르는 이름이다 — read 가 실패하면 시험 실패."], "state": "시험 코드.", "safety": ["시험 코드 — 생산 경로 없음."]},
]

def _census_bundle() -> dict:
    """census 시험(자리 목록 핀)의 경량 번들 — 분기는 전부 이 시험 자신의 판정 갈래라 AST 의 id 를 그대로 쓴다(손으로 고르지 않음)."""
    file, func = "internal/app/engine/strategy_dispatch_handoff_guard_test.go", "TestNoProductionSiteDiscardsTheSeamsAdmissionAnswer"
    ast, _ = base.ast_for(file, func)
    ids = [b["id"] for b in ast.get("branches") or []]
    return {"bundle": "internal-app-engine--testnoproductionsitediscardstheseamsadmissionanswer", "file": file, "func": func,
            "title": func + " (시험)",
            "revision": "자리 목록 핀(want)에 a112 7.3 의 새 읽기 자리 `strategy_lane_projection.go:strategyCoordinatorProjection` 을 이름 대어 더했다(셋 → 넷). 판정 갈래 불변.",
            "scen": {i: ("census 판정 갈래(답 미결속 · 버림 · `_ =` 침묵 · 자리 목록 대조)", "이 시험 자신", "no — 시험 코드",
                         "yes — 편집 전 7.3 GREEN 판에서 「sites=[… strategyCoordinatorProjection …]」 로 실패 관측(`green-7.3-first.log`)") for i in ids},
            "invariants": ["새 자리는 두 값을 다 받고 `!admitted` 로 거른다 — 핀은 숫자가 아니라 이름 목록이라 다음 자리도 이름을 대야 통과한다."],
            "state": "시험 코드.", "safety": ["시험 코드 — 생산 경로 없음."]}


if __name__ == "__main__":
    specs = FUNCS + [_census_bundle()]
    base.main(specs, tag="a112 7.3", pre="lot-7.3", ledger=LEDGER)
    # 게이트의 BTM 좌표 칸 문법(role_check.CELL_COORD `[A-Za-z][\w ]*\bat N:N`)은 하이픈을 받지 않는다 — AST 종류 `type-switch` 의
    # 칸만 `type switch at` 로 적는다(종류 자체는 FLM 표의 따로 된 칸과 ast.json 에 그대로).
    for spec in specs:
        btm = base.FL / spec["bundle"] / "branch-test-map.md"
        btm.write_text(btm.read_text().replace("| type-switch at ", "| type switch at "))
