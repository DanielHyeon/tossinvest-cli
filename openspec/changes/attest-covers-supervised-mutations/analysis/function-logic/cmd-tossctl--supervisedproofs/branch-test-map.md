# Branch Test Map: `supervisedProofs`

- Source: `cmd/tossctl/soak.go`
- Function: `cmd/tossctl/soak.go:supervisedProofs`

RED/GREEN은 이 change를 구현하며 실제로 관측한 것이다. RED `no`는 이 change가 그 분기의
동작을 바꾸지 않아 실패 상태를 따로 만들지 않았다는 뜻이며, Test 열은 지금 그것을 덮는 것을
가리킨다. 이 change의 RED 실행 기록은 `internal-soak--buildattestation`의 branch-test-map에 있다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` line 405: `if len(opts.verifyRecords) > 0 {` | `supervisedProofs` | no | yes |
| B2 | `else` line 411: `} else {` | `supervisedProofs` | no | yes |
| B3 | `range` line 406: `for _, p := range opts.verifyRecords {` | `supervisedProofs` | no | yes |
| B4 | `if` line 407: `if trimmed := strings.TrimSpace(p); trimmed != "" {` | `supervisedProofs` | no | yes |
| B5 | `range` line 412: `for _, market := range []string{verifylive.MarketKR, verifylive.MarketUS} {` | `supervisedProofs` | no | yes |
| B6 | `if` line 414: `if err != nil {` | `supervisedProofs` | no | yes |
| B7 | `range` line 422: `for _, s := range sources {` | `supervisedProofs` | no | yes |
| B8 | `if` line 424: `if err != nil {` | `supervisedProofs` | no | yes |
| B9 | `if` line 425: `if errors.Is(err, os.ErrNotExist) {` | `supervisedProofs` | no | yes |
| B10 | `if` line 433: `if len(evidence.AccountRefs) > 1 {` | `supervisedProofs` | no | yes |
| B11 | `if` line 440: `if len(evidence.AccountRefs) == 1 {` | `supervisedProofs` | no | yes |
| B12 | `range` line 443: `for endpoint, at := range evidence.Endpoints {` | `supervisedProofs` | no | yes |
| B13 | `range` line 457: `for _, e := range soak.LiveOnlyEndpoints() {` | `supervisedProofs` | no | yes |
| B14 | `range` line 461: `for _, p := range proofs {` | `supervisedProofs` | no | yes |
| B15 | `if` line 462: `if allowed[strings.ToUpper(p.Endpoint)] {` | `supervisedProofs` | no | yes |
