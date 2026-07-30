# Branch Test Map: `rankObs`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `true`는 측정된 yes, `false`는 측정된 no | `TestANewlyListedSymbolDoesNotClimbFromLastPlace`(yes) · `TestTheSameNumberOfPlacesIsADifferentMoveInADifferentList`(no) · `TestAChurningSymbolIsMeasuredAgainstWhatWeActuallySaw`(둘 다) | yes (`bool`을 그대로 넘기면 컴파일 실패) | yes |
