# Function Logic Map: `TestConsolePositionPolicyClientCannotReachJournalOrEngineConstructors`

- Source: `internal/app/engine/deps_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| RPC package graph | production imports only | Go source parser | test failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | forbidden dependency reachable | none | test failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `transitiveInternalDeps` | derive import reachability | deterministic/no network | AST |

## State mutations and fallbacks

- Static assertion; it grants no journal capability.

## Safety conclusion

- Safe edit boundary: keep the client dependency closure free of journal, engine, config, and official constructors.
- High-risk impact: no
