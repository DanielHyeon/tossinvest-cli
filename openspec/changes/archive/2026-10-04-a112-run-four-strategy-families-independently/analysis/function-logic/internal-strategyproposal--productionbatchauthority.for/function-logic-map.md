# Function Logic Map: `ProductionBatchAuthority.For`

- Source: `internal/strategyproposal/production.go` (170-176)
- Function: `ProductionBatchAuthority.For` in package `strategyproposal`
- Signature: `ProductionBatchAuthority.For(params=1, results=2)`
- File SHA-256: `b6e54b502e5092745426f8f4a37e4a02777d525a2099aa90de9f7379ee4a2c18`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 1.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Returns the proposal for a symbol only when that symbol has exactly one. Two or more is a refusal, not a choice: choosing here would be the pre-evaluation selection task 4.3.1 removed. The refusal carries no reason, which is why `Ambiguous` was added beside it — a caller that cannot tell 'no proposal' from 'too many' turns one symbol's fail-closed into the market's fail-open (review finding C2).

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- untagged proposal suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyproposal ./internal/strategyproposal/`
- tagged proposal suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/strategyproposal/`
- tagged engine suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- untagged engine suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: the function body was executed 4x (untagged proposal suite); executed 7x (tagged proposal suite); executed 21x (tagged engine suite); not executed (untagged engine suite).

Exact AST return positions: 173:3, 175:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 172:2 | arm entered 2x (untagged proposal suite); arm entered 2x (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); entered by `TestForRefusesASymbolThatHasMoreThanOneLane`, `TestForRefusesAnUnknownSymbol` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `authority.LanesFor` | 171:12 |
| `len` | 172:5 |

## State mutations and fallbacks

- AST assignments: 1. Defers: 0. Goroutine statements: 0.

## Safety conclusion

`For` itself is correct and unchanged. What changed beside it is that the ambiguity is now observable from outside, so the engine can close the market instead of dropping the symbol.
