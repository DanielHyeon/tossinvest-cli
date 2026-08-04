# Function Logic Map: `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`

- Source: `internal/performance/store_test.go`
- Qualified function: `TestStoreSchemaIsSeparateAppendOnlyAndVersioned`
- AST evidence: `ast.json` (current revision)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and fixture/runtime state | exact types and identities declared by the source revision | `internal/performance/store_test.go` plus its direct callers/tests | return/error or test assertion; no value is silently repaired |

## Branches and early returns

| Branches | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| AST `B*` rows | exact condition/loop at the recorded source line | only the source-declared effect; test functions mutate fixtures only | source return or assertion failure | `branch-test-map.md` and affected-package regression |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| direct AST calls | execute the source contract or its assertions | errors remain explicit; no undocumented retry/fallback | `ast.json`, source review and named regression |

## State mutations and fallbacks

- The source revision was reviewed at the function boundary recorded in `ast.json`.
- Test helpers mutate only isolated fixtures; production functions retain only their source-declared durable or pure effects.
- No broker, activation or operating-setting fallback is inferred from this evidence.

## Safety conclusion

- Safe edit boundary: preserve every AST branch and its explicit error/assertion behavior.
- High-risk impact: determined by the production caller; high-risk production functions receive a specific companion refinement before completion.
