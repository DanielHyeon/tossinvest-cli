# Function Logic Map: `TestEngineDependencyGraphExcludesWTSMutators`

- Source: `internal/app/engine/deps_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| engine dependency graph | production imports | Go source parser | test failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | forbidden WTS mutator reachable | none | test failure | this test |\n| B3-B4 | required official/gateway dependency missing | none | test failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `transitiveInternalDeps` | derive engine graph | deterministic/no network | AST |

## State mutations and fallbacks

- Existing static safety assertion remains unchanged.

## Safety conclusion

- Safe edit boundary: preserve official execution authority and reject WTS mutation reachability.
- High-risk impact: yes
