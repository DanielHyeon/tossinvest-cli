# Function Logic Map: `RiskGuardian.IssueEntry`

- Source: `internal/execgw/riskguardian.go` (418-575, 편집 뒤)
- Qualified: `RiskGuardian.IssueEntry`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` 7bf1bbe7623c179e…) · `ast-base.json`(편집 전, base `a000cf29`, `source_sha256` edad4839b273c2b0…, 편집 전 분기수 14) — 현재 branches 14 · returns 13
- Risk scan: `risk-pattern-report.md`

**역할.** 레거시 진입 발급 문(체인 1회 → 원자 발급). 생산 호출 경로 0(Tracer · GuardianAdapter 생산 생성 0) — 잠든 문이지만 깨어나도 0 을 믿지 않게 같은 바닥을 건다.

**a137 편집.** 체인 앞에 `g.MeasureDailyLoss(ctx)` → `g.dailyLossFloor(req.Account, m, now, nil)` 로 `Account.DailyRealizedLoss` 를 바닥으로 바꾸고, 체인은 `evaluateChainWithDailyLoss` 로 돈다(측정 실패면 daily_loss 단 INPUT_UNAVAILABLE 상세를 측정 실패 문구로). 이 문은 FX 권한이 없어 다른 통화 손실이 있으면 측정 불가. 분기 수 불변(14), B4 의 조건 식만 바뀜. 구현 리뷰: 체인 거절은 `dailyLossChainRefusal` 로(측정 실패 표지 — 잠든 문이라 소비자 없음, 일관성). 재리뷰 P2-N1: `verdict, unmeasured := evaluateChainWithDailyLoss(...)` → `dailyLossChainRefusal(verdict, unmeasured)`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `req.Account.DailyRealizedLoss` | 기준 통화 · 음수 아님 | 호출자 | 무효면 대체 없이 체인이 INPUT_UNAVAILABLE |
| 원장 행 | CLOSED 관리 포지션 | `journal.DailyLossRows` | 실패 → 측정 실패 → daily_loss 단 거절 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if req.Collect == nil {` (if 419) | — | 분기 본문 참조 | `go test ./internal/execgw -run RiskGuardian` |
| B2 | `if err != nil {` (if 423) | — | 분기 본문 참조 | `go test ./internal/execgw -run RiskGuardian` |
| B3 | `if intent.Side != risk.SideBuy {` (if 426) | — | 분기 본문 참조 | `go test ./internal/execgw -run RiskGuardian` |
| B4 | `if verdict, unmeasured := evaluateChainWithDailyLoss(in, measureErr); !verdict.Allowed {` (if 444) | — | 분기 본문 참조 | `TestA137TheLedgerLossRefusesAZeroCallerAndEscalates` · `TestA137ALargerValidCallerValueWins` · `TestA137AnInvalidCallerValueIsNotReplaced` · `TestA137AMeasurementFailureNeverFallsBackToTheCaller` · `TestA137TheKillSwitchReasonSurvivesAMeasurementFailure` · `TestReachingTheDailyLossLimitTightensTheOperatingMode`(기존) · `TestA137AnInvalidEquityWithAMeasurementFailureStaysUnmarked` |
| B5 | `if !verdict.Allowed {` (if 453) | — | 분기 본문 참조 | `go test ./internal/execgw -run RiskGuardian` |
| B6 | `if err != nil {` (if 488) | — | 분기 본문 참조 | `go test ./internal/execgw -run RiskGuardian` |
| B7 | `if err != nil {` (if 492) | — | 분기 본문 참조 | `go test ./internal/execgw -run RiskGuardian` |
| B8 | `if err != nil {` (if 497) | — | 분기 본문 참조 | `go test ./internal/execgw -run RiskGuardian` |
| B9 | `if req.strategyPlan == nil {` (if 529) | — | 분기 본문 참조 | `go test ./internal/execgw -run RiskGuardian` |
| B10 | `} else {` (else 531) | — | 분기 본문 참조 | `go test ./internal/execgw -run RiskGuardian` |
| B11 | `if collectErr != nil {` (if 535) | — | 분기 본문 참조 | `go test ./internal/execgw -run RiskGuardian` |
| B12 | `if strategyErr == nil {` (if 540) | — | 분기 본문 참조 | `go test ./internal/execgw -run RiskGuardian` |
| B13 | `if err != nil {` (if 546) | — | 분기 본문 참조 | `go test ./internal/execgw -run RiskGuardian` |
| B14 | `if refused, ok := refusal.(*IssueRefusal); ok && refused.Stage == StageIssuance {` (if 553) | — | 분기 본문 참조 | `go test ./internal/execgw -run RiskGuardian` |

Exact AST return positions: 420 · 424 · 427 · 450 · 455 · 489 · 493 · 498 · 500 · 536 · 538 · 557 · 568

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `g.MeasureDailyLoss` | 원장 측정(a137) | 실패는 토큰에 봉인 | AST calls |
| `g.dailyLossFloor` | 바닥 규칙(a137) | 측정 실패 → 쓸 수 없는 값 | AST calls |
| `evaluateChainWithDailyLoss` | 체인 1회 | 거절 → B4 | AST calls |
| `g.escalateFor` | 한도 도달 강화 | 오류는 거절에 join | AST calls |

## State mutations and fallbacks

- 체인 거절 시 관측 기록 · 강화(기존). a137 은 원장 읽기 하나를 더함(쓰기 없음).

## Safety conclusion

- Safe edit boundary: 체인 입력의 일일 손실 칸과 체인 호출 한 줄. 발급 · 예약 무변경.
- High-risk impact: yes — 진입을 더 막는 방향만. 위험 축소 문(IssueReduction)은 측정하지 않음(`TestA137RiskReductionNeverMeasures`).
