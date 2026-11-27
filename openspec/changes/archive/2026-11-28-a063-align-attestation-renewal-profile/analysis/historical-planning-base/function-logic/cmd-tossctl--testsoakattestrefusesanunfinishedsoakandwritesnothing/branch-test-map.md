# Branch Test Map: `TestSoakAttestRefusesAnUnfinishedSoakAndWritesNothing`

Source: `cmd/tossctl/soak_test.go` (375-394); AST revision: `base`.

| Branch | Raw AST-position condition | Test consequence | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `	if _, _, err := runCLI(t, "--config-dir", configDir, "soak", "run", "--cycles", "1", "--interval", "0"); err != nil {` | `TestSoakAttestRefusesAnUnfinishedSoakAndWritesNothing` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B2 | `	if err == nil {` | `TestSoakAttestRefusesAnUnfinishedSoakAndWritesNothing` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B3 | `	if !strings.Contains(err.Error(), "consecutive") {` | `TestSoakAttestRefusesAnUnfinishedSoakAndWritesNothing` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B4 | `	if _, statErr := os.Stat(filepath.Join(configDir, attest.FileName)); !os.IsNotExist(statErr) {` | `TestSoakAttestRefusesAnUnfinishedSoakAndWritesNothing` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |

This is an AST structural inventory with raw source conditions. It does not turn a current result into historical pre-edit execution evidence.
