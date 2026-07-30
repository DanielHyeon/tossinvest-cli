# Branch Test Map: `consoleLimitSettingsSeam`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | config 경로를 해석할 수 없으면 nil seam → 폼 미렌더·POST 501 | `TestWithoutASeamTheLimitEditorRefusesRatherThanPretends` | yes | yes |
