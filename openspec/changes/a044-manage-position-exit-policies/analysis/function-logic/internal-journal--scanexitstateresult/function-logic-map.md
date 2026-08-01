# Function Logic Map: `scanExitStateResult`

- Source: `internal/journal/exit_snapshot.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| SQL row | exact `exitStateSelect` order | journal schema | scan error |
| nullable v10 evidence | all/none tuple semantics | v10 snapshot contract | partial tuple marked corrupt |
| lifecycle generation | nullable legacy=1, explicit v12 | v12 lifecycle binding | invalid/partial refused |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | no row/scan failure | none | typed/wrapped error | snapshot tests |
| B2 | nullable rung/generation/status | hydrate defaults | result | legacy tests |
| B3 | no v10 evidence | legacy compatibility identity | result | migration tests |
| B4 | partial/full v10 tuple | mark corruption or hydrate | result | corruption tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `row.Scan` | positional hydration | schema order exact | CodeGraph + AST |
| snapshot decoders | validate stored tuple | fail closed as corruption | CodeGraph + AST |

## State mutations and fallbacks

- Read-only hydration; adding lifecycle generation requires coordinated SELECT and Scan order.

## Safety conclusion

- Safe edit boundary: append nullable lifecycle column and default historical rows to generation 1.
- High-risk impact: yes — malformed hydration can cross lifecycle boundaries.
