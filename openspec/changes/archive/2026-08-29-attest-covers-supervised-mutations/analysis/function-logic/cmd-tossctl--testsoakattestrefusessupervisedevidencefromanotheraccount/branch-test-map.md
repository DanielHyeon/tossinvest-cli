# Branch Test Map: `TestSoakAttestRefusesSupervisedEvidenceFromAnotherAccount`

- Source: `cmd/tossctl/soak_test.go`
- Function: `cmd/tossctl/soak_test.go:TestSoakAttestRefusesSupervisedEvidenceFromAnotherAccount`

RED/GREEN은 이 change를 구현하며 실제로 관측한 것이다. RED `no`는 이 change가 그 분기의
동작을 바꾸지 않아 실패 상태를 따로 만들지 않았다는 뜻이며, Test 열은 지금 그것을 덮는 것을
가리킨다. 이 change의 RED 실행 기록은 `internal-soak--buildattestation`의 branch-test-map에 있다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` line 662: `if _, _, err := runCLI(t, "--config-dir", configDir, "soak", "attest"); err == nil {` | `TestSoakAttestRefusesSupervisedEvidenceFromAnotherAccount` | no | yes |
| B2 | `if` line 665: `if _, err := os.Stat(filepath.Join(configDir, attest.FileName)); !os.IsNotExist(err) {` | `TestSoakAttestRefusesSupervisedEvidenceFromAnotherAccount` | no | yes |
