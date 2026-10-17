# Function Logic Map: `NewRefreshingPairedStrategyEntrySupervisor`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (399-428)
- Qualified: `Context.NewRefreshingPairedStrategyEntrySupervisor`
- AST evidence: `ast.json`(현재) — 현재 AST branches 4 · returns 4
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 6f7c999a13c4abdc…) — 편집 전 갈래 수 4
- Risk scan: `risk-pattern-report.md`

**역할.** 생산 감독자(권한 갱신 전용 worker 둘) 생성.

**a146 편집.** `EntrySkipNotice: c.strategyEntrySkipNotice()` 배선 한 줄.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if c == nil \|\| clk == nil {` (if 400) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|ProductionStrategySupervisor'` · `TestA146TheProductionSupervisorCarriesTheSkipNotice` · `TestA137TheProductionSupervisorCarriesTheNotice` |
| B2 | `if c.Entry == nil {` (if 404) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|ProductionStrategySupervisor'` · `TestA146TheProductionSupervisorCarriesTheSkipNotice` · `TestA137TheProductionSupervisorCarriesTheNotice` |
| B3 | `for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {` (range 408) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|ProductionStrategySupervisor'` · `TestA146TheProductionSupervisorCarriesTheSkipNotice` · `TestA137TheProductionSupervisorCarriesTheNotice` |
| B4 | `if err != nil {` (if 421) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|ProductionStrategySupervisor'` · `TestA146TheProductionSupervisorCarriesTheSkipNotice` · `TestA137TheProductionSupervisorCarriesTheNotice` |

Exact AST return positions: 401 · 405 · 422 · 427

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 9종 — 대표: `NewStrategyEntrySupervisor`, `append`, `c.productionStrategyCycle`, `c.strategyDailyLossNotice`, `c.strategyProjectionMu.Lock`, `c.strategyProjectionMu.Unlock`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a146 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a146 편집」에 적은 한 자리.
- High-risk impact: yes(전략 첫 다리 진입 분류 — 주문을 늘리지 않음: 건너뛴 진입은 브로커 0, 매도 · 손절 · 청산 경로 무관)
