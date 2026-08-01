# Function Logic Map: `scanExitEvent`

- Source: `internal/journal/exit_snapshot.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| SQL row | exact ExitEvents select order | journal schema | scan error |
| lifecycle generation | nullable legacy=1, explicit v12 | exit event row | invalid value refused |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | scan fails | none | error | snapshot tests |
| B2 | evidence absent/partial/full | hydrate legacy/corrupt/full | result/error | snapshot tests |
| B3 | lifecycle NULL/positive | default 1 or hydrate | result | migration/lifecycle tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `row.Scan` | hydrate scalar/evidence/generation | positional, no retry | CodeGraph + AST |
| `hydrateExitEventEvidence` | enforce tuple completeness | corruption returned | CodeGraph + AST |

## State mutations and fallbacks

- Read-only; lifecycle generation remains independent of position generation.

## Safety conclusion

- Safe edit boundary: append scan target and validate/default it.
- High-risk impact: yes — history attribution.
