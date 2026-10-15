# Function Logic Map: `recordGateSettings`

- Source: `internal/app/engine/interlock.go` (734-803)
- Qualified: `recordGateSettings`
- AST evidence: `ast.json`(현재) — 현재 AST branches 6 · returns 2
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 389371bc84e21e83…) — 편집 전 갈래 수 6
- Risk scan: `risk-pattern-report.md`

**역할.** 기동 audit — 게이트 설정값 줄들.

**a138 편집.** 목록 끝에 `engine.automation_gate.max_open_positions`(해석 값) 한 줄(순서 무변경).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if adoption.Rejected != "" {` (if 737) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Audit|A074|A138'` · `TestA138TheInterlockAuditsTheOpenPositionMaximum` |
| B2 | `if notifications.Refused != "" {` (if 747) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Audit|A074|A138'` · `TestA138TheInterlockAuditsTheOpenPositionMaximum` |
| B3 | `} else if notifications.PublicService && notifications.TopicConfigured {` (else 750) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Audit|A074|A138'` · `TestA138TheInterlockAuditsTheOpenPositionMaximum` |
| B4 | `} else if notifications.PublicService && notifications.TopicConfigured {` (if 750) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Audit|A074|A138'` · `TestA138TheInterlockAuditsTheOpenPositionMaximum` |
| B5 | `for _, c := range changes {` (range 797) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Audit|A074|A138'` · `TestA138TheInterlockAuditsTheOpenPositionMaximum` |
| B6 | `if _, err := log.RecordChange(c.action, c.setting, c.value, c.detail); err != nil {` (if 798) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Audit|A074|A138'` · `TestA138TheInterlockAuditsTheOpenPositionMaximum` |

Exact AST return positions: 799 · 802

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 8종 — 대표: `gate.EffectiveMaxDailyBuyNotional`, `gate.EffectiveMaxOpenPositions`, `limitString`, `log.RecordChange`, `strconv.FormatBool`, `strings.Join`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a138 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a138 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
