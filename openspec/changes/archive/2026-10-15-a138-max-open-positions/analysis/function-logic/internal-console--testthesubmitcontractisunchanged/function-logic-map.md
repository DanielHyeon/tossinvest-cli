# Function Logic Map: `TestTheSubmitContractIsUnchanged`

- Source: `internal/console/settings_cadence_test.go` (418-457)
- Qualified: `TestTheSubmitContractIsUnchanged`
- AST evidence: `ast.json`(현재) — 현재 AST branches 9 · returns 0
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 1afdd43a25f55990…) — 편집 전 갈래 수 9
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 — 화면별 제출 칸 전수(census).

**a138 편집.** `/settings/limits` 의 칸 목록에 `max_open_positions` 하나.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, path := range settingsTabPaths {` (range 423) | — | 분기 본문 참조 | `go test ./internal/console -run TestTheSubmitContractIsUnchanged` |
| B2 | `for _, form := range formBlock.FindAllStringSubmatch(body(t, h.get(t, path)), -1) {` (range 424) | — | 분기 본문 참조 | `go test ./internal/console -run TestTheSubmitContractIsUnchanged` |
| B3 | `if fields[form[1]] == nil {` (if 425) | — | 분기 본문 참조 | `go test ./internal/console -run TestTheSubmitContractIsUnchanged` |
| B4 | `for _, field := range inputName.FindAllStringSubmatch(form[2], -1) {` (range 428) | — | 분기 본문 참조 | `go test ./internal/console -run TestTheSubmitContractIsUnchanged` |
| B5 | `for action, want := range map[string][]string{` (range 433) | — | 분기 본문 참조 | `go test ./internal/console -run TestTheSubmitContractIsUnchanged` |
| B6 | `if !ok {` (if 444) | — | 분기 본문 참조 | `go test ./internal/console -run TestTheSubmitContractIsUnchanged` |
| B7 | `for _, field := range want {` (range 448) | — | 분기 본문 참조 | `go test ./internal/console -run TestTheSubmitContractIsUnchanged` |
| B8 | `if !got[field] {` (if 449) | — | 분기 본문 참조 | `go test ./internal/console -run TestTheSubmitContractIsUnchanged` |
| B9 | `if len(got) != len(want) {` (if 453) | — | 분기 본문 참조 | `go test ./internal/console -run TestTheSubmitContractIsUnchanged` |

Exact AST return positions: (없음)

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 8종 — 대표: `body`, `formBlock.FindAllStringSubmatch`, `fullSettingsHarness`, `h.authenticate`, `h.get`, `inputName.FindAllStringSubmatch`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a138 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a138 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
