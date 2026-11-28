# Branch Test Map: `recordGateSettings`

편집 전 좌표(AST 6 분기, `internal/app/engine/interlock.go` 774–848). 「미측정」 = 덮는 시험을 이름으로 특정하지 않았음(실행 커버리지 미측정).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 거부된 편입 블록 → detail 에 사유 | 미측정 | n/a | no |
| B2 | 거부된 알림 블록 → detail 에 사유 | `TestARefusedNotificationBlockIsAudited` | n/a | no |
| B3 | 알림 블록 수용 | `TestNotificationSettingsAreAudited` | n/a | no |
| B4 | 공개 ntfy + topic → 경고 detail | 미측정 | n/a | no |
| B5 | 표 순서대로 기록 | `TestExistingAuditEntriesKeepTheirOrder` | n/a | no |
| B6 | 기록 실패 → 기동 거절 | 미측정 | n/a | no |

a208 새 행: `TestA208FloorOnlyIsAuditedAtStartup`(끝 행 · 앞 행 = a163 · 같은 값 재기동 0 행 · 변경 1 행) — RED 관측 `analysis/tdd/red.txt`.
