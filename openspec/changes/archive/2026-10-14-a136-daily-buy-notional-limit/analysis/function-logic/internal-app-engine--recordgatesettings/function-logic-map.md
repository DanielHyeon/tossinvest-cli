# Function Logic Map: `recordGateSettings`

- Source: `internal/app/engine/interlock.go` (723-789, 편집 뒤)
- Qualified: `recordGateSettings`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` a2277d4bd3a76152…, 편집 전 분기 수 6) — 현재 branches 6 · returns 2
- Risk scan: `risk-pattern-report.md`

**역할.** 기동마다 운영 설정 변화를 audit(RecordChange — 바뀐 값만).

**a136 편집.** 변화 목록 **끝에** `engine.automation_gate.max_daily_buy_notional`(해석 값) 추가 — 「Appended, never interleaved」 규칙. config 수동 편집의 이전·새 값이 여기 남음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `log` · `gate` | — | 기동 | 오류 반환 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if adoption.Rejected != "" {` (if 726) | — | 분기 본문 참조 | `TestA136TheInterlockAuditsTheDailyBuyCeiling` |
| B2 | `if notifications.Refused != "" {` (if 736) | — | 분기 본문 참조 | `TestA136TheInterlockAuditsTheDailyBuyCeiling` |
| B3 | `} else if notifications.PublicService && notifications.TopicConfigured {` (else 739) | — | 분기 본문 참조 | `TestA136TheInterlockAuditsTheDailyBuyCeiling` |
| B4 | `} else if notifications.PublicService && notifications.TopicConfigured {` (if 739) | — | 분기 본문 참조 | `TestA136TheInterlockAuditsTheDailyBuyCeiling` |
| B5 | `for _, c := range changes {` (range 783) | — | 분기 본문 참조 | `TestA136TheInterlockAuditsTheDailyBuyCeiling` |
| B6 | `if _, err := log.RecordChange(c.action, c.setting, c.value, c.detail); err != nil {` (if 784) | — | 분기 본문 참조 | `TestA136TheInterlockAuditsTheDailyBuyCeiling` |

Exact AST return positions: 785 · 788

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `log.RecordChange` | 변화 기록 | 오류 반환 | AST |

## State mutations and fallbacks

- audit 파일 append.

## Safety conclusion

- Safe edit boundary: 목록 끝 항목 하나.
- High-risk impact: yes(§0.5 audit)
