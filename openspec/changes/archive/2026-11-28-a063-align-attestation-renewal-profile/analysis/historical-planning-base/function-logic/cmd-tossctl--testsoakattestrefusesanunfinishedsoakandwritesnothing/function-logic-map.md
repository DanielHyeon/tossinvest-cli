# Function Logic Map: `TestSoakAttestRefusesAnUnfinishedSoakAndWritesNothing`

- Source: `cmd/tossctl/soak_test.go` (375-394)
- AST evidence: `ast.json`; revision: `base`.

## Inputs and invariants

This test assembles local fixtures and fails through `testing.T` when an assertion is unmet. It does not place orders or alter runtime configuration.

## Branches and early returns

| Branch | AST kind and raw condition at AST position | Test failure consequence |
|---|---|---|
| B1 | `	if _, _, err := runCLI(t, "--config-dir", configDir, "soak", "run", "--cycles", "1", "--interval", "0"); err != nil {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |
| B2 | `	if err == nil {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |
| B3 | `	if !strings.Contains(err.Error(), "consecutive") {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |
| B4 | `	if _, statErr := os.Stat(filepath.Join(configDir, attest.FileName)); !os.IsNotExist(statErr) {` | The test reports a failure if this assertion/control path is reached with an unmet expectation. |

## Calls and live bindings

The extractor records the local test helper/assertion calls in `ast.json`. These calls create only isolated fixtures and have no live runtime binding.

## State mutations and fallbacks

The function may write only isolated test fixtures. Its fallback is a test failure; no broker, timer, engine, or operating toggle is invoked.

## Safety conclusion

The AST is a `base` revision. It records required source structure and raw branch conditions; it is not evidence that a pre-edit map was captured during implementation.
