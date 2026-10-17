# Function Logic Map: `TestA066LossLockAndBucketFailureNeverBlockRiskReducingPaths`

- Source: `internal/execgw/a066_entry_loss_lock_red_test.go` (212-345)
- Qualified: `TestA066LossLockAndBucketFailureNeverBlockRiskReducingPaths`
- AST evidence: `ast.json`(현재) — 현재 AST branches 18 · returns 1
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 8e40c50d9aab168a…) — 편집 전 갈래 수 18
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 — 손실 lock 이 위험 축소 경로를 막지 않음.

**a138 편집.** 요청에 `OpenPositionLimit` 넉넉한 칸 하나.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil {` (if 221) | — | 분기 본문 참조 | `go test ./internal/execgw -run A066` |
| B2 | `if err != nil {` (if 225) | — | 분기 본문 참조 | `go test ./internal/execgw -run A066` |
| B3 | `if err != nil {` (if 229) | — | 분기 본문 참조 | `go test ./internal/execgw -run A066` |
| B4 | `if err != nil {` (if 236) | — | 분기 본문 참조 | `go test ./internal/execgw -run A066` |
| B5 | `if err != nil {` (if 254) | — | 분기 본문 참조 | `go test ./internal/execgw -run A066` |
| B6 | `if !errors.As(err, &rejected) \|\| rejected.Reason == execgw.ReasonGuardianNonceReused \|\| after != before {` (if 261) | — | 분기 본문 참조 | `go test ./internal/execgw -run A066` |
| B7 | `for _, market := range []riskbucket.Market{riskbucket.MarketKR, riskbucket.MarketUS} {` (range 269) | — | 분기 본문 참조 | `go test ./internal/execgw -run A066` |
| B8 | `for _, horizon := range []riskbucket.Horizon{riskbucket.HorizonShort, riskbucket.HorizonMedium} {` (range 270) | — | 분기 본문 참조 | `go test ./internal/execgw -run A066` |
| B9 | `if err := activateEntryLossLock(ctx, j, "acct-7", market, horizon, clk.Now()); err != nil {` (if 271) | — | 분기 본문 참조 | `go test ./internal/execgw -run A066` |
| B10 | `if out, err := gw.Place(stopCtx, execgw.PlaceRequest{Intent: stop, Decision: exitDecision(t, j, clk, journal.KindPlace, stop.Market, stop.Symbol, stop.Side, stop.Quantity)}); err != nil \|\| out.State != journal.StateConfirmed {` (if 286) | — | 분기 본문 참조 | `go test ./internal/execgw -run A066` |
| B11 | `if out, err := gw.Place(emergencyCtx, execgw.PlaceRequest{Intent: emergency, Decision: exitDecision(t, j, clk, journal.KindPlace, emergency.Market, emergency.Symbol, emergency.Side, emergency.Quantity)}); err != nil \|\| out.State != journal.StateConfirmed {` (if 294) | — | 분기 본문 참조 | `go test ./internal/execgw -run A066` |
| B12 | `if _, err := gw.Cancel(cancelCtx, execgw.CancelRequest{` (if 300) | — | 분기 본문 참조 | `go test ./internal/execgw -run A066` |
| B13 | `if places, cancels, _ := broker.totals(); places != 2 \|\| cancels != 1 {` (if 307) | — | 분기 본문 참조 | `go test ./internal/execgw -run A066` |
| B14 | `if _, entered, err := j.EnterReconcile(ctx, journal.EnterReconcileRequest{` (if 312) | — | 분기 본문 참조 | `go test ./internal/execgw -run A066` |
| B15 | `if err != nil \|\| fill.FailClosed \|\| fill.Delta != "2" {` (if 326) | — | 분기 본문 참조 | `go test ./internal/execgw -run A066` |
| B16 | `for _, market := range []riskbucket.Market{riskbucket.MarketKR, riskbucket.MarketUS} {` (range 335) | — | 분기 본문 참조 | `go test ./internal/execgw -run A066` |
| B17 | `for _, horizon := range []riskbucket.Horizon{riskbucket.HorizonShort, riskbucket.HorizonMedium} {` (range 336) | — | 분기 본문 참조 | `go test ./internal/execgw -run A066` |
| B18 | `if err != nil \|\| changed \|\| held.Cause != "a066 2.7 contract test" {` (if 340) | — | 분기 본문 참조 | `go test ./internal/execgw -run A066` |

Exact AST return positions: 280

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 32종 — 대표: `Add`, `Format`, `UTC`, `a136ZeroManualBuy`, `activateEntryLossLock`, `assertRaisingRefused`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a138 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a138 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
