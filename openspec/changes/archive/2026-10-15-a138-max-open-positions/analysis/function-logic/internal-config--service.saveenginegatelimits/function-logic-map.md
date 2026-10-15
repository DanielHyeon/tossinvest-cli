# Function Logic Map: `SaveEngineGateLimits`

- Source: `internal/config/limits_io.go` (99-168)
- Qualified: `Service.SaveEngineGateLimits`
- AST evidence: `ast.json`(현재) — 현재 AST branches 14 · returns 10
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 6c8fdc5d37028679…) — 편집 전 갈래 수 13
- Risk scan: `risk-pattern-report.md`

**역할.** 편집기 저장 — 검증 · 잠금 · 새 파일 갈래 · 키 단위 쓰기.

**a138 편집.** 새 파일(skeleton) 갈래에 `OpenPositions` 를 옮기는 분기 하나(구현 리뷰 P2-1 — 빠뜨리면 운영자 값이 기본 6 으로 사라짐).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err := next.Validate(); err != nil {` (if 103) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B2 | `if violations := next.CeilingViolations(); len(violations) > 0 {` (if 107) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B3 | `if err != nil {` (if 113) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B4 | `switch {` (switch 119) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B5 | `case errors.Is(err, os.ErrNotExist):` (case 120) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B6 | `if next.DailyBuy != nil {` (if 132) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B7 | `if next.OpenPositions != nil {` (if 136) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B8 | `if err != nil {` (if 140) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B9 | `case err != nil:` (case 144) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B10 | `if !json.Valid(data) {` (if 148) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B11 | `if prev, err := documentGateCurrency(data); err != nil {` (if 157) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B12 | `} else if prev != next.Currency && next.DailyBuy == nil {` (else 159) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B13 | `} else if prev != next.Currency && next.DailyBuy == nil {` (if 159) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B14 | `if err != nil {` (if 164) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |

Exact AST return positions: 104 · 108 · 114 · 141 · 143 · 145 · 149 · 158 · 165 · 167

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 17종 — 대표: `DefaultFile`, `acquireConfigLock`, `append`, `documentGateCurrency`, `errors.Is`, `fmt.Errorf`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a138 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a138 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
