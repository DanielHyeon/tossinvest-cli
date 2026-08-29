# Branch Test Map: `Runner.preflightStatic`

- Source: `internal/verifylive/runner.go`
- Function: `internal/verifylive/runner.go:Runner.preflightStatic`

이 change의 RED는 실제로 관측했다. 전체 실행 기록은
`internal-verifylive--runnermutationsymbol/branch-test-map.md`에 원문으로 있다.
RED `no`는 이 change가 그 분기의 동작을 바꾸지 않아 실패 상태를 따로 만들지 않았다는 뜻이며,
Test 열은 지금 그 분기를 덮는 것을 가리킨다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` line 531: `if step.Deferred != "" {` | `Runner.preflightStatic` 계열 테스트 | 아래 RED 기록 | yes |
| B2 | `if` line 537: `if step.OptIn != "" && !r.optedIn(step) {` | `Runner.preflightStatic` 계열 테스트 | 아래 RED 기록 | yes |
| B3 | `range` line 540: `for _, dep := range step.DependsOn {` | `Runner.preflightStatic` 계열 테스트 | 아래 RED 기록 | yes |
| B4 | `if` line 541: `if !passed(dep) {` | `Runner.preflightStatic` 계열 테스트 | 아래 RED 기록 | yes |
| B5 | `if` line 545: `if step.NeedsHolding && r.holdingSymbol == "" {` | `Runner.preflightStatic` 계열 테스트 | 아래 RED 기록 | yes |
| B6 | `if` line 550: `if symbol := r.mutationSymbol(step); step.Mutates && !SameMarket(MarketOf(symbol), r.market) {` | `Runner.preflightStatic` 계열 테스트 | 아래 RED 기록 | yes |
