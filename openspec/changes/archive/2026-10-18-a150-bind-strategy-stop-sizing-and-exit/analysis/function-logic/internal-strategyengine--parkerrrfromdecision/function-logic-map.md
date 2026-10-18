# Function Logic Map: `parkerrrfromdecision`

- Source: `internal/strategyengine/rr_diagnostic.go`; pre-edit base 28a800f5.
- AST: `ast.json`; risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Caller-supplied numbers are unverified; planned risk is per share, realized denominator is filled account risk. Missing evidence is unknown and invalid evidence is invalid. Sealed decision, lane parser and runtime authority stay unchanged.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return | Required test |
|---|---|---|---|---|
| B1 | currency supplied | local result only | unit is currency/share | existing diagnostic tests; TestRRDiagnosticDecimalBoundary |
| B2 | adjusted flag | local result only | basis marks adjusted | existing diagnostic tests; TestRRDiagnosticDecimalBoundary |
| B3 | planned mandatory evidence missing | local result only | both planned metrics unknown | existing diagnostic tests; TestRRDiagnosticDecimalBoundary |
| B4 | nonpositive, inverted, adjusted, wrong lane/version | local result only | both planned metrics invalid | existing diagnostic tests; TestRRDiagnosticDecimalBoundary |
| B5 | numerator missing | local result only | that metric unknown | existing diagnostic tests; TestRRDiagnosticDecimalBoundary |
| B6 | numerator not positive decimal | local result only | that metric invalid | existing diagnostic tests; TestRRDiagnosticDecimalBoundary |
| B7 | space metric | local result only | entry times percentage / 100 | existing diagnostic tests; TestRRDiagnosticDecimalBoundary |
| B8 | target metric | local result only | target minus entry | existing diagnostic tests; TestRRDiagnosticDecimalBoundary |
| B9 | numerator not positive | local result only | that metric invalid | existing diagnostic tests; TestRRDiagnosticDecimalBoundary |

## Calls and live bindings

`positive`, `decimal`, `nonnegative` call shared exact rational parser; guard inputs locally before these calls. `big.Rat` computes exact ratios. No config, I/O, mutation or network.

## State mutations and fallbacks

Only local output changes; missing inputs return unknown. No fallback to current FX or inferred fills.

## Safety conclusion

Safe boundary: local bounded decimal validation before shared parser; high-risk input validation yes. Preserve all other outcomes and unchanged decision seal.

Source/function binding: `internal/strategyengine/rr_diagnostic.go` / `ParkerRRFromDecision`.


Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.




Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.
