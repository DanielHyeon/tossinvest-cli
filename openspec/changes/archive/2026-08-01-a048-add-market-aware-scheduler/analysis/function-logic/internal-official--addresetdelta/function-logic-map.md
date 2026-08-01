# Function Logic Map: `addResetDelta`

- Source: `internal/official/ratebudget.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| observedAt | non-zero UTC-convertible response instant | parser caller | caller rejects zero before this function |
| seconds | nonnegative whole seconds below official epoch threshold and within duration range | parsed reset raw | invalid range returns false/zero |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | seconds negative or larger than `time.Duration` seconds capacity | none | zero,false | overflow tests |
| B2 | UTC add/sub round trip differs from requested delta | none | zero,false | wrapping/extreme-time tests |
| B3 | conversion and round trip are exact | none | derived UTC,true | valid delta tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `time.Time.UTC/Add/Sub` | strip monotonic component, add delta, and prove exact round trip | pure; mismatch fails closed | AST + parser tests |

## State mutations and fallbacks

- Pure arithmetic helper; it never returns a partially trusted instant.

## Safety conclusion

- Safe edit boundary: official delta derivation.
- High-risk impact: yes, due to reset authority.
