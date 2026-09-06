# Function Logic Map: `TestSoakAttestWritesAVerifiableAttestation`

- Source: `cmd/tossctl/soak_test.go` (399-434)
- AST evidence: `ast.json`; revision: `base`.

## Inputs and invariants

This test assembles local fixtures and fails through `testing.T` when an assertion is unmet. It does not place orders or alter runtime configuration.

## Branches and early returns

| Branch | AST kind and raw condition at AST position | Test failure consequence |
|---|---|---|
| B1 | `	if err != nil {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |
| B2 | `	if err != nil {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |
| B3 | `	if err := a.Verify(time.Now(), "123-45-678901", soak.RequiredEndpoints()); err != nil {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |
| B4 | `	if a.SoakDays < 3 {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |
| B5 | `	for _, e := range a.Endpoints {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |
| B6 | `		if !strings.HasPrefix(e, "GET ") {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |
| B7 | `	for _, remaining := range soak.LiveOnlyEndpoints() {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |
| B8 | `		if !strings.Contains(out, remaining) {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |
| B9 | `	if strings.Contains(out, "678901") {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |

## Calls and live bindings

The extractor records the local test helper/assertion calls in `ast.json`. These calls create only isolated fixtures and have no live runtime binding.

## State mutations and fallbacks

The function may write only isolated test fixtures. Its fallback is a test failure; no broker, timer, engine, or operating toggle is invoked.

## Safety conclusion

The AST is a `base` revision. It records required source structure and raw branch conditions; it is not evidence that a pre-edit map was captured during implementation.
