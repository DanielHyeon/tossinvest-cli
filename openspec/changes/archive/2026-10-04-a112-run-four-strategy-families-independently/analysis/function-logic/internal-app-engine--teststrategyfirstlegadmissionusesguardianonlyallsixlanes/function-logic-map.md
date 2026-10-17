# Function Logic Map: `TestStrategyFirstLegAdmissionUsesGuardianOnlyAllSixLanes`

- Source: `internal/app/engine/strategy_first_leg_admission_test.go` (32-57)
- Function: `TestStrategyFirstLegAdmissionUsesGuardianOnlyAllSixLanes` in package `engine`
- Signature: `TestStrategyFirstLegAdmissionUsesGuardianOnlyAllSixLanes(params=1, results=0)`
- File SHA-256: `367a5e296f03b904d72489ab4dd505f8d0ce93cc74b5a80cd29971ac3434f60f`
- Pinned revision: `base` — the AST and the SHA-256 above are `a8c3d067470fe9cd00523a7629ee93ee05de8e5c`'s file, because the checker requires this record at the frozen comparison base (the function moved or was renamed).
- AST evidence: `ast.json` — AST branches 4.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

The frozen-base record of the guardian-only admission matrix test. Decision 50 extended L3's ledger to this file: the test iterates strategyflow.Descriptors(), so growing the frozen matrix from six to eight made its literal `markets["KR"] != 3 || markets["US"] != 3` assertion false while every subtest still passed. In this worktree it is renamed to `TestStrategyFirstLegAdmissionUsesGuardianOnlyAllEightLanes` and asserts four per market.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: none.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | range | 34:2 | no coverage block maps to this position |
| B2 | if | 44:4 | no coverage block maps to this position |
| B3 | if | 47:4 | no coverage block maps to this position |
| B4 | if | 54:2 | no coverage block maps to this position |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `allSixStrategyFirstLegResults` | 34:23 |
| `t.Run` | 36:3 |
| `firstLegBridgeIssuanceFixture` | 37:16 |
| `string` | 41:13 |
| `admit` | 43:15 |
| `newStrategyFirstLegAdmissionBridge` | 43:15 |
| `context.Background` | 43:72 |
| `t.Fatalf` | 45:5 |
| `t.Fatalf` | 49:5 |
| `string` | 51:12 |
| `t.Fatalf` | 55:3 |

## State mutations and fallbacks

- AST assignments: 7. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
