# Function Logic Map: `consoleOpenAPISeam.Check`

- Source: `cmd/tossctl/console_openapi.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| effective credential source | environment pair or protected file | `official.LoadCredentials` | missing/rejected/unavailable classified without spawn |
| pending setup generation | absent or owner-only marker | onboarding pending file | present marker fails closed as `saved_not_started` |
| validation token cache | normal cache for file mode, isolated cache for environment mode | credential source | environment replacement cannot reuse an old token |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | credential source is file-managed | inspect pending marker | continue or fail closed | pending-generation recovery test |
| B2 | marker inspection fails | none | unavailable | marker failure test |
| B3 | file-generation marker exists | none | `saved_not_started` | pending-generation recovery test |
| B4 | credential load fails | none | unavailable | preflight classification test |
| B5 | effective credentials are missing | none | missing | preflight classification test |
| B6 | credential source is environment-managed | allocate isolated token path | continue or unavailable | environment replacement test |
| B7 | isolated token path allocation fails | none | unavailable | injected temp failure test |
| B8 | isolated cleanup is present | remove temporary cache | continue or unavailable | environment replacement test |
| B9 | isolated cleanup fails | no spawn | unavailable | injected cleanup failure test |
| B10 | official validation fails to execute | no spawn | unavailable | transient preflight test |
| B11 | environment validation is ready | invalidate normal cache | continue or unavailable | environment replacement test |
| B12 | normal-cache invalidation fails | no spawn | unavailable | injected remove failure test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `official.LoadCredentials` | use the exact same source precedence as soak child | no retry here; caller remains stopped | CodeGraph + AST |
| `deps.temp` | isolate environment-generation validation from an old normal token | owner-only temporary path, cleanup required | CodeGraph + AST |
| `deps.validate` | validate/refresh with the source-appropriate token cache | fixed secret-free result | CodeGraph + AST |
| `deps.remove` | prevent the child from reusing an old environment-generation token | any error fails closed | CodeGraph + AST |
| pending-marker reader | prevent incomplete saved generation from restarting | any read error fails closed | AST + focused test |

## State mutations and fallbacks

- File mode may refresh the normal access-token cache.
- Environment mode validates through an isolated cache, cleans it, then removes
  the normal cache before restart so the child must mint a token for the checked
  pair.
- Environment-managed credentials ignore but never mutate a dormant
  file-generation marker because the child uses the environment pair.

## Safety conclusion

- Safe edit boundary: add the persisted pending-generation gate before file credentials are loaded.
- High-risk impact: yes — this function decides whether a background API process may start.
