# Branch Test Map: `groupDecimalText`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 음수 값의 부호가 구분 앞에 남는다 | `TestOnlyOneGuardianAxisIsComputedAndTheRestSayWhyNot`(손실 축의 음수 렌더) | — | yes |
| B2 | 정수부 순회 | 모든 숫자 렌더 경로 | — | yes |
| B3 | 4자리 이상에 `,`가 들어간다 | `TestThePositionsScreenShowsTheExitLineOfAManagedPosition`의 평가액 대조 | — | yes |
| B4 | 소수부가 있으면 그대로 재부착된다 | US 소수 단가 렌더(`us_market_test.go` 경로) + `TestANullHoldTimeIsAnEmDashAndNotZeroSeconds`의 자매 렌더 대조 | — | yes |
