# Branch Test Map: `TestBuildAttestationRefusesAnIncompleteSoak`

Source: `internal/soak/attest_test.go` (233-252); AST revision: `current`.

| Branch | Raw AST-position condition | Test consequence | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `	if err == nil {` | `TestBuildAttestationRefusesAnIncompleteSoak` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B2 | `	if !errors.Is(err, soak.ErrIncomplete) {` | `TestBuildAttestationRefusesAnIncompleteSoak` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B3 | `	if !errors.As(err, &incomplete) {` | `TestBuildAttestationRefusesAnIncompleteSoak` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B4 | `	if !strings.Contains(err.Error(), "unattended credential refresh is proven") {` | `TestBuildAttestationRefusesAnIncompleteSoak` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B5 | `	if codes := incomplete.ReasonCodes(); len(codes) == 0 || codes[0] != soak.ReasonStreak {` | `TestBuildAttestationRefusesAnIncompleteSoak` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |

This is an AST structural inventory with raw source conditions. It does not turn a current result into historical pre-edit execution evidence.
