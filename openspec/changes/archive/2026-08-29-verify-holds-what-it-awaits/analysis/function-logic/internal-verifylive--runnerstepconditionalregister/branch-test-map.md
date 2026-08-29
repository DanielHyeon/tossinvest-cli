# Branch Test Map: `Runner.stepConditionalRegister`

- Source: `internal/verifylive/steps.go`
- Function: `internal/verifylive/steps.go:Runner.stepConditionalRegister`

이 change의 RED는 실제로 관측했다. 전체 실행 기록은
`internal-verifylive--cleanupfrom/branch-test-map.md`에 원문으로 있다.
RED `no`는 이 change가 그 분기의 동작을 바꾸지 않아 실패 상태를 따로 만들지 않았다는 뜻이며,
Test 열은 지금 그 분기를 덮는 것을 가리킨다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` line 558: `if err != nil {` | `Runner.stepConditionalRegister` 계열 테스트 | 아래 RED 기록 | yes |
| B2 | `if` line 561: `if sellable < MinQuantity {` | `Runner.stepConditionalRegister` 계열 테스트 | 아래 RED 기록 | yes |
| B3 | `if` line 568: `if err != nil {` | `Runner.stepConditionalRegister` 계열 테스트 | 아래 RED 기록 | yes |
| B4 | `if` line 572: `if err != nil {` | `Runner.stepConditionalRegister` 계열 테스트 | 아래 RED 기록 | yes |
| B5 | `if` line 587: `if err != nil {` | `Runner.stepConditionalRegister` 계열 테스트 | 아래 RED 기록 | yes |
| B6 | `switch` line 600: `switch {` | `Runner.stepConditionalRegister` 계열 테스트 | 아래 RED 기록 | yes |
| B7 | `case` line 601: `case isGateError(replayErr):` | `Runner.stepConditionalRegister` 계열 테스트 | 아래 RED 기록 | yes |
| B8 | `case` line 603: `case replayErr != nil:` | `Runner.stepConditionalRegister` 계열 테스트 | 아래 RED 기록 | yes |
| B9 | `case` line 606: `case replayID == id:` | `Runner.stepConditionalRegister` 계열 테스트 | 아래 RED 기록 | yes |
| B10 | `case` line 608: `default:` | `Runner.stepConditionalRegister` 계열 테스트 | 아래 RED 기록 | yes |
| B11 | `if` line 611: `if replayID != "" {` | `Runner.stepConditionalRegister` 계열 테스트 | 아래 RED 기록 | yes |
| B12 | `if` line 614: `if err := r.cancelConditional(ctx, sr, replayID, symbol,` | `Runner.stepConditionalRegister` 계열 테스트 | 아래 RED 기록 | yes |
| B13 | `if` line 622: `if co, err := r.readConditional(ctx, sr, id); err == nil {` | `Runner.stepConditionalRegister` 계열 테스트 | 아래 RED 기록 | yes |
| B14 | `else` line 629: `} else {` | `Runner.stepConditionalRegister` 계열 테스트 | 아래 RED 기록 | yes |
| B15 | `if` line 642: `if err != nil {` | `Runner.stepConditionalRegister` 계열 테스트 | 아래 RED 기록 | yes |
| B16 | `else` line 644: `} else {` | `Runner.stepConditionalRegister` 계열 테스트 | 아래 RED 기록 | yes |
| B17 | `if` line 652: `if chain == "" {` | `Runner.stepConditionalRegister` 계열 테스트 | 아래 RED 기록 | yes |
