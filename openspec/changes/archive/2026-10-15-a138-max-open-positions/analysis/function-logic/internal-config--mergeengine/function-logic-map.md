# Function Logic Map: `mergeEngine`

- Source: `internal/config/engine.go` (324-358)
- Qualified: `mergeEngine`
- AST evidence: `ast.json`(현재) — 현재 AST branches 5 · returns 2
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 06a665a9b9d06e0f…) — 편집 전 갈래 수 5
- Risk scan: `risk-pattern-report.md`

**역할.** 파싱 모양 → 설정 구조체 복사.

**a138 편집.** `MaxOpenPositions` 복사 한 줄.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if raw == nil {` (if 325) | — | 분기 본문 참조 | `go test ./internal/config -run 'Engine|A138'` · `TestA138OpenPositionMaximumIsParsedFromTheGateBlock` |
| B2 | `if raw.Autostart != nil {` (if 328) | — | 분기 본문 참조 | `go test ./internal/config -run 'Engine|A138'` · `TestA138OpenPositionMaximumIsParsedFromTheGateBlock` |
| B3 | `if raw.ExitPolicy != nil {` (if 332) | — | 분기 본문 참조 | `go test ./internal/config -run 'Engine|A138'` · `TestA138OpenPositionMaximumIsParsedFromTheGateBlock` |
| B4 | `if raw.AutomationGate == nil {` (if 342) | — | 분기 본문 참조 | `go test ./internal/config -run 'Engine|A138'` · `TestA138OpenPositionMaximumIsParsedFromTheGateBlock` |
| B5 | `if gate.Enabled != nil {` (if 346) | — | 분기 본문 참조 | `go test ./internal/config -run 'Engine|A138'` · `TestA138OpenPositionMaximumIsParsedFromTheGateBlock` |

Exact AST return positions: 326 · 343

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 4종 — 대표: `mergeAdoption`, `mergeNotifications`, `next.validate`, `strings.TrimSpace`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a138 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a138 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
