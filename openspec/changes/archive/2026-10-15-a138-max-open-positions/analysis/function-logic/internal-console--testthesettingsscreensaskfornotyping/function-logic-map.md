# Function Logic Map: `TestTheSettingsScreensAskForNoTyping`

- Source: `internal/console/settings_card_test.go` (310-349)
- Qualified: `TestTheSettingsScreensAskForNoTyping`
- AST evidence: `ast.json`(현재) — 현재 AST branches 8 · returns 0
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 4db183adada6bfd2…) — 편집 전 갈래 수 8
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 — 설정 화면에 사유 · 확인 문구 칸이 없음.

**a138 편집.** 화면이 소유한 설정 목록에 `max_open_positions` 하나(확인 문구가 아님).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, path := range settingsTabPaths {` (range 326) | — | 분기 본문 참조 | `go test ./internal/console -run TestTheSettingsScreensAskForNoTyping` |
| B2 | `if strings.Contains(page, "<textarea") {` (if 328) | — | 분기 본문 참조 | `go test ./internal/console -run TestTheSettingsScreensAskForNoTyping` |
| B3 | `if strings.Contains(page, `type="password"`) {` (if 331) | — | 분기 본문 참조 | `go test ./internal/console -run TestTheSettingsScreensAskForNoTyping` |
| B4 | `for _, form := range formBlock.FindAllStringSubmatch(page, -1) {` (range 334) | — | 분기 본문 참조 | `go test ./internal/console -run TestTheSettingsScreensAskForNoTyping` |
| B5 | `for _, field := range inputName.FindAllStringSubmatch(form[2], -1) {` (range 335) | — | 분기 본문 참조 | `go test ./internal/console -run TestTheSettingsScreensAskForNoTyping` |
| B6 | `if !allowed[field[1]] {` (if 336) | — | 분기 본문 참조 | `go test ./internal/console -run TestTheSettingsScreensAskForNoTyping` |
| B7 | `for _, word := range []string{"입력하세요", "정확히 입력", "사유를 입력"} {` (range 343) | — | 분기 본문 참조 | `go test ./internal/console -run TestTheSettingsScreensAskForNoTyping` |
| B8 | `if strings.Contains(page, word) {` (if 344) | — | 분기 본문 참조 | `go test ./internal/console -run TestTheSettingsScreensAskForNoTyping` |

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
