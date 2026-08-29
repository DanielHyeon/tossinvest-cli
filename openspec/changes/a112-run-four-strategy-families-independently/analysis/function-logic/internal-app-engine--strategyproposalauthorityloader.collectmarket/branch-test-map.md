# Branch Test Map: `strategyProposalAuthorityLoader.collectMarket`

- Source: `internal/app/engine/strategy_proposal_authority.go`; file SHA-256 `61724700e3ecd848b656a37d5f8632a17aed7e02639179d1a5f62a1f508bd27e`. AST branch positions are authoritative.
- Rows carry measured counts from Go coverage profiles, count mode.
- engine tagged suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine,./internal/strategyarbiter,./internal/strategyproposal,./internal/strategyrouter ./internal/app/engine/`
- engine untagged suite: `go test -count=1 -covermode=count -coverpkg=./internal/app/engine,./internal/strategyarbiter,./internal/strategyproposal,./internal/strategyrouter ./internal/app/engine/`
- proposal tagged suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine,./internal/strategyarbiter,./internal/strategyproposal,./internal/strategyrouter ./internal/strategyproposal/`
- Per-test attribution set: the seven `Test*` functions that can reach `strategyProposalAuthorityLoader.collectMarket` — the six in `a112_arbitration_test.go` and `strategy_proposal_authority_test.go` plus none elsewhere, because no other engine test constructs a proposal loader or a production assembly. This is the complete reaching set, not a sample.

Mutation receipts for this function (production source mutated, run, restored from a pristine copy taken before the mutation):

| # | mutation | result | killed by |
|---|---|---|---|
| M-E1 | turn the arbitration refusal into `continue` so only that symbol is dropped | KILLED | `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol` |
| M-E2a | always take `lanes[0]` instead of the arbitrated index | KILLED | `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket` |
| M-E2b | always take `lanes[len(lanes)-1]` instead of the arbitrated index | KILLED | `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket` |

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | if at 170:2 | arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); entered by `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused` |
| B2 | if at 173:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); no per-test profile in the attribution set entered it |
| B3 | if at 176:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); no per-test profile in the attribution set entered it |
| B4 | if at 181:2 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); no per-test profile in the attribution set entered it |
| B5 | if at 185:2 | arm entered 6x (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); entered by `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated` |
| B6 | range at 191:2 | arm entered 17x (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); entered by `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated` |
| B7 | if at 193:3 | arm not entered (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); no per-test profile in the attribution set entered it |
| B8 | if at 205:2 | arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); entered by `TestStrategyProposalAuthorityKeepsMarketFailureLocal` |
| B9 | range at 215:2 | arm entered 14x (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); entered by `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated` |
| B10 | if at 217:3 | arm entered 3x (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); entered by `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated` |
| B11 | if at 222:3 | arm entered 3x (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); entered by `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused` |
| B12 | if at 233:2 | arm entered 2x (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); entered by `TestSymbolsWithNoProposalAtAllAreCountedRefusedRatherThanArbitrated` |
| B13 | range at 240:2 | arm entered 8x (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); entered by `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
