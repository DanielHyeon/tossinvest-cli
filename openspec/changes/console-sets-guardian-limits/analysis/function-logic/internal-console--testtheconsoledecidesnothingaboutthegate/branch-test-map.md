# Branch Test Map: `TestTheConsoleDecidesNothingAboutTheGate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 패키지 전 파일이 검사된다 | 자기 자신 (면제 파일이 존재해야 한다는 짝 테스트가 보완) | yes (면제 전 3파일 실패) | yes |
| B2 | 면제 파일에는 블록 이름이 허용된다 | `TestTheGateEditingExemptionIsNotIdle` | yes | yes |
| B3 | 금지어 전부가 순회된다 | 자기 자신 | yes | yes |
| B4 | 비면제 파일이 블록을 이름하면 실패 | 자기 자신 (templates_settings.go가 실제로 실패했고 문안 수정으로 닫음) | yes | yes |
