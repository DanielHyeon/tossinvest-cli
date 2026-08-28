# Function Logic Map: `validScopes`

- Source: `internal/strategyproposal/production.go` (529-544)
- Function: `validScopes` in package `strategyproposal`
- Signature: `validScopes(params=2, results=1)`
- File SHA-256: `e2285c5ef57e399bf3bf2ca3a0e91b7449b2c152dd9623d5a617454f934082ad`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 3.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Validates the manifest scope list: bounded length, ascending unique key, and every identity field present. L3 changed the ordering and uniqueness unit from the symbol to the (symbol, laneID) pair, because one symbol may now carry several families.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyproposal/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: the function body executed **6x** under the package suite.

Exact AST return positions: 531:3, 539:4, 543:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 530:2 | arm never entered: count 0 in every profile measured for this function |
| B2 | range | 534:2 | arm entered 10x (package suite); entered by `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots`, `TestValidScopesAcceptsSeveralFamiliesForOneSymbolAndStillRejectsDuplicateLanes` |
| B3 | if | 538:3 | arm entered 2x (package suite); entered by `TestValidScopesAcceptsSeveralFamiliesForOneSymbolAndStillRejectsDuplicateLanes` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 530:5 |
| `len` | 530:25 |
| `strings.ToUpper` | 538:44 |
| `strings.TrimSpace` | 538:60 |
| `laneMatchesMarket` | 538:336 |

## State mutations and fallbacks

- AST assignments: 3. Defers: 0. Goroutine statements: 0.
- None.

## Safety conclusion

- Strictly ascending keys make duplicates impossible rather than merely detected. Had the unit stayed the symbol, a second family's scope would have compared as out-of-order and the whole manifest would have been refused.
