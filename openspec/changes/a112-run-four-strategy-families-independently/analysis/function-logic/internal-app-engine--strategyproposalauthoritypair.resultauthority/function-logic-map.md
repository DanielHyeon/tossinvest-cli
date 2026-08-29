# Function Logic Map: `strategyProposalAuthorityPair.ResultAuthority`

- Source: `internal/app/engine/strategy_proposal_authority.go` (96-104)
- Function: `strategyProposalAuthorityPair.ResultAuthority` in package `engine`
- Signature: `strategyProposalAuthorityPair.ResultAuthority(params=0, results=1)`
- File SHA-256: `88e06b6c841ba30cb1c3107fba33c134c82b34f871dec646ee92b739a2e58c94`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 1.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Projects a market's proposal entries to at most one result, and only when the market has exactly one valid entry.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- untagged proposal suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyproposal ./internal/strategyproposal/`
- tagged proposal suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/strategyproposal/`
- tagged engine suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- untagged engine suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: the function body was not executed (untagged proposal suite); not executed (tagged proposal suite); executed 2x (tagged engine suite); not executed (untagged engine suite).

Exact AST return positions: 99:4, 101:3, 103:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 98:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 98:6 |
| `ValidProposal` | 98:34 |
| `value.entries.authority.Proposal` | 98:34 |
| `value.entries.authority.Proposal` | 101:77 |
| `convert` | 103:70 |
| `convert` | 103:110 |

## State mutations and fallbacks

- AST assignments: 1. Defers: 0. Goroutine statements: 0.

## Safety conclusion

Not edited by this lot; the bundle is refreshed because the file hash moved. Its `len(entries) != 1` gate is one of the three C2 readers and is now fed a list that no longer shrinks on an ambiguity refusal.
