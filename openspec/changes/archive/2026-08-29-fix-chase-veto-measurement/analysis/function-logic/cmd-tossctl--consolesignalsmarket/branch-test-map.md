# Branch Test Map: `consoleSignalsMarket`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 읽지 못한 시장은 사유와 함께 남는다 | **커버 없음** — 렌더 쪽만 `internal/console`의 `TestEveryUnmeasuredStateOnTheSignalsScreenCarriesACodeAndASentence`가 `Markets[0].Why`를 직접 세워 잡는다 | no | no |
| (임계) | 두 표면이 같은 임계를 쓴다 | `TestTheTwoSurfacesApplyTheSameThresholds` · `TestTheSignalsSeamReadsTheSameConstructor`(AST: 이 함수가 `candidateVetoThresholds`를 명명한다) · `TestOnlyOneFileInThisCommandBuildsTheVetoThresholds` | yes | yes |
| (Sightings) | seam이 census를 싣는다 | `internal/console`의 `TestTheSignalsPageAttributesTheRefusalsToTheSourceThatProducedThem`(렌더 쪽) — seam의 대입 자체는 `TestTheSignalsSeamReadsTheStoreAndCallsNoSource`가 지나가지만 단언하지 않는다 | yes | yes |
