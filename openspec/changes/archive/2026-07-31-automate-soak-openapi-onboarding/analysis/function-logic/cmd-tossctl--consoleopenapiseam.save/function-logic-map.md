# Function Logic Map: `consoleOpenAPISeam.Save`

- Source: `cmd/tossctl/console_openapi.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| submitted key/secret | nonblank after trim | HTTPS bounded form | reject before persistence |
| effective source | environment or file | `environmentManaged` | environment pair cannot be replaced in browser |
| setup generation | pending or complete | owner-only pending marker | only complete generation returns ready |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | environment-managed or blank input | none | fixed rejection | environment/input tests |
| B2 | isolated validation setup/probe/cleanup fails | temporary 0600 cache only | unavailable/rejected | isolated token tests |
| B3 | pending marker cannot be created | none | unavailable | marker helper test |
| B4 | credential persistence returns an error after being attempted | marker remains | saved-not-started | marker failure test |
| B5 | token invalidation, audit, or final marker clear fails | saved credentials + pending marker remain | saved-not-started | pending-generation recovery test |
| B6 | every step succeeds | credential save, token removal, audit, marker clear | ready | one-action save test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `deps.validate` | verify submitted pair using isolated cache | no old token may satisfy validation | CodeGraph + AST |
| `deps.save` | protected 0600 persistence | must precede save-success audit | CodeGraph + AST |
| `deps.remove` | invalidate normal access token | failure leaves pending marker | CodeGraph + AST |
| `deps.audit` | secret-free save event | failure leaves pending marker | CodeGraph + AST |
| marker helpers | persist incomplete generation across restarts | owner-only, fixed content, fail closed | AST + focused test |

## State mutations and fallbacks

- The pending marker is written before persistence and removed only after token invalidation and audit both succeed.
- No error path includes credential-derived text.

## Safety conclusion

- Safe edit boundary: wrap the existing save/invalidate/audit sequence with a durable generation marker.
- High-risk impact: yes — partial persistence must never be mistaken for a start-ready generation.
