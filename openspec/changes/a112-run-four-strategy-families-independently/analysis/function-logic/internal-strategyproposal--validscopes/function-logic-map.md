# Function Logic Map: `validScopes`

- Source: `internal/strategyproposal/production.go` (552-567)
- Function: `validScopes` in package `strategyproposal`
- Signature: `validScopes(params=2, results=1)`
- File SHA-256: `43ebb628cdfef4f891b652e81dc71c677063d0ad4cbbc9d0d3bc3b3cdcb52236`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 3.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Validates the signed proposal manifest's scope list. Keys on (symbol, lane) so one symbol may legitimately carry several families while a duplicated lane is still refused.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- untagged proposal suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyproposal ./internal/strategyproposal/`
- tagged proposal suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/strategyproposal/`
- tagged engine suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- untagged engine suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: the function body was not executed (untagged proposal suite); executed 6x (tagged proposal suite); not executed (tagged engine suite); not executed (untagged engine suite).

Exact AST return positions: 554:3, 562:4, 566:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 553:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B2 | range | 557:2 | arm not entered (untagged proposal suite); arm entered 10x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestProductionProposalAuthorityFailureIsMarketLocal`, `TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots`, `TestValidScopesAcceptsSeveralFamiliesForOneSymbolAndStillRejectsDuplicateLanes` |
| B3 | if | 561:3 | arm not entered (untagged proposal suite); arm entered 2x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestValidScopesAcceptsSeveralFamiliesForOneSymbolAndStillRejectsDuplicateLanes` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 553:5 |
| `len` | 553:25 |
| `strings.ToUpper` | 561:44 |
| `strings.TrimSpace` | 561:60 |
| `laneMatchesMarket` | 561:336 |

## State mutations and fallbacks

- AST assignments: 3. Defers: 0. Goroutine statements: 0.

## Safety conclusion

Unchanged by this lot; the bundle is refreshed because the file hash moved.
