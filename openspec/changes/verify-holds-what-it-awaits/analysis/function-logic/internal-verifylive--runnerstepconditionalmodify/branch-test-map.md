# Branch Test Map: `Runner.stepConditionalModify`

- Source: `internal/verifylive/steps.go`
- Function: `internal/verifylive/steps.go:Runner.stepConditionalModify`

이 change의 RED는 실제로 관측했다. 전체 실행 기록은
`internal-verifylive--cleanupfrom/branch-test-map.md`에 원문으로 있다.
RED `no`는 이 change가 그 분기의 동작을 바꾸지 않아 실패 상태를 따로 만들지 않았다는 뜻이며,
Test 열은 지금 그 분기를 덮는 것을 가리킨다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` line 731: `if !ok {` | `Runner.stepConditionalModify` 계열 테스트 | 아래 RED 기록 | yes |
| B2 | `if` line 736: `if err != nil {` | `Runner.stepConditionalModify` 계열 테스트 | 아래 RED 기록 | yes |
| B3 | `if` line 740: `if newTrigger <= 0 {` | `Runner.stepConditionalModify` 계열 테스트 | 아래 RED 기록 | yes |
| B4 | `if` line 753: `if err != nil {` | `Runner.stepConditionalModify` 계열 테스트 | 아래 RED 기록 | yes |
| B5 | `if` line 759: `if newID != id {` | `Runner.stepConditionalModify` 계열 테스트 | 아래 RED 기록 | yes |
| B6 | `else` line 769: `} else {` | `Runner.stepConditionalModify` 계열 테스트 | 아래 RED 기록 | yes |
| B7 | `if` line 760: `if _, err := r.readConditional(ctx, sr, id); err != nil {` | `Runner.stepConditionalModify` 계열 테스트 | 아래 RED 기록 | yes |
| B8 | `else` line 763: `} else {` | `Runner.stepConditionalModify` 계열 테스트 | 아래 RED 기록 | yes |
| B9 | `if` line 773: `if co, err := r.readConditional(ctx, sr, newID); err == nil {` | `Runner.stepConditionalModify` 계열 테스트 | 아래 RED 기록 | yes |
