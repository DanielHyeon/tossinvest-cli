# Function Logic Map: `TestBuildAttestationRefusesAnIncompleteSoak`

- Source: `internal/soak/attest_test.go` (233-252)
- AST evidence: `ast.json`; revision: `current`.

## Inputs and invariants

This test assembles local fixtures and fails through `testing.T` when an assertion is unmet. It does not place orders or alter runtime configuration.

## Branches and early returns

| Branch | AST kind and raw condition at AST position | Test failure consequence |
|---|---|---|
| B1 | `	if err == nil {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |
| B2 | `	if !errors.Is(err, soak.ErrIncomplete) {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |
| B3 | `	if !errors.As(err, &incomplete) {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |
| B4 | `	if !strings.Contains(err.Error(), "unattended credential refresh is proven") {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |
| B5 | `	if codes := incomplete.ReasonCodes(); len(codes) == 0 || codes[0] != soak.ReasonStreak {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |

## Calls and live bindings

The extractor records the local test helper/assertion calls in `ast.json`. These calls create only isolated fixtures and have no live runtime binding.

## State mutations and fallbacks

The function may write only isolated test fixtures. Its fallback is a test failure; no broker, timer, engine, or operating toggle is invoked.

## Safety conclusion

The AST is a `current` revision. It records required source structure and raw branch conditions; it is not evidence that a pre-edit map was captured during implementation.
