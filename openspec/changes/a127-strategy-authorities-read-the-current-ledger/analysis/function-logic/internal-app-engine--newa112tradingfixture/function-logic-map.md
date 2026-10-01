# Function Logic Map: `newA112TradingFixture`

- Source: `internal/app/engine/a112_owner_scope_trading_test.go` (`83`–`166`)
- Qualified: `newA112TradingFixture`
- AST evidence: `ast.json` (`source_sha256` 9f190f0d32244592…, revision current) — **편집 뒤**(시험 전용)
- Risk scan: `risk-pattern-report.md`
- AST branches 13 · return 3 · 호출 39

**역할.** a112 거래 픽스처. 비례 원칙(WORKFLOW 「비례 원칙」): 시험 전용 — 무거운 규율 대상 아님. 게이트가 수정 함수로 세므로 경량 번들을 둔다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 시험 임시 원장 · fixture | 시험 | 생산 작성자 경로 + 손상만 SQL | `t.Fatal`/`t.Errorf` |

## Branches and early returns

> AST 분기 열거(시험 코드 — 커버리지 블록 없음).

| Branch | 종류 | 조건 (원문) |
|---|---|---|
| B1 | if | `:86` `if symbols == nil {` |
| B2 | range | `:94` `for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {` |
| B3 | if | `:98` `if !ok {` |
| B4 | if | `:104` `if market == StrategyMarketKR {` |
| B5 | else | `:106` `} else {` |
| B6 | if | `:112` `if options.secondLane != nil {` |
| B7 | if | `:119` `if options.coordinatorOrder {` |
| B8 | if | `:125` `if err != nil {` |
| B9 | if | `:131` `if options.maxOpenExposure != "" {` |
| B10 | if | `:134` `if options.riskBudget != "" {` |
| B11 | if | `:139` `if err != nil {` |
| B12 | if | `:150` `if options.disallowFor != "" {` |
| B13 | if | `:154` `if err != nil \|\| config.Symbol != options.disallowFor {` |

## Calls and live bindings

시험 fixture · 생산 journal/riskbucket API 호출. 브로커 · 실계좌 호출 없음.

## State mutations and fallbacks

시험 임시 원장만.

## Safety conclusion

- **Safe edit boundary**: 위험 적재기를 실제 원장(admission 과 같은 `journal.Open` 원장)으로 단일화, 축소 stub 경로는 riskStub 에 보관(손상 주입 시험 전용) — a127 D4.
- **High-risk impact**: no — 시험. 시험 전용 — 생산 동작 무관.
