# Branch Test Map: `candidateCycleOptions`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 두 표면이 같은 임계를 적용한다 | `TestTheTwoSurfacesApplyTheSameThresholds`(두 seam의 결과를 비교) · `TestOnlyOneFileInThisCommandBuildsTheVetoThresholds`(AST: 생산 파일에 다른 `VetoThresholds{…}` 복합 리터럴 금지) · `TestTheSignalsSeamReadsTheSameConstructor` | yes (리터럴이 둘일 때 AST 테스트가 실패) | yes |
