# Branch Test Map: `TestSoakAndLiveEndpointsCoverTheEngineInterlock`

Source: `cmd/tossctl/soak_test.go` (638-657); AST revision: `current`.

| Branch | Raw AST-position condition | Test consequence | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `	for _, e := range append(soak.RequiredEndpoints(), soak.LiveOnlyEndpoints()...) {` | `TestSoakAndLiveEndpointsCoverTheEngineInterlock` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B2 | `	for _, want := range engine.RequiredEndpoints() {` | `TestSoakAndLiveEndpointsCoverTheEngineInterlock` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B3 | `		if !covered[want] {` | `TestSoakAndLiveEndpointsCoverTheEngineInterlock` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B4 | `	for _, endpoint := range engine.RequiredEndpoints() {` | `TestSoakAndLiveEndpointsCoverTheEngineInterlock` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B5 | `		if endpoint == "GET /api/v1/exchange-rate" {` | `TestSoakAndLiveEndpointsCoverTheEngineInterlock` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |
| B6 | `	if covered["GET /api/v1/exchange-rate"] {` | `TestSoakAndLiveEndpointsCoverTheEngineInterlock` fails when this assertion/control path exposes a mismatch. | not recorded | reported separately |

This is an AST structural inventory with raw source conditions. It does not turn a current result into historical pre-edit execution evidence.
