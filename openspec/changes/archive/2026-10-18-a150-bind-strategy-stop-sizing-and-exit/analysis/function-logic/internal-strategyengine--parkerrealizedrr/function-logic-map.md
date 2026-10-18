# Function Logic Map: `parkerrealizedrr`

- Source: `internal/strategyengine/rr_diagnostic.go`; pre-edit base 28a800f5.
- AST: `ast.json`; risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Caller-supplied numbers are unverified; planned risk is per share, realized denominator is filled account risk. Missing evidence is unknown and invalid evidence is invalid. Sealed decision, lane parser and runtime authority stay unchanged.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return | Required test |
|---|---|---|---|---|
| B1 | required fields missing or finality false | local result only | unknown; do not parse | existing diagnostic tests; TestRRDiagnosticDecimalBoundary |
| B2 | invalid decimal, risk, unit, FX, currency or timestamp | local result only | invalid; no result | existing diagnostic tests; TestRRDiagnosticDecimalBoundary |

## Calls and live bindings

`positive`, `decimal`, `nonnegative` call shared exact rational parser; guard inputs locally before these calls. `big.Rat` computes exact ratios. No config, I/O, mutation or network.

## State mutations and fallbacks

Only local output changes; missing inputs return unknown. No fallback to current FX or inferred fills.

## Safety conclusion

Safe boundary: local bounded decimal validation before shared parser; high-risk input validation yes. Preserve all other outcomes and unchanged decision seal.

Source/function binding: `internal/strategyengine/rr_diagnostic.go` / `ParkerRealizedRR`.


Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.




Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.
