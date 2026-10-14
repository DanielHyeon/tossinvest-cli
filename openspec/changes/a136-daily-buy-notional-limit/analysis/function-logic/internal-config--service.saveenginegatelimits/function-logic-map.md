# Function Logic Map: `SaveEngineGateLimits`

- Source: `internal/config/limits_io.go` (97-158, 편집 뒤)
- Qualified: `Service.SaveEngineGateLimits`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 9d08b616b9a8c621…, 편집 전 분기 수 9) — 현재 branches 12 · returns 10
- Risk scan: `risk-pattern-report.md`

**역할.** 콘솔 한도 편집기의 키 단위 쓰기(enabled 비기록).

**a136 편집.** freeze 리뷰 P1-2 수리: 저장하는 통화가 파일의 현재 `limit_currency` 와 다르면 `max_daily_buy_notional` 을 0 으로 지우는 멤버를 함께 씀 — 옛 통화로 넣은 숫자가 새 통화로 재해석되지 않게(KRW→USD 면 기동 거부, USD→KRW 면 기본값).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `l` | Validate · ceiling 통과 | 콘솔 | 오류 |
| 파일 바이트 | 유효 JSON | 디스크 | 오류 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err := next.Validate(); err != nil {` (if 101) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limit / TestA136'` |
| B2 | `if violations := next.CeilingViolations(); len(violations) > 0 {` (if 105) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limit / TestA136'` |
| B3 | `if err != nil {` (if 111) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limit / TestA136'` |
| B4 | `switch {` (switch 117) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limit / TestA136'` |
| B5 | `case errors.Is(err, os.ErrNotExist):` (case 118) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limit / TestA136'` |
| B6 | `if err != nil {` (if 131) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limit / TestA136'` |
| B7 | `case err != nil:` (case 135) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limit / TestA136'` |
| B8 | `if !json.Valid(data) {` (if 139) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limit / TestA136'` |
| B9 | `if prev, err := documentGateCurrency(data); err != nil {` (if 148) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limit / TestA136'` |
| B10 | `} else if prev != next.Currency {` (else 150) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limit / TestA136'` |
| B11 | `} else if prev != next.Currency {` (if 150) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limit / TestA136'` |
| B12 | `if err != nil {` (if 154) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limit / TestA136'` |

Exact AST return positions: 102 · 106 · 112 · 132 · 134 · 136 · 140 · 149 · 155 · 157

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `next.Validate` · `CeilingViolations` | 저장 전 거부 | 오류 | AST |
| `spliceGateLimits` | 키 단위 쓰기 | 오류 | AST |

## State mutations and fallbacks

- config.json 원자 교체(installBytes).

## Safety conclusion

- Safe edit boundary: 통화 변경 감지 + 지움 멤버 추가 — 다섯 한도 · enabled 무변경.
- High-risk impact: yes(위험 한도 설정 쓰기, §0.5)
