# Branch Test Map: `Runner.Plan`

- Source: `internal/verifylive/plan.go`
- Function: `internal/verifylive/plan.go:Runner.Plan`

이 change의 RED는 실제로 관측했다. 전체 실행 기록은
`internal-verifylive--runnermutationsymbol/branch-test-map.md`에 원문으로 있다.
RED `no`는 이 change가 그 분기의 동작을 바꾸지 않아 실패 상태를 따로 만들지 않았다는 뜻이며,
Test 열은 지금 그 분기를 덮는 것을 가리킨다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `range` line 542: `for _, line := range r.planCleanup() {` | `Runner.Plan` 계열 테스트 | 아래 RED 기록 | yes |
| B2 | `range` line 547: `for _, step := range Steps() {` | `Runner.Plan` 계열 테스트 | 아래 RED 기록 | yes |
| B3 | `if` line 548: `if settled, verdict := r.settled(step.ID); settled {` | `Runner.Plan` 계열 테스트 | 아래 RED 기록 | yes |
| B4 | `if` line 549: `if step.Mutates {` | `Runner.Plan` 계열 테스트 | 아래 RED 기록 | yes |
| B5 | `if` line 558: `if reason, skip := r.preflightStatic(step, passed); skip {` | `Runner.Plan` 계열 테스트 | 아래 RED 기록 | yes |
| B6 | `if` line 559: `if step.Mutates {` | `Runner.Plan` 계열 테스트 | 아래 RED 기록 | yes |
| B7 | `if` line 565: `if !step.Mutates {` | `Runner.Plan` 계열 테스트 | 아래 RED 기록 | yes |
| B8 | `if` line 570: `if strings.TrimSpace(symbol) == "" {` | `Runner.Plan` 계열 테스트 | 아래 RED 기록 | yes |
| B9 | `if` line 585: `if !ok {` | `Runner.Plan` 계열 테스트 | 아래 RED 기록 | yes |
| B10 | `range` line 594: `for _, line := range lines {` | `Runner.Plan` 계열 테스트 | 아래 RED 기록 | yes |
