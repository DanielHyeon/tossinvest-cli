# Function Logic Map: `strategyProjectionFromAssembly`

- Source: `internal/app/engine/strategy_runtime_projection.go`
- Current-base source SHA-256: `6f8241d38004892cdf51bb71f597b368545e3d18e362ecd036e666462de42fce`
- Signature: `strategyProjectionFromAssembly(params=1, results=1)`
- Source range: `81:1`–`138:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `137:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | range | 84:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 91:3 | planned targeted RED before any edit; not run by L0 |
| B3 | switch | 93:4 | planned targeted RED before any edit; not run by L0 |
| B4 | case | 94:4 | planned targeted RED before any edit; not run by L0 |
| B5 | case | 96:4 | planned targeted RED before any edit; not run by L0 |
| B6 | case | 98:4 | planned targeted RED before any edit; not run by L0 |
| B7 | if | 105:3 | planned targeted RED before any edit; not run by L0 |
| B8 | if | 114:3 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| assembly.Schedule.ObservedAt.UTC | 82:14 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyprojection.DormantSnapshot | 83:14 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyprojection.Market | 85:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| assembly.Schedule.For | 86:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| assembly.Candidate.For | 87:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| assembly.Proposal.For | 88:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| assembly.Risk.For | 89:11 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| assembly.Supervisor.Snapshot | 90:17 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyprojection.WithMarketFailure | 101:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| assembly.proposals.forMarket | 104:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 105:6 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| ValidProposal | 105:38 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| authority.entries.authority.Proposal | 105:38 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyprojection.WithMarketFailure | 106:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| authority.entries.authority.Proposal | 110:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| projectionDigest | 113:58 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| projectionDigest | 115:21 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strconv.Itoa | 117:44 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 118:28 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| string | 119:59 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| projectionDigest | 120:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
