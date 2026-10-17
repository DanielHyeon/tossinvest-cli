# Function Logic Map: `NewStrategyEntrySupervisor`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (662-740)
- Qualified: `NewStrategyEntrySupervisor`
- AST evidence: `ast.json`(현재) — 현재 AST branches 18 · returns 14
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 6f7c999a13c4abdc…) — 편집 전 갈래 수 18
- Risk scan: `risk-pattern-report.md`

**역할.** 닫힌 KR/US 조립 검증과 감독자 생성.

**a146 편집.** 옵션 `EntrySkipNotice` 를 감독자 칸으로 옮김(한 칸).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if depth == 0 {` (if 664) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|Supervisor'` · `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| B2 | `if depth < 1 \|\| depth > MaximumStrategyQueueDepth {` (if 667) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|Supervisor'` · `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| B3 | `if cycleLimit == 0 {` (if 671) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|Supervisor'` · `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| B4 | `if cycleLimit < 0 \|\| cycleLimit > MaximumStrategyCycleLimit {` (if 674) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|Supervisor'` · `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| B5 | `if len(opts.Workers) != 2 {` (if 677) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|Supervisor'` · `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| B6 | `if clk == nil {` (if 681) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|Supervisor'` · `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| B7 | `if now.IsZero() {` (if 685) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|Supervisor'` · `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| B8 | `for _, descriptor := range opts.Workers {` (range 690) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|Supervisor'` · `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| B9 | `if !validStrategyMarket(descriptor.Market) {` (if 691) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|Supervisor'` · `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| B10 | `if _, duplicate := workers[descriptor.Market]; duplicate {` (if 694) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|Supervisor'` · `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| B11 | `if descriptor.Effective && descriptor.Cycle == nil {` (if 697) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|Supervisor'` · `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| B12 | `if descriptor.Effective && (descriptor.AuthorityGeneration == 0 \|\| descriptor.AuthorityExpiresAt.IsZero() \|\|` (if 700) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|Supervisor'` · `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| B13 | `if descriptor.Effective && ((descriptor.RestartAttempt == 0 && (descriptor.FirstRefusal != StrategyWorkerRefusalNone \|\| !descriptor.RestartNotBefore.IsZero())) \|\|` (if 704) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|Supervisor'` · `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| B14 | `if descriptor.PollInterval < 0 \|\| descriptor.PollInterval > MaximumStrategyCycleLimit \|\|` (if 708) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|Supervisor'` · `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| B15 | `if !descriptor.Effective && (descriptor.AuthorityGeneration != 0 \|\| !descriptor.AuthorityExpiresAt.IsZero() \|\| descriptor.EvidenceDigest != "" \|\| descriptor.LatchRevision != 0) {` (if 714) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|Supervisor'` · `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| B16 | `if !descriptor.Effective && (descriptor.FirstRefusal != StrategyWorkerRefusalNone \|\| descriptor.RestartAttempt != 0 \|\| !descriptor.RestartNotBefore.IsZero()) {` (if 717) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|Supervisor'` · `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| B17 | `for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {` (range 730) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|Supervisor'` · `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| B18 | `if workers[market] == nil {` (if 731) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|Supervisor'` · `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |

Exact AST return positions: 668 · 675 · 678 · 686 · 692 · 695 · 698 · 702 · 706 · 712 · 715 · 718 · 732 · 736

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 13종 — 대표: `clk.Now`, `clock.System`, `descriptor.AuthorityExpiresAt.IsZero`, `descriptor.RestartNotBefore.IsZero`, `errors.New`, `fmt.Errorf`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a146 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a146 편집」에 적은 한 자리.
- High-risk impact: yes(전략 첫 다리 진입 분류 — 주문을 늘리지 않음: 건너뛴 진입은 브로커 0, 매도 · 손절 · 청산 경로 무관)
