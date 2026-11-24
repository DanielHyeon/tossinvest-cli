# Branch Test Map: `recordGateSettings`

**좌표**(AST 6 분기, base `b93ed133a`, `interlock.go` `774`–`848`). 무편집 대상(design D8) — 실행 0, 시험 이름은 rg 로 고름.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 편입 블록 거절 상세 | 미특정 | no | no |
| B2 | 알림 블록 거절 상세 | `TestARefusedNotificationBlockIsAudited` | no | no |
| B3 | else | `TestNotificationSettingsAreAudited` | no | no |
| B4 | 공개 ntfy topic 상세 | `TestNotificationSettingsAreAudited` | no | no |
| B5 | 행 순회 · 순서 | `TestExistingAuditEntriesKeepTheirOrder` · `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` | no | no |
| B6 | 기록 실패 반환 | 미특정 | no | no |
