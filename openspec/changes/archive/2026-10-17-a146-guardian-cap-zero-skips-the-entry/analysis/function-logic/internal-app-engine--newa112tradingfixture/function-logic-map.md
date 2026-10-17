# Function Logic Map: `newA112TradingFixture`

- Source: `internal/app/engine/a112_owner_scope_trading_test.go` (85-171)
- Qualified: `newA112TradingFixture`
- AST evidence: `ast.json`(현재) — 현재 AST branches 14 · returns 3
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 1ede06c9b62eeb92…) — 편집 전 갈래 수 13
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 fixture — 생산 적재기 · 실제 Guardian · 실제 원장 · dispatch 주기.

**a146 편집.** 옵션 `maxOrderNotional`(주문 금액 상한) 분기 하나 — notional 상한 0 실제 경로 시험용(a146 사람 결정 1).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if symbols == nil {` (if 88) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A112'` · `TestA146ANotionalCapZeroIsTheSameSkipOnTheRealPath` · `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` |
| B2 | `for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {` (range 96) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A112'` · `TestA146ANotionalCapZeroIsTheSameSkipOnTheRealPath` · `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` |
| B3 | `if !ok {` (if 100) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A112'` · `TestA146ANotionalCapZeroIsTheSameSkipOnTheRealPath` · `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` |
| B4 | `if market == StrategyMarketKR {` (if 106) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A112'` · `TestA146ANotionalCapZeroIsTheSameSkipOnTheRealPath` · `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` |
| B5 | `} else {` (else 108) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A112'` · `TestA146ANotionalCapZeroIsTheSameSkipOnTheRealPath` · `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` |
| B6 | `if options.secondLane != nil {` (if 114) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A112'` · `TestA146ANotionalCapZeroIsTheSameSkipOnTheRealPath` · `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` |
| B7 | `if options.coordinatorOrder {` (if 121) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A112'` · `TestA146ANotionalCapZeroIsTheSameSkipOnTheRealPath` · `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` |
| B8 | `if err != nil {` (if 127) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A112'` · `TestA146ANotionalCapZeroIsTheSameSkipOnTheRealPath` · `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` |
| B9 | `if options.maxOpenExposure != "" {` (if 133) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A112'` · `TestA146ANotionalCapZeroIsTheSameSkipOnTheRealPath` · `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` |
| B10 | `if options.riskBudget != "" {` (if 136) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A112'` · `TestA146ANotionalCapZeroIsTheSameSkipOnTheRealPath` · `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` |
| B11 | `if options.maxOrderNotional != "" {` (if 139) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A112'` · `TestA146ANotionalCapZeroIsTheSameSkipOnTheRealPath` · `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` |
| B12 | `if err != nil {` (if 144) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A112'` · `TestA146ANotionalCapZeroIsTheSameSkipOnTheRealPath` · `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` |
| B13 | `if options.disallowFor != "" {` (if 155) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A112'` · `TestA146ANotionalCapZeroIsTheSameSkipOnTheRealPath` · `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` |
| B14 | `if err != nil \|\| config.Symbol != options.disallowFor {` (if 159) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A112'` · `TestA146ANotionalCapZeroIsTheSameSkipOnTheRealPath` · `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` |

Exact AST return positions: 160 · 164 · 170

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 36종 — 대표: `a112AccountLoaderWith`, `a112ExtraEntryKRWith`, `a136ZeroManualBuyBoardInternal`, `accountLoader.collect`, `authority.AccountState`, `authority.FreshUntil`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a146 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a146 편집」에 적은 한 자리.
- High-risk impact: yes(전략 첫 다리 진입 분류 — 주문을 늘리지 않음: 건너뛴 진입은 브로커 0, 매도 · 손절 · 청산 경로 무관)
