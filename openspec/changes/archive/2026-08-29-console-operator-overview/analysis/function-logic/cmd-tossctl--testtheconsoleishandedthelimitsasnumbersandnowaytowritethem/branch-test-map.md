# Branch Test Map: `TestTheConsoleIsHandedTheLimitsAsNumbersAndNoWayToWriteThem`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 해석 가능한 config 디렉터리에서 seam이 만들어진다 | 이 테스트 | yes | yes |
| B2 | 메서드 정확히 하나 | 동일 | yes (Save 추가 시 FAIL) | yes |
| B3 | 메서드 이름 수집(진단) | 동일 | no (진단) | yes |
| B4 | `console.Options`가 `GateLimits`를 받는다 | 동일 | yes (배선 전 FAIL) | yes |
