# Function Logic Map: `canonicalDescriptor`

- Source: `internal/strategyflow/registry.go` (112-120)
- Function: `canonicalDescriptor` in package `strategyflow`
- Signature: `canonicalDescriptor(params=1, results=2)`
- File SHA-256: `c7cfd15029a18c87f4de9ff2cb2730280cd1345a6d182b0eee687a11348cbdda`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 2.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Maps a route decision onto the one frozen descriptor with the same key, or refuses. Its loop now walks eight entries rather than six; the comparison itself is unchanged.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyflow/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: the function body executed **40x** under the package suite.

Exact AST return positions: 116:4, 119:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | range | 114:2 | arm entered 146x (package suite); entered by `TestAcceptedProjectionCoversAllFourFamiliesInBothMarketsTogether`, `TestAcceptedProjectionRejectsRefusedImpureAndMutatedResults`, `TestApprovedCandidatesRouteAndEvaluateAllPairedBindingsWithCompleteLineage` |
| B2 | if | 115:3 | arm entered 39x (package suite); entered by `TestAcceptedProjectionCoversAllFourFamiliesInBothMarketsTogether`, `TestAcceptedProjectionRejectsRefusedImpureAndMutatedResults`, `TestApprovedCandidatesRouteAndEvaluateAllPairedBindingsWithCompleteLineage` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `descriptorFor` | 113:11 |
| `keyFor` | 115:6 |
| `keyFor` | 115:28 |

## State mutations and fallbacks

- AST assignments: 1. Defers: 0. Goroutine statements: 0.
- None. Pure lookup over the package-level frozen slice.

## Safety conclusion

- Fails closed: an unknown or partially-matching key returns `false` and the caller must refuse rather than guess a family.
