# Branch Test Map: `mergeEngine`

**편집 전 좌표**(AST 7 분기, base `b93ed133a`, `config/engine.go` `336`–`378`). 실행 0 — 시험 이름은 rg 로 골랐고 덮음 여부는 미측정.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | raw 없음 → 기본값 | `TestA163TheSourceAgeReportToggleDefaultsOff`(키 없음 갈래) | no | no |
| B2 | autostart 반영 | `TestAutostartDefaultsOffAndLoadsExplicitValues` | no | no |
| B3 | exit 정책 반영 | 미특정 | no | no |
| B4 | a163 토글 반영 | `TestA163TheSourceAgeReportToggleDefaultsOff` | no | no |
| B5 | a197 토글 반영(automation_gate 키 없어도) | `TestA197TheExternalBuyToggleDefaultsOff` | yes | yes |
| B6 | automation_gate 없음 → 조기 반환 | 미특정 | no | no |
| B7 | gate enabled 반영 | 미특정 | no | no |
