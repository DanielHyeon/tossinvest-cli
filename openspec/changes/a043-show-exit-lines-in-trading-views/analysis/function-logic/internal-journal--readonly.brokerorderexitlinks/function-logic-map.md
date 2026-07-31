# Function Logic Map: `ReadOnly.BrokerOrderExitLinks`

- Source: `internal/journal/account_views.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| requested scopes | 1..256 opaque broker-id/account/canonical-market/market-day tuples | filtered visible rows | zero-length id skips no bytes; whitespace-only is valid; invalid/oversized returns typed error |
| origin | a `CONFIRMED` PLACE on the visible order or its validated single AMEND ancestry | mutation lifecycle + lineage | CANCEL/AMEND-only, RECORDED, and FAILED never imply engine origin |
| lineage | exactly one `CONFIRMED` AMEND edge per recursive node and matching scoped intent | journal identities + JSON-array exact visited set | cycle, branch, depth >32 or scope mismatch returns no event |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B7 | empty/bounded/validated scope and JSON setup | local allocation only; broker id copied byte-exact | nil or typed validation error | empty/bounds/opaque-id tests |
| B8-B14 | one bounded recursive SQL query, result setup and exact composite scan | read-only query | read error, row bound or fail-closed identity | direct/collision/amend tests |
| B15-B23 | cycle/branch/scope/depth/cardinality/hydration classification | aggregate local rows only | typed unknown reason | pure cycle, depth, branch, cross-account, corruption tests |
| B24-B28 | terminal SQL error and one-result-per-scope projection | none | exact link or typed fail-closed marker | state/duplicate/unlinked tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| SQLite recursive CTE | recurse only when the correlated valid-parent count is exactly one; visited IDs are a JSON array compared by exact value | scope/depth/width bounded during recursion; no retry | current AST + 1,000-branch, opaque-path, pure-cycle and depth tests |

## State mutations and fallbacks

- Exact SQL lineage joins opaque scoped broker order → CONFIRMED attempt/intent → sole exit event; JSON encoding never changes identity and delimiter substrings cannot create or hide cycles.
- Exactly 32 ancestry edges remain readable when terminal; a node with further ancestry at the bound fails closed without expanding it.
- No live order call, operating-toggle write, or policy recomputation is introduced by this function change.

## Safety conclusion

- Safe edit boundary: read-model/view/test contract for a043 only.
- High-risk impact: no; account mutation and execution paths are unreachable.
