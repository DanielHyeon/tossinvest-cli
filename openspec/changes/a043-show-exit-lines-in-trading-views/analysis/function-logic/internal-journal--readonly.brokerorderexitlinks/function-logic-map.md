# Function Logic Map: `ReadOnly.BrokerOrderExitLinks`

- Source: `internal/journal/account_views.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| requested scopes | 1..256 exact broker-id/account/trading-day tuples | visible broker rows | empty skips SQL; invalid/oversized returns typed error |
| lineage | validated `replaces` edge with matching AMEND attempt+intent scope | journal FK identities | cycle, branch, depth or scope mismatch returns no event |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B6 | empty/bounded/validated scope setup | local allocation only | nil or typed validation error | empty/bounds tests |
| B7-B13 | one bounded recursive SQL query and composite identity scan | read-only query | read error or fail-closed identity | direct/collision/amend tests |
| B14-B20 | lineage and event integrity classification | aggregate local rows only | typed unknown reason | cycle/branch/cross-account/corruption tests |
| B21-B25 | final read checks and one-result-per-scope projection | none | exact link or no-event fail-closed marker | duplicate/unlinked tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| SQLite recursive CTE | walk only validated AMEND ancestors and hydrate event evidence in the same set query | bounded rows/depth; no retry | current AST and focused tests |

## State mutations and fallbacks

- Exact SQL lineage joins scoped broker order → validated attempt/intent → exit event; no all-event materialization or N+1 lookup remains.
- No live order call, operating-toggle write, or policy recomputation is introduced by this function change.

## Safety conclusion

- Safe edit boundary: read-model/view/test contract for a043 only.
- High-risk impact: no; account mutation and execution paths are unreachable.
