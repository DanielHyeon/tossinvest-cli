# Function Logic Map: `productionProtectionAssemblies`

- Source: `internal/app/engine/protection_wiring.go`
- Current-base source SHA-256: `2f664449ec103c181e275b6530aed0f768512ca9437ecfdb9f081bdae8359ca9`
- Signature: `productionProtectionAssemblies(params=1, results=1)`
- Source range: `39:1`–`44:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `40:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| — | — | — | no branch node |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| digestProtectionIdentity | 41:59 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| digestProtectionIdentity | 42:59 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
