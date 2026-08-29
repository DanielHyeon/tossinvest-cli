# Branch Test Map: `mutationNote`

- Source: `internal/soak/attest.go`
- Function: `internal/soak/attest.go:mutationNote`

RED/GREEN은 이 change를 구현하며 실제로 관측한 것이다. RED `no`는 이 change가 그 분기의
동작을 바꾸지 않아 실패 상태를 따로 만들지 않았다는 뜻이며, Test 열은 지금 그것을 덮는 것을
가리킨다. 이 change의 RED 실행 기록은 `internal-soak--buildattestation`의 branch-test-map에 있다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `range` line 303: `for _, p := range accepted {` | `mutationNote` | no | yes |
| B2 | `range` line 307: `for _, e := range LiveOnlyEndpoints() {` | `mutationNote` | no | yes |
| B3 | `if` line 308: `if have[normaliseEndpoint(e)] {` | `mutationNote` | no | yes |
| B4 | `else` line 310: `} else {` | `mutationNote` | no | yes |
| B5 | `switch` line 314: `switch {` | `mutationNote` | no | yes |
| B6 | `case` line 315: `case len(missing) == 0:` | `mutationNote` | no | yes |
| B7 | `case` line 318: `case len(covered) == 0:` | `mutationNote` | no | yes |
| B8 | `case` line 321: `default:` | `mutationNote` | no | yes |
