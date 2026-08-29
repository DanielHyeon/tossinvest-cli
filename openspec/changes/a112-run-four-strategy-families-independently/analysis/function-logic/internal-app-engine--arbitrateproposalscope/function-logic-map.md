# Function Logic Map: `arbitrateProposalScope`

- Source: `internal/app/engine/strategy_market_coordinator.go` (19-29)
- Function: `arbitrateProposalScope` in package `engine`
- Signature: `arbitrateProposalScope(params=5, results=1)`
- File SHA-256: `7eab58c6298ac39de1c5d20eba9a196e9211e6a5d691b1e26481533213d900c6`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 1.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

한 종목의 모든 가족 제안을 `strategyarbiter.Arbitrate` 가 읽을 모양으로 모아 넘긴다. 여기서는 아무것도 고르지 않는다.
기대 범위의 종목은 경로 권한이 아니라 **승인된 후보**에서 읽는다 — 권한이 스스로 말한 종목으로 그 권한을 검사하면
어긋남을 잡으려던 검사가 언제나 참이 되어 아무것도 잡지 못한다.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- engine tagged suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine,./internal/strategyarbiter,./internal/strategyproposal,./internal/strategyrouter ./internal/app/engine/`
- engine untagged suite: `go test -count=1 -covermode=count -coverpkg=./internal/app/engine,./internal/strategyarbiter,./internal/strategyproposal,./internal/strategyrouter ./internal/app/engine/`
- proposal tagged suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine,./internal/strategyarbiter,./internal/strategyproposal,./internal/strategyrouter ./internal/strategyproposal/`
- Per-test attribution set: the seven `Test*` functions that can reach `strategyProposalAuthorityLoader.collectMarket` — the six in `a112_arbitration_test.go` and `strategy_proposal_authority_test.go` plus none elsewhere, because no other engine test constructs a proposal loader or a production assembly. This is the complete reaching set, not a sample.
- Measured entry: executed 14x in the engine tagged suite; never in the untagged suites.

Exact AST return positions: 26:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | range | 23:2 | arm entered 14x (engine tagged suite); arm not entered (engine untagged suite); no cover block (proposal tagged suite); entered by `TestThreeFamiliesOnOneSymbolNowSelectTheHighestScoreInsteadOfClosingTheMarket`, `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol`, `TestAnUncalibratedMarketRefusesEvenASingleProposal`, `TestAProposalMeasuredAgainstAnotherSymbolsRouteAuthorityIsRefused`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently`, `TestStrategyProposalAuthorityKeepsMarketFailureLocal` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `make` | 22:15 |
| `len` | 22:51 |
| `append` | 24:15 |
| `lane.Proposal` | 24:66 |
| `strategyarbiter.Arbitrate` | 26:9 |
| `strategyRouterMarket` | 26:91 |
| `route.approved.Symbol` | 27:11 |
| `route.route.Request` | 27:56 |

## State mutations and fallbacks

- AST assignments: 2. Defers: 0. Goroutine statements: 0.

## Safety conclusion

선택 규칙을 복제하지 않는다. 규칙이 두 곳에 있으면 언젠가 두 곳이 서로 다른 답을 낸다.
