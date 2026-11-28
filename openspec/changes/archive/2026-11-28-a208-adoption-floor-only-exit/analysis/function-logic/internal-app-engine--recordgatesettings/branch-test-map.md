# Branch Test Map: `recordGateSettings`

편집 뒤 좌표(a208 3.3 재추출 · AST 6 분기, `internal/app/engine/interlock.go` 774–850). 「미측정」 = 덮는 시험을 이름으로 특정하지 않았음(실행 커버리지 미측정).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 거부된 편입 블록 → detail 에 사유 | 미측정 | n/a | 미측정(덮는 시험 미특정) |
| B2 | 거부된 알림 블록 → detail 에 사유 | `TestARefusedNotificationBlockIsAudited` | n/a | yes — 선택 시험 통과(3.1) |
| B3 | 알림 블록 수용 | `TestNotificationSettingsAreAudited` | n/a | yes — 선택 시험 통과(3.1) |
| B4 | 공개 ntfy + topic → 경고 detail | 미측정 | n/a | 미측정(덮는 시험 미특정) |
| B5 | 표 순서대로 기록 | `TestExistingAuditEntriesKeepTheirOrder` | n/a | yes — 선택 시험 통과(3.1) |
| B6 | 기록 실패 → 기동 거절 | 미측정 | n/a | 미측정(덮는 시험 미특정) |

a208 새 행: `TestA208FloorOnlyIsAuditedAtStartup`(끝 행 · 앞 행 = a163 · 같은 값 재기동 0 행 · 변경 1 행) — RED 관측 `analysis/tdd/red.txt`.

a208 F3: 무시 행 `TestA208IgnoredFloorOnlyIsAuditedAtStartup`(true → ignored, detail = 로더 사유) — RED 관측 `analysis/tdd/red-f3.txt`.
