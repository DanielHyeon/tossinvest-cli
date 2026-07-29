# Branch Test Map: `acceptSupervised`

- Source: `internal/soak/attest.go`
- Function: `internal/soak/attest.go:acceptSupervised`

RED/GREEN은 이 change를 구현하며 실제로 관측한 것이다. RED `no`는 이 change가 그 분기의
동작을 바꾸지 않아 실패 상태를 따로 만들지 않았다는 뜻이며, Test 열은 지금 그것을 덮는 것을
가리킨다. 이 change의 RED 실행 기록은 `internal-soak--buildattestation`의 branch-test-map에 있다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `range` line 259: `for _, e := range LiveOnlyEndpoints() {` | `acceptSupervised` | no | yes |
| B2 | `range` line 265: `for _, p := range supervised {` | `acceptSupervised` | no | yes |
| B3 | `if` line 267: `if endpoint == "" {` | `acceptSupervised` | no | yes |
| B4 | `if` line 271: `if !allowed[key] {` | `acceptSupervised` | no | yes |
| B5 | `if` line 278: `if !attest.SameAccountMasked(accountRef, p.AccountRef) {` | `acceptSupervised` | no | yes |
| B6 | `if` line 286: `if age < 0 \|\| age >= validity {` | `acceptSupervised` | no | yes |
| B7 | `if` line 289: `if seen[key] {` | `acceptSupervised` | no | yes |
