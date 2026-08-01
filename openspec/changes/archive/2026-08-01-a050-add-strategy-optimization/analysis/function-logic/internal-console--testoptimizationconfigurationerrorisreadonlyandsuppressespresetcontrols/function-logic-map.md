# Function Logic Map: `TestOptimizationConfigurationErrorIsReadOnlyAndSuppressesPresetControls`

- Source: `internal/console/optimization_ui_contract_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| invalid known descriptor | registry preserves key with `ConfigurationError`, clears options, marks read-only | `BuildRegistry` | UI must show alert and zero mutation controls |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B8 | field lookup/build/parse checks; alert marker loop; aria-readonly count; forbidden control loop | test assertions only | exact fail-closed regression | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `BuildRegistry` / HTML parser | construct invalid owner metadata and inspect rendered DOM | test-local; no external IO/retry | AST |

## State mutations and fallbacks

- Mutates only the fake commander's test view. No apply/preview call is made.

## Safety conclusion

- Safe edit boundary: descriptor error must suppress all preset forms.
- High-risk impact: no mutation; fail-closed safety test.
