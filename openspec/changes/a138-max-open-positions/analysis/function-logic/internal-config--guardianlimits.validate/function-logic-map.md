# Function Logic Map: `Validate`

- Source: `internal/config/limits.go` (329-359)
- Qualified: `GuardianLimits.Validate`
- AST evidence: `ast.json`(현재) — 현재 AST branches 9 · returns 7
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 15189779559e480c…) — 편집 전 갈래 수 7
- Risk scan: `risk-pattern-report.md`

**역할.** 편집기 저장 전 검증.

**a138 편집.** `OpenPositions` 가 음수 · 비정수 · NaN · Inf 면 거부 — 새 분기 하나.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, f := range guardianFields() {` (range 330) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B2 | `switch {` (switch 332) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B3 | `case math.IsNaN(v) \|\| math.IsInf(v, 0):` (case 333) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B4 | `case v <= 0:` (case 335) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B5 | `if l.MaxDailyLossRatio > 1 {` (if 340) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B6 | `if strings.TrimSpace(l.Currency) == "" {` (if 344) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B7 | `if l.DailyBuy != nil && (math.IsNaN(*l.DailyBuy) \|\| math.IsInf(*l.DailyBuy, 0) \|\| *l.DailyBuy < 0) {` (if 348) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B8 | `if l.OpenPositions != nil && *l.OpenPositions != 0 {` (if 352) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B9 | `if _, ok := WholeOpenPositions(*l.OpenPositions); !ok {` (if 353) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |

Exact AST return positions: 334 · 336 · 341 · 345 · 349 · 354 · 358

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 8종 — 대표: `WholeOpenPositions`, `errors.New`, `f.Get`, `fmt.Errorf`, `guardianFields`, `math.IsInf`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a138 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a138 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
