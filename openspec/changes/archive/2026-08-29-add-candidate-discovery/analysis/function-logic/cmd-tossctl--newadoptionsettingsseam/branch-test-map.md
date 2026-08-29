# Branch Test Map: `newAdoptionSettingsSeam`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | config 경로 미해석 → 구체 포인터 nil → `consoleSettingsSeam`이 리터럴 nil을 넘긴다 | `TestATypedNilSeamNeverReachesTheInterface` | yes | yes |
