# Branch Test Map: `TestSoakAttestIgnoresSupervisedEndpointsTheGateDoesNotRequire`

- Source: `cmd/tossctl/soak_test.go`
- Function: `cmd/tossctl/soak_test.go:TestSoakAttestIgnoresSupervisedEndpointsTheGateDoesNotRequire`

RED/GREEN은 이 change를 구현하며 실제로 관측한 것이다. RED `no`는 이 change가 그 분기의
동작을 바꾸지 않아 실패 상태를 따로 만들지 않았다는 뜻이며, Test 열은 지금 그것을 덮는 것을
가리킨다. 이 change의 RED 실행 기록은 `internal-soak--buildattestation`의 branch-test-map에 있다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` line 682: `if err != nil {` | `TestSoakAttestIgnoresSupervisedEndpointsTheGateDoesNotRequire` | no | yes |
| B2 | `if` line 685: `if err := rec.Append(verifylive.Entry{` | `TestSoakAttestIgnoresSupervisedEndpointsTheGateDoesNotRequire` | no | yes |
| B3 | `if` line 698: `if _, _, err := runCLI(t, "--config-dir", configDir, "soak", "attest"); err != nil {` | `TestSoakAttestIgnoresSupervisedEndpointsTheGateDoesNotRequire` | no | yes |
| B4 | `if` line 702: `if err != nil {` | `TestSoakAttestIgnoresSupervisedEndpointsTheGateDoesNotRequire` | no | yes |
| B5 | `range` line 705: `for _, e := range a.Endpoints {` | `TestSoakAttestIgnoresSupervisedEndpointsTheGateDoesNotRequire` | no | yes |
| B6 | `if` line 706: `if strings.Contains(e, "conditional-orders") {` | `TestSoakAttestIgnoresSupervisedEndpointsTheGateDoesNotRequire` | no | yes |
| B7 | `range` line 710: `for _, p := range a.SupervisedBy {` | `TestSoakAttestIgnoresSupervisedEndpointsTheGateDoesNotRequire` | no | yes |
| B8 | `if` line 711: `if strings.HasPrefix(p.Endpoint, "GET ") {` | `TestSoakAttestIgnoresSupervisedEndpointsTheGateDoesNotRequire` | no | yes |
