# Function Logic Map: `TestA066LossLockAndBucketFailureNeverBlockRiskReducingPaths`

- Source: `internal/execgw/a066_entry_loss_lock_red_test.go` (212-344, 편집 뒤)
- Qualified: `TestA066LossLockAndBucketFailureNeverBlockRiskReducingPaths`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 43d523c8b1d98695…, 편집 전 분기 수 18) — 현재 branches 18 · returns 1
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 fixture/시험 함수(생산 코드 아님).

**a136 편집.** 시험 fixture 에 a136 의 새 필수 칸을 채움 — EXPOSURE_RAISING 예약에 넉넉한 `DailyBuyLimit`(이 시험이 재는 것과 무관)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| — | — | 시험 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil {` (if 221) | — | 분기 본문 참조 | `go test ./internal/execgw` (이 함수를 쓰는 기존 시험) |
| B2 | `if err != nil {` (if 225) | — | 분기 본문 참조 | `go test ./internal/execgw` (이 함수를 쓰는 기존 시험) |
| B3 | `if err != nil {` (if 229) | — | 분기 본문 참조 | `go test ./internal/execgw` (이 함수를 쓰는 기존 시험) |
| B4 | `if err != nil {` (if 236) | — | 분기 본문 참조 | `go test ./internal/execgw` (이 함수를 쓰는 기존 시험) |
| B5 | `if err != nil {` (if 253) | — | 분기 본문 참조 | `go test ./internal/execgw` (이 함수를 쓰는 기존 시험) |
| B6 | `if !errors.As(err, &rejected) \|\| rejected.Reason == execgw.ReasonGuardianNonceReused \|\| after != before {` (if 260) | — | 분기 본문 참조 | `go test ./internal/execgw` (이 함수를 쓰는 기존 시험) |
| B7 | `for _, market := range []riskbucket.Market{riskbucket.MarketKR, riskbucket.MarketUS} {` (range 268) | — | 분기 본문 참조 | `go test ./internal/execgw` (이 함수를 쓰는 기존 시험) |
| B8 | `for _, horizon := range []riskbucket.Horizon{riskbucket.HorizonShort, riskbucket.HorizonMedium} {` (range 269) | — | 분기 본문 참조 | `go test ./internal/execgw` (이 함수를 쓰는 기존 시험) |
| B9 | `if err := activateEntryLossLock(ctx, j, "acct-7", market, horizon, clk.Now()); err != nil {` (if 270) | — | 분기 본문 참조 | `go test ./internal/execgw` (이 함수를 쓰는 기존 시험) |
| B10 | `if out, err := gw.Place(stopCtx, execgw.PlaceRequest{Intent: stop, Decision: exitDecision(t, j, clk, journal.KindPlace, stop.Market, stop.Sy` (if 285) | — | 분기 본문 참조 | `go test ./internal/execgw` (이 함수를 쓰는 기존 시험) |
| B11 | `if out, err := gw.Place(emergencyCtx, execgw.PlaceRequest{Intent: emergency, Decision: exitDecision(t, j, clk, journal.KindPlace, emergency.` (if 293) | — | 분기 본문 참조 | `go test ./internal/execgw` (이 함수를 쓰는 기존 시험) |
| B12 | `if _, err := gw.Cancel(cancelCtx, execgw.CancelRequest{` (if 299) | — | 분기 본문 참조 | `go test ./internal/execgw` (이 함수를 쓰는 기존 시험) |
| B13 | `if places, cancels, _ := broker.totals(); places != 2 \|\| cancels != 1 {` (if 306) | — | 분기 본문 참조 | `go test ./internal/execgw` (이 함수를 쓰는 기존 시험) |
| B14 | `if _, entered, err := j.EnterReconcile(ctx, journal.EnterReconcileRequest{` (if 311) | — | 분기 본문 참조 | `go test ./internal/execgw` (이 함수를 쓰는 기존 시험) |
| B15 | `if err != nil \|\| fill.FailClosed \|\| fill.Delta != "2" {` (if 325) | — | 분기 본문 참조 | `go test ./internal/execgw` (이 함수를 쓰는 기존 시험) |
| B16 | `for _, market := range []riskbucket.Market{riskbucket.MarketKR, riskbucket.MarketUS} {` (range 334) | — | 분기 본문 참조 | `go test ./internal/execgw` (이 함수를 쓰는 기존 시험) |
| B17 | `for _, horizon := range []riskbucket.Horizon{riskbucket.HorizonShort, riskbucket.HorizonMedium} {` (range 335) | — | 분기 본문 참조 | `go test ./internal/execgw` (이 함수를 쓰는 기존 시험) |
| B18 | `if err != nil \|\| changed \|\| held.Cause != "a066 2.7 contract test" {` (if 339) | — | 분기 본문 참조 | `go test ./internal/execgw` (이 함수를 쓰는 기존 시험) |

Exact AST return positions: 279

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | fixture 구성 | — | AST |

## State mutations and fallbacks

- 시험 데이터만.

## Safety conclusion

- Safe edit boundary: 칸 추가 — 시험이 단언하는 값 무변경.
- High-risk impact: no(시험 전용, 비례 원칙)
