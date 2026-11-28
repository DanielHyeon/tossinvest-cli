# Function Logic Map: `TestSoakAndLiveEndpointsCoverTheEngineInterlock`

- Source: `cmd/tossctl/soak_test.go` (638-657)
- AST evidence: `ast.json`; revision: `current`.

## Inputs and invariants

This test assembles local fixtures and fails through `testing.T` when an assertion is unmet. It does not place orders or alter runtime configuration.

## Branches and early returns

| Branch | AST kind and raw condition at AST position | Test failure consequence |
|---|---|---|
| B1 | `	for _, e := range append(soak.RequiredEndpoints(), soak.LiveOnlyEndpoints()...) {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |
| B2 | `	for _, want := range engine.RequiredEndpoints() {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |
| B3 | `		if !covered[want] {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |
| B4 | `	for _, endpoint := range engine.RequiredEndpoints() {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |
| B5 | `		if endpoint == "GET /api/v1/exchange-rate" {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |
| B6 | `	if covered["GET /api/v1/exchange-rate"] {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |

## Calls and live bindings

The extractor records the local test helper/assertion calls in `ast.json`. These calls create only isolated fixtures and have no live runtime binding.

## State mutations and fallbacks

The function may write only isolated test fixtures. Its fallback is a test failure; no broker, timer, engine, or operating toggle is invoked.

## Safety conclusion

The AST is a `current` revision. It records required source structure and raw branch conditions; it is not evidence that a pre-edit map was captured during implementation.
