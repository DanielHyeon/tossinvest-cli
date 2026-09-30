#!/usr/bin/env python3
"""a094 단위 ② · ③ 번들의 FLM/BTM 을 쓴다 — 분기 표는 flm_tables.py 가 ast.json · 소스 · 커버리지에서 만들고, 이 파일은
Test 열과 산문만 준다(2026-09-30).

사용: python3 write_unit2_bundles.py <coverage-dir>   (cov_engine.out · cov_journal.out · cov_execgw.out · cov_reconcile.out)
"""
from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
FL = HERE.parent / "function-logic"
TABLES = HERE / "flm_tables.py"

E = "TestA094"
BUNDLES: dict[str, dict] = {
    "internal-app-engine--exitobserver.clearthesymbol": {
        "pkg": "engine",
        "role": "청산과 충돌할 미체결 주문을 치우고, 보호 청산을 제출해도 되는지(와 왜 아닌지)를 돌려준다.",
        "inputs": [
            ("`live`", "원장의 미체결 목록(엔진 귀속 주문만)", "`Journal.LiveOrdersForSymbol`", "읽기 실패는 오류 반환(B1, 종전)"),
            ("`unsettled`", "같은 종목 미종결 attempt", "`Submit.UnsettledOnSymbol` = `Gateway.unsettledFor`(주문 경로와 같은 판정)", "읽기 실패는 오류 반환(B2)"),
            ("`withPending`", "`CancelPendingFirst`", "`record`", "B5 · B14"),
            ("`m.state.PendingIntentID`", "무장된 발의의 intent", "`exit_states`", "비면 해제하지 않음(B15)"),
        ],
        "calls": "`LiveOrdersForSymbol` · `Submit.UnsettledOnSymbol` · `Journal.ConfirmedCancelOf`(매도마다) · `Issuer.IssueReduction` · `floatOf` · "
                 "`workingOrderPrice`(빈 가격 0) · `Submit.Cancel`(유일한 브로커 mutation — 취소만) · `Journal.ReleaseClearedExitProposal`. "
                 "새 브로커 **조회**는 0 — 추가된 호출은 전부 원장 읽기다.",
        "mut": "브로커 취소(B11 앞) · 발의 해제(B16 뒤 — 원장 판정이 허락할 때만). 신규 · 정정 주문 없음.",
        "safety": "a094 D−4.3 · D−4.7 · D−3.2. 제출을 여는 `cleared=true` 는 (1) 같은 종목 미종결 0, (2) 치운 주문이 전부 매수이거나 치울 것이 없고, "
                  "(3) 발의가 없거나 원장 판정이 해제를 허락했을 때만이다. 매도 취소 접수 · 확정 취소 대기 매도 · park/모호/접수 대기 발의는 전부 미완료. "
                  "High-risk: yes — 그 판정이 손절 제출을 연다/막는다.",
        "tests": {
            "B1": ["TestAWorkingEntryIsCancelledBeforeTheLiquidation"], "B2": [E + "AnotherIntentInFlightStopsTheArmReleaseLoop"],
            "B3": [E + "AnotherIntentInFlightStopsTheArmReleaseLoop", E + "AnInFlightEngineCancelIsNotCountedButAnInDoubtOneIs"],
            "B4": ["TestAWorkingEntryIsCancelledBeforeTheLiquidation"], "B5": ["TestABreachDisplacesAnOutstandingTakeProfit"],
            "B6": [E + "ACancelledSellIsNotClearedUntilItsClosingRecord"], "B7": [E + "ACancelledSellIsNotClearedUntilItsClosingRecord"],
            "B8": [E + "ACancelledSellIsNotClearedUntilItsClosingRecord"], "B9": [E + "ConsecutiveClearFailuresRaiseAnEarlierAlert"],
            "B10": [E + "AnEmptyPriceIsNotAClearFailure"], "B11": ["TestAnUncancellableEntryWithholdsTheLiquidationAndAlertsPastTheBound"],
            "B12": [E + "ACancelledSellIsNotClearedUntilItsClosingRecord", "TestABreachDisplacesAnOutstandingTakeProfit"],
            "B13": [E + "AClosingRecordThatNeverComesHoldsAndAlerts"], "B14": [E + "AParkedTakeProfitIsNotClearedAndIsNamed"],
            "B15": [E + "AParkedTakeProfitIsNotClearedAndIsNamed"], "B16": [E + "AParkedTakeProfitIsNotClearedAndIsNamed"],
            "B17": [E + "AParkedTakeProfitIsNotClearedAndIsNamed"],
        },
    },
    "internal-app-engine--exitobserver.judge": {
        "pkg": "engine",
        "role": "포지션 하나를 판정한다. a094 는 진입에 `noteHeldProposal`(park 원인 · 종결 증거 대기 명명 critical) 한 호출을 더했다.",
        "inputs": [("`m`", "관리 포지션", "working set", "identityErr 면 거부 알림(B2)"), ("`quote`", "관측 가격", "observe", "쓸 수 없으면 반환(B1)")],
        "calls": "`quoteUsable` · **`noteHeldProposal`(원장 읽기 + RecordCritical 창 0, 반환값 없음 — 결과를 바꾸지 않음)** · `alertRefused` · "
                 "`StampExitSnapshotQuarantineSelector` · `breakEven` · `judgeLadder`/`judgeRatchet`.",
        "mut": "a094 몫 없음(알림 기록뿐). 기존: 재판정 도장(B4).",
        "safety": "noteHeldProposal 은 B1 뒤 · B2 앞(억제 · 조기 반환 앞)에 둬서 손절 자신이 무장 발의일 때도 닿는다(D−4.4 · R10). 오류를 반환하지 않으므로 "
                  "판정 순서 · 결과 무변화(3.R2a). High-risk: yes(판정 진입) — 편집은 호출 한 줄.",
        "tests": {"B1": [E + "TheParkCheckChangesNoOutcome"], "B2": ["TestARungTableSwappedUnderALivePositionIsRefused"],
                  "B3": ["TestCrossingTheFirstTakeProfitKeepsThePositionUnderJudgement"], "B4": ["TestCrossingTheFirstTakeProfitKeepsThePositionUnderJudgement"],
                  "B5": [E + "TheParkCheckChangesNoOutcome"], "B6": [E + "AParkedStopIsNamedOnceAndStaysSuppressed"],
                  "B7": [E + "AnUnacceptedStopIsReleasedAndProposedAgain"], "B8": [E + "AParkedStopIsNamedOnceAndStaysSuppressed"]},
    },
    "internal-app-engine--exitobserver.submit": {
        "pkg": "engine",
        "role": "무장된 발의를 브로커로 보낸다. a094: 해제는 intent 를 넘겨 원장 판정이 하고(4.3c), attempt 가 기록됐는데 비수용 종결이 아니면 무장 유지.",
        "inputs": [("`intentID`", "무장된 발의의 intent", "record", "해제 판정의 기대 intent"), ("`out`", "게이트웨이 결과", "`Submit.Place`", "갈래 B7~B12")],
        "calls": "`applyFloor` · `IssueReduction` · `AttachExitIntent` · `sellIntent` · `Submit.Place` · `release`(= `ReleaseUnacceptedExitProposal`) · `noteDelay` · `alertProposalRefused`.",
        "mut": "발의 해제(release — 원장 판정이 허락할 때만). 주문 1(Place).",
        "safety": "B10(a094): `AttemptID != \"\"` 이고 상태가 NOT_DISPATCHED · FAILED_CONFIRMED 가 아니면 무장 유지 + 오류 — 결과를 못 쓴 제출 위에 두 번째 매도를 얹지 않는다(D−2.5). "
                  "B12 default 는 R1 의 FAILED_CONFIRMED 가 가는 곳(2.9). High-risk: yes.",
        "tests": {"B1": ["TestAFloorThatCannotBeComputedSellsNothing"], "B2": ["TestAZeroFloorSubmitsNothingAndLeavesTheLevelProposable"],
                  "B3": [E + "AnUnacceptedStopIsReleasedAndProposedAgain"], "B4": ["TestABaselineBreachProposesTheWholePosition"],
                  "B5": ["TestABaselineBreachProposesTheWholePosition"], "B6": ["TestABaselineBreachProposesTheWholePosition"],
                  "B7": ["TestABaselineBreachProposesTheWholePosition"], "B8": ["TestAnInDoubtSubmissionKeepsTheProposalArmed"],
                  "B9": ["TestABaselineBreachProposesTheWholePosition"], "B10": [E + "AnUnrecordedOutcomeKeepsTheProposalArmed"],
                  "B11": [E + "AnUnrecordedOutcomeKeepsTheProposalArmed"], "B12": [E + "AnUnacceptedStopIsReleasedAndProposedAgain", "TestARefusedProposalReleasesTheLevelAndAlerts"],
                  "B13": ["TestARefusedProposalReleasesTheLevelAndAlerts"]},
    },
    "internal-app-engine--exitobserver.release": {
        "pkg": "engine",
        "role": "제출하지 못했거나 비수용으로 끝난 발의를 푼다 — 판정은 원장 한 곳(`ReleaseUnacceptedExitProposal`).",
        "inputs": [("`intentID`", "무장된 발의의 intent", "submit", "다르면 원장이 아무것도 안 바꿈")],
        "calls": "`Journal.ReleaseUnacceptedExitProposal`(판정 읽기 + 해제 쓰기 한 트랜잭션).",
        "mut": "발의 해제 · exit_events 한 행(판정이 허락할 때).",
        "safety": "해제 여부를 호출자가 정하지 않는다 — intent 의 attempt 가 전부 비수용이거나 없을 때만. High-risk: yes.",
        "tests": {"B1": [E + "AnUnacceptedStopIsReleasedAndProposedAgain", "TestARefusedProposalReleasesTheLevelAndAlerts"]},
    },
    "internal-app-engine--exitobserver.record": {
        "pkg": "engine",
        "role": "판정 결과를 원장에 기록하고, 제출 가능하면 발의를 무장해 submit 에 넘긴다. a094: 청소 결과(clearResult)를 받아 지연 사유 · 연속 실패 · 연속 종료를 처리.",
        "inputs": [("`cleared`", "청소 결과(cleared · countable · awaitingClose)", "clearTheSymbol", "미완료면 무장 안 함(B9)")],
        "calls": "`clearTheSymbol` · `noteDelay`(사유 = `clearResult.why()` — 형태 B 면 복구 절차 문장) · **`noteClearFailure`** · `clearDelay` · **`endClearStreak`** · "
                 "`RecordExitJudgementResult` · `submit`.",
        "mut": "판정 기록 · 발의 무장(원장) · 관측자 연속 상태(메모리).",
        "safety": "B9: 미완료면 무장 · 제출 없음(종전). 연속 실패 경보는 기존 지연 타이머(delayedSince · delayAlerted)를 건드리지 않는 별개 key(D−2.7). High-risk: yes.",
        "tests": {"B1": [E + "TheParkCheckChangesNoOutcome"], "B2": ["TestABaselineBreachProposesTheWholePosition"],
                  "B3": ["TestABaselineBreachProposesTheWholePosition"], "B4": ["TestABaselineBreachProposesTheWholePosition"],
                  "B5": ["TestAWorkingEntryIsCancelledBeforeTheLiquidation"], "B6": ["TestCrossingTheFirstTakeProfitKeepsThePositionUnderJudgement"],
                  "B7": ["TestAWorkingEntryIsCancelledBeforeTheLiquidation"], "B8": [E + "AnotherIntentInFlightStopsTheArmReleaseLoop"],
                  "B9": [E + "ConsecutiveClearFailuresRaiseAnEarlierAlert", E + "AClosingRecordThatNeverComesHoldsAndAlerts"],
                  "B10": ["TestAWorkingEntryIsCancelledBeforeTheLiquidation"], "B11": ["TestABaselineBreachProposesTheWholePosition"],
                  "B12": ["TestABaselineBreachProposesTheWholePosition"], "B13": ["TestAnUnresolvedProposalSuppressesTheNextOne"],
                  "B14": ["TestAnUnresolvedProposalSuppressesTheNextOne"], "B15": ["TestAnUnresolvedProposalSuppressesTheNextOne"],
                  "B16": ["TestABaselineBreachProposesTheWholePosition"]},
    },
    "internal-app-engine--context.exitobserver": {
        "pkg": "engine",
        "role": "exit 관측 루프를 조립한다. a094: `opts.Critical = c.Notifier`(새 critical 의 단일 입구, 호출자 값과 무관하게 덮음).",
        "inputs": [("`c.Notifier`", "알림기", "Context", "nil 이면 Critical 도 nil(경고 로그만)")],
        "calls": "`obs.RecordOnly` · `exitSideRetrier` · `exitSideFloor` · `NewExitObserver`.",
        "mut": "없음(조립).",
        "safety": "B5 안에서 Alerts · Announcer · Critical 셋을 같은 알림기로 덮는다. High-risk: no(배선) — 단 배선 누락은 알림 침묵이므로 시험으로 고정.",
        "tests": {"B1": ["TestA092ExitObserverGetsRecordOnlyAlertPaths"], "B2": ["TestA092ExitObserverGetsRecordOnlyAlertPaths"],
                  "B3": ["TestA092ExitObserverGetsRecordOnlyAlertPaths"], "B4": ["TestA092ExitObserverGetsRecordOnlyAlertPaths"],
                  "B5": ["TestA094TheExitObserverRecordsCriticalsThroughTheNotifier", "TestA092ExitObserverGetsRecordOnlyAlertPaths"], "B6": ["TestA092ExitObserverGetsRecordOnlyAlertPaths"]},
    },
    "internal-execgw--gateway.checksymbolfree": {
        "pkg": "execgw",
        "role": "같은 종목의 미종결 · park attempt 가 이 mutation 을 막는지. a094: 미종결 판정을 `unsettledFor` 로 추출(exit 청소와 공유).",
        "inputs": [("`plan`", "시장 · 종목 · 노출 증가 여부", "Gateway.submit", "")],
        "calls": "`unsettledFor`(PendingAttempts + attemptTargets) · `UnresolvedAttempts` · `attemptTargets`.",
        "mut": "없음(판정).",
        "safety": "옛 판본은 첫 일치에서 거절했고 새 판본은 끝까지 모은 뒤 첫 것을 이름으로 거절한다 — 뒤 행의 intent 읽기 실패가 거절 대신 오류가 될 수 있다(둘 다 발송 안 함). High-risk: yes.",
        "tests": {"B1": ["TestTransportOutcomeTable"], "B2": [E + "AnotherIntentInFlightStopsTheArmReleaseLoop", "TestGatewayRefusesFailClosedBranchesBeforeDispatch"],
                  "B3": ["TestTransportOutcomeTable"], "B4": ["TestTransportOutcomeTable"], "B5": ["TestTransportOutcomeTable"],
                  "B6": ["TestTransportOutcomeTable"], "B7": ["TestTransportOutcomeTable"]},
    },
    "internal-execgw--gateway.confirmcreatedorder": {
        "pkg": "execgw",
        "role": "발주 직후 확인 — 이제 `ConfirmPlacedOrder`(기동 ACKED 확정과 공유) 한 호출. 응답 종목이 비면 확인 실패(D−5.1).",
        "inputs": [("`brokerOrderID`", "브로커가 준 번호", "dispatch", "바이트 일치 요구")],
        "calls": "`ConfirmPlacedOrder`(OrderRaw · roundTripTimeout · parseOrderFacts · 번호 바이트 일치 · 종목 비공백 일치).",
        "mut": "없음.",
        "safety": "보수 방향 강화(종목 공백 = 실패 → IN_DOUBT). High-risk: yes(접수 확정).",
        "tests": {"B1": ["TestA094TheReadBackNeedsANamedSymbol"]},
    },
    "internal-journal--journal.resolveexitproposal": {
        "pkg": "journal",
        "role": "판정 없이 발의를 비우는 쓰기 — 이제 기대 intent 필수, 쓰기는 `clearExitProposalTx` 한 곳.",
        "inputs": [("`expectedIntentID`", "무장된 발의의 intent", "호출자", "비면 ErrInvalidRequest(B1), 다르면 무변화")],
        "calls": "`proposalResolutionAction` · `clearExitProposalTx`(기대 intent 대조 · rung 되돌림 · exit_events).",
        "mut": "exit_states 발의 컬럼 NULL · exit_events 한 행 · (사다리) active_rung.",
        "safety": "늦게 도착한 해제가 다른 발의를 지우지 않는다(4.3a). 손절 가격 컬럼 무접촉(4.5). High-risk: yes.",
        "tests": {"B1": [E + "ALateReleaseDoesNotClearAnotherProposal"], "B2": ["TestResolvingNothingIsNotAnError"],
                  "B3": ["TestResolvingNothingIsNotAnError"], "B4": ["TestResolvingNothingIsNotAnError", E + "ALateReleaseDoesNotClearAnotherProposal"],
                  "B5": ["TestARefusalReArmsTheLevel", "TestACancelledRungIsProposableAgain"]},
    },
    "internal-journal--journal.liveordersforsymbol": {
        "pkg": "journal",
        "role": "엔진 귀속 미체결 주문 목록. a094: 종결 증거 술어를 공유 상수(`confirmedOrderTerminalEvidence`)로 — SQL 바이트는 같은 술어.",
        "inputs": [("`accountRef` · `market` · `symbol`", "범위", "호출자", "")],
        "calls": "`guardTrackedFillIdentity`(같은 범위 소유 모호 = 오류) · 질의 · `ResolveCurrentOrderIDScoped`.",
        "mut": "없음.",
        "safety": "동작 무변(술어 문자열을 상수로 옮김). 해제 판정과 같은 술어를 쓰게 해 판정을 둘로 두지 않는다. High-risk: yes(청소의 목록).",
        "tests": {"B1": [E + "AnAmbiguouslyOwnedOrderIsStillAwaitingClose"], "B2": ["TestAWorkingEntryIsCancelledBeforeTheLiquidation"],
                  "B3": ["TestAWorkingEntryIsCancelledBeforeTheLiquidation"], "B4": ["TestAWorkingEntryIsCancelledBeforeTheLiquidation"],
                  "B5": ["TestAWorkingEntryIsCancelledBeforeTheLiquidation"], "B6": ["TestAWorkingEntryIsCancelledBeforeTheLiquidation"],
                  "B7": ["TestAWorkingEntryIsCancelledBeforeTheLiquidation"]},
    },
    "internal-reconcile--recovery.run": {
        "pkg": "reconcile",
        "role": "재시작 복구 순서(재시작 규칙 → 미종결 정산 → 계좌 읽기 → 재구성). a094: ACKED 갈래에서 발주를 기록 번호로 한 번 읽어 바이트 일치면 확정(confirmAcked), 아니면 종전대로 StillPending + 명명 critical.",
        "inputs": [("`pending`", "미종결 attempt", "`PendingAttempts`", "읽기 실패는 ErrRecoveryIncomplete(종전)"),
                   ("`rec.State == ACKED`", "접수 뒤 확정 전 죽은 attempt", "원장", "confirmAcked — 실패는 복구를 실패시키지 않음")],
        "calls": "`RecoverPending` · `PendingAttempts` · **`confirmAcked`**(LookupIntent · `execgw.ConfirmPlacedOrder` · Resume · ResolveConfirmed · RecordCritical 창 0) · `blockedSymbol` · `replay` · `Resolver.Resolve` · `stableSnapshot` · `LocalStateFromJournal` · Comparer · Gate.",
        "mut": "attempt ACKED→CONFIRMED(바이트 일치일 때만) · 그 밖 종전.",
        "safety": "ACKED 는 목록 대조 해소로 보내지 않는다(matcher 번호 판별자 부재). 새 오류 반환 경로 0 — 관측 루프가 뜨지 않는 경로를 만들지 않음(D−4.2-2). 브로커 호출은 ACKED PLACE 행마다 읽기 1(§0.4 D−5.5). High-risk: yes.",
        "tests": {f"B{i}": ["TestRecoveryReleasesTheLatchOnlyWhenItCompletes"] for i in range(1, 14)} | {
            "B4": ["TestA094AnAckedPlaceIsConfirmedByItsRecordedNumber", "TestA094AnUnconfirmedAckedPlaceStaysAndIsNamed"],
            "B5": ["TestA094AnAckedPlaceIsConfirmedByItsRecordedNumber"],
            "B6": ["TestA094AnUnconfirmedAckedPlaceStaysAndIsNamed", "TestA094AnAckedCancelOrNumberlessPlaceIsOnlyNamed"],
            "B7": ["TestCrashMidDispatchBecomesInDoubtAndIsResolved"], "B8": ["TestCrashMidDispatchBecomesInDoubtAndIsResolved"]},
    },
    "internal-app-engine--context.recovery": {
        "pkg": "engine",
        "role": "재시작 복구를 조립한다. a094: Alerts(엔진 알림기)와 CatchUp(기동 따라잡기 클로저)을 호출자 값과 무관하게 덮는다.",
        "inputs": [("`c.Notifier`", "알림기", "Context", "nil 이면 Alerts 없음 · 따라잡기 critical nil")],
        "calls": "`reconcile.New` · `catchUpExitProposals`(클로저 — 실행은 복구 뒤).",
        "mut": "없음(조립).",
        "safety": "배선 누락은 침묵이므로 역할 시험으로 고정(`CatchUpWired`). High-risk: no(배선).",
        "tests": {f"B{i}": ["TestA094TheRecoveryCarriesTheBootCatchUp"] for i in range(1, 6)},
    },
    "internal-app-engine--startpositionpolicycommandserver": {
        "pkg": "engine",
        "role": "엔진 제어 endpoint 를 연다. a094: park 해동 route 를 capability 발견으로 등록(3줄, a066 블록 옆).",
        "inputs": [("`commands`", "명령 서비스", "engine run", "attemptThawCommands 를 구현하면 route 추가")],
        "calls": "`registerAttemptThawRoute`(신설) 외 종전.",
        "mut": "없음(등록).",
        "safety": "capability 없는 빌드는 route 집합 불변. 콘솔은 route 를 부르지 않음(시험). High-risk: no(등록) — 명령 자체는 attempt_thaw_command.go.",
        "tests": {f"B{i}": ["TestA094AnOperatorThawClosesTheParkAndReleasesTheProposal"] for i in range(1, 19)},
    },
}

TEST_BUNDLES = {
    "internal-app-engine--testabreachdisplacesanoutstandingtakeprofit": ("engine", "기존 시험 — a094 D−4.3 로 손절이 취소한 주기 다음 주기에 나감(+1 관측). 이름 붙은 대가(design.md:403 · :311)."),
    "internal-journal--testarefusalrearmsthelevel": ("journal", "기존 시험 — ResolveExitProposal 이 기대 intent 를 요구하므로 발의에 intent 를 달고 그 intent 로 해제(단언 무변)."),
    "internal-journal--testresolvingnothingisnotanerror": ("journal", "기존 시험 — 기대 intent 인자 추가(단언 무변: 빈 상태의 해제는 오류 아님)."),
    "internal-journal--testacancelledrungisproposableagain": ("journal", "기존 시험 — 발의에 intent 를 달고 그 intent 로 해제(단언 무변: rung 되돌림 · 기준선 유지)."),
    "internal-journal--testarefusedproposalisrearmedafterarestart": ("journal", "기존 시험 — 발의에 intent 를 달고 그 intent 로 해제(단언 무변)."),
}


def write_test_bundles(cov_dir: Path) -> None:
    for name, (pkg, why) in TEST_BUNDLES.items():
        bundle = FL / name
        ast = json.loads((bundle / "ast.json").read_text())
        fn = ast["function"]
        tests = {b["id"]: [fn] for b in (ast.get("branches") or [])} or {"B1": [fn]}
        tests_file = bundle / ".tests.json"
        tests_file.write_text(json.dumps(tests))
        out = subprocess.run([sys.executable, str(TABLES), str(bundle), str(cov_dir / f"cov_{pkg}.out"), str(tests_file)],
                             capture_output=True, text=True, check=True).stdout
        tests_file.unlink()
        branch_table, btm_rows = out.split("\n\n", 1)
        (bundle / "function-logic-map.md").write_text(f"""# Function Logic Map: `{fn}`

- Source: `{ast['file']}` (`{ast['start']['line']}`–`{ast['end']['line']}`)
- Qualified: `{fn}`
- AST evidence: `ast.json` (`source_sha256` {ast['source_sha256'][:16]}…)
- Risk scan: `risk-pattern-report.md`

**역할.** {why}

## Inputs and invariants

시험 픽스처(임시 원장). 생산 코드 아님 — 비례 원칙상 이 번들은 게이트의 요구 집합(수정된 기존 함수)을 채우는 기록이다.

## Branches and early returns

{branch_table.strip()}

## Calls and live bindings

시험 대상 API(원장 · 관측 루프)만 부른다. 브로커 호출 0.

## State mutations and fallbacks

임시 원장만.

## Safety conclusion

- High-risk impact: no(시험). 단언은 약화되지 않았다 — 바뀐 것은 새 호출 형태(기대 intent) 또는 이름 붙은 대가의 +1 관측뿐.
""")
        (bundle / "branch-test-map.md").write_text(f"""# Branch Test Map: `{fn}`

- Source: `{ast['file']}`

{btm_rows.strip()}
""")
        print("wrote", name)


def main() -> int:
    cov_dir = Path(sys.argv[1])
    only = set(sys.argv[2:])
    for name, spec in BUNDLES.items():
        if only and name not in only:
            continue
        bundle = FL / name
        ast = json.loads((bundle / "ast.json").read_text())
        tests_file = bundle / ".tests.json"
        tests_file.write_text(json.dumps(spec["tests"]))
        out = subprocess.run([sys.executable, str(TABLES), str(bundle), str(cov_dir / f"cov_{spec['pkg']}.out"), str(tests_file)],
                             capture_output=True, text=True, check=True).stdout
        tests_file.unlink()
        branch_table, btm_rows = out.split("\n\n", 1)
        nb = len(ast.get("branches") or [])
        fn = ast["function"]
        qualified = ((ast.get("receiver") or "") + "." if ast.get("receiver") else "") + fn
        inputs = "\n".join(f"| {a} | {b} | {c} | {d} |" for a, b, c, d in spec["inputs"])
        flm = f"""# Function Logic Map: `{qualified}`

- Source: `{ast['file']}` (`{ast['start']['line']}`–`{ast['end']['line']}`)
- Qualified: `{qualified}`
- AST evidence: `ast.json` (`source_sha256` {ast['source_sha256'][:16]}…) — 구현 로트(2026-09-30) 편집 뒤. 편집 전 AST 는 `analysis/implementation/pre-edit/`
- Risk scan: `risk-pattern-report.md`
- 분기 {nb}

**역할.** {spec['role']}

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
{inputs}

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 `go test ./internal/<pkg>/ -count=1 -covermode=set` 프로파일(2026-09-30, `analysis/harness/flm_tables.py`)에서 그 줄로 시작하는 블록의 count.

{branch_table.strip()}

## Calls and live bindings

{spec['calls']}

## State mutations and fallbacks

{spec['mut']}

## Safety conclusion

- {spec['safety']}
"""
        btm = f"""# Branch Test Map: `{qualified}`

- Source: `{ast['file']}`

> Test 열은 그 분기를 지나는 시험(현존 · 통과)이다. 「미진입」 분기의 인용은 그 함수를 도는 시험이며 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> RED 는 새 a094 시험에 한해 변이 원장(`analysis/implementation/mutation-ledger.md`)이 잰다.

{btm_rows.strip()}
"""
        (bundle / "function-logic-map.md").write_text(flm)
        (bundle / "branch-test-map.md").write_text(btm)
        print("wrote", name)
    if not only:
        write_test_bundles(cov_dir)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
