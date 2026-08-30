# Branch Test Map: `ProductionBatchAuthority.For`

- Source: `internal/strategyproposal/production.go`; file SHA-256 `43ebb628cdfef4f891b652e81dc71c677063d0ad4cbbc9d0d3bc3b3cdcb52236`. AST branch positions are authoritative.
- Rows carry measured counts from Go coverage profiles, count mode.
- untagged proposal suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyproposal ./internal/strategyproposal/`
- tagged proposal suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/strategyproposal/`
- tagged engine suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- untagged engine suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`

Mutation receipts for this function (production source mutated, run, restored from a pristine copy taken before the mutation):

| # | mutation | result | killed by |
|---|---|---|---|
| M2 | return `values[0]` whenever the symbol has any lane (previous lot) | KILLED | `TestForRefusesASymbolThatHasMoreThanOneLane` |

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | if at 117:2 | arm entered 2x (untagged proposal suite); arm entered 2x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestForRefusesASymbolThatHasMoreThanOneLane`, `TestForRefusesAnUnknownSymbol` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
