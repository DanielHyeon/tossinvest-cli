# Branch Test Map: `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten`

편집 뒤 좌표(a208 3.3 재추출 · AST 8 분기, `internal/app/engine/a163_source_age_audit_test.go` 16–55). 「미측정」 = 덮는 시험을 이름으로 특정하지 않았음(실행 커버리지 미측정).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 기록 실패 | `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` | n/a | yes — 선택 시험 통과(3.1) |
| B2 | Entries 실패 | `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` | n/a | yes — 선택 시험 통과(3.1) |
| B3 | a163 행 위치 · action · 값 | `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` | yes — a208 행 추가로 실패하는 단언을 첨자 갱신 | yes — 3.1 |
| B4 | a163 행 앞 = a139 | `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` | n/a | yes — 선택 시험 통과(3.1) |
| B5 | action 철자 | `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` | n/a | yes — 선택 시험 통과(3.1) |
| B6 | 토글 켬 = 1 행 | `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` | n/a | yes — 선택 시험 통과(3.1) |
| B7 | 변경 행 old→new | `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` | n/a | yes — 선택 시험 통과(3.1) |
| B8 | 같은 값 재기동 0 행 | `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` | n/a | yes — 선택 시험 통과(3.1) |
