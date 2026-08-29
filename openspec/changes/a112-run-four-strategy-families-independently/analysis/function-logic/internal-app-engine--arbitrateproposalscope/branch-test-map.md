# Branch Test Map: `arbitrateProposalScope`

- Source: `internal/app/engine/strategy_market_coordinator.go`; file SHA-256 `7eab58c6298ac39de1c5d20eba9a196e9211e6a5d691b1e26481533213d900c6`. AST branch positions are authoritative.
- Rows carry measured counts from Go coverage profiles, count mode.
- engine tagged suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine,./internal/strategyarbiter,./internal/strategyproposal,./internal/strategyrouter ./internal/app/engine/`
- engine untagged suite: `go test -count=1 -covermode=count -coverpkg=./internal/app/engine,./internal/strategyarbiter,./internal/strategyproposal,./internal/strategyrouter ./internal/app/engine/`
- proposal tagged suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine,./internal/strategyarbiter,./internal/strategyproposal,./internal/strategyrouter ./internal/strategyproposal/`
- Per-test attribution set: the seven `Test*` functions that can reach `strategyProposalAuthorityLoader.collectMarket` — the six in `a112_arbitration_test.go` and `strategy_proposal_authority_test.go` plus none elsewhere, because no other engine test constructs a proposal loader or a production assembly. This is the complete reaching set, not a sample.

Mutation receipts for this function (production source mutated, run, restored from a pristine copy taken before the mutation):

| # | mutation | result | killed by |
|---|---|---|---|
| M-E3 | read the expected symbol from the route authority instead of the approved candidate | KILLED | `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused` |

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | range at 23:2 | arm entered 14x (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); entered by `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
