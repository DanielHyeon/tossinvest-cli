# Branch Test Map: `mergeEngine`

편집 뒤 좌표(a208 F3 재추출 · AST 7 분기, `internal/config/engine.go` 349–393). 분기 수 무변(타입 주석 5 줄로 함수가 +5, B3 뒤 두 줄로 B4 이후 +2 더).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if raw == nil` | 미측정 | n/a | 미측정 |
| B2 | Autostart 있음 | 미측정 | n/a | 미측정 |
| B3 | ExitPolicy 있음 | `TestUnknownCommonExitPolicyIsRetainedAsRejected` · `TestA208FloorOnlyCopyIsNotAConfigurableCommonPolicy` | n/a | yes — 선택 시험 통과(F3) |
| B4 | ExitSourceAgeReport 있음 | 미측정(a163 시험) | n/a | 미측정 |
| B5 | CancelExternalOppositeBuy 있음 | 미측정(a197 시험) | n/a | 미측정 |
| B6 | AutomationGate 없음 | 미측정 | n/a | 미측정 |
| B7 | gate.Enabled 있음 | 미측정 | n/a | 미측정 |

a208 F3 분기 밖 한 줄: `TestA208FloorOnlyIgnoredReasonFollowsTheCommonPolicy` — RED 관측 red-f3.txt(FAIL), GREEN 통과.
