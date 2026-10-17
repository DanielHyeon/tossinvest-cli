# Function Logic Map: `ValidateDescriptors`

- Source: `internal/strategyflow/registry.go` (28-45)
- Function: `ValidateDescriptors` in package `strategyflow`
- Signature: `ValidateDescriptors(params=1, results=1)`
- File SHA-256: `c7cfd15029a18c87f4de9ff2cb2730280cd1345a6d182b0eee687a11348cbdda`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 4.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Refuses any descriptor set that is not exactly the frozen matrix: right count, no unknown lane, no duplicate, and every field equal. L3 changed the frozen matrix from six to eight, so the count this function enforces moved with it.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyflow/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: the function body executed **6x** under the package suite.

Exact AST return positions: 30:3, 40:4, 44:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 29:2 | arm entered 1x (package suite); entered by `TestValidateDescriptorsRejectsPartialDuplicateUnknownAndMismatched` |
| B2 | range | 33:2 | arm entered 40x (package suite); entered by `TestPairedRegistryCoversAllFourFamiliesInBothMarkets`, `TestPairedRegistryCoversKRUSContinuationReversalWeeklyAndBreakout`, `TestValidateDescriptorsRejectsPartialDuplicateUnknownAndMismatched` |
| B3 | range | 37:2 | arm entered 26x (package suite); entered by `TestPairedRegistryCoversAllFourFamiliesInBothMarkets`, `TestPairedRegistryCoversKRUSContinuationReversalWeeklyAndBreakout`, `TestValidateDescriptorsRejectsPartialDuplicateUnknownAndMismatched` |
| B4 | if | 39:3 | arm entered 3x (package suite); entered by `TestValidateDescriptorsRejectsPartialDuplicateUnknownAndMismatched` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 29:5 |
| `len` | 29:25 |
| `fmt.Errorf` | 30:10 |
| `len` | 30:67 |
| `len` | 30:85 |
| `make` | 32:14 |
| `len` | 32:42 |
| `make` | 36:10 |
| `len` | 36:32 |
| `fmt.Errorf` | 40:11 |

## State mutations and fallbacks

- AST assignments: 5. Defers: 0. Goroutine statements: 0.
- Builds two local maps; mutates nothing outside.

## Safety conclusion

- This is the atomicity gate for the 8-set. It is what makes 'exactly four families in both markets' a checked property rather than a convention.
