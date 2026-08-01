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
| B1 | exact AST `range` at source line 155: `for name, text := range productionSources(t) {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B2 | exact AST `if` at source line 156: `if strings.Contains(text, "INSERT INTO positions") &&` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B3 | exact AST `range` at source line 160: `for _, stmt := range updateStatements(text) {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B4 | exact AST `if` at source line 161: `if exactEntryDecisionID.MatchString(stmt) {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B5 | exact AST `if` at source line 168: `if inserters == 0 {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| production source scanner/update statement parser | scan SQL-bearing Go sources | parse/read error fails test | static suite |
| exact-token regexp | avoid false positives on strategy provenance column names | no substring fallback | regression test run |

## State mutations and fallbacks

- Test-local counter/errors only. It does not mutate production state and has no allowlist fallback.

## Safety conclusion

- Preserves the original position immutability guard while distinguishing the separate `entry_decision_identity` provenance column.
