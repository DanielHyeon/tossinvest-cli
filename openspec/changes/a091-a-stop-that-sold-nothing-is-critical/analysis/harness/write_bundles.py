#!/usr/bin/env python3
"""a091 2차 판 FLM/BTM 쓰기 — 분기 표 · 진입 실측은 ast.json · 소스 원문 · 커버리지 프로파일에서 **측정으로** 만들고,
이 파일은 역할 · 입력 · 호출 계약(수치) · 결론 산문과 Test 열만 준다(2026-10-01, base b30318d6).

사용 (저장소 루트, base 커밋의 깨끗한 워크트리에서):
  python3 openspec/changes/a091-…/analysis/harness/write_bundles.py <coverage-dir>
    <coverage-dir> 에 engine.out · obs.out · riskcalc.out · reconcile.out · exitpolicy.out
    (`go test -count=1 -coverprofile=<pkg>.out ./internal/<pkg>/` — 기본 covermode set)
쓰기: analysis/function-logic/<bundle>/{function-logic-map.md, branch-test-map.md, risk-pattern-report.md}.
ast.json 은 이 스크립트가 쓰지 않는다(`go run ./tools/logic-map --file … --func …` 가 쓴다).
"""
from __future__ import annotations

import json
import re
import subprocess
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
FL = HERE.parent / "function-logic"
ROOT = Path.cwd()

# 호출 계약의 공통 수치 — 전부 base b30318d6 소스에서 읽은 값이다(좌표는 각 행).
FLOOR_CONTRACT = (
    "생산 = `reconcileFloor`(exit 사본, `exitSideFloor` — `exit_record_only.go:26`). RECONCILE 블록이 그 종목을 덮지 않으면 "
    "브로커 0회(`exitwiring.go:196-200`). 덮으면 `Retrier.Query` **2회**(Holdings `exitwiring.go:207` · SellableQuantity `:231` — 둘째는 첫째가 성공할 때만) + 원장 읽기 1"
    "(`localOpenSells` → `LiveOrdersForSymbol`). 정책 `DefaultRetryPolicy`(`retry.go:129-137`, 배선 `exitwiring.go:48`): 최대 **3**시도, "
    "대기 400ms → 800ms(±25%, 상한 3s), 예산 **8s** 는 **대기만** 자른다(`sleepWithin` `:444-452` — 시도 자체는 자르지 않음). "
    "한 시도 = 공식 클라이언트 `send`(`client.go:320-360`): 아래 요청 열. HTTP 요청 하나의 시한은 `http.Client.Timeout` **15s**(`client.go:20` · `:131`). **상한은 없다(4판 정정 — 2라운드 R2-11)**: "
    "토큰 관리자 잠금 `tm.mu` 은 캐시 파일 읽기 · 교환 · 저장을 쥐고(`token.go:61-78` · `:109-125` · `:172-223`) 기한이 없다. "
    "실현 가능한 요청 열(한 시도): 토큰 캐시 유효 → GET 1 / 캐시 무효 → 교환 1 + GET 1 / GET 이 401 → refresh 1(채택이면 교환 0, 아니면 1) + GET 1, "
    "채택한 토큰도 401 이면 refresh 한 번 더(이번엔 교환) + GET 1(`client.go:344-360` — 둘째 refresh 는 첫째가 채택일 때만). HTTP 만 세면 "
    "한 시도 ≤ 교환 2 + GET 3 = 5 요청 × 15s, 한 Query ≤ 3 시도 — **가정이 붙은 HTTP 추정이지 벽시계 상한이 아니다**. 401/403 은 재시도 없이 게이트 래치 + 모드 강화(통지는 기록 전용 — a092). "
    "**a091 은 이 호출을 바꾸지 않는다**(기존 비용 — §0.3 의 새 항 아님)"
)
ALERT_CONTRACT = (
    "`o.opts.Alerts` = `obs.RecordOnly{N, Relay}`(생산 배선 `exitwiring.go:348-349`). 일반 등급 → `NormalRelay.Offer`"
    "(`normal_relay.go:43-58` — 비차단 `select`, 버퍼가 차면 버림을 로그로 기록). critical → `n.mu` 아래 "
    "`Journal.RecordAlert`(`record_only.go:136-137` — SQLite `BEGIN IMMEDIATE`, `busy_timeout` **5s**(`journal.go:36`) · "
    "`synchronous=FULL` fsync, 원격 전송 0). `n.mu` 대기는 **기한 없음**(보유자는 로컬 원장 연산만 — a092 정본 「등급화된 알림」 잠금 문단). "
    "재알림 창 `DefaultRemindAfter` **1h**(`notifier.go:59`). 기록 실패 → 그 자리에서 게이트 래치 + 승격 시도, 오류 반환 → `o.alert` 가 "
    "`logErr` 한 줄로 삼킴(`exitloop.go:1818-1819`)"
)

B: dict[str, dict] = {
    "internal-app-engine--exitobserver.applyfloor": {
        "pkg": "engine",
        "edit": "**a091 편집 대상**(tasks 3.x) — 보호 여부 인자 하나 추가 · 0주 두 경로(B2 · 끝)의 보고 종류 · 문구. 반환값 `(수량, capped, err)` 무변경.",
        "role": "RECONCILE 확정 하한으로 청산 수량을 자른다. 0주는 두 자리에서만 나온다: B2(하한 계산 실패 → 리터럴 `\"0\"`) · 끝(`floor.Quantity` 가 원안보다 작고 그것이 `\"0\"`).",
        "inputs": [
            ("`quantity`", "양의 정수 정규형(`big.Int.String`)", "`record` → `snapshot.ProjectedQuantity` ← `ProjectWholeShares`(`snapshot.go:93`)", "0 이면 `orderable=false` 라 이 함수에 오지 않는다(아래 M1)"),
            ("`floor.Quantity`", "음 아닌 십진 정규형, **0 가능**", "`riskcalc.ConfirmedFloorQuantity` — `MaxDecimal(\"0\", …)` → `CanonicalDecimal`(`decimal.go:92-101`) 또는 `zeroFloor` 리터럴 `\"0\"`(`confirmed_floor.go:236-243`)", "0 이면 끝 경로가 `\"0\"` 반환"),
            ("`o.opts.Floor`", "nil 허용", "주입(`exitSideFloor`)", "nil 이면 무캡(B1)"),
            ("**보호/익절 맥락**", "—", "**인자에 없다**", "⚠ 이 함수는 제안이 손절인지 모른다 — a091 이 `submit` 에서 `isProtective(proposal)` 를 넘긴다(D2)"),
        ],
        "calls": [
            ("`o.opts.Floor.ConfirmedFloor`", "1621", "RECONCILE 확정 하한", FLOOR_CONTRACT),
            ("`o.logErr`", "1626", "B2 의 유일한 기록(오류 객체)", "구조화 로그 한 줄(`exitloop.go:1834-1839`), 반환 없음. **종류가 `EventExitProposalCapped`** — H2 가 바꾸는 자리"),
            ("`riskcalc.CompareDecimal`", "1633", "하한 vs 원안", "순수 계산, 오류 → B4"),
            ("`fmt.Errorf`", "1635", "B4 오류 감싸기", "순수"),
            ("`riskcalc.SubDecimal`", "1640", "잔여", "순수 계산, 오류 → B6"),
            ("`fmt.Errorf`", "1642", "B6 오류 감싸기", "순수"),
            ("`o.alert`", "1644", "캡 알림(부분 · 0주 공통 — 현행)", ALERT_CONTRACT + ". **현행 종류 `EventExitProposalCapped` = normal** → Relay 경로(outbox 행 0)"),
            ("`string`", "1646", "event key 조립 `type|position`", "순수"),
            ("`o.label`", "1647", "종목 표시명", "메모리 조회"),
            ("`fmt.Sprintf`", "1648", "본문", "순수"),
        ],
        "mut": "상태 변경 없음. 반환값과 로그 · 알림이 전부. B2 의 fail-closed(0 으로 봄) 방향은 옳다 — 문제는 보고의 등급 · 종류 · 문구다.",
        "safety": "제출 수량 계산(B1~B6 · 끝의 반환)은 a091 이 건드리지 않는다(§0.3 · §0.9). 편집은 B2 의 오류 줄(종류 · 계좌 가림) · B2 알림 추가 · 끝의 원인 분류 · 종류/문구 분기 · 알림 켜짐 게이트뿐(design D1 · D3 · D8). High-risk: yes(**확정 하한이** 손절을 0주로 깎는 유일한 자리 — 0 투영 보호 액션은 이 함수에 오지 않는다, record B11).",
        "tests": {
            "B1": ["TestNoFloorSourceCapsNothing"],
            "B2": ["TestAFloorThatCannotBeComputedSellsNothing"],
            "B3": ["TestAZeroFloorSubmitsNothingAndLeavesTheLevelProposable"],
            "B4": ["TestTheConfirmedFloorCapsTheLiquidation"],
            "B5": ["TestTheConfirmedFloorCapsTheLiquidation"],
            "B6": ["TestTheConfirmedFloorCapsTheLiquidation"],
        },
    },
    "internal-obs--severityof": {
        "pkg": "obs",
        "edit": "함수 본문은 편집하지 않는다 — a091 은 등급표 `criticalEvents`(`event.go:337-361`)에 새 종류 한 줄을 더한다(tasks 2.4). 판정 방식의 근거.",
        "role": "등급은 종류에만 붙는다: `criticalEvents` 맵 조회 하나. 미등록은 normal(기본값 — 주석 `event.go:363-370`).",
        "inputs": [
            ("`t`", "임의 문자열", "호출자", "미등록 → normal"),
            ("`criticalEvents`", "**19** 종(base b30318d6: 주문 · 브로커 · 알림 · 모드 · 루프 13 + exit 관측 5 + a095 편입 실패 1)", "`event.go:337-361`", "—"),
        ],
        "calls": [],
        "mut": "없음.",
        "safety": "같은 종류 안에서 등급을 나눌 수 없다 → 0주만 critical 로 올리려면 새 종류가 필요하다(D1 B). a095 가 같은 모양으로 `EventExitPositionAdoptionFailed` 를 신설했다(선례). High-risk: yes(등급표 = 진입 차단 스위치).",
        "tests": {"B1": ["TestAQuarantineCreationIsCritical", "TestMeasurementEventsAreNeverCritical"]},
    },
    "internal-app-engine--exitobserver.submit": {
        "pkg": "engine",
        "edit": "**a091 편집 대상**(tasks 3.6) — `applyFloor` 호출에 `isProtective(proposal)` 를 넘긴다. 분기 · 반환 무변경.",
        "role": "무장된 발의를 브로커로 보낸다. 0주(B2)면 제출 없이 해제한다.",
        "inputs": [
            ("`proposal`", "`Action.Orderable()` 5종 중 하나", "`record`", "`isProtective` = BaselineBreach · LadderStop(`exitloop.go:1375-1377`)"),
            ("`quantity`", "양의 정수 정규형", "`record` ← `snapshot.ProjectedQuantity`", "—"),
            ("`submitQuantity`", "`applyFloor` 반환", "—", "`isZeroQuantity` 가 0 판정(수치 비교 — 아래)"),
        ],
        "calls": [
            ("`o.applyFloor`", "1397", "확정 하한", "위 applyFloor 번들의 계약 — RECONCILE 에서 브로커 읽기 ≤2 Query"),
            ("`isZeroQuantity`", "1401", "0주 판정", "순수. **수치 비교**(`CompareDecimal(q,\"0\") <= 0`, 파싱 실패 · 빈 문자열도 0)"),
            ("`o.release`", "1406", "0주 → 발의 해제(`ProposalRefused`)", "원장 트랜잭션(`ReleaseUnacceptedExitProposal`), busy_timeout 5s"),
            ("`o.opts.Issuer.IssueReduction`", "1409", "Guardian 축소 발행", "로컬 판정(브로커 0)"),
            ("`costs.Market`", "1412", "시장 변환", "순수"),
            ("`strings.ToLower`", "1412", "—", "순수"),
            ("`strings.TrimSpace`", "1412", "—", "순수"),
            ("`fmt.Sprintf`", "1418", "사유 문구", "순수"),
            ("`o.alertProposalRefused`", "1422", "발행 거절 알림", "critical → 기록 전용(ALERT 계약)"),
            ("`err.Error`", "1422", "—", "순수"),
            ("`o.release`", "1423", "해제", "원장"),
            ("`o.opts.Journal.AttachExitIntent`", "1430", "intent 부착", "원장"),
            ("`fmt.Errorf`", "1431", "—", "순수"),
            ("`o.sellIntent`", "1434", "주문 모양", "순수"),
            ("`o.alertRefused`", "1436", "판정 거절 알림", "critical → 기록 전용"),
            ("`o.release`", "1437", "해제", "원장"),
            ("`o.opts.Submit.Place`", "1439", "**유일한 브로커 mutation**(매도 1)", "게이트웨이 계약(a094 · order-execution) — a091 무변경"),
            ("`o.log`", "1447", "체결 확정 로그", "로그 한 줄"),
            ("`string`", "1450", "—", "순수"),
            ("`o.noteDelay`", "1460", "심볼 점유 지연", "critical 지연 알림(기록 전용)"),
            ("`o.release`", "1461", "해제(취소)", "원장"),
            ("`fmt.Errorf`", "1467", "—", "순수"),
            ("`fmt.Errorf`", "1469", "—", "순수"),
            ("`err.Error`", "1474", "—", "순수"),
            ("`o.alertProposalRefused`", "1476", "거절 알림", "critical → 기록 전용"),
            ("`o.release`", "1477", "해제", "원장"),
        ],
        "mut": "발의 해제(원장) · 주문 1(Place). 0주 경로(B2)는 제출 0 · 해제 1.",
        "safety": "a091 은 `applyFloor` 호출의 인자 하나만 바꾼다 — B1~B13 무변경. High-risk: yes.",
        "tests": {
            "B1": ["TestAFloorThatCannotBeComputedSellsNothing"],
            "B2": ["TestAZeroFloorSubmitsNothingAndLeavesTheLevelProposable"],
            "B3": ["TestA094AnUnacceptedStopIsReleasedAndProposedAgain"],
            "B4": ["TestABaselineBreachProposesTheWholePosition"],
            "B5": ["TestABaselineBreachProposesTheWholePosition"],
            "B6": ["TestABaselineBreachProposesTheWholePosition"],
            "B7": ["TestABaselineBreachProposesTheWholePosition"],
            "B8": ["TestAnInDoubtSubmissionKeepsTheProposalArmed"],
            "B9": ["TestABaselineBreachProposesTheWholePosition"],
            "B10": ["TestA094AnUnrecordedOutcomeKeepsTheProposalArmed"],
            "B11": ["TestA094AnUnrecordedOutcomeKeepsTheProposalArmed"],
            "B12": ["TestA094AnUnacceptedStopIsReleasedAndProposedAgain", "TestARefusedProposalReleasesTheLevelAndAlerts"],
            "B13": ["TestARefusedProposalReleasesTheLevelAndAlerts"],
        },
    },
    "internal-app-engine--exitobserver.alert": {
        "pkg": "engine",
        "edit": "편집하지 않는다 — 새 critical 이 타는 기존 경로의 증거(design D5).",
        "role": "exit 관측 goroutine 의 알림 입구. 오류는 로그 한 줄로 삼킨다.",
        "inputs": [("`o.opts.Alerts`", "nil 허용", "생산: `obs.RecordOnly`(`exitwiring.go:349`)", "nil 이면 무동작(B1)")],
        "calls": [
            ("`o.opts.Alerts.Notify`", "1818", "기록(critical) 또는 이관(normal)", ALERT_CONTRACT),
            ("`o.logErr`", "1819", "기록 실패 로그", "로그 한 줄 — 종류는 사건의 종류(`e.Type`)"),
        ],
        "mut": "원장 outbox 행(critical) 또는 이관 버퍼(normal). 기록 실패는 알림기가 이미 게이트를 잠갔다.",
        "safety": "새 종류는 이 경로를 그대로 탄다 — 발송 경로를 만들지 않는다(D5). 원격 전송 0.",
        "tests": {"B1": ["TestNoFloorSourceCapsNothing"], "B2": ["TestA092ACappedLiquidationDoesNotWaitForTheTransport"]},
    },
    "internal-obs--recordonly.notify": {
        "pkg": "obs",
        "edit": "편집하지 않는다 — 새 critical 종류의 기록 · 로그 경로 증거(D5 · H2).",
        "role": "사건을 로그 한 줄로 남기고(`logEvent` — 종류 = `e.Type`), critical 이면 원장에 기록만, 일반이면 유계 버퍼에 넘긴다.",
        "inputs": [("`r.N`", "nil 허용", "배선", "nil → 무동작(B1)"), ("`e.Type`", "종류", "호출자", "`SeverityOf` 로 등급")],
        "calls": [
            ("`SeverityOf`", "50", "등급", "순수(맵 조회)"),
            ("`n.logEvent`", "51", "구조화 로그 한 줄", "로컬 로그 — **알림과 같은 종류**(H2 근거: 알림 경로 자체는 이미 한 종류)"),
            ("`withoutFields`", "51", "필드 제거(계좌 원문 차단)", "순수"),
            ("`n.logNormalDrop`", "54", "릴레이 없음 → 버림 기록", "로그 한 줄"),
            ("`r.Relay.Offer`", "57", "일반 등급 이관", "비차단(`normal_relay.go:53-57`)"),
            ("`n.recordCritical`", "60", "critical 기록", "`n.mu` + `Journal.RecordAlert`(busy_timeout 5s, 원격 0), 실패 → 게이트 래치 + 승격 시도"),
            ("`n.remindAfter`", "60", "재알림 창", "`DefaultRemindAfter` 1h"),
        ],
        "mut": "outbox 행 삽입 또는 재무장(critical) · 버퍼 적재(normal).",
        "safety": "새 종류가 critical 로 등록되면 이 함수의 B2(normal 갈래)를 건너 `recordCritical` 로 간다 — 그 사실 하나가 a091 의 이익(원장 흔적)이다.",
        "tests": {
            "B1": ["TestA092RecordOnlyCriticalNeverPublishes"],
            "B2": ["TestA092ANormalAlertIsHandedOffNotSent"],
            "B3": ["TestA092WithoutARelayANormalAlertIsDroppedNotSent"],
        },
    },
    "internal-app-engine--iszeroquantity": {
        "pkg": "engine",
        "edit": "편집하지 않는다 — M1(0주 판정의 모양) 증거.",
        "role": "0 판정. 공백 제거 뒤 빈 문자열이면 0(B1), 아니면 **수치 비교** `CompareDecimal(q, \"0\") <= 0`, 파싱 실패도 0.",
        "inputs": [("`q`", "십진 문자열", "`applyFloor` 반환 · 포지션 수량", "파싱 실패 → 0(fail-closed)")],
        "calls": [("`strings.TrimSpace`", "1872", "—", "순수"), ("`riskcalc.CompareDecimal`", "1876", "수치 비교", "순수")],
        "mut": "없음.",
        "safety": "첫 리뷰 M1 이 적은 「정확히 `\"0\"` 문자열 비교」는 base b30318d6 에서 **더 이상 참이 아니다** — `\"0.0\"` · `\" 0\"` 도 0 이다. 0주 판정은 철자에 의존하지 않는다.",
        "tests": {"B1": ["TestAZeroFloorSubmitsNothingAndLeavesTheLevelProposable"]},
    },
    "internal-app-engine--exitobserver.record": {
        "pkg": "engine",
        "edit": "편집하지 않는다 — M1: `submit` 은 `orderable` 일 때만 불린다(B11)는 증거.",
        "role": "판정을 원장에 무장하고 제출로 넘긴다. `orderable = snapshot.Orderable && !proposal.Zero()`(`:1233`) 가 거짓이면 제출이 없다.",
        "inputs": [("`snapshot.ProjectedQuantity`", "정수 정규형", "`EvaluateLadderSnapshot` · `EvaluateRatchetSnapshot`", "`\"0\"` 이면 `Orderable=false`")],
        "calls": [("`o.submit`", "1353", "제출", "submit 번들")],
        "calls_note": "호출 좌표 전수는 `ast.json` `calls`(24) — 이 번들이 근거로 쓰는 것은 B11 → `o.submit`(`:1353`) 하나다.",
        "mut": "원장 무장 · 해제(판정 트랜잭션).",
        "safety": "0 투영 수량은 `orderable` 을 거짓으로 만들어 `submit` 에 닿지 않는다 — `applyFloor` 에 들어오는 `quantity` 는 항상 양수다.",
        "tests": {
            "B1": ["TestA094TheParkCheckChangesNoOutcome"], "B2": ["TestABaselineBreachProposesTheWholePosition"],
            "B3": ["TestABaselineBreachProposesTheWholePosition"], "B4": ["TestABaselineBreachProposesTheWholePosition"],
            "B5": ["TestAWorkingEntryIsCancelledBeforeTheLiquidation"], "B6": ["TestCrossingTheFirstTakeProfitKeepsThePositionUnderJudgement"],
            "B7": ["TestAWorkingEntryIsCancelledBeforeTheLiquidation"], "B8": ["TestA094AnotherIntentInFlightStopsTheArmReleaseLoop"],
            "B9": ["TestA094ConsecutiveClearFailuresRaiseAnEarlierAlert"], "B10": ["TestAWorkingEntryIsCancelledBeforeTheLiquidation"],
            "B11": ["TestABaselineBreachProposesTheWholePosition"], "B12": ["TestABaselineBreachProposesTheWholePosition"],
            "B13": ["TestAnUnresolvedProposalSuppressesTheNextOne"], "B14": ["TestAnUnresolvedProposalSuppressesTheNextOne"],
            "B15": ["TestAnUnresolvedProposalSuppressesTheNextOne"], "B16": ["TestABaselineBreachProposesTheWholePosition"],
        },
    },
    "internal-exitpolicy--exitlinesnapshot.executableproposal": {
        "pkg": "exitpolicy",
        "edit": "편집하지 않는다 — M1 증거.",
        "role": "주문 가능하지 않거나 투영 수량이 `\"0\"` 이면 빈 제안(B1).",
        "inputs": [("`s.ProjectedQuantity`", "`big.Int.String` 정규형", "`ProjectWholeShares`", "`\"0\"` → 빈 제안")],
        "calls": [],
        "mut": "없음.",
        "safety": "문자열 `\"0\"` 비교이지만 입력이 `big.Int.String` 이라 0 의 철자는 `\"0\"` 하나뿐이다(ProjectWholeShares 번들).",
        "tests": {"B1": ["TestOneShareIntermediateLadderTargetIsStateOnly"]},
    },
    "internal-exitpolicy--projectwholeshares": {
        "pkg": "exitpolicy",
        "edit": "편집하지 않는다 — M1: 투영 수량의 철자 출처.",
        "role": "잔여 × 비율(상한 1)을 유리수로 곱해 내림한 **정수**를 `units.String()`(`:93`)로 낸다. 음수면 0(B5).",
        "inputs": [("`remaining`", "양의 유리수 정규형", "`canonicalSnapshotContext` 의 `positive` + `RatString`", "파싱 실패 → 오류(B1)")],
        "calls": [("`units.String`", "93", "정수 정규형 철자", "순수 — 0 은 `\"0\"`")],
        "calls_note": "호출 좌표 전수는 `ast.json` `calls`(21) — 근거는 반환 철자(`:93`) 하나다.",
        "mut": "없음.",
        "safety": "반환은 `big.Int` 의 10진 철자 — 선행 0 · 소수점 · 공백이 없다.",
        "tests": {
            "B1": ["TestOneShareFinalAndBreachProjectExactlyOne"], "B2": ["TestOneShareFinalAndBreachProjectExactlyOne"],
            "B3": ["TestOneShareFinalAndBreachProjectExactlyOne"], "B4": ["TestOneShareFinalAndBreachProjectExactlyOne"],
            "B5": ["TestOneShareFinalAndBreachProjectExactlyOne"],
        },
    },
    "internal-exitpolicy--canonicalsnapshotcontext": {
        "pkg": "exitpolicy",
        "edit": "편집하지 않는다 — M1: 잔여 수량은 양수여야 하고 `RatString` 정규형으로 넘어간다.",
        "role": "스냅숏 문맥을 정규화한다. 잔여 수량은 `positive(\"remaining quantity\", …)`(`:263`) — 0 · 음수 · 비수치는 오류(B2).",
        "inputs": [("`ctx.RemainingQuantity`", "포지션 수량(원장 투영)", "`ConvergeQuantities` 가 쓴 계좌 값 등", "0/음수/비수치 → 오류")],
        "calls": [("`positive`", "263", "양수 강제", "순수"), ("`quantity.RatString`", "267", "정규형", "순수")],
        "calls_note": "호출 좌표 전수는 `ast.json` `calls`(8).",
        "mut": "없음.",
        "safety": "포지션 수량의 철자(원장에 어떤 모양으로 있든)는 여기서 유리수로 파싱되고 `RatString` 으로 바뀐다 — 철자가 0 판정으로 새지 않는다.",
        "tests": {"B1": ["TestSnapshotIdentityIsDeterministicAndObservationBound"], "B2": ["TestSnapshotIdentityIsDeterministicAndObservationBound"]},
    },
    "internal-exitpolicy--evaluateladdersnapshot": {
        "pkg": "exitpolicy",
        "edit": "편집하지 않는다 — M1: 사다리 스냅숏의 `orderable = projected != \"0\"`(`:143`).",
        "role": "사다리 판정의 스냅숏. 주문 가능 액션이면 `ProjectWholeShares` 로 투영하고 0 이면 state-only.",
        "inputs": [("`in.Context.RemainingQuantity`", "양수", "canonicalSnapshotContext", "—")],
        "calls": [("`ProjectWholeShares`", "139", "투영", "순수")],
        "calls_note": "호출 좌표 전수는 `ast.json` `calls`(15).",
        "mut": "없음.",
        "safety": "0 투영은 `orderable=false` — 제출 경로에 닿지 않는다.",
        "tests": {f"B{i}": ["TestOneShareIntermediateLadderTargetIsStateOnly"] for i in range(1, 12)},
    },
    "internal-exitpolicy--evaluateratchetsnapshot": {
        "pkg": "exitpolicy",
        "edit": "편집하지 않는다 — M1: 래칫 스냅숏의 `orderable = projected != \"0\"`(`:208`).",
        "role": "래칫 판정의 스냅숏. 사다리와 같은 투영 · 주문 가능성 규칙.",
        "inputs": [("`in.Context.RemainingQuantity`", "양수", "canonicalSnapshotContext", "—")],
        "calls": [("`ProjectWholeShares`", "204", "투영", "순수")],
        "calls_note": "호출 좌표 전수는 `ast.json` `calls`(13).",
        "mut": "없음.",
        "safety": "0 투영은 `orderable=false`.",
        "tests": {f"B{i}": ["TestRatchetOneSharePartialIsStateOnlyButBreachIsFull"] for i in range(1, 11)},
    },
    "internal-riskcalc--confirmedfloorquantity": {
        "pkg": "riskcalc",
        "edit": "편집하지 않는다 — M1: `floor.Quantity` 의 철자 출처.",
        "role": "확정 하한 = max(0, min(보유, 매도가능) − 로컬 미체결 매도). 스냅숏 부재 · 낡음은 `zeroFloor`(리터럴 `\"0\"`, B4 · B6), 계산 경로는 `MaxDecimal(\"0\", …)` → `CanonicalDecimal`.",
        "inputs": [("`in.Holdings` · `in.Sellable`", "스냅숏(나이 ≤ 한계)", "`reconcileFloor` 의 두 브로커 읽기", "부재 · 낡음 → 0")],
        "calls": [("`MaxDecimal`", "180", "0 하한", "순수 — 정규 철자 반환(`decimal.go:92-101`)")],
        "calls_note": "호출 좌표 전수는 `ast.json` `calls`(19).",
        "mut": "없음.",
        "safety": "반환 `Quantity` 의 0 은 `\"0\"` 한 철자다. 첫 리뷰 M1 의 전제(정규형)는 여기서 성립한다.",
        "tests": {
            "B1": ["TestTheFloorIsTheFormula"], "B2": ["TestAnUnknownLocalQuantityIsRefused"], "B3": ["TestMalformedQuantitiesAreRefused"],
            "B4": ["TestAnAbsentSnapshotIsZero"], "B5": ["TestMalformedQuantitiesAreRefused"], "B6": ["TestAStaleSnapshotIsZero"],
            "B7": ["TestTheFloorIsTheFormula"], "B8": ["TestTheSellableQuantityOnlyLowers"], "B9": ["TestTheFloorIsTheFormula"],
            "B10": ["TestTheFloorIsTheFormula"], "B11": ["TestALocalQuantityCanNeverRaiseTheFloor"],
        },
    },
    "internal-reconcile--converger.convergequantities": {
        "pkg": "reconcile",
        "edit": "편집하지 않는다 — M1(Manager 지시): 포지션 수량의 생산자. 수량 불일치에서 원장 투영을 계좌 값으로 맞춘다.",
        "role": "불일치마다 `ApplyPositionAdjustment(NewQuantity: mismatch.Authority())`(`:218` — `Authority()` = `m.Broker`, `compare.go:266`). 0 으로 수렴하면 exit 상태를 닫는다(B13).",
        "inputs": [("`mismatch.Broker`", "계좌가 보고한 수량(철자는 대사 비교가 만든 값)", "`reconcile.Compare`", "—")],
        "calls": [("`c.Journal.ApplyPositionAdjustment`", "209", "투영 수렴", "원장 트랜잭션")],
        "calls_note": "호출 좌표 전수는 `ast.json` `calls`(31).",
        "mut": "포지션 수량 · exit 상태(0 수렴 시 종료).",
        "safety": "이 값의 철자는 0 판정에 닿지 않는다: exit 루프는 `isZeroQuantity(p.Quantity)`(수치) 포지션을 건너뛰고(`exitloop.go:541`), 나머지는 `canonicalSnapshotContext` 가 양수 강제 + `RatString` 으로 바꾼다.",
        "tests": {
            "B1": ["TestNoMismatchesIsANoOp"], "B2": ["TestAQuantityMismatchConvergesToTheAccount"], "B3": ["TestAQuantityMismatchConvergesToTheAccount"],
            "B4": ["TestAQuantityMismatchConvergesToTheAccount"], "B5": ["TestConvergenceMakesTheBlockReleasable"], "B6": ["TestConvergenceMakesTheBlockReleasable"],
            "B7": ["TestAQuantityMismatchConvergesToTheAccount"], "B8": ["TestAQuantityMismatchConvergesToTheAccount"],
            "B9": ["TestASymbolWithNoLiveInstanceIsRefusedRatherThanFolded"], "B10": ["TestAQuantityMismatchConvergesToTheAccount"],
            "B11": ["TestAStaleAdjustmentStopsThePass"], "B12": ["TestAStaleAdjustmentStopsThePass"],
            "B13": ["TestConvergingAManagedPositionToZeroAlerts"], "B14": ["TestConvergingAManagedPositionToZeroAlerts"],
            "B15": ["TestConvergingAManagedPositionToZeroAlerts"],
        },
    },
    "internal-obs--notifier.escalate": {
        "pkg": "obs",
        "edit": "**a091 편집 대상(5판, R3-3)** — 두 로그 줄(`:433` 실패 · `:440` 승격)에서 `FieldAccount` 원문을 뺀다. 판정 · 반환 · 원장 호출 무변경.",
        "role": "critical 전달 실패(또는 기록 실패)의 운영 모드 승격을 원장에 남긴다. 통지하지 않는다(전송 수단이 방금 실패).",
        "inputs": [("`n.AccountRef`", "계좌 참조(생산 = 계좌번호, `interlock.go:680-684`)", "배선", "빈 값 → 승격 없음(B1)"), ("`e.Type`", "촉발 사건 종류", "호출자", "로그 필드")],
        "calls": [
            ("`strings.TrimSpace`", "426", "빈 계좌 판정", "순수"),
            ("`n.Journal.EscalateOperatingMode`", "429", "ENTRY_BLOCKED 승격", "원장 트랜잭션(`busy_timeout` 5s · 연결 풀 대기 기한 없음), 원격 0"),
            ("`n.Log.Error`", "433", "승격 실패 로그", "로그 한 줄 — **`FieldAccount` 원문**(a091 이 뺀다)"),
            ("`MaskAccount`", "433", "오류 속 계좌 가림", "순수"),
            ("`string`", "435", "—", "순수"),
            ("`n.Log.Warn`", "440", "승격 로그", "로그 한 줄 — **`FieldAccount` 원문**(a091 이 뺀다)"),
        ],
        "mut": "운영 모드 행(원장).",
        "safety": "호출자 셋(`notifier.go:238` · `:261` · `record_only.go:157`)이 공유하는 함수다 — a091 의 편집은 로그 필드 하나를 빼는 것뿐이고 판정 · 반환은 같다. 계좌는 한 알림기 하나이므로 필드를 빼도 줄의 뜻은 같다. High-risk: 아니오(로그), 단 모드 승격 경로에 있다.",
        "tests": {
            "B1": ["TestA092RecordOnlyFailureLatchesAndEscalates"],
            "B2": ["TestA092TheEscalationFailureErrorMasksTheAccount"],
            "B3": ["TestA092TheEscalationFailureErrorMasksTheAccount"],
            "B4": ["TestA092RecordOnlyFailureLatchesAndEscalates"],
        },
    },
}


def coverage(cov_dir: Path, source: str) -> dict[int, int]:
    out: dict[int, int] = {}
    pat = re.compile(r"^(?:.*/)?" + re.escape(source) + r":(\d+)\.\d+,\d+\.\d+ \d+ (\d+)$")
    for profile in cov_dir.glob("*.out"):
        for line in profile.read_text().splitlines():
            m = pat.match(line)
            if m:
                s, c = int(m.group(1)), int(m.group(2))
                out[s] = max(out.get(s, 0), c)
    return out


def esc(s: str) -> str:
    return s.replace("|", "\\|")


def write(name: str, spec: dict, cov_dir: Path) -> None:
    d = FL / name
    ast = json.loads((d / "ast.json").read_text())
    src = ast["file"]
    lines = (ROOT / src).read_text().splitlines()
    cov = coverage(cov_dir, src)
    fn = (ast.get("receiver") + "." if ast.get("receiver") else "") + ast["function"]
    br = ast.get("branches") or []
    rows = []
    for b in br:
        n = b["at"]["line"]
        entered = "예" if cov.get(n, 0) > 0 else ("아니오" if n in cov else "—")
        rows.append((b["id"], b["kind"], n, esc(lines[n - 1].strip()), entered))
    rets = ", ".join(f"`{r['at']['line']}:{r['at']['column']}`" for r in (ast.get("returns") or []))
    head = [
        f"# Function Logic Map: `{fn}`",
        "",
        f"- Source: `{src}` (`{ast['start']['line']}`–`{ast['end']['line']}`)",
        f"- Qualified: `{fn}`",
        f"- AST evidence: `ast.json` (`source_sha256` {ast['source_sha256'][:16]}…) — base `b30318d6` 에서 `go run ./tools/logic-map`",
        "- Risk scan: `risk-pattern-report.md`",
        f"- 분기 {len(br)} · 반환 {len(ast.get('returns') or [])}",
        "",
        f"**편집.** {spec['edit']}",
        "",
        f"**역할.** {spec['role']}",
        "",
        "## Inputs and invariants",
        "",
        "| Input/state | Valid range | Source of truth | Failure behavior |",
        "|---|---|---|---|",
    ]
    head += [f"| {esc(a)} | {esc(b)} | {esc(c)} | {esc(e)} |" for a, b, c, e in spec["inputs"]]
    head += [
        "",
        "## Branches and early returns",
        "",
        "> 조건은 소스 원문, 진입 실측은 base `b30318d6` 의 `go test -count=1 -coverprofile`(covermode set, 2026-10-01)에서 그 줄로 시작하는 블록의 count "
        "(`analysis/harness/write_bundles.py`). engine 패키지 실행은 `-trimpath` 로 `TestA111…` 두 시험이 소스 경로를 못 찾아 실패했다 — 커버리지 프로파일은 그대로 쓰인다(두 시험은 이 함수들과 무관한 AST 핀).",
        "",
        "| Branch | 종류 | 조건 (원문) | 진입 실측 |",
        "|---|---|---|---|",
    ]
    if not rows:
        head.append("| — | — | 분기 없음 — 유일한 경로 | — |")
    head += [f"| {i} | {k} | `:{n}` `{t}` | {e} |" for i, k, n, t, e in rows]
    head += ["", f"Exact AST return positions: {rets}" if rets else "Exact AST return positions: none", "", "## Calls and live bindings", ""]
    if spec.get("calls_note"):
        head += [spec["calls_note"], ""]
    if spec["calls"]:
        head += ["| Callee | Line | Why called | Error/timeout/retry contract |", "|---|---|---|---|"]
        head += [f"| {esc(c)} | `:{ln}` | {esc(w)} | {esc(k)} |" for c, ln, w, k in spec["calls"]]
    else:
        head.append("호출 없음(`ast.json` `calls` 0).")
    head += ["", "## State mutations and fallbacks", "", spec["mut"], "", "## Safety conclusion", "", f"- {spec['safety']}", ""]
    (d / "function-logic-map.md").write_text("\n".join(head))

    t = spec["tests"]
    btm = [
        f"# Branch Test Map: `{fn}`",
        "",
        f"- Source: `{src}`",
        "",
        "> Test 열은 그 함수를 지나는 현존 · 통과 시험이다. 「아니오」 분기의 인용은 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).",
        "> a091 의 새 RED 는 구현 로트의 변이 원장이 잰다 — 이 표는 편집 **전** base 의 사실이다.",
        "",
        "| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |",
        "|---|---|---|---|---|---|",
    ]
    if not rows:
        btm.append(f"| B1 | 분기 없음 — 유일한 경로 | — | {' · '.join(f'`{x}`' for x in t.get('B1', []))} | n/a | yes |")
    for i, _k, n, txt, e in rows:
        btm.append(f"| {i} | `:{n}` `{txt}` | {e} | {' · '.join(f'`{x}`' for x in t.get(i, []))} | n/a | yes |")
    (d / "branch-test-map.md").write_text("\n".join(btm) + "\n")

    subprocess.run(
        [sys.executable, "tools/logic-map/risk_pattern_report.py", src, "--output", str(d / "risk-pattern-report.md")],
        check=True,
    )


def main() -> int:
    cov_dir = Path(sys.argv[1])
    for name, spec in B.items():
        write(name, spec, cov_dir)
        print("wrote", name)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
