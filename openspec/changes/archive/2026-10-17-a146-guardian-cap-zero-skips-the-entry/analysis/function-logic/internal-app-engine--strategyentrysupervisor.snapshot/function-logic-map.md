# Function Logic Map: `Snapshot`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (797-820)
- Qualified: `StrategyEntrySupervisor.Snapshot`
- AST evidence: `ast.json`(현재) — 현재 AST branches 2 · returns 3
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 6f7c999a13c4abdc…) — 편집 전 갈래 수 2
- Risk scan: `risk-pattern-report.md`

**역할.** 시장별 읽기 전용 관측 값.

**a146 편집.** `GuardianCapZeroSkips` · `FirstGuardianCapZeroSkip` 두 칸 복사.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if s == nil \|\| !validStrategyMarket(market) {` (if 798) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|Snapshot'` · `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| B2 | `if worker == nil {` (if 804) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|Snapshot'` · `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |

Exact AST return positions: 799 · 805 · 807

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 5종 — 대표: `cap`, `len`, `s.mu.RLock`, `s.mu.RUnlock`, `validStrategyMarket`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a146 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a146 편집」에 적은 한 자리.
- High-risk impact: yes(전략 첫 다리 진입 분류 — 주문을 늘리지 않음: 건너뛴 진입은 브로커 0, 매도 · 손절 · 청산 경로 무관)
