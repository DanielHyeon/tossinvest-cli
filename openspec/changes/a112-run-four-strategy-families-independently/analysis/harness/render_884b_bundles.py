#!/usr/bin/env python3
"""a112 8.8.4 로트 B — 활성화 거절의 필드명(항목 1, Q-B1=(c)) · 관문 계산 이동(항목 2, Q-B2)의 편집 뒤 번들(render_5222_bundles.main 재사용).

편집 전 번들: `analysis/measurements/lot-8.8.4-B/pre-edit/`(HEAD f473d815). RED: `lot-8.8.4-B/red-8.8.4-B.log`. 변이: `lot-8.8.4-B/mutation-8.8.4-B.tsv`.
재번호: `lot-8.8.4-B/renumber.txt`(편집 전/뒤 ast 를 difflib 으로 정렬 — 손 재번호 아님).
"""
from __future__ import annotations

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import render_5222_bundles as base  # noqa: E402

LEDGER = "`analysis/measurements/lot-8.8.4-B/mutation-8.8.4-B.tsv`"
F = "`a112_activation_error_fields_test.go` `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel`"
D = "`a112_activation_error_fields_test.go` `TestTheDocumentPathNamesItsRefusalToo`"
E = "`a112_proposal_closure_carriage_test.go` `TestEveryReachableProposalClosureCarriesTheGatesActivationExceptRouteNotReady`"
CENSUS = "`a112_proposal_closure_carriage_test.go` `TestTheThirteenProposalClosuresKeepTheirOrderAndTheGateIsComputedRightAfterRouteReadiness`"
RED = "yes — 편집 전 메시지에 필드명 없음(`red-8.8.4-B.log`)"
SAME = "no — 갈래 불변"
PFA = "internal/strategyrouter/production_family_activation.go"
WRAP_SAFETY = ["판정 · 수락 집합 불변 — 같은 입력이 같은 sentinel 로 거절된다(errors.Is 보존, `==` 비교 0 — grep). 메시지에 필드 이름만 더해진다.",
               "복합 결속은 분기 하나 그대로 두고 그 안에서 필드별 비교를 모은다(Q-B1=(c)); 항은 전부 순수 비교 · 부작용 없는 검사라 무조건 평가해도 판정이 같다(2026-10-04 단락 평가 전제 확인)."]
FUNCS = [
    {"bundle": "internal-strategyrouter--familyactivationdocument.body", "file": PFA, "func": "FamilyActivationDocument.body", "title": "FamilyActivationDocument.body",
     "revision": "분기 불변(5). B1 · B3 의 맨 sentinel 반환에 이유(시장 표 없음 · 모르는 가족 이름)를 `%w` 로 붙였다.",
     "scen": {
         "B1": ("서술자 표가 없는 시장 → `%w: market … has no descriptor table`", D, RED, "yes"),
         "B2": ("켤 가족 순회", "`TestTheCommittedGoldenManifestMatchesItsPinAndPromotesTheFourLanes` · " + D, SAME, "yes"),
         "B3": ("모르는 가족 이름 → `%w: on: unknown family …`", D, RED, "yes"),
         "B4": ("표 순회로 네 서술자 생성", "`TestTheAuthoringEncoderStillProducesTheCommittedGoldenBytes`", SAME, "yes"),
         "B5": ("켠 가족이면 ON", "`TestTheAuthoringEncoderStillProducesTheCommittedGoldenBytes`", SAME, "yes"),
     },
     "invariants": ["도구 경로(매니페스트 작성)의 거절이 이유를 말한다 — 검사 자체는 LoadProductionFamilyActivation 하나가 한다(이 함수 머리말)."],
     "state": "상태 변경 없음.", "safety": WRAP_SAFETY[:1]},
    {"bundle": "internal-strategyrouter--familyactivationremaining", "file": PFA, "func": "familyActivationRemaining", "title": "familyActivationRemaining",
     "revision": "분기 불변(1). 만료 반환에 `expires_at` 과 두 시각을 `%w` 로 붙였다(ErrProductionFamilyActivationExpired 보존).",
     "scen": {"B1": ("만료(지금이 expires_at 이상) → `%w: expires_at … is not after …`", F + "(expired) · `TestTheLeaseCeilingOnlyEverShrinks` · `TestLoadingAndLeasingJudgeExpiryAtTheSameInstant`",
                     f"yes — 편집 전 필드명 없음; 변이 B10(sentinel 탈락) CAUGHT({LEDGER})", "yes")},
     "invariants": ["이 패키지의 유일한 만료 판정 — 적재와 lease 상한이 같이 부른다(판정 불변)."],
     "state": "상태 변경 없음.", "safety": WRAP_SAFETY[:1]},
    {"bundle": "internal-strategyrouter--loadproductionfamilyactivation", "file": PFA, "func": "LoadProductionFamilyActivation", "title": "LoadProductionFamilyActivation",
     "revision": "편집 전 9 분기 → 10(재번호 `lot-8.8.4-B/renumber.txt`, difflib 정렬): 편집 전 B5(`err != nil || digest != pin`)가 B5(읽기 결함 — 읽기 함수의 오류를 `%w` "
                 "사슬에 보존) · B6(새 — 핀 불일치, `manifest_digest`)으로 갈렸다(Manager 판정 — 결함과 불일치는 다른 종류). 편집 전 B6~B9 → B7~B10. B4(설정 결속)는 같은 분기 "
                 "하나로 조건을 `len(failedFields(...)) != 0` 으로 바꿔 어긋난 필드 전부를 싣는다. 나머지 맨 sentinel 반환은 이유를 `%w` 로 붙였다.",
     "scen": {
         "B1": ("핀이 비었으면 미선언 → `%w(Undeclared): manifest_digest pin is empty` — **맨 앞 순서 불변**(엔진 판별이 errors.Is 로 의존)",
                F + "(undeclared pin) · `TestOnlyAnEmptyPinMeansTheActivationWasNeverDeclared`", f"yes — 변이 B1(sentinel 탈락) CAUGHT({LEDGER})", "yes"),
         "B2": ("ctx 가 nil → `%w: context is nil`", D, RED, "yes"),
         "B3": ("ctx 취소 → ctx.Err() 그대로", "`TestACancelledContextPromotesNothing`", SAME, "yes"),
         "B4": ("설정 결속(복합 — 분기 하나) → `%w: config binding: <어긋난 필드 전부>`",
                F + "(config binding: … 아홉 단일 + 세 필드 동시)", f"yes — 변이 B2 · B3 · B4 · B5 CAUGHT({LEDGER})", "yes"),
         "B5": ("매니페스트 파일 읽기 결함 → `%w: manifest file …: %w(읽기 함수 오류)` — 불일치와 다른 종류", F + "(manifest file missing)",
                f"yes — 편집 전에는 불일치와 한 갈래; 변이 B6(%v) · B7(불일치로 위장) CAUGHT({LEDGER})", "yes"),
         "B6": ("**(새)** 핀이 파일 바이트와 다름 → `%w: manifest_digest: …`", F + "(pin does not match the file bytes) · `TestBytesThatChangedAfterTheDeploymentPinnedThemPromoteNothing`",
                RED, "yes"),
         "B7": ("해석 거절 → 해석기가 붙인 이유 그대로(sentinel 포함)", F + "(bytes not canonical · trailing data · unknown field)", f"yes — 변이 B8(맨 sentinel) CAUGHT({LEDGER})", "yes"),
         "B8": ("폐기 → `%w(Revoked): revoked=true`", F + "(revoked)", RED, "yes"),
         "B9": ("검증 거절 → 검증기 오류 그대로", F + "(body binding · lifetime · descriptor · expired)", SAME, "yes"),
         "B10": ("끝의 ctx 취소 재확인", "`TestACancelledContextPromotesNothing`", SAME, "도달 — 측정은 기존 BTM"),
     },
     "invariants": ["미선언 판정이 맨 앞 — 엔진 `familyGateFor` 의 `errors.Is(err, ErrProductionFamilyActivationUndeclared)` 판별 불변.",
                    "읽기 결함(공유 읽기 함수 `readProductionRouteFile` 의 오류)과 핀 불일치는 다른 메시지 · 다른 사슬. 그 읽기 함수는 OS 원인을 자기 sentinel 하나로 접는다 — OS 원인 노출은 이 로트 밖(잔여)."],
     "state": "상태 변경 없음 — 파일 읽기 · 검증.", "safety": WRAP_SAFETY + ["B5/B6 분리로 분기 하나가 늘었지만 둘 다 거절(같은 sentinel)이라 수락 집합 불변."]},
    {"bundle": "internal-strategyrouter--decodeproductionfamilyactivation", "file": PFA, "func": "decodeProductionFamilyActivation", "title": "decodeProductionFamilyActivation",
     "revision": "분기 불변(4). 네 거절이 이유(크기 · json · 뒤 데이터 · 정규 직렬화 아님)를 `%w` 로 싣는다 — json 오류는 원래 오류도 사슬에 남긴다. 호출자(Load B7)는 이 오류를 그대로 돌려준다.",
     "scen": {
         "B1": ("크기 0 또는 상한 초과 → `%w: manifest size …`", "읽기 함수가 크기를 먼저 막아 Load 경로로는 도달 불가 — 정규 직렬화 등식이 같은 몫", SAME, "도달 불가(Load 경로)"),
         "B2": ("json 해석 실패(모르는 필드 포함) → `%w: manifest json: %w`", F + "(an unknown field)", RED, "yes"),
         "B3": ("문서 뒤 데이터 → `%w: trailing data …`", F + "(trailing data after the document)", RED, "yes"),
         "B4": ("정규 직렬화와 다름 → `%w: manifest bytes are not …canonical…`", F + "(bytes not canonical) · `TestAnActivationWhoseBytesAreNotCanonicalPromotesNothing`", RED, "yes"),
     },
     "invariants": ["판정 불변 — 같은 바이트가 같은 sentinel 로 거절."], "state": "상태 변경 없음.", "safety": WRAP_SAFETY[:1]},
    {"bundle": "internal-strategyrouter--validateproductionfamilyactivation", "file": PFA, "func": "validateProductionFamilyActivation", "title": "validateProductionFamilyActivation",
     "revision": "분기 불변(8, difflib 정렬 — B1 · B2 · B5 의 조건만 다시 씀). B1(몸통 결속) · B2(수명) · B5(서술자 필드)는 같은 분기 하나로 `len(failedFields(...)) != 0` — 어긋난 "
                 "필드 전부. B5 의 표 대조 셋은 `known &&` 로 묶어 모르는 레인이 표 필드까지 탓하지 않게 했다(판정은 앞 판 `!known || …` 과 같다). B6 · B7 · B8 은 이유를 `%w` 로.",
     "scen": {
         "B1": ("몸통 결속(복합) → `%w: body binding: <어긋난 필드 전부>`", F + "(body binding: … 열하나 단일 + 세 필드 동시) · `TestAnActivationWhoseBindingsDoNotMatchPromotesNothing`",
                f"yes — 변이 B2 · B3 CAUGHT({LEDGER})", "yes"),
         "B2": ("수명(복합) → `%w: lifetime: <어긋난 항목 전부>`", F + "(lifetime: … 넷 + 둘 동시) · `TestAnActivationOutsideItsApprovedLifetimePromotesNothing`", RED, "yes"),
         "B3": ("만료 → familyActivationRemaining 의 Expired 오류 그대로", F + "(expired)", SAME, "yes"),
         "B4": ("서술자 순회", "`TestAVerifiedFourFamilyActivationPromotesExactlyTheLanesItNames`", SAME, "yes"),
         "B5": ("서술자 필드(복합) → `%w: descriptors[lane_id=…]: <어긋난 필드>`", F + "(descriptor: unknown lane · horizon drift)", f"yes — 변이 B9(known 가드 탈락) CAUGHT({LEDGER})", "yes"),
         "B6": ("effective ON 인데 desired ON 아님 → 이유", F + "(descriptor: effective without desired)", RED, "yes"),
         "B7": ("중복 레인 → `%w: descriptors: duplicate lane_id …`", F + "(descriptor: duplicate lane)", RED, "yes"),
         "B8": ("네 레인이 아님 → `%w: descriptors: N of M lanes`", F + "(descriptor: three of four) · `TestAnActivationWhoseDescriptorSetIsNotExactlyTheFourPromotesNothing`", RED, "yes"),
     },
     "invariants": ["판정 불변 — 복합 결속의 결정식은 앞 판 OR 의 항과 같다(`len(fields) != 0`)."], "state": "상태 변경 없음.", "safety": WRAP_SAFETY},
    {"bundle": "internal-strategyrouter--failedfields", "file": PFA, "func": "failedFields", "title": "failedFields (새 함수)",
     "revision": "새 함수(a112 8.8.4 로트 B) — 편집 전 없음. 복합 결속의 진단: 실패한 이름만 순서대로 모은다.",
     "scen": {
         "B1": ("검사 순회", F, f"yes — 새 함수; 변이 B2(첫 실패만) CAUGHT({LEDGER})", "yes"),
         "B2": ("실패한 검사의 이름을 모음", F + "(세 필드 동시 · 두 항목 동시)", f"yes — 변이 B3(아무것도 안 모음 → 결속 통과) CAUGHT({LEDGER})", "yes"),
     },
     "invariants": ["결정은 호출자의 `len(fields) != 0` 하나 — 이 함수가 비면 결속이 통과하므로(B3) 이름 수집이 곧 판정 입력이다."],
     "state": "상태 변경 없음.", "safety": ["판정 입력이다 — 변이 B3 이 그것을 잡는다."]},
    {"bundle": "internal-app-engine--strategyproposalauthorityloader.collectmarket", "file": "internal/app/engine/strategy_proposal_authority.go",
     "func": "strategyProposalAuthorityLoader.collectMarket", "title": "strategyProposalAuthorityLoader.collectMarket",
     "revision": "분기 불변(15). 관문 **계산**(`gate = loader.familyGateFor(...)`)만 제안 조정 직전에서 B1(ROUTE_NOT_READY) 뒤로 옮겼다(Manager 판정 Q-B2 — 계산/판정 분리). "
                 "13 닫힘의 실패 kind 순서 · FAMILY_GATE_CLOSED 판정 자리는 그대로이고, 옮긴 값은 닫힘 갈래가 싣는 활성화와 조정 관문으로만 쓰인다. 결과: B1 만 영값, 나머지 열두 닫힘이 관문 활성화를 싣는다(앞 판은 일곱이 영값).",
     "scen": {
         "B1": ("경로 · 스케줄 미준비 → ROUTE_NOT_READY(**영값** — 관문 계산 전; 결속 값이 없거나 믿을 수 없음)", E + "(route not ready — 적재 0 회 · 영값) · " + CENSUS,
                f"no — 갈래 불변; 변이 E2(관문을 이 가드 앞으로) CAUGHT({LEDGER})", "yes"),
         "B2": ("FX 미준비 → FX_NOT_READY(관문 활성화)", E + "(fx not ready)", RED.replace("메시지에 필드명 없음", "영값을 실음"), "yes"),
         "B3": ("적재기 설정 결손 → INTERNAL_FAILURE(관문 활성화)", E + "(loader misconfigured)", RED.replace("메시지에 필드명 없음", "영값을 실음"), "yes"),
         "B4": ("제안 공개 열쇠 무효 → AUTHORITY_INVALID(관문 활성화)", E + "(proposal public key invalid)", RED.replace("메시지에 필드명 없음", "영값을 실음"), "yes"),
         "B5": ("US 시장 env 이름", "`TestStrategyProposalAuthorityLoadsKRUSConcurrently`", SAME, "yes"),
         "B6": ("경로 항목 순회", CENSUS, SAME, "yes"),
         "B7": ("종목 빈값 · 중복 → INTERNAL_FAILURE(관문 활성화)", E + "(duplicate routed symbol)", RED.replace("메시지에 필드명 없음", "영값을 실음"), "yes"),
         "B8": ("제안 적재 실패 · digest 불일치 → AUTHORITY_INVALID(관문 활성화)", E + "(proposal load failed)", RED.replace("메시지에 필드명 없음", "영값을 실음"), "yes"),
         "B9": ("범위가 제안을 잃음 → PROPOSAL_PRODUCTION_FAULT(관문 활성화)", E + "(a scope lost its proposal) · `TestALostProposalClosesTheMarketInsteadOfReleasingTheOtherSymbol`",
                RED.replace("메시지에 필드명 없음", "영값을 실음"), "yes"),
         "B10": ("관문이 범위를 통째로 지움 → FAMILY_GATE_CLOSED(판정 자리 불변)", "`TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve` · " + CENSUS, SAME, "yes"),
         "B11": ("계보 신원 충돌 → INTERNAL_FAILURE", CENSUS + "(입력으로 도달 어려움 — census 만)", SAME, "census"),
         "B12": ("조정자 넘침 → QUEUE_OVERFLOW", "`TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne` · " + CENSUS, SAME, "yes"),
         "B13": ("중재 거절 → ARBITRATION_REFUSED", "`TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol` · " + CENSUS, SAME, "yes"),
         "B14": ("선택을 되돌리지 못함 → INTERNAL_FAILURE", CENSUS + "(이 갈래는 census 만 — 술어 `entries()` 는 `TestASelectionWithNoLaneToComeBackToClosesInsteadOfShrinkingTheList` 가 단위로)", SAME, "census"),
         "B15": ("받아들인 범위 0 → NO_ACCEPTED_SCOPE(관문 뒤 대조군)", E + "(no scope accepted)", SAME, "yes"),
     },
     "invariants": ["13 닫힘의 kind 순서와 관문 계산 자리 하나 · fail 클로저 하나 · 성공 반환 하나가 gate.activation 을 싣는다 — census 가 못 박음(새 갈래는 표 편집 강제).",
                    "familyGateFor 는 읽기 전용 — env 읽기, `LoadProductionFamilyActivation`(매니페스트 파일 읽기 · 검증), 레인 목록 조회. 원장 · 브로커 · 토글 · 게이트웨이 쓰기 0."],
     "state": "상태 변경 없음 — 권위 값을 조립해 돌려준다.",
     "safety": ["관문 계산이 앞당겨져 FX · 설정 · 열쇠 · 중복 · 적재 · 고장으로 닫히는 주기에도 활성화 매니페스트 읽기가 돈다(시장당 파도당 소형 파일 1회, 읽기 전용). "
                "진입 경로라 손절 즉시성과 무관하다(Manager 판정 Q-B2 수용).",
                "닫힌 시장은 항목이 0 이라 실은 활성화로 주문이 나가지 않는다 — 실은 활성화는 그 주기의 레인 관측(승격)에만 쓰인다(TestAClosedMarketStillCarriesTheGatesActivation)."]},
]


if __name__ == "__main__":
    for spec in FUNCS:
        (base.FL / spec["bundle"]).mkdir(parents=True, exist_ok=True)
    base.main(FUNCS, tag="a112 8.8.4-B", pre="lot-8.8.4-B", ledger=LEDGER)
