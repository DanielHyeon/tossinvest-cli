# Function Logic Map: `TestOptimizationUsesOnePresetPreviewFlowWithoutClientDraft`

- Source: `internal/console/optimization_ui_contract_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| rendered exit-protection page | exactly three server preset forms/buttons, no select/textarea/script/client draft | fake owner registry + template | test fails on extra/missing control or state |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B9 | parse success; exact form/button counts; forbidden element loop; required marker loop; invented-state loop | test assertions only | exact regression message | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| HTML parser / DOM helpers | inspect actual response DOM | local/pure; no retry | AST |

## State mutations and fallbacks

- Test-local state only. Explicitly rejects local/session storage and multi-field draft affordances.

## Safety conclusion

- Safe edit boundary: one-field server-preset UI contract.
- High-risk impact: no mutation; protects high-risk UI boundary.
