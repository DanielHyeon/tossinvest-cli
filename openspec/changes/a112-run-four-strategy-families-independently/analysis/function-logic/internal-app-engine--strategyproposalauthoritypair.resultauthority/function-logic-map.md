# Function Logic Map: `ResultAuthority`

- Source: `internal/app/engine/strategy_proposal_authority.go`
- Current-base source SHA-256: `9e8816b71972be1678026da1c774934c85adc33872eed6e3a616abf9fa73dc2b`
- Signature: `strategyProposalAuthorityPair.ResultAuthority(params=0, results=1)`
- Source range: `93:1`–`101:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `96:4, 98:3, 100:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 95:3 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| len | 95:6 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| ValidProposal | 95:34 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| value.entries.authority.Proposal | 95:34 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| value.entries.authority.Proposal | 98:77 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| convert | 100:70 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| convert | 100:110 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
