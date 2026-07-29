# Branch Test Map: `runSoakAttest`

- Source: `cmd/tossctl/soak.go`
- Function: `cmd/tossctl/soak.go:runSoakAttest`

RED/GREEN은 이 change를 구현하며 실제로 관측한 것이다. RED `no`는 이 change가 그 분기의
동작을 바꾸지 않아 실패 상태를 따로 만들지 않았다는 뜻이며, Test 열은 지금 그것을 덮는 것을
가리킨다. 이 change의 RED 실행 기록은 `internal-soak--buildattestation`의 branch-test-map에 있다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` line 472: `if err != nil {` | `runSoakAttest` | no | yes |
| B2 | `if` line 475: `if opts.validity > 0 {` | `runSoakAttest` | no | yes |
| B3 | `if` line 480: `if base := soakSurveyedBase(root); base != "" {` | `runSoakAttest` | no | yes |
| B4 | `if` line 486: `if err != nil {` | `runSoakAttest` | no | yes |
| B5 | `if` line 491: `if err != nil {` | `runSoakAttest` | no | yes |
| B6 | `if` line 497: `if err != nil {` | `runSoakAttest` | no | yes |
| B7 | `if` line 500: `if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {` | `runSoakAttest` | no | yes |
| B8 | `if` line 503: `if err := attest.Save(path, attestation); err != nil {` | `runSoakAttest` | no | yes |
| B9 | `range` line 516: `for _, p := range attestation.SupervisedBy {` | `runSoakAttest` | no | yes |
| B10 | `if` line 518: `if market == "" {` | `runSoakAttest` | no | yes |
| B11 | `if` line 526: `if len(missing) > 0 {` | `runSoakAttest` | no | yes |
| B12 | `range` line 529: `for _, e := range missing {` | `runSoakAttest` | no | yes |
