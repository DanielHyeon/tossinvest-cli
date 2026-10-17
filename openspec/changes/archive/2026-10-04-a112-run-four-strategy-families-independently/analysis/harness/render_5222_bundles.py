#!/usr/bin/env python3
"""a112 5.2.2.2 — 본문을 바꾼 아홉 함수의 편집 뒤 번들(FLM/BTM)을 ast.json 에서 채운다.

좌표 · 호출 · return 은 AST(`go run ./tools/logic-map`, 현재 작업 트리)에서만 읽고, 저자는 분기의 뜻 · 시험 이름 · 편집 설명만 쓴다.
편집 전 번들은 `analysis/measurements/lot-5.2.2.2/pre-edit/`(HEAD e8d56d49 기준 여덟 + admit 은 base 016da624 번들 사본).
변이 원장은 `analysis/measurements/lot-5.2.2.2/mutation-5.2.2.2.tsv`(X01~X22 CAUGHT 21 · M20 SURVIVED 예상).

사용: python3 render_5222_bundles.py   (저장소 루트에서)
"""
from __future__ import annotations

import json
import subprocess
from pathlib import Path

ROOT = Path(subprocess.run(["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True, check=True).stdout.strip())
A = Path(__file__).resolve().parents[1]
FL = A / "function-logic"
E = "internal/app/engine/"
LEDGER = "`analysis/measurements/lot-5.2.2.2/mutation-5.2.2.2.tsv`"
RED = "`analysis/measurements/lot-5.2.2.2/red-5.2.2.2.log`"
T = "`a112_owner_scope_trading_test.go`"
SEAL = "`a112_first_leg_owner_scope_seal_test.go`"
HO = "`a112_owner_scope_handoff_test.go`"
CEN = "`a112_scope_refusal_census_test.go`"
SAME = ("분기 불변 — 편집 전 번들 서술(`pre-edit/`) 그대로", "no — 이 로트가 바꾸지 않음", "편집 전 번들의 측정")

FUNCS = [
    {
        "bundle": "internal-app-engine--buildproductionstrategymarketworker", "file": E + "strategy_entry_supervisor.go",
        "func": "buildProductionStrategyMarketWorker", "title": "buildProductionStrategyMarketWorker",
        "revision": "편집 전 6 분기 → 8: 편집 전 B2(준비 + handoff 승인)에서 handoff 조건을 떼어 범위 순회(B3)와 범위별 승인 · 유효(B4)로 옮기고, "
                    "편집 전 B3(봉인 깨짐)을 B4 에 합쳤다. 편집 전 B4 · B5(보호 · 진입 관문 관측 실패 → dormant)는 B5 · B6(그 범위만 건너뜀)이 되고, "
                    "승격 근거가 된 범위가 없으면 B7 이 dormant. 편집 전 B6 은 B8(불변).",
        "scen": {
            "B1": ("배선 미완/nil → dormant(편집 전 B1 불변)", "`TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure`", "no — 편집 없음", "yes"),
            "B2": ("시장 권한(일정 · 후보 · 경로 · 환율 · 위험 · 계좌) 준비 미완 → dormant. **편집: handoff 조건을 뺐다**(B3~B4 로)",
                   "`TestARefusedHandoffLeavesTheWorkerDormant`", "no — 준비 조건 불변", "yes"),
            "B3": ("**(새)** 주문 경로와 같은 handoff 목록(`dispatchHandoffs`) 순회 — 활성화 없는 시장은 시장 단위 하나(오늘), 서명 활성화 시장은 범위마다 하나",
                   f"{T} `TestAnActivatedTwoScopeMarketPromotesItsWorkerPerScope`", f"yes — 편집 전 활성화 두 범위 시장 dormant(FAIL 관측), 변이 X16 CAUGHT({LEDGER})", "yes"),
            "B4": ("**(새, 편집 전 B2 의 handoff 절반 + 편집 전 B3)** 그 handoff 가 거절했거나 봉인 깨진 제안 → 그 범위는 승격 근거가 못 됨(continue)",
                   "`TestARefusedHandoffLeavesTheWorkerDormant`(시장 단위 상한 거절 → 모든 범위 건너뜀 → B7 dormant)", "no — 동작 보존(거절 handoff 는 편집 전에도 dormant)", "yes"),
            "B5": ("보호 관측 실패 → **그 범위만** 건너뜀(편집 전 B4 는 시장 dormant — 범위 하나면 B7 로 같은 결과)",
                   "`TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure`", "no — 범위 하나에서는 동작 동일", "yes"),
            "B6": ("진입 관문 관측 실패 → **그 범위만** 건너뜀(J3 — 한 범위의 거절이 다른 범위의 승격을 굶기지 않음)",
                   f"{T} `TestAnActivatedTwoScopeMarketPromotesItsWorkerPerScope`(005930 차단 → 000660 으로 승격, 둘 다 차단 → dormant)",
                   f"yes — 변이 X15(`return dormant`) · X17(관문 무시) CAUGHT({LEDGER})", "yes"),
            "B7": ("**(새)** 승격 근거가 된 범위 없음 → dormant", f"{T} `TestAnActivatedTwoScopeMarketPromotesItsWorkerPerScope`(모든 범위 차단) · `TestARefusedHandoffLeavesTheWorkerDormant`",
                   f"yes — 변이 X17 CAUGHT", "yes"),
            "B8": ("digest / revision / 만료 → dormant(편집 전 B6 불변)",) + SAME,
        },
        "invariants": ["활성화 없는 시장(오늘 생산 전부)은 handoff 가 하나라 편집 전과 같은 판정이다(토글 OFF = upstream).",
                       "관측(보호 · 진입 관문)은 준비 확인일 뿐이고 주문마다 제출 경로가 범위 단위로 다시 검사한다(`withStrategyEntryGateAuthority`)."],
        "safety": ["High-risk 인접(승격은 화면 · 주기 가동만 움직이고 주문은 dispatch 가 낸다). 편집은 활성화 시장에서만 승격을 **넓힌다** — 한 범위라도 모든 관측을 통과해야 한다. "
                   "새로 승격되는 입력: 서명 활성화된 두 범위 시장(편집 전 OverCapacity 로 dormant). 활성화 없는 시장의 새 통과 입력 0."],
    },
    {
        "bundle": "internal-app-engine--productionstrategyfirstlegauthorityloader.collectstrategyfirstlegauthority", "file": E + "strategy_account_first_leg_authority.go",
        "func": "productionStrategyFirstLegAuthorityLoader.collectStrategyFirstLegAuthority", "title": "collectStrategyFirstLegAuthority",
        "revision": "편집 전 11 분기 → 13: 편집 전 B4(시장 단위 개수 관문 `len(proposal.entries) != 1`)를 **지우고**, identity 대조 뒤에 범위별 위험 권한 부재(B5) · "
                    "범위별 계좌 권한 부재(B6) · 계보 시장 통화 미지(B7)를 더했다. 편집 전 B5(identity)는 B4, B6~B11 은 B8~B13(조건의 권한 출처만 범위 번들로).",
        "scen": {
            "B1": ("loader · ctx · 시계 · 원장 · Guardian 부재 → 발급 불가",) + SAME,
            "B2": ("시장 권한(위험 · 환율 · 계좌 · 일정) 준비 미완 → `paired production authority is incomplete for market`(개수 조건은 6.2 에서 이미 뗌)",) + SAME,
            "B3": ("소유자 범위 선택 실패(0 또는 복수) → `…owner scope is not uniquely authorized by the assembly` — **범위 거절 타입 아님**(위조 의심 → 주기 멈춤)",
                   f"{SEAL} `TestTheFirstLegSealRefusesEveryForgeryAxis`(범위 하나 · 둘 · 조정자 순서 세 쌍 × 미선택 범위 · 타 시장)",
                   "no — 분기 불변; 두 범위 쌍 재실행은 이 로트", "yes"),
            "B4": ("봉인된 identity 대조(편집 전 B5) — 불일치는 **타입 없는 오류**(범위 거절 아님)",
                   f"{SEAL} `TestTheFirstLegSealRefusesEveryForgeryAxis`(세 쌍 × 같은 범위 패자 · 게이트된 레인 · 조건 재작성) · {T} `TestAForgedScopeStopsTheCycleBeforeTheNextValidScope`",
                   f"yes — 변이 X07(identity 불일치를 범위 거절로) CAUGHT 14({LEDGER})", "yes"),
            "B5": ("**(새)** 그 범위의 준비된 위험 권한 없음 → `*strategyScopeRefusal`(그 범위만 거절 · 봉투 폴백 없음, J3)",
                   f"{HO} `TestTwoOwnerScopesTradeOnlyWhereEachHasItsOwnAuthority`(두 순서) · {SEAL} `TestTheFirstLegSealSelectsByScopeInATwoScopePair`",
                   f"yes — {RED}(편집 전 개수 관문 거절) · 변이 X08(시장 번들 폴백) · X22(타입 없는 오류) CAUGHT", "yes"),
            "B6": ("**(새)** 그 범위의 준비된 계좌 권한 없음 → `*strategyScopeRefusal`",
                   f"{T} `TestAScopeWithoutItsOwnAccountAuthorityIsRefusedAloneAndRecorded`", "yes — 변이 X09(시장 권한 폴백) · X21(타입 없는 오류) CAUGHT", "yes"),
            "B7": ("**(새)** 계보 시장 통화 미지(KR/US 밖) → 발급 불가 — 통화는 봉투가 아니라 `Lineage.Market` 에서 유도(A#3)",
                   f"{T} `TestTheFirstLegCurrencyComesFromTheLineageNotTheEnvelope`(유도 경로; 미지 시장 갈래는 시험 seam 으로 못 만듦 — 진입 0)",
                   "yes — 변이 X10(봉투 통화) CAUGHT", "유도 경로 yes · 거절 갈래 진입 0"),
            "B8": ("위험 권한 범위 불일치(편집 전 B6) — 이제 **그 범위의** 번들로 대조",
                   f"{T} `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope`(범위마다 자기 번들로 통과)", "no — 조건 불변, 출처만 범위 번들", "yes"),
            "B9": ("포지션 캠페인 CAS 변경(편집 전 B7)",) + SAME,
            "B10": ("위험 버킷 항목 순회(편집 전 B8) — 범위 번들",) + SAME,
            "B11": ("가격 단위 무효(편집 전 B9)",) + SAME,
            "B12": ("노출 스냅숏 만료(collect 클로저, 편집 전 B10) — 범위 계좌 권한의 FreshUntil",) + SAME,
            "B13": ("예약 버전 읽기 실패(collect 클로저, 편집 전 B11)",) + SAME,
        },
        "invariants": ["범위 거절 타입(`strategyScopeRefusal`)을 만드는 자리는 B5 · B6 둘뿐이다 — 언급 census "
                       f"{CEN} `TestTheScopeRefusalTypeIsMadeOnlyWhereTheCensusSaysItIs`.",
                       "시장 단위 개수 관문이 지키던 것(관문 전수표 (a)~(e))은 review 「5.2.2.2」 절의 대체 수단 · 시험으로 옮겼다."],
        "safety": ["High-risk(1차 레그 발급). 편집은 개수 관문을 걷어 **범위 둘인 활성화 시장의 발급을 연다** — 각 범위는 자기 위험 · 계좌 권한으로만 발급되고, "
                   "권한이 없는 범위는 그 범위만 타입 거절된다. 활성화 없는 시장은 결과 · 계좌 권한이 여전히 항목 하나를 요구하므로 새 통과 입력 0.",
                   "위조 다섯 축은 범위 하나 · 둘(두 순서) 쌍에서 모두 거절(행동 시험), 오분류 · 폴백 변이 X05~X11 · X21 · X22 CAUGHT."],
    },
    {
        "bundle": "internal-app-engine--strategyaccountauthorityloader.collectmarket", "file": E + "strategy_account_first_leg_authority.go",
        "func": "strategyAccountAuthorityLoader.collectMarket", "title": "strategyAccountAuthorityLoader.collectMarket",
        "revision": "편집 전 4 분기 → 6: 편집 전 B1(항목 정확히 하나 · 유효)을 활성화 여부로 가르고(B1), 계좌 권한 적재를 **항목(범위)마다** 한다(B4 순회 · B5 범위 유효 · "
                    "B6 적재 성공). 편집 전 B4(적재 실패 → 시장 전체 AuthorityUnavailable)는 범위별 준비 안 됨으로 바뀌고, 준비된 범위가 없으면 "
                    "`strategyAccountMarketFromScopes` 가 첫 범위의 사유로 시장 전체를 준비 안 됨으로 둔다(범위 하나면 편집 전과 같은 사유).",
        "scen": {
            "B1": ("항목 없음, 또는 **활성화 없는 시장**에서 항목이 정확히 하나가 아니거나 무효 → `StrategyAccountProposalNotReady`(토글 OFF = upstream)",
                   "`TestProductionFirstLegAuthorityLoaderPairedKRUS`(시장당 하나 — 오늘 경로)", "no — 활성화 없는 시장 동작 불변", "yes"),
            "B2": ("loader 구성 불완전 → `StrategyAccountInternalFailure`",) + SAME,
            "B3": ("시장이 US 면 계좌 시장 US",) + SAME,
            "B4": ("**(새)** 항목(소유자 범위)마다 계좌 권한 적재 — 적재 종목은 `entries[0]` 이 아니라 **그 범위의 종목**(A#5)",
                   f"{T} `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope` · `TestAScopeWithoutItsOwnAccountAuthorityIsRefusedAloneAndRecorded`",
                   f"yes — 변이 X14(첫 항목 종목) CAUGHT({LEDGER})", "yes"),
            "B5": ("**(새)** 범위 키 정규화 실패 또는 무효 제안 → 그 범위만 `ProposalNotReady`", "진입 0 — 조립이 무효 제안을 항목에 싣지 않음(시험 seam 으로 못 만듦)", "no", "진입 0"),
            "B6": ("적재 성공 · 시장 · 매니페스트 일치 → 그 범위 준비(편집 전 B4 의 반대편); 실패는 그 범위만 `AuthorityUnavailable`",
                   f"{T} `TestAScopeWithoutItsOwnAccountAuthorityIsRefusedAloneAndRecorded`(005930 적재 실패 → 000660 만 거래)", "yes — 변이 X09 · X14 CAUGHT", "yes"),
        },
        "invariants": ["시장 칸(`authority` · `snapshot`)은 첫 준비된 범위의 것 — 범위 하나면 편집 전과 같은 값.",
                       "범위의 계좌 권한은 `forScope(key)` 로만 1차 레그에 건너간다(봉투 폴백 없음)."],
        "safety": ["High-risk 인접(1차 레그의 계좌 권한 출처). 활성화 없는 시장의 새 통과 입력 0. 활성화 시장에서 한 범위의 적재 실패가 시장 전체를 닫지 않게 됐다(J3) — "
                   "그 범위는 1차 레그에서 타입 거절된다."],
    },
    {
        "bundle": "internal-app-engine--strategydispatchcycle.dispatch", "file": E + "strategy_dispatch_cycle.go",
        "func": "strategyDispatchCycle.dispatch", "title": "dispatch",
        "revision": "편집 전 19 분기 → 22: 편집 전 B12(admission 거절 → 오류)안에 범위 거절 타입을 `%w` 로 싣는 B13 을 더했고, 편집 전 B14 앞의 "
                    "위험 세대 읽기를 시장 번들에서 **그 범위의 번들**로 옮겼다(B15 범위 키 · B16 범위 번들). 편집 전 B13 → B14, B14 → B17, B15~B19 → B18~B22.",
        "scen": {
            **{f"B{i}": ("편집 전과 같은 분기(좌표만)",) + SAME for i in range(1, 12)},
            "B12": ("admission 거절 → 주기 오류(편집 전 B12)", f"{T} `TestAForgedScopeStopsTheCycleBeforeTheNextValidScope`", "no — 조건 불변", "yes"),
            "B13": ("**(새)** 거절이 범위 거절이면 타입을 `%w` 로 싣는다 — 문구는 같음(J4 ①: 분류는 타입으로)",
                    f"{HO} `TestTwoOwnerScopesTradeOnlyWhereEachHasItsOwnAuthority` · {T} `TestAScopeWithoutItsOwnAccountAuthorityIsRefusedAloneAndRecorded`",
                    f"yes — 변이 X06(모든 거절에 타입) CAUGHT({LEDGER})", "yes"),
            "B14": ("Guardian 결정 세대 없음(편집 전 B13)",) + SAME,
            "B15": ("**(새)** 계보의 범위 키 정규화", f"{T} `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope`", "no — 경로 추가", "yes"),
            "B16": ("**(새)** 그 범위의 준비된 위험 번들에서 세대를 읽음(범위 번들이 없으면 세대 0 → B17 거절)",
                    f"{T} `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope`",
                    "M20(시장 번들 세대) SURVIVED — **예상**: 모든 범위 번들이 같은 서명 시장 매니페스트에서 와서 세대가 구조상 시장 단위다", "yes"),
            "B17": ("서명 위험 정책 세대 없음(편집 전 B14)",) + SAME,
            **{f"B{i}": ("편집 전과 같은 분기(좌표만, 편집 전 B" + str(i - 3) + ")",) + SAME for i in range(18, 23)},
        },
        "invariants": ["dispatch 는 범위 거절을 만들지 않는다 — admit 이 1차 레그 수집 오류에서 꺼낸 값을 나를 뿐(census)."],
        "safety": ["High-risk(주문 경로). 새로 통과하는 입력 0 — 추가는 오류에 타입을 싣는 것과 세대 출처를 범위 번들로 좁힌 것뿐."],
    },
    {
        "bundle": "internal-app-engine--strategyfirstlegadmissionbridge.admit", "file": E + "strategy_first_leg_admission.go",
        "func": "strategyFirstLegAdmissionBridge.admit", "title": "admit",
        "revision": "편집 전 6 분기 → 7: 편집 전 B3(권한 수집 실패 → AuthorityCollectionFailed) 안에 범위 거절 타입을 결과에 싣는 B4 를 더했다. "
                    "편집 전 B4~B6 → B5~B7. 편집 전 번들은 base 016da624 기준(`pre-edit/internal-app-engine--strategyfirstlegadmissionbridge.admit`).",
        "scen": {
            "B1": ("결과 검증 거절",) + SAME,
            "B2": ("bridge · loader · Guardian 부재",) + SAME,
            "B3": ("1차 레그 권한 수집 실패 → `AuthorityCollectionFailed`(문구 = 수집 오류)", f"{T} `TestAForgedScopeStopsTheCycleBeforeTheNextValidScope`", "no — 조건 불변", "yes"),
            "B4": ("**(새)** 수집 오류가 `*strategyScopeRefusal` 이면(errors.As — 타입) 결과에 싣는다; 그 밖의 수집 오류는 싣지 않음(J4)",
                   f"{T} `TestAScopeWithoutItsOwnAccountAuthorityIsRefusedAloneAndRecorded` · `TestAForgedScopeStopsTheCycleBeforeTheNextValidScope` · {CEN} census",
                   f"yes — 변이 X05(모든 수집 오류에 타입) CAUGHT({LEDGER})", "yes"),
            "B5": ("권한 불일치(편집 전 B4)",) + SAME,
            "B6": ("Guardian precheck 실패(편집 전 B5)",) + SAME,
            "B7": ("원자 admission 실패(편집 전 B6) — 원장 `BUCKET_USAGE_STALE` 은 여기서 **타입 없이** 올라감",
                   f"{T} `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope`(같은 파도 둘째 범위 — 타입 아님 단언)", "no — 조건 불변", "yes"),
        },
        "invariants": ["범위 거절 타입은 수집 오류에서 **꺼내기만** 한다 — 만들지 않는다(census)."],
        "safety": ["High-risk(admission). 새로 통과하는 입력 0 — 결과에 분류 값 하나를 더할 뿐, 거절 코드 · 문구는 불변."],
    },
    {
        "bundle": "internal-app-engine--strategyprojectionfromassembly", "file": E + "strategy_runtime_projection.go",
        "func": "strategyProjectionFromAssembly", "title": "strategyProjectionFromAssembly",
        "revision": "편집 전 8 분기 → 10: 편집 전 B7(시장 단위 handoff 거절 또는 봉인 깨짐 → EvidenceStale)을 handoff 목록 순회(B7) · 첫 승인 유효 범위 선택(B8) · "
                    "없음(B9)으로 나눴다. 편집 전 B8 → B10.",
        "scen": {
            **{f"B{i}": ("편집 전과 같은 분기",) + SAME for i in range(1, 7)},
            "B7": ("**(새)** 주문 경로와 같은 handoff 목록 순회", f"{T} `TestAnActivatedTwoScopeMarketPromotesItsWorkerPerScope`(projection 단언)",
                   f"yes — 변이 X19(시장 단위 handoff) CAUGHT({LEDGER})", "yes"),
            "B8": ("**(새)** 조정자 순서의 첫 승인 · 유효 범위를 보임", f"{T} `TestAnActivatedTwoScopeMarketPromotesItsWorkerPerScope`", "yes — X19", "yes"),
            "B9": ("승인 범위 없음 → EvidenceStale(편집 전 B7 의 결과)", "진입 0 — 이 로트의 시험 없음(편집 전에도 진입 0)", "no", "진입 0"),
            "B10": ("레인 증거 다이제스트 부재(편집 전 B8)",) + SAME,
        },
        "invariants": ["읽기 전용 화면 — 주문 · 토글 · 원장 쓰기 없음. 두 범위 시장은 첫 범위만 보인다(범위별 행은 review 잔여)."],
        "safety": ["High-risk 아님(화면). 활성화 없는 시장은 handoff 하나라 편집 전과 같은 값."],
    },
    {
        "bundle": "internal-app-engine--strategyproposalauthorityloader.collectmarket", "file": E + "strategy_proposal_authority.go",
        "func": "strategyProposalAuthorityLoader.collectMarket", "title": "strategyProposalAuthorityLoader.collectMarket",
        "revision": "편집 전 16 분기 → 15: 제안 집합 digest 를 손으로 적던 순회(편집 전 B16 `range entries`)를 지우고 A-lite 계약과 같은 함수 "
                    "`strategyProposalSetDigest(entries)` 를 부른다(A#6 — digest 식 단일 출처). B1~B15 는 편집 전과 같은 분기 · 좌표 이동만.",
        "scen": {
            **{f"B{i}": ("편집 전과 같은 분기(좌표만)",) + SAME for i in range(1, 16)},
        },
        "invariants": ["제안 집합 digest 의 식은 `strategyProposalSetDigest` 하나다 — 조립과 dispatch 대조가 같은 함수를 부른다."],
        "safety": ["동작 불변 리팩터(같은 식) — `TestTheProposalSetDigestMatchesWhatTheAssemblyRecords` 가 실제 중재 경로로 잰다; 변이 X18(조립 쪽만 바뀜) CAUGHT."],
    },
    {
        "bundle": "internal-app-engine--strategyproposalauthoritypair.resultauthority", "file": E + "strategy_proposal_authority.go",
        "func": "strategyProposalAuthorityPair.ResultAuthority", "title": "strategyProposalAuthorityPair.ResultAuthority",
        "revision": "편집 전 1 분기 → 4: 시장 단위 handoff 하나(`dispatchHandoff().Single()`) 대신 주문 경로와 같은 목록(`dispatchHandoffs`)을 순회하고(B1), "
                    "하나라도 거절 · 무효면 시장 준비 안 됨(B2 — 편집 전 B1), 목록이 비면 준비 안 됨(B3), 범위가 둘 이상이거나 활성화 시장이면 범위별 결과를 싣는다(B4).",
        "scen": {
            "B1": ("**(새)** handoff 목록 순회", f"{T} `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope`", "yes — 변이 X13 CAUGHT", "yes"),
            "B2": ("handoff 거절 또는 무효 제안 → 시장 결과 권한 준비 안 됨(편집 전 B1 — 목록의 일부만 넘기지 않음)",
                   "`TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`", "no — 동작 보존", "yes"),
            "B3": ("**(새)** 목록 없음 → 준비 안 됨", "진입 0 — dispatchHandoffs 는 항상 하나 이상을 돌려준다(거절 handoff 포함)", "no", "진입 0"),
            "B4": ("**(새)** 범위 둘 이상 또는 활성화 시장 → 범위별 결과(`scoped`)를 싣는다 — 위험 적재기가 범위마다 번들을 만든다",
                   f"{T} `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope`", f"yes — 변이 X13(`if false`) CAUGHT({LEDGER})", "yes"),
        },
        "invariants": ["결과 권한이 보는 범위 집합 = 주문 경로가 받는 handoff 집합(같은 함수)."],
        "safety": ["High-risk 인접(위험 권한 입력). 활성화 없는 시장은 handoff 하나 — `scoped` 없음 → 위험 적재기가 편집 전과 같은 결과 하나를 본다."],
    },
    {
        "bundle": "internal-app-engine--strategyriskauthorityloader.collectmarket", "file": E + "strategy_risk_authority.go",
        "func": "strategyRiskAuthorityLoader.collectMarket", "title": "strategyRiskAuthorityLoader.collectMarket",
        "revision": "편집 전 5 분기 → 6: 번들 하나 적재(편집 전 B4 적재 실패 · B5 범위 불일치 → 시장 AuthorityUnavailable)를 결과 권한의 범위마다(B4 순회 · B5 키 · "
                    "B6 적재 성공)로 바꿨다. 준비된 범위가 없으면 `strategyRiskMarketFromScopes` 가 시장 전체를 AuthorityUnavailable 로 둔다(편집 전 사유 그대로).",
        "scen": {
            "B1": ("결과 권한 준비 안 됨 → LaneNotReady",) + SAME,
            "B2": ("환율 준비 안 됨 → FXNotReady",) + SAME,
            "B3": ("시장이 US 면 버킷 시장 US",) + SAME,
            "B4": ("**(새)** 결과 권한의 범위마다 번들 하나(`result.results()` — 활성화 없는 시장은 하나)",
                   f"{T} `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope`(범위 번들 digest 둘)", f"yes — 변이 X12(첫 범위만) CAUGHT({LEDGER})", "yes"),
            "B5": ("**(새)** 범위 키 정규화 성공 시에만 적재 — 실패면 그 범위 AuthorityUnavailable", "진입 0(거짓 갈래) — 조립 결과는 항상 유효 키", "no", "진입 0"),
            "B6": ("적재 성공 · 시장 · 계좌 · 시각 · 항목 5 일치 → 그 범위 준비(편집 전 B4 · B5 의 반대편)",
                   f"{T} `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope` · `TestTheRiskStubBridgeIsStillNeededBecauseTheLoaderRefusesTheRealJournal`(**a127 `82080177` 에서 제거됨** — 다리 제거 조건 이행, 대체 양성 시험 `TestTheRiskLoaderReadsTheRealJournal`: 실원장 → 범위 준비)(실원장 → 전 범위 준비 안 됨)",
                   "yes — X12", "yes"),
        },
        "invariants": ["범위 번들은 모두 같은 서명 시장 매니페스트에서 온다(세대 · 정책은 시장 단위, 버킷 사용량 스냅숏은 적재 시점의 원장)."],
        "safety": ["High-risk(위험 권한). 한 범위 적재 실패는 그 범위만 준비 안 됨(J3). 오늘 생산에서는 스키마 핀 27 결함으로 전 범위가 준비 안 됨 — ROADMAP a112 이월."],
    },
]

# 시험 파일의 도우미 · 시험(비례 원칙 — High-risk 아님: 생산 동작 없음). 본문 편집은 한두 줄이고, 번들은 게이트가 요구하는 모양만 채운다.
TEST_NOTE = "시험 코드 — 생산 동작 없음(비례 원칙: FLM 의무의 무거운 규율 대상 아님, 게이트 모양만 채움)"
FUNCS += [
    {
        "bundle": "internal-app-engine--testproductionfirstlegauthorityloaderpairedkrus", "file": E + "strategy_account_first_leg_authority_test.go",
        "func": "TestProductionFirstLegAuthorityLoaderPairedKRUS", "title": "TestProductionFirstLegAuthorityLoaderPairedKRUS",
        "revision": "a112 5.2.2.2: 손으로 만든 계좌 권한에 범위 목록을 싣는 한 줄(`a112ScopedAccount`)만 더했다 — 1차 레그 권한이 이제 범위별 계좌 권한을 "
                    "`forScope` 로 고르므로(봉투 폴백 없음) 범위 없는 fixture 는 거절된다. 분기 불변.",
        "scen": {f"B{i}": ("시험 본문 분기(편집 불변)", "이 시험 자신", "no — 시험 코드", "yes") for i in range(1, 12)},
        "invariants": [TEST_NOTE],
        "safety": ["High-risk 아님. 편집은 fixture 가 생산 모양(범위별 계좌 권한)을 갖게 할 뿐 — 단언 불변."],
        "state": "시험 fixture 만 바꾼다.",
    },
    {
        "bundle": "internal-app-engine--strategydispatchgatewayspy.observestrategyentrygate", "file": E + "strategy_dispatch_cycle_test.go",
        "func": "strategyDispatchGatewaySpy.ObserveStrategyEntryGate", "title": "strategyDispatchGatewaySpy.ObserveStrategyEntryGate",
        "revision": "a112 5.2.2.2: 종목 단위 진입 관문 거절(`failEntryGateSymbol`)을 더했다(B2) — worker 승격이 한 범위의 관문 거절로 다른 범위를 굶기지 "
                    "않는지 재려고. 시장 단위 거절(B1)은 불변.",
        "scen": {
            "B1": ("시장 단위 거절(편집 불변)", "`TestProductionStrategyWorkersPromoteKRUSInSameWaveAndIsolateProtectionFailure` 등 기존 사용처", "no — 시험 코드", "yes"),
            "B2": ("**(새)** 종목 단위 거절", f"{T} `TestAnActivatedTwoScopeMarketPromotesItsWorkerPerScope`", "no — 시험 코드", "yes"),
        },
        "invariants": [TEST_NOTE],
        "safety": ["High-risk 아님(시험 스파이). 실주문 없음."],
        "state": "스파이의 관측 횟수 map 만 쓴다.",
    },
    {
        "bundle": "internal-app-engine--newstrategyriskloaderfixture", "file": E + "strategy_risk_authority_test.go",
        "func": "newStrategyRiskLoaderFixture", "title": "newStrategyRiskLoaderFixture",
        "revision": "a112 5.2.2.2: 본문을 `newStrategyRiskLoaderFixtureWith(t, nil)` 위임 한 줄로 바꿨다 — KR 서명 위험 정책에 종목을 더할 수 있는 판을 "
                    "새 함수로 떼고, 기존 호출자는 같은 fixture 를 받는다(extraKR 이 nil 이면 편집 전과 같은 매니페스트). 분기 없음.",
        "scen": {},
        "happy": ("`newStrategyRiskLoaderFixtureWith(t, nil)` 위임", "`TestStrategyRiskAuthorityLoaderPairedKRUSSameWave` · `TestStrategyRiskAuthorityLoaderPreservesPeerOnMarketFailure` 등 기존 호출자 전부(엔진 태그 스위트 PASS)"),
        "invariants": [TEST_NOTE],
        "safety": ["High-risk 아님. 기존 호출자가 받는 fixture 는 바이트 단위로 같은 매니페스트(append 할 것이 없음)."],
        "state": "시험 임시 디렉터리에 stub 원장 · 서명 매니페스트를 쓴다(변경 없음).",
    },
]


def ast_for(file: str, func: str) -> dict:
    out = subprocess.run(["go", "run", "./tools/logic-map", "--file", file, "--func", func], cwd=ROOT, capture_output=True, text=True, check=True,
                         env={**__import__("os").environ, "GOFLAGS": "-trimpath"}).stdout
    return json.loads(out), out


def main(funcs=None, tag: str = "a112 5.2.2.2", pre: str = "lot-5.2.2.2", ledger: str = LEDGER) -> None:
    for spec in FUNCS if funcs is None else funcs:
        bundle = FL / spec["bundle"]
        ast, raw = ast_for(spec["file"], spec["func"])
        pos = lambda n: f"{n['at']['line']}:{n['at']['column']}"
        ast["branches"] = ast.get("branches") or []
        ast["returns"] = ast.get("returns") or []
        ast["calls"] = ast.get("calls") or []
        ids = [b["id"] for b in ast["branches"]]
        if set(ids) != set(spec["scen"]):
            raise SystemExit(f"{spec['bundle']}: AST branches {ids} != authored {sorted(spec['scen'])}")
        (bundle / "ast.json").write_text(raw)
        btm = [f"# Branch Test Map: `{spec['title']}`", "",
               f"- Source SHA-256: `{ast['source_sha256']}`; AST branch locations are authoritative.",
               f"- Revision: **modified ({tag}, 2026-10-01).** {spec['revision']}",
               f"- 편집 전 번들: `analysis/measurements/{pre}/pre-edit/{spec['bundle']}/`. 변이 원장 {ledger}.", "",
               "| Branch | Scenario anchor | Test | RED observed | GREEN observed |", "|---|---|---|---|---|"]
        for b in ast["branches"]:
            scenario, test, red, green = spec["scen"][b["id"]]
            btm.append(f"| {b['id']} | {b['kind']} at {pos(b)} — {scenario} | {test} | {red} | {green} |")
        if not ast["branches"]:  # 분기 없는 함수 — 게이트가 요구하는 happy-path 한 행
            btm.append(f"| B1 | happy path (분기 없음) — {spec['happy'][0]} | {spec['happy'][1]} | no — 시험 코드 | yes |")
        (bundle / "branch-test-map.md").write_text("\n".join(btm) + "\n")
        flm = [f"# Function Logic Map: `{spec['title']}`", "",
               f"- Source: `{ast['file']}`", f"- Source SHA-256: `{ast['source_sha256']}`",
               f"- Signature: `{ast['signature']}`", f"- Source range: `{ast['start']['line']}:1`–`{ast['end']['line']}:2`",
               f"- AST evidence: `ast.json` — **편집 뒤**({tag}).", "- Risk scan: `risk-pattern-report.md`.", "",
               "## Inputs and invariants", ""] + [f"- {line}" for line in spec["invariants"]] + [
               "", "## Branches and early returns", "",
               "- Exact AST return nodes: `" + ", ".join(pos(r) for r in ast["returns"]) + "`.", "",
               "| Branch | AST kind | Source location | Meaning |", "|---|---|---|---|"]
        flm += [f"| {b['id']} | {b['kind']} | {pos(b)} | {spec['scen'][b['id']][0]} |" for b in ast["branches"]]
        flm += ["", "## Calls and live bindings", "", "| Callee expression | Position |", "|---|---|"]
        flm += [f"| `{c.get('text', '(unnamed)')}` | {pos(c)} |" for c in ast["calls"]]
        flm += ["", "## State mutations and fallbacks", "", f"- {spec.get('state', '이 함수 자신은 원장을 쓰지 않는다. 봉투 · 시장 값으로의 폴백 없음(범위 권한이 없으면 거절 또는 준비 안 됨).')}"]
        flm += ["", "## Safety conclusion", ""] + [f"- {line}" for line in spec["safety"]]
        (bundle / "function-logic-map.md").write_text("\n".join(flm) + "\n")
        subprocess.run(["python3", "tools/logic-map/risk_pattern_report.py", spec["file"], "--output", str(bundle / "risk-pattern-report.md")],
                       cwd=ROOT, check=True, capture_output=True)
        print("rendered", spec["bundle"], len(ids), "branches")


if __name__ == "__main__":
    main()
