# Branch Test Map: `TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun`

Source: `cmd/tossctl/soak_test.go` (803-834); AST revision: `current`.

| Branch | Raw AST-position condition | Test consequence | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `	if _, _, err := runCLI(t, "--config-dir", configDir, "soak", "attest"); err != nil {` | `TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B2 | `	if err != nil {` | `TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B3 | `	if missing := a.MissingEndpoints(engine.RequiredEndpoints()); len(missing) != 0 {` | `TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B4 | `	if len(a.SupervisedBy) != len(soak.LiveOnlyEndpoints()) {` | `TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B5 | `	for _, p := range a.SupervisedBy {` | `TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B6 | `		if p.Source == "" {` | `TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B7 | `	for _, endpoint := range a.Endpoints {` | `TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B8 | `		if endpoint == "GET /api/v1/exchange-rate" {` | `TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |

This is an AST structural inventory with raw source conditions. It does not turn a current result into historical pre-edit execution evidence.
