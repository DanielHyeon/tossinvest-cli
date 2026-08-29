# Branch Test Map: `heldAfter`

- Source: `internal/verifylive/cleanup.go`
- Function: `internal/verifylive/cleanup.go:heldAfter`

이 change의 RED는 실제로 관측했다. 전체 실행 기록은
`internal-verifylive--cleanupfrom/branch-test-map.md`에 원문으로 있다.
RED `no`는 이 change가 그 분기의 동작을 바꾸지 않아 실패 상태를 따로 만들지 않았다는 뜻이며,
Test 열은 지금 그 분기를 덮는 것을 가리킨다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `range` line 185: `for i := range entries {` | `heldAfter` 계열 테스트 | 아래 RED 기록 | yes |
| B2 | `if` line 186: `if entries[i].StepID == gate {` | `heldAfter` 계열 테스트 | 아래 RED 기록 | yes |
