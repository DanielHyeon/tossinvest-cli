# Branch Test Map: `decidedAfter`

- Source: `internal/verifylive/cleanup.go`
- Function: `internal/verifylive/cleanup.go:decidedAfter`

이 change의 RED는 실제로 관측했다. 전체 실행 기록은
`internal-verifylive--cleanupfrom/branch-test-map.md`에 원문으로 있다.
RED `no`는 이 change가 그 분기의 동작을 바꾸지 않아 실패 상태를 따로 만들지 않았다는 뜻이며,
Test 열은 지금 그 분기를 덮는 것을 가리킨다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `range` line 152: `for i := range entries {` | `decidedAfter` 계열 테스트 | 아래 RED 기록 | yes |
| B2 | `range` line 153: `for _, x := range entries[i].Artifacts {` | `decidedAfter` 계열 테스트 | 아래 RED 기록 | yes |
| B3 | `if` line 154: `if x.Kind == a.Kind && x.ID == a.ID {` | `decidedAfter` 계열 테스트 | 아래 RED 기록 | yes |
| B4 | `if` line 159: `if created >= 0 {` | `decidedAfter` 계열 테스트 | 아래 RED 기록 | yes |
| B5 | `if` line 163: `if created < 0 {` | `decidedAfter` 계열 테스트 | 아래 RED 기록 | yes |
| B6 | `range` line 167: `for i := range entries {` | `decidedAfter` 계열 테스트 | 아래 RED 기록 | yes |
| B7 | `if` line 168: `if entries[i].StepID == id {` | `decidedAfter` 계열 테스트 | 아래 RED 기록 | yes |
