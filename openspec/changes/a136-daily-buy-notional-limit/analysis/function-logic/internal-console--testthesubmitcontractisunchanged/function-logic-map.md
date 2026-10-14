# Function Logic Map: `TestTheSubmitContractIsUnchanged`

- Source: `internal/console/settings_cadence_test.go` (418-457, 편집 뒤)
- Qualified: `TestTheSubmitContractIsUnchanged`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` b697789a28ef026f…, 편집 전 분기 수 9) — 현재 branches 9 · returns 0
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
| B1 | `for _, path := range settingsTabPaths {` (range 423) | — | 분기 본문 참조 | `go test ./internal/console` |
| B2 | `for _, form := range formBlock.FindAllStringSubmatch(body(t, h.get(t, path)), -1) {` (range 424) | — | 분기 본문 참조 | `go test ./internal/console` |
| B3 | `if fields[form[1]] == nil {` (if 425) | — | 분기 본문 참조 | `go test ./internal/console` |
| B4 | `for _, field := range inputName.FindAllStringSubmatch(form[2], -1) {` (range 428) | — | 분기 본문 참조 | `go test ./internal/console` |
| B5 | `for action, want := range map[string][]string{` (range 433) | — | 분기 본문 참조 | `go test ./internal/console` |
| B6 | `if !ok {` (if 444) | — | 분기 본문 참조 | `go test ./internal/console` |
| B7 | `for _, field := range want {` (range 448) | — | 분기 본문 참조 | `go test ./internal/console` |
| B8 | `if !got[field] {` (if 449) | — | 분기 본문 참조 | `go test ./internal/console` |
| B9 | `if len(got) != len(want) {` (if 453) | — | 분기 본문 참조 | `go test ./internal/console` |

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
