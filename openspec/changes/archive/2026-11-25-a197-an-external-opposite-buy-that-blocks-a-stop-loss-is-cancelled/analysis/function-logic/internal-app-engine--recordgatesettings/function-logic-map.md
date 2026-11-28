# Function Logic Map: `recordGateSettings`

- Source: `internal/app/engine/interlock.go` (`774`–`848`, base `b93ed133a`)
- Qualified: `recordGateSettings`
- AST evidence: `ast.json` — branches 6 · returns 2
- Risk scan: `risk-pattern-report.md`
- 작성 시점: 편집 전(a197 재개 개정 2026-11-25, Manager 결정 O-1). **설계 결론: 이 함수는 편집하지 않는다**(아래 Safety conclusion) — Manager 지시로 번들만 만들어 둔다.

**역할.** 기동 설정을 audit log 에 값이 바뀐 기동만 한 행씩 기록한다(`log.RecordChange`). 호출자: 생산 `NewContext` `engine.go:486` 하나, 시험 함수 5(`a074_notification_audit_test.go` 4 · `a163_source_age_audit_test.go` 1).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `log` · `gate` · `adoption` · `notifications` · `attestationFile` · `sourceAgeReport` | 로드된 설정 | `NewContext` | — |
| 불변식 | 새 설정은 목록 **끝에** 덧붙이고 기존 순서 · 이름을 바꾸지 않는다(a074 `TestExistingAuditEntriesKeepTheirOrder`), 마지막 행은 a163 `engine.exit_source_age_report`(a163 핀 — 그 앞은 `per_trade_risk_budget`) | 시험 핀 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:777` `adoption.Rejected != ""` | 상세 문장 | — | 미특정 |
| B2 | `:787` `notifications.Refused != ""` | 상세 문장 | — | `TestARefusedNotificationBlockIsAudited` |
| B3 | `:790` else | — | — | — |
| B4 | `:790` 공개 ntfy ∧ topic | 상세 문장 | — | `TestNotificationSettingsAreAudited` |
| B5 | `:842` `range changes` | — | — | `TestExistingAuditEntriesKeepTheirOrder` |
| B6 | `:843` `RecordChange` 오류 | — | `return err` `@844` | 미특정 |

정상 `return nil` `@847`.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `log.RecordChange` | 값 변경 기록 | 오류 → B6 반환 → 호출자 기동 거절 | AST `:843` |
| `strconv.FormatBool` · `limitString` · `perTradeRiskBudgetAudit` 등 | 값 문자열 | — | AST |

## State mutations and fallbacks

- audit log 행 기록만.

## Safety conclusion

- **a197 은 이 함수를 편집하지 않는다.** 끝에 행을 덧붙이면 a163 핀 `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten`(마지막 행 = `engine.exit_source_age_report`)이 깨지고, 중간에 넣으면 같은 핀(그 앞 행 = `per_trade_risk_budget`)과 a074 순서 핀의 정신이 깨지며, 인자를 늘리면 시험 함수 5 개를 고쳐야 한다(gate ⑤ 번들 5). 같은 실패 의미를 `NewContext` 의 B5 바로 뒤 한 줄로 얻는다(design D8).
- High-risk impact: no(무편집).
