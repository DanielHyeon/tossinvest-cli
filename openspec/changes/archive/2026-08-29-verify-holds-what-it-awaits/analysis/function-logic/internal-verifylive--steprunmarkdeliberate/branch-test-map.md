# Branch Test Map: `stepRun.markDeliberate`

- Source: `internal/verifylive/runner.go`
- Function: `internal/verifylive/runner.go:stepRun.markDeliberate`

이 change의 RED는 실제로 관측했다. 전체 실행 기록은
`internal-verifylive--cleanupfrom/branch-test-map.md`에 원문으로 있다.
RED `no`는 이 change가 그 분기의 동작을 바꾸지 않아 실패 상태를 따로 만들지 않았다는 뜻이며,
Test 열은 지금 그 분기를 덮는 것을 가리킨다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `range` line 731: `for i := range sr.artifacts {` | `stepRun.markDeliberate` 계열 테스트 | 아래 RED 기록 | yes |
| B2 | `if` line 732: `if sr.artifacts[i].Kind == kind && sr.artifacts[i].ID == id && !sr.artifacts[i].Cancelled {` | `stepRun.markDeliberate` 계열 테스트 | 아래 RED 기록 | yes |
