# Branch Test Map: `TestAnOrdinaryStartIsStillNotARemeasure`

- Source: `internal/console/remeasure_test.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `for _, o := range final.Summary.Outcomes {` (internal/console/remeasure_test.go:275) | 이 함수 자체가 테스트다 | yes | yes |
| B2 | `if o.Step == verifylive.StepConditionalPersist {` (internal/console/remeasure_test.go:276) | 이 함수 자체가 테스트다 | yes | yes |
| B3 | `if !o.AlreadySettled {` (internal/console/remeasure_test.go:279) | 이 함수 자체가 테스트다 | yes | yes |
