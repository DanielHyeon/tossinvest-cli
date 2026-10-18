# Function Logic Map: `NewPairedStrategyEntryProductionAssembly`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (325-406)
- Qualified: `Context.NewPairedStrategyEntryProductionAssembly`
- AST evidence: `ast.json`(편집 전(= base 소스)) — 현재 AST branches 8 · returns 6
- 편집 전 기준: `ast-base.json`(base `7b1748481`, `source_sha256` 714ca3ce3fed5bce…) — 편집 전 갈래 수 8
- Risk scan: `risk-pattern-report.md`

**역할.** 조립 감독자 생성 — 권한 파도(일정 · 후보 · 경로 · FX · 레인 · 제안 · 위험 · 계좌)를 모아 첫 다리 승인 · dispatch 주기 · 두 시장 worker 를 세우고 감독자를 만들어 투영을 발표함. 감독자는 투영에만 쓰이고 Run 되지 않음(design F9).

**a157 편집.** 394 행 감독자 옵션에 `DailyLossNotice: c.strategyDailyLossNotice()` · `EntrySkipNotice: c.strategyEntrySkipNotice()` 를 더하고, `c.Entry != nil` 일 때만 `opts.EntryGate = c.Entry`(타입 있는 nil 을 인터페이스에 넣지 않음 — design F10 · D4 · D5). 새 분기 하나.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD(설계 근거는 design F1~F10) | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if c == nil {` (if 326) | — | 조립 불가 오류(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A112|Refresh|Wave|ProductionAssembly'` · `TestTheMarketThatLeadsAWaveAlwaysPublishesIt` |
| B2 | `if c.Journal != nil && strings.TrimSpace(c.Journal.Path()) != "" {` (if 332) | — | 원장 경로 유도(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A112|Refresh|Wave|ProductionAssembly'` · `TestTheMarketThatLeadsAWaveAlwaysPublishesIt` |
| B3 | `if journalPath != "" {` (if 343) | — | 증거 DB 경로 유도(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A112|Refresh|Wave|ProductionAssembly'` · `TestTheMarketThatLeadsAWaveAlwaysPublishesIt` |
| B4 | `if err != nil {` (if 353) | — | 레인 · 감독자 · 발표 오류 그대로(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A112|Refresh|Wave|ProductionAssembly'` · `TestAFailedWaveCarriesNoAssemblyForEitherMarket` |
| B5 | `if clk != nil {` (if 374) | dispatch 주기에 실시계 | 무편집 | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A112|Refresh|Wave|ProductionAssembly'` · `TestTheProductionAssemblyGivesTheDispatchCycleTheRealClock` |
| B6 | `for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {` (range 390) | worker 둘 | 무편집 | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A112|Refresh|Wave|ProductionAssembly'` · `TestTheMarketThatLeadsAWaveAlwaysPublishesIt` |
| B7 | `if err != nil {` (if 395) | — | 레인 · 감독자 · 발표 오류 그대로(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A112|Refresh|Wave|ProductionAssembly'` · `TestAFailedWaveCarriesNoAssemblyForEitherMarket` |
| B8 | `if err := c.publishStrategyRuntime(assembly); err != nil {` (if 402) | — | 레인 · 감독자 · 발표 오류 그대로(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A112|Refresh|Wave|ProductionAssembly'` · `TestAFailedWaveCarriesNoAssemblyForEitherMarket` |

Exact AST return positions: 327 · 354 · 380 · 396 · 403 · 405

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 34종 — 대표: `NewStrategyEntrySupervisor`, `accountAuthority.Snapshot`, `append`, `c.Journal.Path`, `c.productionStrategyLanes`, `c.productionStrategyWorker`, `c.publishStrategyRuntime`, `candidateAuthority.Snapshot`) | 본문 참조 | 소스 참조 — 원장 · 브로커 · 파일 호출은 design F2 · F9 | AST |

## State mutations and fallbacks

- 편집은 위 「a157 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a157 편집」에 적은 자리.
- High-risk impact: yes(Guardian 첫 다리 판정 순서 · 일일 손실 분류 · 감독자 조립 — 받아들이는 요청 집합 불변(design D2), 주문을 늘리는 갈래 없음,
  일일 손실 강화 보존(1-1), 매도 · 손절 · 청산 경로 무관)
