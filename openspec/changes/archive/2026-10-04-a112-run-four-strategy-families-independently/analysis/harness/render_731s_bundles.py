#!/usr/bin/env python3
"""a112 7.3.1 SHADOW 로트 — 편집 뒤 번들(FLM/BTM)을 ast.json 과 **측정된** 분기 진입에서 채운다.

좌표 · 호출 · return 은 AST(현재 작업 트리)에서만 읽는다. 분기 행의 「시험」 칸은 손으로 고르지 않고 측정에서 읽는다:
`measurements/lot-7.3.1-shadow/branch-coverage-engine.json`(branch_coverage.py — shadow 시험 하나씩 + 패키지 합집합, 태그 빌드)과
`branch-coverage-projection.json`. 저자가 쓰는 것은 분기의 뜻과 편집 설명뿐이다.

시험 함수(서명 변경에 맞춘 호출 수정)는 `TEST_FUNCS` — 판정 불변, 행은 시험 자신의 갈래다.
편집 전 번들: `measurements/lot-7.3.1-shadow/pre-edit/`(HEAD 9e5f3ccf). 변이 원장: `measurements/lot-7.3.1-shadow/mutation-7.3.1-S.tsv`.

사용: python3 render_731s_bundles.py   (저장소 루트에서)
"""
from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import render_5222_bundles as base  # noqa: E402

A = Path(__file__).resolve().parents[1]
M = A / "measurements" / "lot-7.3.1-shadow"
LEDGER = "`analysis/measurements/lot-7.3.1-shadow/mutation-7.3.1-S.tsv`"
RED = "`analysis/measurements/lot-7.3.1-shadow/red-7.3.1-shadow.log`"
E = "internal/app/engine/"
P = "internal/strategyprojection/"
UNCHANGED = "no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건)"
NEW = f"yes — {RED}(편집 전 기호 없음 · 컴파일 RED)"


def measured(cov: dict, bundle: str, branch: str) -> tuple[str, str]:
    row = cov["bundles"].get(bundle, {}).get(branch)
    if row is None:
        return "측정 없음", "unmeasured"
    tests = row["entered_by"]
    union = row["union"]
    shown = ", ".join(f"`{name}`" for name in tests[:6]) + (f" 외 {len(tests) - 6}" if len(tests) > 6 else "")
    if tests:
        return shown, f"yes — shadow 시험 {len(tests)}개가 진입(측정)" + ("" if union is None else f"; 패키지 합집합 진입={union}")
    if union:
        return "shadow 시험 밖의 패키지 시험(합집합 측정)", "yes — 패키지 합집합 진입(측정)"
    return "진입 없음", "no — 측정상 진입 0(커버리지 공백, 통과가 아님)"


def prod(bundle: str, file: str, func: str, revision: str, meaning: dict[str, str], new: set[str], invariants: list[str],
         safety: list[str], state: str, cov_key: str = "engine") -> dict:
    return {"bundle": bundle, "file": file, "func": func, "title": func.rsplit(".", 1)[-1] if "." in func else func,
            "revision": revision, "meaning": meaning, "new": new, "invariants": invariants, "safety": safety, "state": state, "cov": cov_key}


SAFE_CARRY = ["High-risk(조정 · 제안 권한 · 조립) 경로의 편집은 운반뿐이다 — 조정 · admit · Submit · Arbitrate · dispatch 의 입력 · 순서 · 반환은 편집 전과 "
              "같고(차등 dispatch 시험 · 변이 S01~S08), shadow 값은 authority 구조체에 들어가지 않는다(census ② `TestOnlyTheAllowedFunctionsEverTouchAShadowType`)."]
PROD = [
    prod("internal-app-engine--coordinatemarketproposals", E + "strategy_market_coordinator.go", "coordinateMarketProposals",
         "분기 불변(8). 반환에 shadow 묶음을 더했다: 시작에 `shadow := strategyShadowBatch{observed: true}`, 안쪽 루프 `gate.admit` 앞에 수집 helper 문장 하나, "
         "계보 충돌 return 은 부재 값, 정상 return 만 수집 묶음(브리프 v3.3 §4 ⓐⓑⓒ).",
         {"B1": "경로(종목) 순회", "B2": "레인 없는 종목 → 거절 계수", "B3": "그 종목의 레인 제안 순회 — **편집: admit 앞에서 shadow.collect**",
          "B4": "관문이 멈춤 → gated 기록", "B5": "관문 EMITTED → 레인 봉투", "B6": "조정자 Submit 거절", "B7": "계보 신원 충돌 → **부재 값**으로 반환(편집)",
          "B8": "범위 전부 관문에 빼앗김 → erasedScopes"}, set(), ["조정 경로(admit · Submit · 계보 색인 · Arbitrate)의 식에 shadow 사용 0 — AST 핀."],
         SAFE_CARRY, "조정자 · arbitration · 지역 shadow 묶음만 쓴다(원장 · 브로커 0)."),
    prod("internal-app-engine--strategyproposalauthorityloader.collectmarket", E + "strategy_proposal_authority.go",
         "strategyProposalAuthorityLoader.collectMarket",
         "분기 불변(15). out 인자 `shadow` 를 더했다: 첫 문장에서 부재 값으로 비우고(조정 앞 닫힘 일곱 = 관측 없음), 조정 바로 뒤 대입 하나로 수집 묶음 + "
         "결속 설정(`loader.shadowConfig`)을 싣는다. 반환 갈래 열다섯 불변.",
         {"B1": "경로 · 스케줄 미준비 → ROUTE_NOT_READY", "B2": "FX 미준비", "B3": "적재기 설정 결함", "B4": "제안 공개 열쇠 무효", "B5": "US env 이름",
          "B6": "경로 항목 순회", "B7": "종목 중복", "B8": "제안 적재 실패 · digest 불일치", "B9": "받아들인 범위가 제안을 잃음",
          "B10": "관문이 범위를 지움 → FAMILY_GATE_CLOSED(묶음 실음)", "B11": "계보 충돌(조정자의 부재 값 그대로)", "B12": "큐 넘침", "B13": "중재 거절",
          "B14": "미해결 선택", "B15": "받아들인 범위 0"}, set(), ["shadow 대입은 둘뿐(첫 문장 부재 값 · 조정 뒤) — AST 핀."], SAFE_CARRY,
         "호출자의 shadow 칸 두 번 대입(그 밖 쓰기 0)."),
    prod("internal-app-engine--strategyproposalauthorityloader.collect", E + "strategy_proposal_authority.go", "strategyProposalAuthorityLoader.collect",
         "분기 불변(6). 두 번째 반환값 strategyShadowPair 를 더했다 — 시장 goroutine 이 자기 묶음을 싣고, recover 갈래는 묶음을 부재 값으로 되돌린다.",
         {"B1": "입력 결함 → 실패 짝 + 부재 shadow 짝", "B2": "시장 순회(KR · US goroutine)", "B3": "recover — **편집: shadow 를 부재 값으로**",
          "B4": "결과 두 개 수신", "B5": "KR 결과", "B6": "US 결과"}, set(), ["recover 갈래의 부재 값 대입은 AST 핀(`TestTheRemainingCarryAndPublishSitesKeepTheirShape`)."],
         SAFE_CARRY, "짝 두 개를 채운다(원장 0)."),
    prod("internal-app-engine--context.newpairedstrategyentryproductionassembly", E + "strategy_entry_supervisor.go",
         "Context.NewPairedStrategyEntryProductionAssembly",
         "분기 불변(8). collect 의 둘째 값을 `shadowAuthority` 로 받아 조립 리터럴의 별개 필드 `shadow` 에 담는다 — dispatch · worker · 결과 권한에는 넘기지 않는다.",
         {"B1": "Context nil", "B2": "원장 경로", "B3": "evidence 경로", "B4": "레인 런타임 오류", "B5": "시계 → dispatch now", "B6": "시장 worker 순회",
          "B7": "supervisor 생성 오류", "B8": "투영 발행 오류"}, set(), ["shadow 필드 대입 자리는 AST 핀."], SAFE_CARRY, "조립 값 · 발행(편집 전과 같음)."),
    prod("internal-app-engine--context.runproductionstrategymarketcycle", E + "strategy_entry_supervisor.go", "Context.runProductionStrategyMarketCycle",
         "분기 불변(4). evaluate 에 인자 index 5 `fresh.shadow.forMarket(market)` 를 더했다(Manager 판정 (A)). 마지막 문장 dispatch · Args[2] 불변.",
         {"B1": "refresh 오류", "B2": "레인 런타임 오류", "B3": "evaluate 오류(durable latch)", "B4": "dispatch 없는 조립 → nil"}, set(),
         ["본문의 shadow 타입 식은 Args[5] 하나 · 그 유일한 호출은 forMarket — 타입 규칙 핀 `TestTheMarketCycleCarriesShadowOnlyAsTheLastEvaluateArgument`."],
         ["High-risk(시장 주기) — dispatch 문장 · 원천 · 순서 무변경(:169-175 핀 · 차등 dispatch 시험). dispatch 앞 shadow 일은 값 운반 하나."],
         "레인 관측 · dispatch(편집 전과 같음)."),
    prod("internal-app-engine--context.newrefreshingpairedstrategyentrysupervisor", E + "strategy_entry_supervisor.go",
         "Context.NewRefreshingPairedStrategyEntrySupervisor",
         "분기 불변(4). worker 의 Cycle 클로저를 `c.productionStrategyCycle(clk, market)`(shadow 래퍼 — 성공 플래그 · 경과 판정 · 실패 폐기 defer · nil 뒤 비동기 시작)로 바꿨다.",
         {"B1": "Context · 시계 nil", "B2": "진입 관문 없음", "B3": "시장 순회", "B4": "supervisor 생성 오류"}, set(),
         ["두 생산 자리가 같은 래퍼를 쓴다 — AST 핀 `TestTheCycleClosureStartsTheShadowOnlyAfterANilCycleAndDiscardsOtherwise`."],
         ["High-risk(supervisor 배선) — 주기 오류 · panic 전파 · 회복 · latch · 삼킴 경로 불변(핀 (iv) · (v) · central-integrity 신원)."], "supervisor 등록."),
    prod("internal-app-engine--context.productionstrategyworker", E + "strategy_entry_supervisor.go", "Context.productionStrategyWorker",
         "분기 불변(1). 마지막 인자(cycle)를 `c.productionStrategyCycle(clk, market)` 로 바꿨다.", {"B1": "Context nil → 빈 worker"}, set(),
         ["래퍼 하나 — 위와 같은 핀."], ["High-risk(worker 배선) — wiringReady 판정 · 인자 불변."], "없음(값 구성)."),
    prod("internal-app-engine--newstrategylaneruntime", E + "strategy_lane_runtime.go", "newStrategyLaneRuntime",
         "분기 불변(1). shadow 맵 다섯(cells · epochs · observed · inFlight · skipped)을 생성자에서 만든다 — nil 맵 대입 panic 이 주기 경로의 오류를 덮지 않게(v3.3 N5).",
         {"B1": "시계 nil → nil 런타임"}, set(), ["맵 초기화만 더함."], ["레인 목록 · 원장 · 계좌 무변경."], "새 런타임 값."),
    prod("internal-app-engine--strategylaneruntime.evaluate", E + "strategy_lane_runtime.go", "strategyLaneRuntime.evaluate",
         "분기 불변(8). 6번째 인자 shadow 를 받아 record 로 넘기기만 한다(④ — 호출 · 순회 0).",
         {"B1": "런타임 nil", "B2": "복구 오류", "B3": "레인 순회(동시)", "B4": "입력 찾기", "B5": "레인 소유", "B6": "panic 수거 순회", "B7": "panic 재던짐",
          "B8": "latch 저장 오류"}, set(), ["묶음 사용은 record 인자 한 자리 — AST 핀."], ["레인 실행 · 관측 · latch 저장 불변."], "record 위임."),
    prod("internal-app-engine--strategylaneruntime.record", E + "strategy_lane_runtime.go", "strategyLaneRuntime.record",
         "분기 불변(4). 파도 증가 바로 뒤 같은 잠금에서 `shadowCells[market] = {wave, batch, activation}` 을 덮어쓴다(⑤).",
         {"B1": "관측 없음 → 물결 아님(칸도 그대로)", "B2": "파도 맵 지연 생성", "B3": "파도 증가(포화)", "B4": "관측 덮어쓰기"}, set(),
         ["칸 대입은 Lock 뒤 · 파도 증가 뒤 · defer Unlock — AST 핀."], ["레인 관측 기록(편집 전과 같음) + 값 보관."], "observed · waves · shadowCells."),
    prod("internal-app-engine--strategylaneruntime.projection", E + "strategy_lane_projection.go", "strategyLaneRuntime.projection",
         "편집 전 2 분기 → 3: 런타임 시계로 `shadowObservationUsable`(파도 등식 ∧ 미만료 ∧ 나이 상한)을 판정해 쓸 수 있는 shadow 관측을 넘긴다(B3 새).",
         {"B1": "런타임 nil", "B2": "레인 순회", "B3": "**(새)** 쓸 수 있는 shadow 관측"}, {"B3"},
         ["판정 함수 하나 — 경계 핀 (e)(f) · 나이 등식 · 의도된 간극."], ["읽기 전용 — 레인 · 관측을 바꾸지 않는다."], "없음(읽기)."),
    prod("internal-app-engine--strategylaneprojection", E + "strategy_lane_projection.go", "strategyLaneProjection",
         "편집 전 5 분기 → 6: 쓸 수 있는 shadow 관측 ∧ 관측 값 OFF/OFF 이면 runtime=SHADOW · shadowOutcome(B6 새).",
         {"B1": "미관측 레인", "B2": "투입이 들어간 물결", "B3": "연 사이클", "B4": "결과 있음", "B5": "REFUSED 결과만 거절 코드",
          "B6": "**(새)** SHADOW — 관측된 OFF/OFF 레인만"}, {"B6"}, ["ON 관측은 SHADOW 가 아니다 — `TestNeitherTheStepNorTheProjectionShadowsAnOnLane`(변이 S30)."],
         ["읽기 전용 — SHADOW 는 승격하지 않는다(desired/effective 는 관측 값 그대로)."], "없음(값 구성)."),
    prod("internal-app-engine--context.read", E + "strategy_runtime_projection.go", "Context.Read",
         "편집 전 8 분기 → 9: supervisor 가 그 시장 평가를 abandon 으로 기록했으면 그 시장 레인의 SHADOW 를 지운다(R1 두 시계 — B7 새). 이후 갈래 번호가 하나씩 밀렸다.",
         {"B1": "Context · ctx nil", "B2": "투영 저장소 없음", "B3": "저장소 읽기 오류", "B4": "레인 런타임 있음 → 레인 덧씌움", "B5": "supervisor 없음",
          "B6": "시장 순회", "B7": "**(새)** abandon 기록 시장 → SHADOW 지움", "B8": "worker 없음 · 잠기지 않음(편집 전 B7)", "B9": "현재 시장 레코드 → 실패 표시(편집 전 B8)"},
         {"B7"}, ["R1 시험 `TestAMarketWhoseEvaluationWasAbandonedShowsNoShadow`(변이 S31)."], ["읽기 전용 투영."], "없음(읽기)."),
]
PROJ = [
    prod("internal-strategyprojection--validatelane", P + "lanes.go", "validateLane",
         "편집 전 15 분기 → 17: runtime 어휘 {UNOBSERVED, SHADOW}(B1 조건), SHADOW 교차 규칙(B2 새) · shadowOutcome 짝(B3 새). 이후 번호 둘씩 밀림.",
         {"B1": "desired/effective/runtime 어휘", "B2": "**(새)** SHADOW ⇒ OFF/OFF ∧ health ∧ cycleGeneration>0", "B3": "**(새)** shadowOutcome ⇔ SHADOW",
          "B4": "거절 코드 짝", "B5": "미관측", "B6": "미관측 사실 추론", "B7": "건강 어휘", "B8": "정책", "B9": "정규 값", "B10": "물결 ⇔ 트리거",
          "B11": "트리거 없음", "B12": "트리거 없는 사실", "B13": "트리거 어휘", "B14": "시작 ⇔ ENQUEUED", "B15": "시작 어휘", "B16": "결과 ⇔ ADMITTED", "B17": "결과 어휘"},
         {"B2", "B3"}, ["거부 표 시험 `TestValidateRefusesShadowOutsideTheOneCrossRule`(변이 V01 · V02)."], ["외부 경계 검증기 — 기존 거절 규칙 불변."], "없음.", "projection"),
    prod("internal-strategyprojection--lanejsonfields", P + "lanes.go", "LaneJSONFields", "분기 없음. 계약 JSON 이름 목록에 `shadowOutcome` 을 더했다.",
         {}, set(), ["이름 목록은 실제 직렬화(`TestLaneAndCoordinatorJSONNamesAreTheContract`)와 OpenAPI(`TestOpenAPIDocumentsTheLaneAndCoordinatorChildrenByTheirExactNames`)가 잰다."],
         ["읽기 전용 목록."], "없음.", "projection"),
    prod("internal-strategyprojection--selectedscopejsonfields", P + "lanes.go", "SelectedScopeJSONFields",
         "본문 불변 — 바로 뒤에 LaneRuntimes · LaneShadowOutcomes 를 더한 diff 조각이 이 함수의 끝 줄과 맞닿아 수정 함수로 잡혔다.", {}, set(),
         ["목록 불변 — `TestLaneAndCoordinatorJSONNamesAreTheContract`."], ["읽기 전용 목록."], "없음.", "projection"),
    prod("internal-strategyprojection--clonelanes", P + "lanes.go", "cloneLanes", "분기 불변(2). shadowOutcome 포인터도 깊은 복사한다.",
         {"B1": "nil", "B2": "레인 순회"}, set(), ["복사 시험 `TestTheShadowVocabularyAndTheWireShape`."], ["읽기 전용 복사."], "새 슬라이스.", "projection"),
]


def render_prod(spec: dict, covs: dict) -> None:
    cov = covs[spec["cov"]]
    ast, raw = base.ast_for(spec["file"], spec["func"])
    ids = [b["id"] for b in ast.get("branches") or []]
    if set(ids) != set(spec["meaning"]):
        raise SystemExit(f"{spec['bundle']}: AST branches {ids} != authored {sorted(spec['meaning'])}")
    scen = {}
    for branch in ids:
        tests, green = measured(cov, spec["bundle"], branch)
        scen[branch] = (spec["meaning"][branch], tests, NEW if branch in spec["new"] else UNCHANGED, green)
    rendered = {"bundle": spec["bundle"], "file": spec["file"], "func": spec["func"], "title": spec["title"], "revision": spec["revision"],
                "scen": scen, "invariants": spec["invariants"], "safety": spec["safety"], "state": spec["state"]}
    if not ids:
        rendered["happy"] = ("분기 없음 — " + spec["revision"], "; ".join(spec["invariants"]))
    (base.FL / spec["bundle"]).mkdir(parents=True, exist_ok=True)
    base.main([rendered], tag="a112 7.3.1 SHADOW", pre="lot-7.3.1-shadow", ledger=LEDGER)


def render_tests(names: list[tuple[str, str]]) -> None:
    for file, func in names:
        bundle = base.bundle_name(file, func) if hasattr(base, "bundle_name") else (
            file.rsplit("/", 1)[0].replace("/", "-") + "--" + func.lower())
        ast, _ = base.ast_for(file, func)
        ids = [b["id"] for b in ast.get("branches") or []]
        why = ("a112 7.3.1 서명 변경(evaluate 6번째 인자 · collect 두 반환값 · collectMarket out 인자)에 맞춘 호출 수정 — 이 시험의 판정 · 단언은 바뀌지 않았다."
               if "shadow" not in file else "a112 7.3.1 SHADOW 시험(새 판정 반전 · 신규 단언).")
        rendered = {"bundle": bundle, "file": file, "func": func, "title": func + " (시험)", "revision": why,
                    "scen": {i: ("시험 자신의 갈래", "이 시험 자신", "no — 시험 코드", "yes — 엔진 태그 스위트 PASS") for i in ids},
                    "happy": ("분기 없음 — 시험 본문", "이 시험 자신(엔진 태그 스위트 PASS)"),
                    "invariants": ["시험 코드 — 생산 판정 없음."], "safety": ["시험 코드 — 실주문 · 원장 쓰기 없음(fixture 원장 · Gateway 스파이)."],
                    "state": "시험 fixture."}
        (base.FL / bundle).mkdir(parents=True, exist_ok=True)
        base.main([rendered], tag="a112 7.3.1 SHADOW", pre="lot-7.3.1-shadow", ledger=LEDGER)


if __name__ == "__main__":
    covs = {"engine": json.loads((M / "branch-coverage-engine.json").read_text()),
            "projection": json.loads((M / "branch-coverage-projection.json").read_text())}
    for spec in PROD + PROJ:
        render_prod(spec, covs)
    tests = [tuple(line.split("\t")) for line in (M / "modified-test-functions.tsv").read_text().split("\n") if line.strip()]
    render_tests(tests)
