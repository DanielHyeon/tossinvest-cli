# Function Logic Map: `TestTheTwoSurfacesApplyTheSameThresholds`

- Source: `cmd/tossctl/vetothresholds_source_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| scan surface thresholds | fresh `candidate.VetoThresholds` value | `candidateCycleOptions` wiring | B1 fails if scan no longer uses the single source |
| shared constructor thresholds | all fields absent until human activation | a046 review dormant scope | B2 fails if any legacy or invented number becomes effective |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | scan value differs from constructor value | testing failure only | `t.Errorf` | this test, first assertion |
| B2 | any of seen_late, extended, near_high is non-empty | testing failure only | `t.Errorf` | this test, dormant assertion |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `candidateCycleOptions` | observe scan wiring | pure option assembly in this fixture; no I/O | CodeGraph caller + AST |
| `candidateVetoThresholds` | observe single threshold source | pure leaf; no errors/retry | CodeGraph definition + AST |
| `t.Errorf` | record invariant violations without masking the second check | test failure | AST |

## State mutations and fallbacks

- Local values only. No production mutation, transport, clock, account, or config binding.
- There is deliberately no fallback assertion to the legacy 2.0 value.

## Safety conclusion

- Safe edit boundary: update the expected runtime state from the legacy near-high value to the approved dormant all-absent state.
- High-risk impact: no; this is a static/wiring test and makes the candidate path more fail-closed.
