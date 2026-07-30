# Branch Test Map: `TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun`

- Source: `cmd/tossctl/soak_test.go`
- Function: `cmd/tossctl/soak_test.go:TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun`

RED/GREEN은 이 change를 구현하며 실제로 관측한 것이다. RED `no`는 이 change가 그 분기의
동작을 바꾸지 않아 실패 상태를 따로 만들지 않았다는 뜻이며, Test 열은 지금 그것을 덮는 것을
가리킨다. 이 change의 RED 실행 기록은 `internal-soak--buildattestation`의 branch-test-map에 있다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` line 631: `if _, _, err := runCLI(t, "--config-dir", configDir, "soak", "attest"); err != nil {` | `TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun` | no | yes |
| B2 | `if` line 635: `if err != nil {` | `TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun` | no | yes |
| B3 | `if` line 639: `if missing := a.MissingEndpoints(engine.RequiredEndpoints()); len(missing) != 0 {` | `TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun` | no | yes |
| B4 | `if` line 642: `if len(a.SupervisedBy) != len(soak.LiveOnlyEndpoints()) {` | `TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun` | no | yes |
| B5 | `range` line 646: `for _, p := range a.SupervisedBy {` | `TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun` | no | yes |
| B6 | `if` line 647: `if p.Source == "" {` | `TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun` | no | yes |
