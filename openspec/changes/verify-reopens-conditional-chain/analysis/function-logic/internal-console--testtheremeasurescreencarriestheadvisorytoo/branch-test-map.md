# Branch Test Map: `TestTheRemeasureScreenCarriesTheAdvisoryToo`

- Source: `internal/console/remeasure_test.go`
- Function: `internal/console/remeasure_test.go:TestTheRemeasureScreenCarriesTheAdvisoryToo` (base revision `de14674974ab`)

이 change는 이 함수를 수정하지 않았다. RED 열이 전부 `no`인 것은 그 사실의 기록이다 —
실패 상태를 만들지 않았고, 이 함수는 구현 전후 모두 통과했다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` base line 357: `if !strings.Contains(page, "order-hours-closed") {` | `TestTheRemeasureScreenCarriesTheAdvisoryToo` | no | yes |
| B2 | `if` base line 360: `if !strings.Contains(page, "재측정 1단계") {` | `TestTheRemeasureScreenCarriesTheAdvisoryToo` | no | yes |
