# Function Logic Map: `runMarket`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (926-1016)
- Qualified: `StrategyEntrySupervisor.runMarket`
- AST evidence: `ast.json`(현재) — 현재 AST branches 17 · returns 11
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 6f7c999a13c4abdc…) — 편집 전 갈래 수 17
- Risk scan: `risk-pattern-report.md`

**역할.** 시장 하나의 평가 루프 — 만료 · 허용 판정, 주기 실행(감시견), 결과에 따라 계속 · 삼킴 · 중앙 고장 · 시장 잠금.

**a146 편집.** 주기 ctx 에 a146 스킵 기록자를 싣는 문장 하나(a137 기록자 옆). 분기 · 잠금 · 삼킴 판정 무편집.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `select {` (select 927) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Supervisor|Fault'` · `TestA146ASkipNeverLatchesOrSwallowsOnEitherLaneKind` · `TestA137AnUnmarkedFaultStillLatchesOrIsSwallowedAsToday` · `TestA146ACapZeroEntryNoLongerLatchesAnEffectiveMarket` |
| B2 | `for {` (for 932) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Supervisor|Fault'` · `TestA146ASkipNeverLatchesOrSwallowsOnEitherLaneKind` · `TestA137AnUnmarkedFaultStillLatchesOrIsSwallowedAsToday` · `TestA146ACapZeroEntryNoLongerLatchesAnEffectiveMarket` |
| B3 | `select {` (select 933) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Supervisor|Fault'` · `TestA146ASkipNeverLatchesOrSwallowsOnEitherLaneKind` · `TestA137AnUnmarkedFaultStillLatchesOrIsSwallowedAsToday` · `TestA146ACapZeroEntryNoLongerLatchesAnEffectiveMarket` |
| B4 | `if expired {` (if 942) | latchMarket(만료) | 무편집 | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Supervisor|Fault'` · `TestA146ASkipNeverLatchesOrSwallowsOnEitherLaneKind` · `TestA137AnUnmarkedFaultStillLatchesOrIsSwallowedAsToday` · `TestA146ACapZeroEntryNoLongerLatchesAnEffectiveMarket` |
| B5 | `if err != nil {` (if 944) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Supervisor|Fault'` · `TestA146ASkipNeverLatchesOrSwallowsOnEitherLaneKind` · `TestA137AnUnmarkedFaultStillLatchesOrIsSwallowedAsToday` · `TestA146ACapZeroEntryNoLongerLatchesAnEffectiveMarket` |
| B6 | `if err := s.waitMarketRestart(ctx, restartNotBefore); err != nil {` (if 948) | — | 무편집 | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Supervisor|Fault'` · `TestA146ASkipNeverLatchesOrSwallowsOnEitherLaneKind` · `TestA137AnUnmarkedFaultStillLatchesOrIsSwallowedAsToday` · `TestA146ACapZeroEntryNoLongerLatchesAnEffectiveMarket` |
| B7 | `if ctx.Err() != nil {` (if 949) | — | return(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Supervisor|Fault'` · `TestA146ASkipNeverLatchesOrSwallowsOnEitherLaneKind` · `TestA137AnUnmarkedFaultStillLatchesOrIsSwallowedAsToday` · `TestA146ACapZeroEntryNoLongerLatchesAnEffectiveMarket` |
| B8 | `if !allowed {` (if 957) | — | continue(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Supervisor|Fault'` · `TestA146ASkipNeverLatchesOrSwallowsOnEitherLaneKind` · `TestA137AnUnmarkedFaultStillLatchesOrIsSwallowedAsToday` · `TestA146ACapZeroEntryNoLongerLatchesAnEffectiveMarket` |
| B9 | `if abandoned {` (if 965) | markAbandoned | 무편집 | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Supervisor|Fault'` · `TestA146ASkipNeverLatchesOrSwallowsOnEitherLaneKind` · `TestA137AnUnmarkedFaultStillLatchesOrIsSwallowedAsToday` · `TestA146ACapZeroEntryNoLongerLatchesAnEffectiveMarket` |
| B10 | `if cancelled {` (if 968) | — | return(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Supervisor|Fault'` · `TestA146ASkipNeverLatchesOrSwallowsOnEitherLaneKind` · `TestA137AnUnmarkedFaultStillLatchesOrIsSwallowedAsToday` · `TestA146ACapZeroEntryNoLongerLatchesAnEffectiveMarket` |
| B11 | `if err == nil {` (if 971) | — | continue — **스킵만 있던 주기는 여기로 옴**(a146 의 효과 지점, 분기 자체는 무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Supervisor|Fault'` · `TestA146ASkipNeverLatchesOrSwallowsOnEitherLaneKind` · `TestA137AnUnmarkedFaultStillLatchesOrIsSwallowedAsToday` · `TestA146ACapZeroEntryNoLongerLatchesAnEffectiveMarket` |
| B12 | `if refreshOnly {` (if 974) | recordSwallowedCycleError | continue(무편집 — 스킵은 오류가 아니므로 여기 오지 않음) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Supervisor|Fault'` · `TestA146ASkipNeverLatchesOrSwallowsOnEitherLaneKind` · `TestA137AnUnmarkedFaultStillLatchesOrIsSwallowedAsToday` · `TestA146ACapZeroEntryNoLongerLatchesAnEffectiveMarket` |
| B13 | `if isCentralStrategyIntegrity(err) && !s.blockEntryOnCentralIntegrity(worker) {` (if 992) | 진입 차단 | 중앙 신호(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Supervisor|Fault'` · `TestA146ASkipNeverLatchesOrSwallowsOnEitherLaneKind` · `TestA137AnUnmarkedFaultStillLatchesOrIsSwallowedAsToday` · `TestA146ACapZeroEntryNoLongerLatchesAnEffectiveMarket` |
| B14 | `if isCentralStrategyIntegrity(err) {` (if 998) | — | 중앙 신호(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Supervisor|Fault'` · `TestA146ASkipNeverLatchesOrSwallowsOnEitherLaneKind` · `TestA137AnUnmarkedFaultStillLatchesOrIsSwallowedAsToday` · `TestA146ACapZeroEntryNoLongerLatchesAnEffectiveMarket` |
| B15 | `if err != nil {` (if 1003) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Supervisor|Fault'` · `TestA146ASkipNeverLatchesOrSwallowsOnEitherLaneKind` · `TestA137AnUnmarkedFaultStillLatchesOrIsSwallowedAsToday` · `TestA146ACapZeroEntryNoLongerLatchesAnEffectiveMarket` |
| B16 | `if err := s.waitMarketRestart(ctx, restartNotBefore); err != nil {` (if 1007) | — | 무편집 | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Supervisor|Fault'` · `TestA146ASkipNeverLatchesOrSwallowsOnEitherLaneKind` · `TestA137AnUnmarkedFaultStillLatchesOrIsSwallowedAsToday` · `TestA146ACapZeroEntryNoLongerLatchesAnEffectiveMarket` |
| B17 | `if ctx.Err() != nil {` (if 1008) | — | return(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Supervisor|Fault'` · `TestA146ASkipNeverLatchesOrSwallowsOnEitherLaneKind` · `TestA137AnUnmarkedFaultStillLatchesOrIsSwallowedAsToday` · `TestA146ACapZeroEntryNoLongerLatchesAnEffectiveMarket` |

Exact AST return positions: 929 · 935 · 946 · 950 · 953 · 969 · 994 · 1000 · 1005 · 1009 · 1012

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 17종 — 대표: `ctx.Done`, `ctx.Err`, `invokeBoundedStrategyCycle`, `isCentralStrategyIntegrity`, `s.blockEntryOnCentralIntegrity`, `s.evaluationState`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a146 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a146 편집」에 적은 한 자리.
- High-risk impact: yes(전략 첫 다리 진입 분류 — 주문을 늘리지 않음: 건너뛴 진입은 브로커 0, 매도 · 손절 · 청산 경로 무관)
