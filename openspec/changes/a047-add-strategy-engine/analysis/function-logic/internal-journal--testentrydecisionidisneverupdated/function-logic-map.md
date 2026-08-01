# Function Logic Map: `TestEntryDecisionIDIsNeverUpdated`

- Source: `internal/journal/adoption_static_test.go`
- CodeGraph callers/callees: package test only
- AST: generated after implementation

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| production SQL text | exact token `entry_decision_id`, not longer provenance names such as `entry_decision_identity` | journal schema invariant | test failure on any UPDATE |
| insert writer count | at least one position INSERT | position journal | test failure if scan is broken |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | production INSERT names exact guarded token | local count | continue | positive control |
| B2 | UPDATE names exact guarded token | test error | fail | static invariant |
| B3 | no inserter found | test fatal | fail | positive control |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| production source scanner/update statement parser | scan SQL-bearing Go sources | parse/read error fails test | static suite |
| exact-token regexp | avoid false positives on strategy provenance column names | no substring fallback | regression test run |

## State mutations and fallbacks

- Test-local counter/errors only. It does not mutate production state and has no allowlist fallback.

## Safety conclusion

- Preserves the original position immutability guard while distinguishing the separate `entry_decision_identity` provenance column.
