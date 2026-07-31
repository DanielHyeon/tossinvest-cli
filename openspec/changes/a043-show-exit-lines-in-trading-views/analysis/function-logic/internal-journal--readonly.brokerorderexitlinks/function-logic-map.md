# Function Logic Map: `ReadOnly.BrokerOrderExitLinks`

- Source: `internal/journal/account_views.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| requested scopes | 1..256 broker-id/account/canonical-market/market-day tuples | filtered visible rows | empty skips SQL; invalid/oversized returns typed error |
| origin | a `CONFIRMED` PLACE on the visible order or its validated single AMEND ancestry | mutation lifecycle + lineage | CANCEL/AMEND-only, RECORDED, and FAILED never imply engine origin |
| lineage | exactly one `CONFIRMED` AMEND edge per recursive node and matching scoped intent | journal identities | cycle, branch, depth or scope mismatch returns no event |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B6 | empty/bounded/validated scope setup | local allocation only | nil or typed validation error | empty/bounds tests |
| B7-B13 | one bounded recursive SQL query and composite identity scan | read-only query | read error or fail-closed identity | direct/collision/amend tests |
| B14-B20 | lineage and event integrity classification | aggregate local rows only | typed unknown reason | cycle/branch/cross-account/corruption tests |
| B21-B28 | place/event candidate cardinality and one-result-per-scope projection | none | exact link or typed fail-closed marker | state/duplicate/unlinked tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| SQLite recursive CTE | recurse only when the correlated valid-parent count is exactly one | scope/depth/width bounded during recursion; no retry | current AST + 1,000-branch test |

## State mutations and fallbacks

- Exact SQL lineage joins scoped broker order → CONFIRMED attempt/intent → sole exit event; no all-event materialization, N+1 lookup, or branch expansion remains.
- No live order call, operating-toggle write, or policy recomputation is introduced by this function change.

## Safety conclusion

- Safe edit boundary: read-model/view/test contract for a043 only.
- High-risk impact: no; account mutation and execution paths are unreachable.
