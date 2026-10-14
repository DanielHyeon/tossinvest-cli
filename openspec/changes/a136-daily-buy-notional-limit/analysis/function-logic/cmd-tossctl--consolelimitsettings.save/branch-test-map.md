# Branch Test Map: `Save`

- Source: `cmd/tossctl/limitsettings.go` (48-71, 편집 뒤)

| Branch | Anchor | Scenario | Test | a136 |
|---|---|---|---|---|
| B1 | if 50 | 기존 동작 | `TestA136TheDailyBuyCeilingSaveIsAuditedFromTheFile` · `TestTheLimitSeamSavesAuditsAndLeavesTheSwitchAlone` | 무편집 분기 |
| B2 | if 56 | 기존 동작 | `TestA136TheDailyBuyCeilingSaveIsAuditedFromTheFile` · `TestTheLimitSeamSavesAuditsAndLeavesTheSwitchAlone` | 무편집 분기 |
| B3 | else 58 | 기존 동작 | `TestA136TheDailyBuyCeilingSaveIsAuditedFromTheFile` · `TestTheLimitSeamSavesAuditsAndLeavesTheSwitchAlone` | 무편집 분기 |
| B4 | if 63 | 기존 동작 | `TestA136TheDailyBuyCeilingSaveIsAuditedFromTheFile` · `TestTheLimitSeamSavesAuditsAndLeavesTheSwitchAlone` | 무편집 분기 |
| B5 | if 65 | 기존 동작 | `TestA136TheDailyBuyCeilingSaveIsAuditedFromTheFile` · `TestTheLimitSeamSavesAuditsAndLeavesTheSwitchAlone` | 무편집 분기 |
