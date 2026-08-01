# Function Logic Map: `AssessApprovedCandidate`

- Source: `internal/candidate/thresholdset.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `set` | a `LoadThresholdSet` result with `valid=true` | a046 design decisions 1, 2, 9; `ThresholdSet.valid` | zero `ApprovedCandidate` + typed `ApprovalError` (`invalid_set`) |
| `input.Candidate.Key` | set market must match exactly; symbol must be non-empty and canonical | `candidate.Key`, `Candidate` lifecycle contract | zero value + typed `scope_mismatch` or `invalid_candidate_life` error |
| `input.Candidate.FirstSeenAt` | non-zero; together with `Key` identifies exactly one candidate life | `candidate.go` D1 lifecycle comments and `Store.Promote` | zero value + typed `invalid_candidate_life` error |
| `input` measurements | each of `seen_late`, `extended`, `near_high` must be measured and clear | a046 spec “승인 후 pass”; `Chase.Passed` | zero value + typed `veto_raised` or `veto_unmeasured` error |
| threshold provenance | version, canonical set digest, evidence digest, approval instant copied from the validated set | a046 design decisions 1, 4, 9 | only exposed through value-returning accessors on a valid `ApprovedCandidate` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | threshold set is absent/unapproved/invalid | none | zero value + typed `invalid_set` | existing invalid-set test plus typed-error assertion |
| B2 | candidate market differs from the set scope | none | zero value + typed `scope_mismatch` | wrong-market assessment test |
| B3 | candidate symbol is empty/non-canonical or `FirstSeenAt` is zero | none | zero value + typed `invalid_candidate_life` | incomplete candidate-life test |
| B4 | `AssessChase` raises one or more vetoes | local value construction only | zero value + typed `veto_raised`, ordered veto codes | dangerous measurement test |
| B5 | no veto is raised but one or more are unmeasured | local value construction only | zero value + typed `veto_unmeasured`, ordered veto codes | unmeasured measurement test |
| B6 | defensive invariant: no raised/unmeasured code exists but `Passed()` is still false | local value construction only | zero value + typed `veto_unmeasured`; never mint on an inconsistent Chase | defensive branch is structurally unreachable under current `Chase`; map and source audit pin it |
| Success | all three vetoes are measured and clear | local value construction only | valid immutable `ApprovedCandidate` carrying candidate-life and threshold approval provenance | pass/provenance/deterministic-identity test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ThresholdSet.VetoThresholds` | copy the three validated decimal thresholds | invalid set returns absent thresholds, but B1 prevents reaching it | CodeGraph + current source |
| `AssessChase` | evaluate the three pure vetoes at injected `input.At` | no I/O, clock, timeout, retry, order, or mutation; each veto is three-state | CodeGraph callees + `veto.go` |
| `Chase.Passed/Raised/NotMeasured` | enforce the pass-only type boundary and produce deterministic refusal detail | ordered by private `OrderedVetoCodes()` copy; no fallback to clear | `veto.go` D3 invariants |
| candidate-life ID constructor | bind only `Key` and normalized `FirstSeenAt` into a domain-separated deterministic digest | rejects incomplete/non-canonical identity; no random/clock input | `candidate.go` D1 lifecycle + `Store.Promote` |

## State mutations and fallbacks

- The function mutates only its local copy of `input` to install thresholds.
- It performs no package/global mutation, I/O, clock read, retry, broker call, order construction, or `RiskIntent` construction.
- There is no partial success: every refusal returns the exact zero `ApprovedCandidate`.
- There is no identity fallback. Candidate life is exactly `(Market, Symbol, FirstSeenAt)`; `LastSeenAt`, source set, assessment time, and random data are excluded.
- There is no provenance fallback. Metadata comes only from the already-bound immutable `ThresholdSet`.

## Safety conclusion

- Safe edit boundary: add fail-closed validation and copied provenance within `internal/candidate`; preserve the package’s no-order/no-risk dependency guards.
- High-risk impact: safety-sensitive future entry boundary, but no current production caller and no live mutation authority. The edit must remain pure and fail closed.
