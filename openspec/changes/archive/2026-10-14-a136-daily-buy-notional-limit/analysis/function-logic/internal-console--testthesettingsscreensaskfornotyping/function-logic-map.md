# Function Logic Map: `TestTheSettingsScreensAskForNoTyping`

- Source: `internal/console/settings_card_test.go` (310-348, 편집 뒤)
- Qualified: `TestTheSettingsScreensAskForNoTyping`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` e08f8acbe6268522…, 편집 전 분기 수 8) — 현재 branches 8 · returns 0
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
| B1 | `for _, path := range settingsTabPaths {` (range 325) | — | 분기 본문 참조 | `go test ./internal/console` |
| B2 | `if strings.Contains(page, "<textarea") {` (if 327) | — | 분기 본문 참조 | `go test ./internal/console` |
| B3 | `if strings.Contains(page, `type="password"`) {` (if 330) | — | 분기 본문 참조 | `go test ./internal/console` |
| B4 | `for _, form := range formBlock.FindAllStringSubmatch(page, -1) {` (range 333) | — | 분기 본문 참조 | `go test ./internal/console` |
| B5 | `for _, field := range inputName.FindAllStringSubmatch(form[2], -1) {` (range 334) | — | 분기 본문 참조 | `go test ./internal/console` |
| B6 | `if !allowed[field[1]] {` (if 335) | — | 분기 본문 참조 | `go test ./internal/console` |
| B7 | `for _, word := range []string{"입력하세요", "정확히 입력", "사유를 입력"} {` (range 342) | — | 분기 본문 참조 | `go test ./internal/console` |
| B8 | `if strings.Contains(page, word) {` (if 343) | — | 분기 본문 참조 | `go test ./internal/console` |

Exact AST return positions: 

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | fixture | — | AST |

## State mutations and fallbacks

- 시험 데이터만.

## Safety conclusion

- Safe edit boundary: 항목 하나.
- High-risk impact: no(시험 전용)
