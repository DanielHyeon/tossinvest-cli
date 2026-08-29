# Branch Test Map: `Plan.Authorises`

- Source: `internal/verifylive/plan.go`
- Function: `internal/verifylive/plan.go:Plan.Authorises`

이 change의 RED는 실제로 관측했다. 전체 실행 기록은
`internal-verifylive--runnermutationsymbol/branch-test-map.md`에 원문으로 있다.
RED `no`는 이 change가 그 분기의 동작을 바꾸지 않아 실패 상태를 따로 만들지 않았다는 뜻이며,
Test 열은 지금 그 분기를 덮는 것을 가리킨다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `range` line 394: `for _, m := range p.Mutations {` | `Plan.Authorises` 계열 테스트 | 아래 RED 기록 | yes |
| B2 | `if` line 395: `if m.Step != step \|\| m.Kind != kind {` | `Plan.Authorises` 계열 테스트 | 아래 RED 기록 | yes |
| B3 | `if` line 403: `if !strings.EqualFold(strings.TrimSpace(m.Symbol), strings.TrimSpace(symbol)) {` | `Plan.Authorises` 계열 테스트 | 아래 RED 기록 | yes |
| B4 | `if` line 406: `if m.Side != "" && !strings.EqualFold(m.Side, strings.TrimSpace(side)) {` | `Plan.Authorises` 계열 테스트 | 아래 RED 기록 | yes |
| B5 | `if` line 409: `if m.MaxQuantity <= 0 {` | `Plan.Authorises` 계열 테스트 | 아래 RED 기록 | yes |
| B6 | `if` line 412: `if quantity > 0 {` | `Plan.Authorises` 계열 테스트 | 아래 RED 기록 | yes |
| B7 | `if` line 417: `if quantity <= m.MaxQuantity+quantityTolerance {` | `Plan.Authorises` 계열 테스트 | 아래 RED 기록 | yes |
