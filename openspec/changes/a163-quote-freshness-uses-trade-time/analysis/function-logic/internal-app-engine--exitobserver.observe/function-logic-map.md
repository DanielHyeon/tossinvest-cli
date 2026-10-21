# Function Logic Map: `ExitObserver.observe`

- Source: `internal/app/engine/exitloop.go` (842-913)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- 기준: 편집 전(base `f7d80c348`)
- AST 분기 15 · 반환 6 · defer 0 · go 0
- 위험 등급: High-risk(손절 판정의 유일한 시세 입력)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `symbols` | 보유·대상 종목 | `workingSet` | — |
| `q.FetchedAt` | 0 또는 받은 시각 | 어댑터 | 0 이면 사이클 대체(`validatedAt`), 미래·15 s 초과면 후보 제외(B7) |
| `q.Last` | > 0 유한 | 어댑터 | 아니면 후보 제외(B4) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 849 `if qerr != nil {` | — | `return qerr` → Retrier 재시도/실패 | `TestAFailedObservationHoldsTheJudgement` |
| B2 | range 858 `for _, symbol := range symbols {` | `wanted` 집합 구성 | — | 기존 시험 |
| B3 | range 862 `for _, q := range quotes {` | 후보 순회 | — | 기존 시험 |
| B4 | if 864 `if _, needed := wanted[symbol]; !needed \|\| q.Last <= 0 \|\| math.IsNaN(q.Last) \|\| math.IsInf(q.Last, 0) {` | 후보 제외 | `continue` | `TestAQuoteWithNoLastTradeIsNotAnObservation` |
| B5 | if 868 `if q.FetchedAt.IsZero() {` | `FetchedAt` 0 → 사이클 대체, `needsFallback` | — | `TestA111FallbackObservationSourceIsDurableAndRestartMonotone` |
| B6 | else 871 `} else if effectiveAt.After(validatedAt) \|\|` | else-if 사슬 | — | `TestA111QuoteEvidenceUsesOnePostBatchClockAndNeverFallsBackFromBadOfficialTime` |
| B7 | if 871 `} else if effectiveAt.After(validatedAt) \|\|` | 미래·15 s 초과 후보 제외 | `continue` | `TestA111QuoteEvidenceUsesOnePostBatchClockAndNeverFallsBackFromBadOfficialTime` |
| B8 | if 878 `if len(next) == 0 {` | — | 후보 0 → `EvidenceInvalid` | 기존 시험 |
| B9 | if 884 `if err != nil {` | — | `return nil, err` | `TestAFailedObservationHoldsTheJudgement` |
| B10 | if 887 `if needsFallback {` | 대체 시퀀스 회복 | — | `TestA111FallbackSequenceRecoveryIsLazyAndPriceEvidenceUsesTheGateDuration` |
| B11 | if 893 `if err != nil {` | — | 원장 오류 반환 | 기존 시험 |
| B12 | for 896 `for {` | CAS 루프 | — | 기존 시험 |
| B13 | if 898 `if current >= maximum \|\| o.observationSequence.CompareAndSwap(current, maximum) {` | 루프 탈출 | — | 기존 시험 |
| B14 | range 904 `for symbol, quote := range out {` | 대체 출처 기입 | — | 기존 시험 |
| B15 | if 905 `if quote.Source != "" {` | 공식 출처는 건너뜀 | `continue` | 기존 시험 |

반환 좌표(AST): 850 · 879 · 882 · 885 · 894 · 912

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.opts.Retrier.Query` | 단일 브로커 읽기 + 신선도 도장 | 재시도 매트릭스(QueryPrice) | AST |
| `o.opts.Prices.Prices` | 공식 `/prices` | 오류는 B1 | AST |
| `clock.LeaseAnchor` | 사용 임대 앵커 | — | AST |
| `o.opts.Journal.MaxExitObservationCycle` | 대체 시퀀스 회복 | 오류는 B11 | AST |

## State mutations and fallbacks

- `out` 맵 · `needsFallback` · `o.observationSequence`(원자).

## a163 편집

- a163 3.2: `next[symbol] = observedQuote{…}` 리터럴에 `SourceTimestamp: q.SourceTimestamp` 하나. 분기 15 · 반환 6 불변(design D7). B4·B7 의 거름은 여전히 `FetchedAt`·`Last` 만 읽음.

## Safety conclusion

- Safe edit boundary: 리터럴 필드 대입 하나. 조건식 무편집.
- High-risk impact: yes — 변이 ④b·⑨ 는 렌더 해시가 아니라 e2e·census·identity 시험이 잡는다.
