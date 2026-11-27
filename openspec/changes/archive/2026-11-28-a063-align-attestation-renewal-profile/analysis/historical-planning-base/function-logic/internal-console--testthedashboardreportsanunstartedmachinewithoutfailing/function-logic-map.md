# Function Logic Map: `TestTheDashboardReportsAnUnstartedMachineWithoutFailing`

- Source: `internal/console/console_test.go` (647-660)
- AST evidence: `ast.json`; revision: `base`.

## Inputs and invariants

This test assembles local fixtures and fails through `testing.T` when an assertion is unmet. It does not place orders or alter runtime configuration.

## Branches and early returns

| Branch | AST kind and raw condition at AST position | Test failure consequence |
|---|---|---|
| B1 | `	for _, want := range []string{"soak", "attestation", "tossctl soak run"} {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |
| B2 | `		if !strings.Contains(page, want) {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |
| B3 | `	if !strings.Contains(page, "게이트를 켜지 않는다") {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |

## Calls and live bindings

The extractor records the local test helper/assertion calls in `ast.json`. These calls create only isolated fixtures and have no live runtime binding.

## State mutations and fallbacks

The function may write only isolated test fixtures. Its fallback is a test failure; no broker, timer, engine, or operating toggle is invoked.

## Safety conclusion

The AST is a `base` revision. It records required source structure and raw branch conditions; it is not evidence that a pre-edit map was captured during implementation.
