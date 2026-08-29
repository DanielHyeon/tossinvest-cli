# Branch Test Map: `ChainOf`

- Source: `internal/verifylive/record.go`
- Function: `internal/verifylive/record.go:ChainOf`

이 change의 RED는 실제로 관측했다. 전체 실행 기록은
`internal-verifylive--cleanupfrom/branch-test-map.md`에 원문으로 있다.
RED `no`는 이 change가 그 분기의 동작을 바꾸지 않아 실패 상태를 따로 만들지 않았다는 뜻이며,
Test 열은 지금 그 분기를 덮는 것을 가리킨다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `range` line 473: `for _, e := range entries {` | `ChainOf` 계열 테스트 | 아래 RED 기록 | yes |
| B2 | `range` line 474: `for _, a := range e.Artifacts {` | `ChainOf` 계열 테스트 | 아래 RED 기록 | yes |
| B3 | `if` line 475: `if a.Kind == kind && a.ID == id && a.ChainID != "" {` | `ChainOf` 계열 테스트 | 아래 RED 기록 | yes |
