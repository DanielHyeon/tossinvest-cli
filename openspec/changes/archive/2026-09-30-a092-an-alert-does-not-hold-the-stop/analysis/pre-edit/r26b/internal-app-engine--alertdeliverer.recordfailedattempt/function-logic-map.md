# Function Logic Map: `alertDeliverer.recordFailedAttempt`

- Source: `internal/app/engine/alertdelivery.go`
- AST evidence: `ast.json` — **편집 전**, :348–376, 분기 7, source_sha256 `1fa0da7e9719…`, 추출 커밋 `e55102f0`.
- Risk scan: `risk-pattern-report.md`
- 편집 목적: 26라운드 codex P0 · P1#4: (1) 반납 결과가 행 없음 · 모르는 결과면 원칙 E 조건부 차단(승격 없음 — N6), 이미 선 한도 판정은 보존 (2) 시도 기록이 `AlreadySettled` · `LeaseLost`(선점)이면 선점 사실을 기록(델타 「선점이 일어났다는 사실은 기록되어야 한다」 — 오늘은 빈 갈래).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `id` · `token` | 이 실행자가 claim 한 행과 임차 | 배달 실행자 사이클 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil` (:350) | — | — | (미실행) |
| B2 | `if err != nil` (:355) | — | — | (미실행) |
| B3 | `switch res.Outcome` (:361) | — | — | `TestADeliveryAndAnEmptyAcknowledgementAfterTheEvidenceDropOnlyTheBlock` |
| B4 | `case journal.SettleApplied:` (:362) | — | — | `TestADeliveryAndAnEmptyAcknowledgementAfterTheEvidenceDropOnlyTheBlock` |
| B5 | `if res.Attempts < alertAttemptLimit` (:365) | — | — | `TestADeliveryAndAnEmptyAcknowledgementAfterTheEvidenceDropOnlyTheBlock` |
| B6 | `case journal.SettleAlreadySettled, journal.SettleLeaseLost:` (:369) | — | — | (미실행) |
| B7 | `default:` (:371) | — | — | (미실행) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ReleaseAlertClaim` · `MarkAlertAttemptFailed` · `judge` · `readEpoch` | 정산 · 반납 · 원칙 E | 결과 분류 | AST |

## State mutations and fallbacks

- 원장 정산 · 반납 · 게이트 래치(judge).

## Safety conclusion

- Safe edit boundary: 반납 결과 분류와 선점 기록만 — 시도 한도 판정 · 승격 규칙 불변.
- High-risk impact: yes — 배달 실행자의 진입 차단 판정(a124 정본 영역, a092 델타 「모든 발송자」 문단).
