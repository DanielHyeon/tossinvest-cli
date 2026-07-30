# Branch Test Map: `Console.handleSettings`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 편입 seam 있음 | `TestTheSettingsScreenShowsTheRawBlockAndTheVerdict` | no (기존 동작) | yes |
| B2 | 편입 Load 실패 | `TestTheSettingsScreenShowsTheRawBlockAndTheVerdict` | no (기존 동작) | yes |
| B3 | 한도 seam 없음 → 폼 미렌더 + POST 501 | `TestWithoutASeamTheLimitEditorRefusesRatherThanPretends` | yes | yes |
| B4 | 한도 Load 실패해도 편입 섹션 생존 | `TestAnUnreadableConfigDoesNotHideTheRestOfTheScreen` | yes | yes |
