#!/usr/bin/env python3
"""a112 5.2.2.2 리뷰 수리 로트 — 편집한 엔진 함수 여섯과 riskbucket 함수 셋의 편집 뒤 번들을 ast.json 에서 채운다(렌더러는 render_5222_bundles.main).

편집 전 번들: `analysis/measurements/lot-5.2.2.2-fix/pre-edit/`(엔진 다섯은 80ae96a5 착지의 편집 뒤 번들 사본, riskbucket 셋은 편집 전 AST + 지도,
admit 은 `lot-5.2.2.2` 의 편집 뒤 번들 = 80ae96a5 상태). 변이 원장 `analysis/measurements/lot-5.2.2.2-fix/mutation-5.2.2.2-fix.tsv`.

사용: python3 render_5222_fix_bundles.py   (저장소 루트에서)
"""
from __future__ import annotations

import importlib.util
from pathlib import Path

HERE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location("render_5222_bundles", HERE / "render_5222_bundles.py")
base = importlib.util.module_from_spec(spec)
spec.loader.exec_module(base)

E, T, SEAL, HO, CEN, SAME = base.E, base.T, base.SEAL, base.HO, base.CEN, base.SAME
LEDGER = "`analysis/measurements/lot-5.2.2.2-fix/mutation-5.2.2.2-fix.tsv`"
RED = "`analysis/measurements/lot-5.2.2.2-fix/red-fix.log`"
RB = "internal/riskbucket/production_snapshot_authority.go"
UNCHANGED = ("편집 없음(좌표만 — 리뷰 수리 로트)", "편집 전 번들 서술 그대로", "no — 이 로트가 바꾸지 않음", "편집 전 번들의 측정")


def same(meaning: str):
    return (meaning,) + UNCHANGED[1:]


FUNCS = [
    {
        "bundle": "internal-app-engine--buildproductionstrategymarketworker", "file": E + "strategy_entry_supervisor.go",
        "func": "buildProductionStrategyMarketWorker", "title": "buildProductionStrategyMarketWorker",
        "revision": "리뷰 수리(A #3): 승격 근거 범위가 자기 위험 · 계좌 권한을 갖도록 B5(키) · B6(위험 forScope) · B7(계좌 forScope)를 더했고, "
                    "만료를 준비된 계좌 범위의 최소값(`earliestFreshUntil`)으로 바꿨다(B11 조건의 만료 항). 나머지는 80ae96a5 와 같은 분기(번호 이동).",
        "scen": {
            "B1": same("배선 미완/nil → dormant"),
            "B2": same("시장 권한 준비 미완 → dormant"),
            "B3": same("주문 경로와 같은 handoff 목록 순회"),
            "B4": same("handoff 거절 · 봉인 깨짐 → 그 범위 건너뜀"),
            "B5": ("**(새)** 범위 키 정규화 실패 → 건너뜀", "진입 0 — 조립 결과는 늘 유효 키(seam 으로 못 만듦)", "no", "진입 0"),
            "B6": ("**(새)** 그 범위의 위험 권한 준비 안 됨 → 건너뜀(승격 근거 아님)",
                   f"{T} `TestAWorkerPromotesOnlyOnAScopeWithBothAuthorities/000660_without_risk_authority` · `TestAnActivatedTwoScopeMarketPromotesItsWorkerPerScope`",
                   f"yes — {RED}(편집 전 권한 없는 범위로 승격) · 변이 Y15 CAUGHT({LEDGER})", "yes"),
            "B7": ("**(새)** 그 범위의 계좌 권한 준비 안 됨 → 건너뜀",
                   f"{T} `TestAWorkerPromotesOnlyOnAScopeWithBothAuthorities/000660_without_account_authority`", "yes — 변이 Y16 CAUGHT", "yes"),
            "B8": same("보호 관측 실패 → 그 범위 건너뜀"),
            "B9": ("진입 관문 관측 실패 → 그 범위 건너뜀(J3)", f"{T} `TestAnActivatedTwoScopeMarketPromotesItsWorkerPerScope`", "80ae96a5 변이 X15 · X17", "yes"),
            "B10": same("승격 근거 범위 없음 → dormant"),
            "B11": ("digest / revision / **만료(준비된 계좌 범위의 최소 FreshUntil)** → dormant",
                    f"{T} `TestAWorkerOverTwoReadyScopesExpiresWithItsEarliestAccountScope`", f"yes — {RED} · 변이 Y17 CAUGHT", "yes"),
        },
        "invariants": ["활성화 없는 시장은 범위 하나 — 최소 만료 = 그 범위의 FreshUntil = 편집 전 값, forScope 는 그 범위 그대로(새 거절 0)."],
        "safety": ["High-risk 인접(승격). 편집은 승격을 **좁히기만** 한다(권한 없는 범위로의 승격 · 늦은 만료 제거). 주문은 dispatch 가 범위마다 다시 검사."],
    },
    {
        "bundle": "internal-app-engine--productionstrategyfirstlegauthorityloader.collectstrategyfirstlegauthority",
        "file": E + "strategy_account_first_leg_authority.go", "func": "productionStrategyFirstLegAuthorityLoader.collectStrategyFirstLegAuthority",
        "title": "collectStrategyFirstLegAuthority",
        "revision": "리뷰 수리: B5(**활성화 없는 시장의 개수 관문** — codex #1, 6.2 위치), B6(키 정규화 실패 → 결함, A #4), B8 · B10(위험 · 계좌 범위 권한 부재가 "
                    "범위 국소 원인일 때만 범위 거절, 아니면 타입 없는 결함 — A #1 · codex #2)를 더했다. 나머지는 80ae96a5 와 같은 분기(번호 이동).",
        "scen": {
            "B1": same("loader · ctx · 시계 · 원장 · Guardian 부재"),
            "B2": same("시장 권한 준비 미완"),
            "B3": same("소유자 범위 선택 실패 — 위조 의심(타입 없음)"),
            "B4": ("봉인된 identity 대조 — 불일치는 타입 없는 오류", f"{SEAL} `TestTheFirstLegSealRefusesEveryForgeryAxis` · {T} `TestAForgedScopeStopsTheCycleBeforeTheNextValidScope`",
                   "80ae96a5 변이 X07", "yes"),
            "B5": ("**(새)** 서명 활성화 **없는** 시장에서 항목이 정확히 하나가 아님 → `paired production authority is incomplete for market`(편집 전 문구)",
                   f"{SEAL} `TestAnUnactivatedMultiEntryPairIsStillRefusedAtTheFirstLeg`",
                   f"yes — {RED}(편집 전 err=nil) · 변이 Y12(제거) · Y13(활성화 시장에도) CAUGHT({LEDGER})", "yes"),
            "B6": ("**(새)** 범위 키 정규화 실패 → 결함(타입 없음)", "진입 0 — 봉인 선택이 정규화를 보장(도달 불가)", "M14 SURVIVED — 예상(도달 불가)", "진입 0"),
            "B7": ("그 범위의 준비된 위험 권한 없음", f"{T} `TestARiskScopeOutsideTheSignedPolicyIsRefusedAloneInEitherOrder` · `TestACorruptLedgerRowStopsTheCycleWithItsCause`",
                   "yes — 아래 B8", "yes"),
            "B8": ("**(새)** 원인이 범위 국소가 아님(원장 결함 · 무결성 · 항목 부재) → 타입 없는 결함 `production risk authority fault …: %w`(주기 멈춤 · 원인 보존); "
                   "범위 국소(`riskbucket.ErrProductionRiskScopeRefused`)면 범위 거절 타입(원인 Unwrap)",
                   f"{T} `TestACorruptLedgerRowStopsTheCycleWithItsCause` · `TestAScopeLatchIsRefusedAloneButALatchReadFaultStops` · `TestARiskScopeOutsideTheSignedPolicyIsRefusedAloneInEitherOrder` · "
                   f"{HO} `TestAScopeTheRiskAuthorityDoesNotHoldIsAFaultInEitherOrder`",
                   f"yes — {RED}(편집 전 손상 행 → placed=[000660] · 범위 거절) · 변이 Y06 · Y07 · Y08 CAUGHT", "yes"),
            "B9": ("그 범위의 준비된 계좌 권한 없음", f"{T} `TestAScopeWithoutItsOwnAccountAuthorityIsRefusedAloneAndRecorded` · `TestAnAccountLoadCancelledByItsContextStopsTheCycle`",
                   "yes — 아래 B10", "yes"),
            "B10": ("**(새)** 계좌 원인이 ctx 종료 · 항목 부재면 결함(타입 없음), 그 밖의 적재 실패(서명 매니페스트)면 범위 거절",
                    f"{T} `TestAnAccountLoadCancelledByItsContextStopsTheCycle` · `TestAScopeWithoutItsOwnAccountAuthorityIsRefusedAloneAndRecorded`",
                    f"yes — {RED}(편집 전 ctx 취소 → placed=[000660]) · 변이 Y09 · Y10 CAUGHT", "yes"),
            "B11": same("계보 시장 통화 미지"),
            "B12": same("위험 권한 범위 불일치(범위 번들)"),
            "B13": same("포지션 캠페인 CAS 변경"),
            "B14": same("위험 버킷 항목 순회"),
            "B15": same("가격 단위 무효"),
            "B16": same("노출 스냅숏 만료(collect 클로저)"),
            "B17": same("예약 버전 읽기 실패(collect 클로저)"),
        },
        "invariants": ["범위 거절 타입을 만드는 자리는 B8 · B10 의 범위 국소 갈래 둘뿐 — census `TestTheScopeRefusalTypeIsMadeOnlyWhereTheCensusSaysItIs` · "
                       "`TestNoEngineErrorTypeImplementsAs`.",
                       "활성화 밖에서 마지막 경계의 수용 집합은 80ae96a5 이전(7ab8cd12)과 같다(B5)."],
        "safety": ["High-risk(1차 레그). 편집은 거절을 **늘리기만** 한다(활성화 없는 다항목 쌍 · 범위 국소가 아닌 결함은 이제 주기를 멈춤). 새로 통과하는 입력 0."],
    },
    {
        "bundle": "internal-app-engine--strategyaccountauthorityloader.collectmarket", "file": E + "strategy_account_first_leg_authority.go",
        "func": "strategyAccountAuthorityLoader.collectMarket", "title": "strategyAccountAuthorityLoader.collectMarket",
        "revision": "리뷰 수리: 적재 결과를 switch(B6~B10)로 나눠 실패 원인을 범위 칸에 운반하고(B7), 실패 뒤 ctx 가 끝났으면 원인을 ctx 로 바꾼다(B8 — 생산 "
                    "적재기는 ctx 종료를 자기 오류로 접으므로, 조건 ④). 준비 판정 자체(B9)는 80ae96a5 와 같다.",
        "scen": {
            "B1": same("항목 없음 · 활성화 없는 시장의 항목 하나 아님/무효"),
            "B2": same("loader 구성 불완전"),
            "B3": same("시장이 US 면 계좌 시장 US"),
            "B4": same("항목(범위)마다 적재"),
            "B5": same("범위 키 · 제안 유효할 때만 적재"),
            "B6": ("**(새)** 적재 결과 분기", f"{T} `TestAScopeWithoutItsOwnAccountAuthorityIsRefusedAloneAndRecorded`", "no — 구조", "yes"),
            "B7": ("**(새)** 적재 실패 → 원인 운반", f"{T} `TestAScopeWithoutItsOwnAccountAuthorityIsRefusedAloneAndRecorded` · `TestAnAccountLoadCancelledByItsContextStopsTheCycle`",
                   "yes — 변이 Y10 CAUGHT(전부 결함이면 J3 시험 FAIL)", "yes"),
            "B8": ("**(새)** 실패 뒤 ctx 종료 → 원인 = ctx 오류(결함)", f"{T} `TestAnAccountLoadThatFailsUnderACancelledContextIsAFault`",
                   "yes — 변이 Y11 CAUGHT", "yes"),
            "B9": same("적재 성공 · 시장 · 매니페스트 일치 → 그 범위 준비"),
            "B10": ("**(새)** 적재 성공인데 시장 · 매니페스트 불일치 → 결함 원인", "진입 0 — 시험 적재기는 늘 일치(seam 으로 못 만듦)", "no", "진입 0"),
        },
        "invariants": ["준비 판정은 편집 전과 같다 — 바뀐 것은 실패 원인의 운반뿐(판정 불변)."],
        "safety": ["High-risk 인접. 새로 통과 · 새로 거절하는 입력 0(원인 분류는 1차 레그에서)."],
    },
    {
        "bundle": "internal-app-engine--strategyriskauthorityloader.collectmarket", "file": E + "strategy_risk_authority.go",
        "func": "strategyRiskAuthorityLoader.collectMarket", "title": "strategyRiskAuthorityLoader.collectMarket",
        "revision": "리뷰 수리: 적재 결과를 switch(B6~B9)로 나눠 실패 원인을 범위 칸에 **그대로** 운반한다(B7 — 편집 전에는 어떤 err 든 「준비 안 됨」 하나로 "
                    "접었다, A #1 · codex #2). 준비 판정(B8)은 80ae96a5 와 같다.",
        "scen": {
            "B1": same("결과 권한 준비 안 됨"),
            "B2": same("환율 준비 안 됨"),
            "B3": same("시장이 US 면 버킷 시장 US"),
            "B4": same("범위마다 번들 하나"),
            "B5": same("범위 키 정규화 성공 시에만 적재"),
            "B6": ("**(새)** 적재 결과 분기", f"{T} `TestACorruptLedgerRowStopsTheCycleWithItsCause`", "no — 구조", "yes"),
            "B7": ("**(새)** 적재 실패 → 적재기 원인 그대로 운반(범위 국소 신원 포함 여부는 1차 레그가 가름)",
                   f"{T} `TestACorruptLedgerRowStopsTheCycleWithItsCause` · `TestARiskScopeOutsideTheSignedPolicyIsRefusedAloneInEitherOrder`",
                   f"yes — {RED} · 변이 Y06(원인을 범위 국소로 덮음) CAUGHT", "yes"),
            "B8": same("적재 성공 · 시장 · 계좌 · 시각 · 항목 5 일치 → 그 범위 준비"),
            "B9": ("**(새)** 적재 성공인데 시장 · 계좌 · 시각 · 항목 수 불일치 → 결함 원인", "진입 0 — 생산 적재기가 같은 값으로 봉인(seam 으로 못 만듦)", "no", "진입 0"),
        },
        "invariants": ["준비 판정 불변 — 바뀐 것은 실패 원인의 운반뿐."],
        "safety": ["High-risk(위험 권한). 새로 통과 · 새로 거절하는 입력 0."],
    },
    {
        "bundle": "internal-app-engine--strategyfirstlegadmissionbridge.admit", "file": E + "strategy_first_leg_admission.go",
        "func": "strategyFirstLegAdmissionBridge.admit", "title": "admit",
        "revision": "리뷰 수리: 80ae96a5 의 B4(수집 오류가 범위 거절 타입이면 결과에 싣기 — errors.As)를 지우고 B3 에서 **수집 오류 그대로**를 `cause` 에 싣는다 — "
                    "범위 거절 타입과 결함 원인의 신원이 모두 dispatch 로 건너간다(문구 불변). 7 분기 → 6.",
        "scen": {
            "B1": same("결과 검증 거절"),
            "B2": same("bridge · loader · Guardian 부재"),
            "B3": ("1차 레그 권한 수집 실패 → `AuthorityCollectionFailed`, 수집 오류를 `cause` 로 운반",
                   f"{T} `TestACorruptLedgerRowStopsTheCycleWithItsCause` · `TestAScopeWithoutItsOwnAccountAuthorityIsRefusedAloneAndRecorded`",
                   f"yes — {RED}(편집 전 원인 신원 소실) · 변이 Y19 CAUGHT", "yes"),
            "B4": same("권한 불일치"),
            "B5": same("Guardian precheck 실패"),
            "B6": same("원자 admission 실패(원장 `BUCKET_USAGE_STALE` 은 cause 없이 문구로 — 타입 없음)"),
        },
        "invariants": ["admit 은 범위 거절 타입을 만들지도 언급하지도 않는다(census)."],
        "safety": ["High-risk(admission). 거절 코드 · 문구 불변, 새로 통과하는 입력 0."],
    },
    {
        "bundle": "internal-app-engine--strategydispatchcycle.dispatch", "file": E + "strategy_dispatch_cycle.go",
        "func": "strategyDispatchCycle.dispatch", "title": "dispatch",
        "revision": "리뷰 수리: B13 의 조건을 `admitted.scope != nil` 에서 `admitted.cause != nil` 로 — 수집 오류를 사슬째(`%w`) 나른다(범위 거절 타입 · 결함 원인 "
                    "모두). 분기 수 · 위치 불변(22).",
        "scen": {
            **{f"B{i}": same("80ae96a5 와 같은 분기") for i in range(1, 13)},
            "B13": ("수집 오류가 있으면 사슬째 `%w`(문구 = Detail), 없으면 문구만", f"{T} `TestACorruptLedgerRowStopsTheCycleWithItsCause` · `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope`(STALE 은 타입 없음)",
                    "yes — 변이 Y20(%v) · Y21(원인 없는 거절에 범위 타입) CAUGHT", "yes"),
            **{f"B{i}": same("80ae96a5 와 같은 분기") for i in range(14, 23)},
        },
        "invariants": ["dispatch 는 범위 거절 타입을 만들지 않는다 — 수집 오류를 나를 뿐."],
        "safety": ["High-risk(주문 경로). 새로 통과하는 입력 0 — 오류 사슬만 보존."],
    },
    {
        "bundle": "internal-riskbucket--loadproductionrisksnapshotauthority", "file": RB, "func": "LoadProductionRiskSnapshotAuthority",
        "title": "LoadProductionRiskSnapshotAuthority",
        "revision": "리뷰 수리(J4 = (A)): B6 · B7 의 감싸기를 `%w: %v` → `%w: %w` 로 — 원인의 신원(범위 국소 sentinel · 원장 결함)을 사슬에 보존. 문구 · 분기 · 판정 불변.",
        "scen": {
            "B1": same("ctx · 관측 시각 부재"), "B2": same("ctx 종료"), "B3": same("구성 · 소유자 · 경로 · digest 형식"),
            "B4": same("매니페스트 파일 · digest 불일치"), "B5": same("매니페스트 해석 · 서명 검증 실패"),
            "B6": ("bind 실패 → `ErrProductionRiskSnapshotUnavailable` + **원인 %w**(범위 국소 sentinel 보존)",
                   f"{T} `TestARiskScopeOutsideTheSignedPolicyIsRefusedAloneInEitherOrder`", f"yes — {RED} · 변이 Y05(%v) CAUGHT({LEDGER})", "yes"),
            "B7": ("원장 항목 적재 실패 → `ErrProductionRiskSnapshotUnavailable` + **원인 %w**(latch sentinel · 원장 결함 원인 보존)",
                   f"{T} `TestACorruptLedgerRowStopsTheCycleWithItsCause` · `TestAScopeLatchIsRefusedAloneButALatchReadFaultStops`", "yes — 변이 Y04(결함에도 sentinel) CAUGHT", "yes"),
        },
        "invariants": ["판정 무변(Manager 조건 ①): 거절되는 입력 집합은 편집 전과 같다 — 감싸기 동사만 바뀌었다(편집 전 번들 `lot-5.2.2.2-fix/pre-edit/`)."],
        "safety": ["High-risk(위험 권한). 신원 배관만 — 새로 통과 · 새로 거절 0."],
        "state": "원장 읽기 전용(`mode=ro` · `query_only`), 쓰기 없음.",
    },
    {
        "bundle": "internal-riskbucket--bindproductionriskinputs", "file": RB, "func": "bindProductionRiskInputs", "title": "bindProductionRiskInputs",
        "revision": "리뷰 수리(J4 = (A)): B4(서명 정책에 그 종목의 섹터 매핑 없음)의 오류를 `ErrProductionRiskScopeRefused` 로 감쌌다 — 범위 국소 신원. 분기 · 판정 불변.",
        "scen": {
            "B1": same("봉인된 결과 불일치(무결성 — 결함)"), "B2": same("지원하지 않는 horizon"), "B3": same("전략 위험 매핑 없음(결함 — 범위 국소로 넓히지 않음)"),
            "B4": ("서명 정책에 그 종목 섹터 매핑 없음 → **`ErrProductionRiskScopeRefused`**(범위 국소)",
                   f"{T} `TestARiskScopeOutsideTheSignedPolicyIsRefusedAloneInEitherOrder`", f"yes — {RED} · 변이 Y01(sentinel 제거) CAUGHT", "yes"),
            "B5": same("최악 체결가 권한 무효"), "B6": same("FX 결속 실패"), "B7": same("정책 창 불일치"), "B8": same("예약 정책 무효"),
            "B9": same("차원 순회"), "B10": same("한도 해석 실패"),
        },
        "invariants": ["범위 국소 신원은 B4 하나 — 나머지 실패는 결함으로 남긴다(넓히지 않음)."],
        "safety": ["High-risk. 판정 불변(같은 입력이 같은 자리에서 거절)."],
        "state": "순수 함수 — 상태 없음.",
    },
    {
        "bundle": "internal-riskbucket--loadproductionriskentries", "file": RB, "func": "loadProductionRiskEntries", "title": "loadProductionRiskEntries",
        "revision": "리뷰 수리(J4 = (A)): 편집 전 B6(`err != nil || scopeLatches != 0` → `scope latch present` 하나)을 B6(조회 결함 → `scope latch unreadable: %w`) "
                    "과 B7(latch 존재 → `ErrProductionRiskScopeRefused`)로 나눴다 — 둘 다 거절(판정 불변), latch 만 범위 국소 신원. 13 분기 → 14.",
        "scen": {
            "B1": same("소유자 UID 없음"), "B2": same("원장 파일 검증 실패"), "B3": same("열기 실패"), "B4": same("ping 실패"),
            "B5": same("스키마 핀 불일치(R1 — 별도 change)"),
            "B6": ("**(분리)** latch 조회 결함 → `scope latch unreadable: %w`(결함)",
                   f"{T} `TestAScopeLatchIsRefusedAloneButALatchReadFaultStops`(표 없음)", f"yes — {RED}(편집 전 범위 거절) · 변이 Y03(결함에 sentinel) CAUGHT", "yes"),
            "B7": ("**(분리)** 그 범위에 scope latch 있음 → `ErrProductionRiskScopeRefused`(범위 국소)",
                   f"{T} `TestAScopeLatchIsRefusedAloneButALatchReadFaultStops`(latch 행)", "yes — 변이 Y02(sentinel 제거) CAUGHT", "yes"),
            "B8": same("권한 창 불일치"), "B9": same("차원 순회"), "B10": same("사용량 읽기 실패(원장 결함 — 손상 행 포함)"),
            "B11": same("latch 된 사용량"), "B12": same("정책 출처 구성 실패"), "B13": same("스냅숏 출처 구성 실패"), "B14": same("권한 항목 구성 실패"),
        },
        "invariants": ["판정 무변 — 편집 전 B6 이 거절하던 입력은 B6 또는 B7 이 같은 결론(거절)으로 거절한다."],
        "safety": ["High-risk. 원장 읽기 전용. 새로 통과 · 새로 거절 0."],
        "state": "원장 읽기 전용(`mode=ro` · `query_only`), 쓰기 없음.",
    },
]

if __name__ == "__main__":
    for spec in FUNCS:
        (base.FL / spec["bundle"]).mkdir(parents=True, exist_ok=True)
    base.main(FUNCS, tag="a112 5.2.2.2 리뷰 수리", pre="lot-5.2.2.2-fix", ledger=LEDGER)
