# Function Logic Map: `NewStrategyEntrySupervisor`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (647-725, 편집 뒤)
- Qualified: `NewStrategyEntrySupervisor`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` 6f7c999a13c4abdc…) · `ast-base.json`(편집 전, base `63fd0c92`, `source_sha256` 9cb510c1c9c7f44a…, 편집 전 분기수 18) — 현재 branches 18 · returns 14
- Risk scan: `risk-pattern-report.md`

**역할.** 감독자 생성 · 검증.

**a137 편집.** 재리뷰 P1-N1: `dailyLossNotice: opts.DailyLossNotice` 한 칸(entry 와 같은 줄). 분기 불변(18).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| opts | 옵션 | 호출자 | 기존 검증 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if depth == 0 {` (if 649) | — | 분기 본문 참조 | 기존 생성 시험 |
| B2 | `if depth < 1 \|\| depth > MaximumStrategyQueueDepth {` (if 652) | — | 분기 본문 참조 | 기존 생성 시험 |
| B3 | `if cycleLimit == 0 {` (if 656) | — | 분기 본문 참조 | 기존 생성 시험 |
| B4 | `if cycleLimit < 0 \|\| cycleLimit > MaximumStrategyCycleLimit {` (if 659) | — | 분기 본문 참조 | 기존 생성 시험 |
| B5 | `if len(opts.Workers) != 2 {` (if 662) | — | 분기 본문 참조 | 기존 생성 시험 |
| B6 | `if clk == nil {` (if 666) | — | 분기 본문 참조 | 기존 생성 시험 |
| B7 | `if now.IsZero() {` (if 670) | — | 분기 본문 참조 | 기존 생성 시험 |
| B8 | `for _, descriptor := range opts.Workers {` (range 675) | — | 분기 본문 참조 | 기존 생성 시험 |
| B9 | `if !validStrategyMarket(descriptor.Market) {` (if 676) | — | 분기 본문 참조 | 기존 생성 시험 |
| B10 | `if _, duplicate := workers[descriptor.Market]; duplicate {` (if 679) | — | 분기 본문 참조 | 기존 생성 시험 |
| B11 | `if descriptor.Effective && descriptor.Cycle == nil {` (if 682) | — | 분기 본문 참조 | 기존 생성 시험 |
| B12 | `if descriptor.Effective && (descriptor.AuthorityGeneration == 0 \|\| descriptor.AuthorityExpiresAt.IsZero() \|\|` (if 685) | — | 분기 본문 참조 | 기존 생성 시험 |
| B13 | `if descriptor.Effective && ((descriptor.RestartAttempt == 0 && (descriptor.FirstRefusal != StrategyWorkerRefusalNone \|\| !descriptor.RestartNotBefore.IsZero())) \|\|` (if 689) | — | 분기 본문 참조 | 기존 생성 시험 |
| B14 | `if descriptor.PollInterval < 0 \|\| descriptor.PollInterval > MaximumStrategyCycleLimit \|\|` (if 693) | — | 분기 본문 참조 | 기존 생성 시험 |
| B15 | `if !descriptor.Effective && (descriptor.AuthorityGeneration != 0 \|\| !descriptor.AuthorityExpiresAt.IsZero() \|\| descriptor.EvidenceDigest != "" \|\| descriptor.LatchRevision != 0) {` (if 699) | — | 분기 본문 참조 | 기존 생성 시험 |
| B16 | `if !descriptor.Effective && (descriptor.FirstRefusal != StrategyWorkerRefusalNone \|\| descriptor.RestartAttempt != 0 \|\| !descriptor.RestartNotBefore.IsZero()) {` (if 702) | — | 분기 본문 참조 | 기존 생성 시험 |
| B17 | `for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {` (range 715) | — | 분기 본문 참조 | 기존 생성 시험 |
| B18 | `if workers[market] == nil {` (if 716) | — | 분기 본문 참조 | 기존 생성 시험 |

Exact AST return positions: 653 · 660 · 663 · 671 · 677 · 680 · 683 · 687 · 691 · 697 · 700 · 703 · 717 · 721

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | — | — | — |

## State mutations and fallbacks

- 감독자 구조체.

## Safety conclusion

- Safe edit boundary: 칸 하나.
- High-risk impact: no — 관측.
