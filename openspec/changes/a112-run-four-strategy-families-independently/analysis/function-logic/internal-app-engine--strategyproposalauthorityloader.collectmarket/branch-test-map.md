# Branch Test Map: `strategyProposalAuthorityLoader.collectMarket`

- Source: `internal/app/engine/strategy_proposal_authority.go`; file SHA-256 `88e06b6c841ba30cb1c3107fba33c134c82b34f871dec646ee92b739a2e58c94`. AST branch positions are authoritative.
- Rows carry measured counts from Go coverage profiles, count mode.
- untagged proposal suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyproposal ./internal/strategyproposal/`
- tagged proposal suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/strategyproposal/`
- tagged engine suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- untagged engine suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`

Mutation receipts for this function (production source mutated, run, restored from a pristine copy taken before the mutation):

| # | mutation | result | killed by |
|---|---|---|---|
| M15 | delete the market-level ambiguity guard | KILLED | `TestProposalCollectMarketClosesTheMarketBeforeBuildingEntries` |
| M15b | ask `Ambiguous` but do not return | KILLED | `TestProposalCollectMarketClosesTheMarketBeforeBuildingEntries` |

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | if at 164:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B2 | if at 167:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B3 | if at 170:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B4 | if at 175:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B5 | if at 179:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm entered 2x (tagged engine suite); arm not entered (untagged engine suite); entered by `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently` |
| B6 | range at 185:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm entered 4x (tagged engine suite); arm not entered (untagged engine suite); entered by `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently` |
| B7 | if at 187:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B8 | if at 199:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm entered 1x (tagged engine suite); arm not entered (untagged engine suite); entered by `TestStrategyProposalAuthorityKeepsMarketFailureLocal` |
| B9 | range at 206:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm entered 3x (tagged engine suite); arm not entered (untagged engine suite); entered by `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently` |
| B10 | if at 207:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B11 | range at 215:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm entered 3x (tagged engine suite); arm not entered (untagged engine suite); entered by `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently` |
| B12 | if at 217:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B13 | if at 224:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B14 | range at 231:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm entered 3x (tagged engine suite); arm not entered (untagged engine suite); entered by `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
