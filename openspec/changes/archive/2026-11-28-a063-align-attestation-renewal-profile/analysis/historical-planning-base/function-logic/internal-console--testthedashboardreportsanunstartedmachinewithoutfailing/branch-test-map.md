# Branch Test Map: `TestTheDashboardReportsAnUnstartedMachineWithoutFailing`

Source: `internal/console/console_test.go` (647-660); AST revision: `base`.

| Branch | Raw AST-position condition | Test consequence | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `	for _, want := range []string{"soak", "attestation", "tossctl soak run"} {` | `TestTheDashboardReportsAnUnstartedMachineWithoutFailing` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B2 | `		if !strings.Contains(page, want) {` | `TestTheDashboardReportsAnUnstartedMachineWithoutFailing` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B3 | `	if !strings.Contains(page, "게이트를 켜지 않는다") {` | `TestTheDashboardReportsAnUnstartedMachineWithoutFailing` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |

This is an AST structural inventory with raw source conditions. It does not turn a current result into historical pre-edit execution evidence.
