# Branch Test Map: `Console.handleSettingsSave`

편집 전 좌표(AST 4 분기, `internal/console/settings.go` 238–273). 「미측정」 = 덮는 시험을 이름으로 특정하지 않았음(실행 커버리지 미측정).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | seam 미배선 → 설명 | `TestAnUnwiredSettingsSeamIsExplained` | n/a | no |
| B2 | 읽기 실패 → 저장 안 함 | 미측정 | n/a | no |
| B3 | 손절 % 해석 실패 → 저장 안 함 | `TestInvalidStopPercentWritesNothing` | n/a | no |
| B4 | 저장 실패 → 설명 | 미측정 | n/a | no |

a208: `TestA208ConsoleSaveKeepsFloorOnly`(참 · 거짓 둘 다 이월) — RED 관측 `analysis/tdd/red.txt`.
