# Branch Test Map: `recordGateSettings`



| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if 742 — 기존 audit 행 + 끝 행(engine.exit_source_age_report) | `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` · `TestExistingAuditEntriesKeepTheirOrder` | yes(2.16 — 골격에서 끝 행 없음으로 실패) | pass |
| B2 | if 752 — 기존 audit 행 + 끝 행(engine.exit_source_age_report) | `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` · `TestExistingAuditEntriesKeepTheirOrder` | yes(2.16 — 골격에서 끝 행 없음으로 실패) | pass |
| B3 | else 755 — 기존 audit 행 + 끝 행(engine.exit_source_age_report) | `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` · `TestExistingAuditEntriesKeepTheirOrder` | yes(2.16 — 골격에서 끝 행 없음으로 실패) | pass |
| B4 | if 755 — 기존 audit 행 + 끝 행(engine.exit_source_age_report) | `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` · `TestExistingAuditEntriesKeepTheirOrder` | yes(2.16 — 골격에서 끝 행 없음으로 실패) | pass |
| B5 | range 807 — 기존 audit 행 + 끝 행(engine.exit_source_age_report) | `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` · `TestExistingAuditEntriesKeepTheirOrder` | yes(2.16 — 골격에서 끝 행 없음으로 실패) | pass |
| B6 | if 808 — 기존 audit 행 + 끝 행(engine.exit_source_age_report) | `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten` · `TestExistingAuditEntriesKeepTheirOrder` | yes(2.16 — 골격에서 끝 행 없음으로 실패) | pass |
