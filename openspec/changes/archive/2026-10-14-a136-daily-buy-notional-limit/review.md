# a136 review

## 사람 결정 기록 (2026-10-14)

Manager 가 전달한 사용자 최종 결정 — 원문은 design.md 「사람 결정」. 요약: D1=D(KST 08:00 경계, `clock.AccountTradingDay` 한 곳 — a137 공유) ·
D2=c(예약 원장 집계, 0.7 실측으로 체결 없음 증명 가능 확인) · D4=≥ · D5=USD base 무값 기동 거부 · D9=b(수동 매수 합산). 2차 결정(같은 날): 「별도 change 로 나눌 필요 없음」 — 수동 매수 합산과 콘솔 칸 모두 a136 안. FX 조항은 a137 문구 그대로 a136 델타로.

## proposal-freeze 리뷰 (2026-10-14)

- 보이스: 분리된 컨텍스트의 적대 리뷰어 1(code-reviewer, 적대 Eng + CEO 범위 + 운영 DX 관점), 읽기 전용, HEAD f504cc52 + 워킹트리 WIP.
- 판정: **APPROVE-WITH-FIXES** — P1 3 · P2 10 · P3 6.
- 수용 · 수리: P1-1(Z3 범위 — design D10, 변이 M3 · M4), P1-2(통화 전환 지움 — `SaveEngineGateLimits`, 변이 M14), P2-1(발급일 ≠ 커밋일 거부, M7),
  P2-2(spec ① RELEASED 한정), P2-3(engine-safety MODIFIED), P2-4(`LimitsDigest` 결속, M15), P2-7(콘솔 칸 — 2차 사용자 결정으로 a136 안),
  P2-8(UNION 두 갈래), P2-9(`measure.gridPolicy`), P2-10(proposal 갱신), P3(기본값 대조 시험 · 인자 수 정리 · 빠진 시나리오 · FAILED_CONFIRMED 포함 사유 ·
  UNKNOWN_BROKER_STATE 매일 셈을 「거절되는 정상 입력」에).
- 잔재(기록만): P2-5(a137 과 일 키 적용 시각 — 같은 함수, 적용 시각은 각 change 정의), P2-6(계산 불가 거부의 관측 sub-reason 없음), P3 「오늘 합계/한도」 읽기 전용 표시.
- **P1-3 단계 순서 위반 — 사실로 인정.** `reserveRows` · `DefaultPolicy` · `Policy.Validate` · `mergeEngine` 은 FLM 전에 편집됐다. 수리: base 트리(f504cc52
  `git archive`)에서 편집 전 AST(`ast-base.json`)를 뽑아 번들을 만들고, 그 뒤의 편집(설정 쓰기 · 인터록 · 발급 문 셋 · digest · gridPolicy)은 번들 뒤에 했다.
  편집 전 AST 와 편집 뒤 AST 의 분기 차이는 각 번들 머리에 적었다. **Manager 지시(2026-10-14): 게이트 전 독립 리뷰가 이 4 함수를 base 리비전 FLM/BTM
  (`ast-base.json`) 기준으로 다시 대조한다.**

## Pre-Edit Gate

```text
Pre-Edit Gate:
- change id / task id: a136-daily-buy-notional-limit / 1.1~1.4
- 대상 심볼(편집하는 기존 함수):
    internal/journal/reservations.go       reserveRows                 (판정 호출 한 줄)
    internal/execgw/riskguardian.go        RiskGuardian.IssueEntry      (collectIssue 클로저에 한도 한 칸)
    internal/execgw/riskguardian_qfinal.go RiskGuardian.IssuePrecheckedQFinalEntry (같음)
    internal/execgw/riskguardian_first_leg.go RiskGuardian.IssuePrecheckedQFinalCampaignFirstLeg (같음)
    internal/risk/input.go                 DefaultPolicy · Policy.Validate
    internal/app/engine/guardian_wiring.go riskPolicyFromAutomationGate
    internal/app/engine/interlock.go       newAutomationStatus · verifyGate · acceptanceDetail · recordGateSettings
    internal/config/engine.go              mergeEngine
  새 파일(새 leaf): internal/clock/account_day.go · internal/journal/daily_buy.go · internal/config/daily_buy.go ·
                   internal/execgw/daily_buy.go · internal/app/engine/daily_buy_interlock.go
- CodeGraph definition/callers/callees/impact: analysis/code-context/codegraph-baseline.md (reserveRows 호출자 5 = 모든 예약 tx)
- CodeGraphContext 후보와 evidence reconciliation: analysis/code-context/ (불일치 0)
- 변이 원장: `analysis/mutation/round1.log`(18 중 3 이 컴파일 실패를 CAUGHT 로 오분류 · 2 SURVIVED) → 하네스 · 시험 수리 →
  `round2.log` **18/18 CAUGHT**(무변이 대조군 GREEN 선행).
- 기존 동작 파악 근거: HEAD 읽기 + analysis/ast-evidence/ + 기존 시험(reservations_test · riskguardian_test · interlock 시험 · guardian_wiring_internal_test)
- Function Logic Map / Branch Test Map: analysis/function-logic/<pkg>--<func>/ (편집 뒤 재추출)
- upstream 상속 테스트 영향: no — 게이트 OFF 는 Guardian 을 만들지 않음(engine.go:536-549), 매도 경로 무편집
- 실패 테스트 선행 작성: yes (analysis/red/)
- 설정·DB·journal 변경과 rollback: 설정 키 1(선택, 없음 = 기본) · 원장 스키마 변경 0 · rollback = 이전 이미지(키는 옛 빌드가 무시)
- 안전 불변식 §0 위반 여부 검토: 통과 — 매도 · 손절 · 청산은 reserveRows 를 지나지 않음(RISK_REDUCING 은 예약 없음), 판정은 거부 방향만 추가
```

## 구현 독립 적대 리뷰 (2026-10-14, tasks 2.3)

- 보이스: 분리된 컨텍스트 적대 리뷰어 1(code-reviewer — 적대 Eng · 안전 불변식 · 시험 반증), 사본(`git archive`)에서 재현 시험 · 자체 변이 실행, 저장소 무변경.
- 판정: **APPROVE-WITH-FIXES** — P0 0 · P1 2 · P2 4 · P3 6. 안전 불변식 4 직접 위반 없음(IssueReduction 무편집 · 예약 없음, 보조 실행자 감독 밖).
- 수용 · 수리(이 커밋):
  - P1-1 콘솔 고급 폼 통화 전환 시 미리 채워진 옛 숫자 이월(≈1,400배 완화) → 핸들러가 저장 전 파일과 비교해 「통화 바뀜 + 숫자 그대로(>0)」면 값을 보내지 않아 writer 가
    지움. audit 값에 통화를 붙여 숫자가 같아도 통화 변화가 남음. 시험 `TestA136ACurrencySwitchDoesNotCarryThePrefilledFigure`, 변이 M39 CAUGHT.
  - P1-2 per-order `"CLOSED"` 는 API 에 없음(그룹 라벨) → OpenAPI OrderStatus 열거로 종결 판정, 픽스처를 실제 열거로 재작성. 시험 `TestA136TerminalStatusesCountOnlyWhatFilled`, M37 CAUGHT.
  - P2-1 금액 기반 시장가 `orderAmount` → max(orderAmount, 체결액). 시험 `TestA136AnAmountBasedMarketBuyCountsItsOrderAmount`, M38 CAUGHT.
  - P2-2 중복 제거를 PLACE · AMEND 로 한정. 시험 `TestA136EngineOrderIDsExcludeCancelTargets`, M40 CAUGHT.
  - P2-3 R3 · R4 · R5 · R7 고정 → `TestA136Z3NeedsASnapshotAfterTheAttemptAndOneOwner` · `TestA136TheManualBuyFreshnessBoundIsTwoMinutes`, M33 · M34 · M35 · M36 CAUGHT.
    R2(`matching != 1` → `< 1`)는 **동등 변이**: scoped_fill_snapshots PK 가 정확히 결합 키(계좌 · 시장 · 거래일 · 종목 · 방향 · 주문 id)라 범위 안 2행이 구조적으로 불가.
    R9(Z2 에 UNRESOLVED 추가)는 그 보유가 HELD 로 남아 HELD 갈래가 세므로 부분 동등(기록만).
  - P3-1 design D6 digest 문장 정정. P3-3 CLOSED 페이지 크기 100 · from/to KST(OpenAPI 명시)로 `[미측정]` 닫음. P3-5 「거절되는 정상 입력」에 추가.
  - P3-6 `reserveRows` BTM 에 편집 전 → 뒤 분기 번호 대응표 · B4 「a136 신규」 표기. 4 함수(FLM 전 편집) 대조: 리뷰어가 ast-base.json 과 대조해 기존 분기 동작 변화 없음 확인
    (`Policy.Validate` 11=11 · `DefaultPolicy` 0=0 · `mergeEngine` 5=5 · `reserveRows` 6→7 은 신규 if 하나와 위치 이동뿐).
- 잔재(기록, Manager 판단):
  - P2-4 쓰기 잠금 안 Z3 후보 질의의 `mutation_attempts.target_order_id` 무색인 스캔 · 30초마다 `EngineBrokerOrderIDs` 전체 스캔 — 시간 `[미측정]`. 색인 추가는 스키마 변경이라 이 change 에서
    하지 않음. 후보는 「오늘 발급 · RELEASED · BROKER_TERMINAL · CONFIRMED PLACE」 행에만 돔.
  - P3-2 FX 검증이 precheck 가 아니라 발급 시점(design D9 구현 기록) — 조항 문구와 위치 차이. **Manager 수용(2026-10-14): 발급 시점 검증은 더 늦고 더 엄격하므로 잔재로 받아들임.**
  - P3-4 발급일 ≠ 커밋일 거부가 LIMIT_REACHED 로 관측되어 실제 한도 도달과 sub-reason 이 같음(detail 은 다름).
- 변이: `analysis/mutation/round5.log` · `round5-m37.log` · `round5-console.log` — 새 변이 M33~M40 전부 CAUGHT(M37 은 컴파일되는 형태로 재측정).
- 착지 주의: a137 이 a136 착지 뒤 `riskguardian.go` · `riskguardian_qfinal.go` · `riskguardian_qfinal_test.go` · `strategy_entry_gate_authority.go` 를 고쳤다
  (a1cfe0c2). a136 번들은 그 파일들을 **a136 의 마지막 Go 커밋 시점 소스**로 고정해 두었으므로(재추출하지 않음), 게이트는 a136 착지 기록
  (`--record-landing`)이나 사람 승인 base 재고정이 필요함 — Manager 절차.

## 사람 승인 base 재고정 (2026-10-14)

- 승인: **Manager, 2026-10-14**(docs/WORKFLOW.md 「사람 승인 base 재고정」 2 — Manager 승인으로 충분).
- 옛 base `f504cc52` → 새 base `6cd36e6d2f70`(a136 의 844a042c 직전 a137 마지막 커밋). 커밋 `1539c675`(base-commit.txt 단독, 영수증 포함).
- 귀속 실측(Manager, 옛 base 에서): required **68** · missing **10** — 빠진 10 은 전부 a137 형제 착지 몫(a1cfe0c2 가 고친 riskguardian.go · riskguardian_qfinal.go ·
  riskguardian_qfinal_test.go · strategy_entry_gate_authority.go 의 함수). 옛 base..새 base 사이 a136 자기 Go 커밋 5(82edde52 · c8783ad9 · de9617ab · 3685583d ·
  a000cf29)가 고친 기존 함수는 전부 fresh 번들로 덮여 있음.
- 새 base 에서의 수리: 844a042c 가 고친 함수 11(생산 5: recordDailyBuySave · manualBuyRefresh.collect · Journal.EngineBrokerOrderIDs · Summarize · valueOf,
  시험 6)에 번들을 만듦(편집 전 AST = 새 base 6cd36e6d). 이웃 편집만 받은 셋(Summarize · 시험 둘)은 검사기 규칙대로 `revision: base`. a137 이 고친 파일의
  번들 다섯은 현재 소스로 재추출(모양 비교: 넷은 분기 · 반환 · 호출 동일, IssueEntry 는 a137 의 일일 손실 측정 호출이 늘었고 a136 편집 지점은 무변경).
- `python3 tools/logic-map/check_analysis.py --change a136-daily-buy-notional-limit` → **rc 0**(「evidence complete or diff-proven exempt」, required 13).

## 없는 시험을 인용했던 기록 (정정)

`RiskGuardian.IssueEntry` · `IssuePrecheckedQFinalEntry` · `IssuePrecheckedQFinalCampaignFirstLeg` 번들의 BTM 이 `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` 을
인용했는데 **그 시험은 어디에도 없었다**(`git log -S` 빈 결과). 번들을 쓰던 때 계획한 이름을 적고 실제 시험은 문별 이름으로 만들었으며, 검사기의
「tracked file 에 없는 시험」 경고를 「커밋하면 풀린다」고 보고 걸러 냈다 — 그 경고 하나는 커밋으로 풀릴 수 없는 것이었다. 인용을 실제 시험으로 바꿈:
`TestA136IssueEntryRefusesTheBuyThatReachesTheDailyCeiling` · `TestA136QFinalEntryRefusesTheBuyThatReachesTheDailyCeiling` · `TestA136QFinalCampaignFirstLegRefusesTheBuyThatReachesTheDailyCeiling`
(tossos_testseams). 세 시험이 문을 실제로 묶는다는 증거는 변이 M11(문이 다른 상한을 싣게 함) CAUGHT — 세 시험 모두 실패(round4.log).

## P1-3 재대조 — FLM 전에 편집된 4 함수 (base 리비전 FLM/BTM 대 편집 뒤 AST)

`ast-base.json`(base f504cc52 트리 `git archive` 에서 추출) 과 `ast.json`(현재) 의 분기 목록을 순서 정렬(difflib)로 맞추고 조건은 소스 줄 그대로 비교함.
기존 분기는 전부 조건 문자열이 같고, 바뀐 분기는 0, 새 분기는 `reserveRows` 의 B4(`checkDailyBuyNotional`) 하나 — 그 함수의 BTM 은 편집 전 → 뒤 번호 대응과
B4 「a136 신규」 를 적고 있음.

#### `reserveRows` (internal/journal/reservations.go) — base 분기 6 → 편집 뒤 7

| base | 편집 뒤 | 조건(소스 줄) | 판정 |
|---|---|---|---|
| B1 | B1 | `if err := checkDecisionReservable(ctx, tx, plan, now); err != nil {` | 같음(조건 문자열 동일) |
| B2 | B2 | `if err != nil {` | 같음(조건 문자열 동일) |
| B3 | B3 | `if err := plan.checkLimits(held); err != nil {` | 같음(조건 문자열 동일) |
| — | B4 | `if err := checkDailyBuyNotional(ctx, tx, req, plan, now); err != nil {` | **a136 신규** |
| B4 | B5 | `for _, row := range plan.rows {` | 같음(조건 문자열 동일) |
| B5 | B6 | `if _, err := tx.ExecContext(ctx,` | 같음(조건 문자열 동일) |
| B6 | B7 | `if err != nil {` | 같음(조건 문자열 동일) |

#### `DefaultPolicy` (internal/risk/input.go) — base 분기 0 → 편집 뒤 0

| base | 편집 뒤 | 조건(소스 줄) | 판정 |
|---|---|---|---|
| — | — | 직선 함수(분기 0) | 분기 없음 — 편집은 필드 대입 |

#### `Validate` (internal/risk/input.go) — base 분기 11 → 편집 뒤 11

| base | 편집 뒤 | 조건(소스 줄) | 판정 |
|---|---|---|---|
| B1 | B1 | `if err != nil {` | 같음(조건 문자열 동일) |
| B2 | B2 | `for _, m := range []struct {` | 같음(조건 문자열 동일) |
| B3 | B3 | `if _, err := requirePositiveMoney(m.label, m.value, currency); err != nil {` | 같음(조건 문자열 동일) |
| B4 | B4 | `if err != nil {` | 같음(조건 문자열 동일) |
| B5 | B5 | `if quantity.Sign() <= 0 {` | 같음(조건 문자열 동일) |
| B6 | B6 | `if err != nil {` | 같음(조건 문자열 동일) |
| B7 | B7 | `if ratio.Sign() <= 0 || ratio.Cmp(big.NewRat(1, 1)) > 0 {` | 같음(조건 문자열 동일) |
| B8 | B8 | `if err != nil {` | 같음(조건 문자열 동일) |
| B9 | B9 | `if minRR.Sign() <= 0 {` | 같음(조건 문자열 동일) |
| B10 | B10 | `if p.Reentry.MaxEntriesPerSymbol < 1 {` | 같음(조건 문자열 동일) |
| B11 | B11 | `if p.Reentry.Cooldown < 0 {` | 같음(조건 문자열 동일) |

#### `mergeEngine` (internal/config/engine.go) — base 분기 5 → 편집 뒤 5

| base | 편집 뒤 | 조건(소스 줄) | 판정 |
|---|---|---|---|
| B1 | B1 | `if raw == nil {` | 같음(조건 문자열 동일) |
| B2 | B2 | `if raw.Autostart != nil {` | 같음(조건 문자열 동일) |
| B3 | B3 | `if raw.ExitPolicy != nil {` | 같음(조건 문자열 동일) |
| B4 | B4 | `if raw.AutomationGate == nil {` | 같음(조건 문자열 동일) |
| B5 | B5 | `if gate.Enabled != nil {` | 같음(조건 문자열 동일) |
