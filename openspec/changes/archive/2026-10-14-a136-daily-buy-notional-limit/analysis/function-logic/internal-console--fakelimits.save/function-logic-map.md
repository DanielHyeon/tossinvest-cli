# Function Logic Map: `Save`

- Source: `internal/console/settings_limits_test.go` (40-63, 편집 뒤)
- Qualified: `fakeLimits.Save`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 0521a39b3c5f6f84…, 편집 전 분기 수 3) — 현재 branches 4 · returns 4
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 fixture/시험(생산 코드 아님).

**a136 편집.** a136 하루 매수 칸을 허용 목록 · fake 저장 · 기대값에 반영 — 시험이 지키는 규칙(확인 문구 없음 · 제출 계약) 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| — | — | 시험 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if f.saveErr != nil {` (if 43) | — | 분기 본문 참조 | `go test ./internal/console` |
| B2 | `if err := l.Validate(); err != nil {` (if 46) | — | 분기 본문 참조 | `go test ./internal/console` |
| B3 | `if v := l.CeilingViolations(); len(v) > 0 {` (if 49) | — | 분기 본문 참조 | `go test ./internal/console` |
| B4 | `if l.DailyBuy != nil { // a136: 폼이 보낸 하루 매수 상한만 씀(nil = 그대로)` (if 59) | — | 분기 본문 참조 | `go test ./internal/console` |

Exact AST return positions: 44 · 47 · 50 · 62

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | fixture | — | AST |

## State mutations and fallbacks

- 시험 데이터만.

## Safety conclusion

- Safe edit boundary: 항목 하나.
- High-risk impact: no(시험 전용)
