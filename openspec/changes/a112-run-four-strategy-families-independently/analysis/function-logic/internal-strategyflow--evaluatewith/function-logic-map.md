# Function Logic Map: `evaluateWith`

- Source: `internal/strategyflow/flow.go` (27-124)
- Function: `evaluateWith` in package `strategyflow`
- Signature: `evaluateWith(params=3, results=1)`
- File SHA-256: `c4e9738af8202122e48460436ce5cf7717b8ec8af4495b1b581171114dfe06ce`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 12.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

The whole composition: validate the approved candidate, route, pick the canonical descriptor, bind the lane input, evaluate, then seal lineage and execution terms. L3 changed its router parameter type to `RouteSetResult` and made it select this request's own decision out of the candidate set rather than accepting a pre-chosen winner.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyflow/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: the function body executed **41x** under the package suite.

Exact AST return positions: 31:3, 37:3, 45:3, 54:3, 59:3, 72:3, 78:3, 86:3, 91:3, 96:3, 109:3, 118:3, 123:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 29:2 | arm never entered: count 0 in every profile measured for this function |
| B2 | if | 34:2 | arm never entered: count 0 in every profile measured for this function |
| B3 | if | 41:2 | arm entered 1x (package suite); entered by `TestRouterRefusalSkipsLaneAndUnsupportedBindingIsTyped` |
| B4 | if | 51:2 | arm entered 1x (package suite); entered by `TestRouterRefusalSkipsLaneAndUnsupportedBindingIsTyped` |
| B5 | if | 56:2 | arm never entered: count 0 in every profile measured for this function |
| B6 | if | 69:2 | arm entered 1x (package suite); entered by `TestWrongMarketLaneInputAndForgedAcceptedLineageFailClosed` |
| B7 | if | 75:2 | arm never entered: count 0 in every profile measured for this function |
| B8 | if | 82:2 | arm entered 1x (package suite); entered by `TestLaneRefusalPreservesFirstTypedCodeAndRouterLaneEvidence` |
| B9 | if | 88:2 | arm entered 1x (package suite); entered by `TestWrongMarketLaneInputAndForgedAcceptedLineageFailClosed` |
| B10 | if | 93:2 | arm entered 1x (package suite); entered by `TestExistingOwnerRoutePinsCampaignAndAcceptsLaneConfigLineage` |
| B11 | if | 105:2 | arm entered 35x (package suite); entered by `TestAcceptedProjectionCoversAllFourFamiliesInBothMarketsTogether`, `TestAcceptedProjectionRejectsRefusedImpureAndMutatedResults`, `TestApprovedCandidatesRouteAndEvaluateAllPairedBindingsWithCompleteLineage` |
| B12 | if | 114:2 | arm entered 1x (package suite); entered by `TestEvaluateRejectsAcceptedLaneWithoutExactExecutionTerms` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `validApproved` | 29:6 |
| `candidateLineage` | 33:13 |
| `validScope` | 34:6 |
| `sealLineage` | 36:20 |
| `route` | 40:12 |
| `string` | 43:23 |
| `sealLineage` | 44:20 |
| `selectRouteDecision` | 50:46 |
| `sealLineage` | 53:20 |
| `validRouteLineage` | 56:99 |
| `sealLineage` | 58:20 |
| `sealLineage` | 71:20 |
| `lanes.lookup` | 74:17 |
| `sealLineage` | 77:20 |
| `binding.evaluate` | 81:15 |
| `sealLineage` | 85:20 |
| `matchingLaneLineage` | 88:6 |
| `sealLineage` | 90:20 |
| `sealLineage` | 95:20 |
| `sealLineage` | 108:20 |
| `sealLineage` | 112:21 |
| `sealExecutionTerms` | 113:15 |
| `sealLineage` | 117:20 |

## State mutations and fallbacks

- AST assignments: 53. Defers: 0. Goroutine statements: 0.
- Builds one `Result` value. No journal, no broker, no goroutine.

## Safety conclusion

- Every refusal path is typed and first-wins, so a later refusal cannot overwrite the first typed code — that ordering is what the lineage evidence depends on. Any new branch must preserve it.
