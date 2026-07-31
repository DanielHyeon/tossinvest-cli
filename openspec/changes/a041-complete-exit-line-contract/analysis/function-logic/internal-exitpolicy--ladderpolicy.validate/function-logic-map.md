# Function Logic Map: `LadderPolicy.Validate`

- Source: `internal/exitpolicy/ladder.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| policy identity/table | non-empty id, semantic version, non-empty rungs, canonical digest | registry/config | `ErrRefused` or identity conflict |
| rung/runner values | targets increasing, stops non-decreasing, ratios [0,1], runner (0,100) | immutable policy | refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | missing id/rungs | none | refusal | existing validation tests |
| B2 | rung invalid/order descends | none | refusal | existing table tests |
| B3 | runner invalid | none | refusal | common policy tests |
| B4 | claimed digest differs from canonical values | none | identity conflict | a041 collision test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Rung.Validate`/decimal parsers | exact numeric/range checks | pure/refusal | CodeGraph + AST |
| `LadderPolicy.Identity` | canonical id/version/digest verification | pure/fail closed | CodeGraph + AST |

## State mutations and fallbacks

- Pure validator; no config write or operational toggle.

## Safety conclusion

- Safe edit boundary: add identity verification after every existing monotonicity/range check.
- High-risk impact: yes — policy loading fail-closed.
