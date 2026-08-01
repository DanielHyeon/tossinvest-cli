# Function Logic Map: `AssessApprovedCandidate`

- Source: `internal/candidate/thresholdset.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| threshold set | sealed and valid | approved config bundle | `ApprovalInvalidSet` |
| candidate life | canonical key, ACTIVE, no old `CooledAt`, ordered timestamps | candidate store snapshot | `ApprovalInvalidCandidateLife` |
| evaluation instant | `LastSeenAt <= At < LastSeenAt+DefaultStalenessTTL` | injected `VetoInputs.At` | `ApprovalInvalidCandidateLife` |
| market scope and veto evidence | exact threshold market; every veto measured and clear | sealed threshold set + `AssessChase` | scope/veto refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | threshold set invalid | none | `ApprovalInvalidSet` | invalid-set table row |
| B2 | candidate identity invalid | none | `ApprovalInvalidCandidateLife` | invalid-life rows |
| B3 | state/timestamps/current-life invalid | none | `ApprovalInvalidCandidateLife` | cooled, expired, stale, exact-expiry, re-cross rows |
| B4 | market differs from threshold scope | none | `ApprovalScopeMismatch` | wrong-market row |
| B5 | chase veto raised/unmeasured | local threshold copy only | typed veto refusal | dangerous/unmeasured rows |
| B6 | all checks pass | none outside returned value | immutable `ApprovedCandidate` with life proof | provenance pass test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `candidateLifeID` | bind approval to key + first sighting | error is converted to typed invalid-life refusal | CodeGraph + AST |
| `set.VetoThresholds` / `AssessChase` | evaluate sealed thresholds | no retry/timeout; pure deterministic call | CodeGraph + AST |

## State mutations and fallbacks

- No external mutation or fallback. The function copies the threshold values into local input and returns a new opaque value.

## Safety conclusion

- Safe edit boundary: fail closed before `AssessChase`; never mint approval from stale/re-crossed state.
- High-risk impact: yes — approval is a strategy activation boundary, covered by exact expiry/current-life tests.
