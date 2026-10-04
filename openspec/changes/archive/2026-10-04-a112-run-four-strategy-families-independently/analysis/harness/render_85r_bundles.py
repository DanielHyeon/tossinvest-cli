#!/usr/bin/env python3
"""a112 8.5 응답 로트(Manager 최종 판정 2026-10-04) — 편집 뒤 번들 넷(render_5222_bundles.main 재사용, 8.8.4-B 서술 위에 이 로트의 변경만 덮음).

편집 전 번들: `analysis/measurements/lot-8.5-R/pre-edit/`(HEAD 178cc196). RED: `lot-8.5-R/red-8.5-R.log`. 변이: `lot-8.5-R/mutation-8.5-R.tsv`.
재번호: `lot-8.5-R/renumber.txt`(편집 전/뒤 ast 를 difflib 으로 정렬 — 손 재번호 아님). 분기가 바뀐 함수는 loadFamilyActivation 하나(B1 새로 앞에).
"""
from __future__ import annotations

import copy
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import render_5222_bundles as base  # noqa: E402
import render_884b_bundles as lotb  # noqa: E402

TAG, PRE = "a112 8.5-R", "lot-8.5-R"
LEDGER = "`analysis/measurements/lot-8.5-R/mutation-8.5-R.tsv`"
RED = "yes — `red-8.5-R.log`"
F = lotb.F
CENSUS = lotb.CENSUS
E = lotb.E
DECIDE = "`a112_gate_decision_recompute_test.go`"
CANCEL = DECIDE + " `TestACancelDuringTheProposalLoadClosesTheMarketAtTheDecision`"
REVOKE = DECIDE + " `TestARevocationDuringTheProposalLoadClosesTheMarketAtTheDecision`"
LEGACY = DECIDE + " `TestAnUndeclaredMarketCancelledDuringTheLoadKeepsTheLegacyPath`"
NILENV = DECIDE + " `TestANilEnvironmentReaderRollsTheGateBackInsteadOfPanicking`"


def spec(bundle: str) -> dict:
    for item in lotb.FUNCS:
        if item["bundle"] == bundle:
            return copy.deepcopy(item)
    raise SystemExit(bundle)


collect = spec("internal-app-engine--strategyproposalauthorityloader.collectmarket")
collect["revision"] = ("분기 불변(15). 관문을 **두 번** 계산한다(8.5 응답 로트 ① — Manager 판정 Y, 보이스 2 P1 · 보이스 1 P2-1): B1 뒤의 조기 계산(8.8.4 항목 2)은 "
                       "조정 앞 닫힘 여섯(B2 · B3 · B4 · B7 · B8 · B9)이 싣는 **진단** 값으로 남기고, 판정 계산을 편집 전(f473d815) 자리 — "
                       "`coordinateMarketProposals` 바로 앞 — 에 되살렸다. 8.8.4 의 「판정 자리 불변 · 옮긴 값은 조정 관문으로도 쓰인다」 서술은 거짓이었다"
                       "(관문 스냅숏이 제안 적재 시간만큼 낡아 적재 중 취소 · 철회가 판정에 안 보임). 이제 판정은 편집 전과 같은 함수 · 같은 순간이다.")
pre_arbitration = {"B2", "B3", "B4", "B7", "B8", "B9"}
for branch in pre_arbitration:
    scenario, test, red, green = collect["scen"][branch]
    collect["scen"][branch] = (scenario.replace("(관문 활성화)", "(조기 · 진단 활성화)"), test, red, green)
collect["scen"]["B2"] = (collect["scen"]["B2"][0], collect["scen"]["B2"][1] + " · " + REVOKE + "(a closure before arbitration carries the diagnostic value — 적재 1 회 · gen 7)",
                         collect["scen"]["B2"][2], "yes")
collect["scen"]["B10"] = ("관문이 범위를 통째로 지움 → FAMILY_GATE_CLOSED(**판정** 관문 — 조정 바로 앞에서 다시 계산; 판정 활성화를 실음)",
                          "`TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve`(판정 활성화 carry 단언 — 8.5 보이스 3 P2-1 · 변이 E9) · " + CANCEL + " · " + REVOKE
                          + "(revoked mid-load) · " + LEGACY + " · " + CENSUS,
                          f"{RED} — 편집 전(HEAD 178cc196) 적재 중 취소 · 철회가 READY · handoff 2 로 통과", "yes")
collect["scen"]["B11"] = ("계보 신원 충돌 → INTERNAL_FAILURE(판정 활성화)",
                          CENSUS + "(census 만 — 아래 한계 참조)", "no — 갈래 불변", "census — 도달 불가 공간")
collect["scen"]["B12"] = ("조정자 넘침 → QUEUE_OVERFLOW(판정 활성화)",
                          "`TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne`(미선언 실행 + 검증된 관문 아래 실행의 판정 활성화 carry 단언 — 8.5 보이스 3 P2-1 · 변이 E7) · " + CENSUS,
                          "no — 갈래 불변(carry 단언은 기존 행동의 핀)", "yes")
collect["scen"]["B14"] = ("선택을 되돌리지 못함 → INTERNAL_FAILURE(판정 활성화)",
                          collect["scen"]["B14"][1].replace("(이 갈래는 census 만", "(이 갈래는 census 만 — 아래 한계 참조"), "no — 갈래 불변", "census — 도달 불가 공간")
collect["scen"]["B15"] = ("받아들인 범위 0 → NO_ACCEPTED_SCOPE(관문 뒤 대조군 — 판정 활성화, 적재 2 회)", E + "(no scope accepted — loads 2) · " + REVOKE + "(replaced mid-load — READY 가 gen 8)",
                          f"{RED} — 편집 전 적재 1 회 · 성공이 조기 값(gen 7)을 실음", "yes")
collect["invariants"] = [
    "관문 대입은 최상위 문장 둘: 조기(B1 가드 바로 다음) · 판정(`coordinateMarketProposals` 바로 앞, 그 호출의 관문 인자가 `gate`) — census 가 이웃 문장까지 못 박음.",
    "13 닫힘의 kind 순서 · fail 클로저 하나 · 성공 반환 하나가 gate.activation 을 싣는다 — census(새 갈래는 표 편집 강제).",
    "조정 앞 닫힘(B2 · B3 · B4 · B7 · B8 · B9)은 조기 값, 조정 뒤 닫힘(B10~B14)과 성공 · B15 는 판정 값을 싣는다. **잔여 비대칭(관측 전용, Manager 판정 수용)**: "
    "적재 중 철회 경합에서 조정 앞 닫힘은 철회 전 조기 값을 싣는다 — 그 갈래는 항목 0 이라 조정 · handoff · 주문에 닿지 않고 그 주기의 레인 관측에만 쓰인다.",
    "**census 의 한계(8.5 보이스 3 P2-1)**: B11 · B14 갈래 안에서 활성화를 다른 값으로 바꾸는 선택자-대입 변이(E5 · E6)는 census 도 행동도 못 본다. "
    "두 갈래는 입력으로 도달 불가한 공간이다(계보 신원이 LaneID · LaneVersion 을 해시 — strategyflow/types.go; 중복 종목은 B7 이 먼저 닫음) — census-only 를 유지(보이스 3 지지).",
    "familyGateFor 는 읽기 전용 — env 읽기, `LoadProductionFamilyActivation`(매니페스트 파일 읽기 · 검증), 레인 목록 조회. 원장 · 브로커 · 토글 · 게이트웨이 쓰기 0."]
collect["safety"] = [
    "판정 = 편집 전과 같은 함수 · 같은 순간 — 적재 중 ctx 취소 · 매니페스트 철회가 다시 그 주기 판정에 보인다(허용 방향 경합 닫힘; 철회는 하류에 다시 막는 자리가 없었다).",
    "미선언(토글 OFF) 시장은 두 계산 모두 미선언 선반환이라 기존 경로 그대로 — 취소 주기에도 스냅숏 동일(" + LEGACY + ").",
    "대가: 조정에 닿는 주기에 활성화 읽기 1 회 추가(64 KiB 이하 로컬 0400 파일, 핀 선언 시장만 — 오늘 생산 핀 0). 진입 경로라 손절 즉시성과 무관.",
    "닫힌 시장은 항목이 0 이라 실은 활성화로 주문이 나가지 않는다 — 실은 활성화는 그 주기의 레인 관측(승격)에만 쓰인다."]

nil_env = {"bundle": "internal-app-engine--strategyproposalauthorityloader.loadfamilyactivation", "file": "internal/app/engine/strategy_family_activation.go",
           "func": "strategyProposalAuthorityLoader.loadFamilyActivation", "title": "strategyProposalAuthorityLoader.loadFamilyActivation",
           "revision": "편집 전 2 분기 → 3(재번호 `lot-8.5-R/renumber.txt`, difflib 정렬): **B1 새로**(getenv nil → Unavailable, 8.5 응답 로트 ② — codex r2 P2). "
                       "편집 전 B1(US digest env) → B2, B2(US 위험 정책 env) → B3. 앞 판은 nil getenv 에서 공황했고 collect 의 recover 가 같은 주기 사유를 INTERNAL_FAILURE 로 덮었다.",
           "scen": {
               "B1": ("**(새)** getenv 가 nil → 맨 `ErrProductionFamilyActivationUnavailable` — 미선언이 아님(핀을 읽을 수 없음), 관문 되돌림(맨 sentinel 인 이유: 이 경로 호출은 `TestTheRollbackPathOnlyReads` 의 읽기 전용 허용 목록으로 묶임)",
                      NILENV, f"{RED} — 편집 전 공황(nil 함수 호출) · collect 사유 INTERNAL_FAILURE", "yes"),
               "B2": ("US 시장이면 US 활성화 digest env", "`TestStrategyProposalAuthorityLoadsKRUSConcurrently` · `TestADeclaredActivationThatLapsesRollsItsMarketBackInsteadOfWidening`",
                      "no — 갈래 불변(편집 전 B1)", "yes"),
               "B3": ("US 시장이면 US 위험 정책 env", "`TestStrategyProposalAuthorityLoadsKRUSConcurrently`", "no — 갈래 불변(편집 전 B2)", "yes"),
           },
           "invariants": ["env 를 읽을 수 없으면 핀이 비었는지 모른다 — 미선언(기존 경로)으로 답하지 않고 Unavailable(되돌림)로 답한다(더 닫힌 쪽).",
                          "생산 생성자 `newStrategyProposalAuthorityLoader` 는 nil 을 `os.Getenv` 로 채운다 — B1 도달은 손으로 만든 적재기뿐."],
           "state": "상태 변경 없음 — env 읽기와 strategyrouter 적재 호출.",
           "safety": ["B1 은 거절만 더한다(수락 집합 불변) — 공황 → 되돌림으로 바뀌어 같은 주기의 앞선 닫힘 사유(FX_NOT_READY 등)가 보존된다.",
                      "B1 은 새 호출을 더하지 않는다 — 되돌림 경로 읽기 전용 허용 목록(`TestTheRollbackPathOnlyReads`) 불변(첫 판의 `fmt.Errorf` 는 그 시험이 막았다 — 변이 대조군 전체 스위트 실측)."]}

load = spec("internal-strategyrouter--loadproductionfamilyactivation")
load["revision"] = ("분기 불변(10). B5(읽기 결함)의 안쪽 오류를 `%w` → `%v` 로(8.5 응답 로트 ④ — 보이스 2 P2-1): 편집 전(f473d815) `errors.Is` 사슬로 복원 — "
                    "읽기 결함이 공유 읽기 함수의 sentinel(ErrProductionRouteUnavailable)까지 만족하던 둘째 신원 제거. B10 행 정정(보이스 3 P2-5).")
load["scen"]["B5"] = ("매니페스트 파일 읽기 결함 → `%w: manifest file …: %v(읽기 함수 오류 문장)` — 불일치와 다른 종류, 사슬에 활성화 sentinel 하나",
                      F + "(manifest file missing — 모든 모양에서 `errors.Is(err, ErrProductionRouteUnavailable)==false`)",
                      f"{RED} — 편집 전 둘째 `%w` 로 ErrProductionRouteUnavailable 도 만족", "yes")
load["scen"]["B4"] = (load["scen"]["B4"][0], load["scen"]["B4"][1], load["scen"]["B4"][2], "yes — market 항은 아래 불변식(가림) 참조")
load["scen"]["B10"] = ("끝의 ctx 취소 재확인", "도달 불가(결정적 입력 없음 — 진입 확인 :440 뒤 ctx 를 보지 않는 동기 읽기 · 검증뿐) — census/검토로 닫음. "
                       "인용하던 `TestACancelledContextPromotesNothing` 은 앞의 확인(B3)만 지난다(보이스 3 커버리지: 이 몸통 count=0)", "no — 갈래 불변", "도달 불가 — census/검토")
load["invariants"] = load["invariants"] + [
    "sentinel 배타성: 각 거절은 자기 sentinel 하나만 만족한다(" + F + " — 네 sentinel 전부를 모양마다 대조; 8.5 보이스 3 P1-2).",
    "**설정 결속의 market 항(`name == \"\"`)은 판정을 바꾸지 못한다 — 기록(8.5 보이스 2 P2-2, 편집 전부터)**. 가림 가드 둘: (1) 디렉터리 읽기 실패 — 항이 없으면 "
    "`filepath.Join(dir, \"\")` 가 디렉터리를 가리켜 `readProductionRouteFile` 이 정규 0400 파일이 아니라고 거절(B5); (2) 몸통 결속 `body.Market != config.Market || "
    "!validMarket(body.Market)`(validateProductionFamilyActivation B1). 이 블록을 닫아 두는 등식: `ProductionFamilyActivationFileName(m) == \"\" ⇔ !validMarket(m)` "
    "(둘 다 {KR, US} 위의 닫힌 대응 — production_family_activation.go `ProductionFamilyActivationFileName` · types.go `validMarket`). 이 항의 유일한 핀은 메시지 시험 "
    + F + "(config binding: unknown market)이다."]

validate = spec("internal-strategyrouter--validateproductionfamilyactivation")
validate["revision"] = ("분기 불변(8). 서술자 거절 셋(B5 · B6 · B7)의 메시지가 lane_id 원문 대신 위치 `descriptors[i]` 와 필드명만 싣는다(8.5 응답 로트 ③ — codex r2 P2: "
                        "매니페스트의 임의 문자열 · 개행이 오류 문장으로 새지 않음). 순회가 색인을 받는다(`for index, descriptor := range`).")
validate["scen"]["B2"] = ("수명(복합) → `%w: lifetime: <어긋난 항목 전부>`",
                          F + "(lifetime: … 넷 + 둘 동시 + 8.5 응답 로트 ⑧ 셋 — issued_at · expires_at 비정규 · 발급 = 만료 · 만료 < 발급) · `TestAnActivationOutsideItsApprovedLifetimePromotesNothing`",
                          "no — 갈래 불변(새 모양은 기존 항의 핀 — 변이 T565 · T566 · T569/m569)", "yes")
validate["scen"]["B5"] = ("서술자 필드(복합) → `%w: descriptors[i]: <어긋난 필드>`(lane_id 원문 없음)",
                          F + "(descriptor: unknown lane · horizon drift · 개행 lane id · effective 열거 밖 — 8.5 보이스 3 P1-1 · 보이스 1 m602)",
                          f"{RED} — 편집 전 `descriptors[lane_id=<원문>]`(개행 포함)", "yes")
validate["scen"]["B6"] = ("effective ON 인데 desired ON 아님 → `descriptors[i]: effective ON without desired ON`", F + "(descriptor: effective without desired)", RED, "yes")
validate["scen"]["B7"] = ("중복 레인 → `%w: descriptors[i]: duplicate lane_id`(원문 없음)", F + "(descriptor: duplicate lane)", RED, "yes")
validate["safety"] = ["판정 · 수락 집합 불변 — 메시지 문자열만 바뀐다(sentinel 그대로, 배타성은 " + F + " 가 모양마다 잰다)."]

FUNCS = [collect, nil_env, load, validate]

if __name__ == "__main__":
    for item in FUNCS:
        (base.FL / item["bundle"]).mkdir(parents=True, exist_ok=True)
    base.main(FUNCS, tag=TAG, pre=PRE, ledger=LEDGER)
