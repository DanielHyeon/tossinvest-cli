# Function Logic Map: `TestLegacyLadderNullPolicyReadsAsDefaultV1`

- Source: `internal/journal/policy_snapshot_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| legacy NULL policy row | pre-v9 evidence with no policy identity | direct fixture update | must remain typed unknown, never registry-backfilled |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B5 | create row, null identity, read and assert explicit unknown | temp journal only | fatal assertion | policy snapshot test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| ExitState | typed v10 semantic read | SQL errors global; semantic absence typed | AST |

## State mutations and fallbacks

- Expected value changed from inferred default_v1 to explicit unknown per a042 manager decision.

## Safety conclusion

- Safe edit boundary: test expectation only.
- High-risk impact: yes semantically; prevents policy meaning fabrication and is covered by the legacy test.
