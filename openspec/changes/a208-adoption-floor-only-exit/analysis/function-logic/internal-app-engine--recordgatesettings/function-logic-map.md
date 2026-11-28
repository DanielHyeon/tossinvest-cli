# Function Logic Map: `recordGateSettings`

- Source: `internal/app/engine/interlock.go` (774-848)
- Qualified: `recordGateSettings`
- AST evidence: `ast.json` (source_sha256 `7d2ce7348681…`) — AST branches 6 · returns 2
- Risk scan: `risk-pattern-report.md`(설정 규칙 일치 0)
- 작성 시점: 편집 **전**(a208 tasks 1.1, 2026-11-28, HEAD `810aec6f7` — base `baeddfe31` 와 본문 동일).

**역할.** 엔진 기동 때 운영 설정을 audit 에 기록하는 표(§0.5). 표 순서대로 `RecordChange` — 직전 값과 같으면 쓰지 않는다. 생산 호출자 `NewContext`(engine.go:487) 하나 + 시험(a074 4 · a163 1).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `gate` | 로드된 automation gate | `NewContext` | — |
| `adoption` | 병합된 `config.Adoption`(거부 블록 전면 0) | `cfg.Engine.Adoption` | — |
| `notifications` | 알림 해석 결과(비밀 없음) | `resolveNotificationPublisher` | — |
| 불변식 | 표는 끝에만 덧붙임(두 기동을 나란히 읽기 위해) | `TestExistingAuditEntriesKeepTheirOrder` · a163 시험 | — |

## Branches and early returns

| Branch | Condition | At | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if adoption.Rejected != ""` | 777:2 | `adoptionDetail` 문장 | — | 미측정 |
| B2 | `if notifications.Refused != ""` | 787:2 | `notificationDetail` = 거부 사유 | — | `TestARefusedNotificationBlockIsAudited` |
| B3 | `else` 갈래 | 790:9 | — | — | `TestNotificationSettingsAreAudited` |
| B4 | `if notifications.PublicService && notifications.TopicConfigured` | 790:9 | 공개 ntfy 경고 detail | — | 미측정 |
| B5 | `for _, c := range changes` | 842:2 | 행마다 `RecordChange` | — | `TestExistingAuditEntriesKeepTheirOrder` |
| B6 | `if _, err := log.RecordChange(...); err != nil` | 843:3 | — | return err (844:4) | 미측정 |

Exact AST return positions: 844:4, 847:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strconv.FormatBool` · `limitString` · `strings.Join` | 값 문자열화 | — | AST |
| `perTradeRiskBudgetAudit` | a139 유도 값 | — | AST |
| `log.RecordChange` | 한 행 기록(값이 바뀐 경우만) | 오류 → B6 return(기동 거절) | AST · HEAD |

Live config binding: 기동 때 로드된 값만. 주기 재읽기 없음.

## State mutations and fallbacks

- audit 로그 append 만. fallback 없음 — 기록 실패는 기동 거절(보수 방향).

## Safety conclusion

- Safe edit boundary(a208): `changes` 표 **끝**에 `engine.adoption.floor_only_exit` 행 하나를 덧붙인다(분기 0 추가, 기존 행 순서 무변). a166 census 는 이 함수 본문을 고정하지 않는다(`shapes.json` 에는 `NewContext` 의 호출 줄만 — 호출 줄 무변).
- 귀결: a163 시험의 「마지막 행」 단언이 「끝에서 둘째」가 된다(같은 change 에서 갱신, 번들 `internal-app-engine--testa163…`).
- High-risk impact: no(audit 기록, 주문 · 판정 무관).
