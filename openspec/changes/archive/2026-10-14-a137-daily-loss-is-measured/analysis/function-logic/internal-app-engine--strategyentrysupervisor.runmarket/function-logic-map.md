# Function Logic Map: `StrategyEntrySupervisor.runMarket`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (910-998, 편집 뒤)
- Qualified: `StrategyEntrySupervisor.runMarket`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` 6f7c999a13c4abdc…) · `ast-base.json`(편집 전, base `63fd0c92`, `source_sha256` 9cb510c1c9c7f44a…, 편집 전 분기수 17) — 현재 branches 17 · returns 11
- Risk scan: `risk-pattern-report.md`

**역할.** 한 시장 감독 루프 — 주기 실행과 오류 갈래(삼킴 · 잠금).

**a137 편집.** 재리뷰 P1-N1: 주기 ctx 에 이 시장의 측정 불가 기록자를 실음(`withStrategyDailyLossRecorder`) — 분기 · 갈래 불변(17). 전달 자리가 조용히 끝낸 측정 불가 주기가 스냅숏 계수 · 계좌 일 키 알림에 남음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| worker | 시장 런타임 | 감독자 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `select {` (select 911) | — | 분기 본문 참조 | 기존 감독자 시험 · `TestA137TheDailyLossStopsNeverLatchOrSwallowOnEitherLaneKind` |
| B2 | `for {` (for 916) | — | 분기 본문 참조 | 기존 감독자 시험 · `TestA137TheDailyLossStopsNeverLatchOrSwallowOnEitherLaneKind` |
| B3 | `select {` (select 917) | — | 분기 본문 참조 | 기존 감독자 시험 · `TestA137TheDailyLossStopsNeverLatchOrSwallowOnEitherLaneKind` |
| B4 | `if expired {` (if 926) | — | 분기 본문 참조 | 기존 감독자 시험 · `TestA137TheDailyLossStopsNeverLatchOrSwallowOnEitherLaneKind` |
| B5 | `if err != nil {` (if 928) | — | 분기 본문 참조 | 기존 감독자 시험 · `TestA137TheDailyLossStopsNeverLatchOrSwallowOnEitherLaneKind` |
| B6 | `if err := s.waitMarketRestart(ctx, restartNotBefore); err != nil {` (if 932) | — | 분기 본문 참조 | 기존 감독자 시험 · `TestA137TheDailyLossStopsNeverLatchOrSwallowOnEitherLaneKind` |
| B7 | `if ctx.Err() != nil {` (if 933) | — | 분기 본문 참조 | 기존 감독자 시험 · `TestA137TheDailyLossStopsNeverLatchOrSwallowOnEitherLaneKind` |
| B8 | `if !allowed {` (if 941) | — | 분기 본문 참조 | 기존 감독자 시험 · `TestA137TheDailyLossStopsNeverLatchOrSwallowOnEitherLaneKind` |
| B9 | `if abandoned {` (if 947) | — | 분기 본문 참조 | 기존 감독자 시험 · `TestA137TheDailyLossStopsNeverLatchOrSwallowOnEitherLaneKind` |
| B10 | `if cancelled {` (if 950) | — | 분기 본문 참조 | 기존 감독자 시험 · `TestA137TheDailyLossStopsNeverLatchOrSwallowOnEitherLaneKind` |
| B11 | `if err == nil {` (if 953) | — | 분기 본문 참조 | `TestA137AnUnmeasurableDayIsCountedAndAnnouncedOncePerAccountDay`(effective · refreshOnly) · `TestA137TheDailyLossStopsNeverLatchOrSwallowOnEitherLaneKind` |
| B12 | `if refreshOnly {` (if 956) | — | 분기 본문 참조 | `TestA137AnUnmarkedFaultStillLatchesOrIsSwallowedAsToday` · 기존 `TestARefreshOnlyWorkerCountsTheCycleErrorsItSwallows` |
| B13 | `if isCentralStrategyIntegrity(err) && !s.blockEntryOnCentralIntegrity(worker) {` (if 974) | — | 분기 본문 참조 | 기존 감독자 시험 · `TestA137TheDailyLossStopsNeverLatchOrSwallowOnEitherLaneKind` |
| B14 | `if isCentralStrategyIntegrity(err) {` (if 980) | — | 분기 본문 참조 | 기존 감독자 시험 · `TestA137TheDailyLossStopsNeverLatchOrSwallowOnEitherLaneKind` |
| B15 | `if err != nil {` (if 985) | — | 분기 본문 참조 | 기존 감독자 시험 · `TestA137TheDailyLossStopsNeverLatchOrSwallowOnEitherLaneKind` |
| B16 | `if err := s.waitMarketRestart(ctx, restartNotBefore); err != nil {` (if 989) | — | 분기 본문 참조 | 기존 감독자 시험 · `TestA137TheDailyLossStopsNeverLatchOrSwallowOnEitherLaneKind` |
| B17 | `if ctx.Err() != nil {` (if 990) | — | 분기 본문 참조 | 기존 감독자 시험 · `TestA137TheDailyLossStopsNeverLatchOrSwallowOnEitherLaneKind` |

Exact AST return positions: 913 · 919 · 930 · 934 · 937 · 951 · 976 · 982 · 987 · 991 · 994

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `withStrategyDailyLossRecorder` | 측정 불가 기록자(a137) | — | AST calls |
| `invokeBoundedStrategyCycle` | 주기 | 기존 | AST calls |

## State mutations and fallbacks

- worker 계수(기록자 경유, g.mu 아래).

## Safety conclusion

- Safe edit boundary: 주기 ctx 하나.
- High-risk impact: yes(전략 감독) — 잠금 · 삼킴 판정 불변, 관측만 추가.
