# Function Logic Map: `openConsolePerformanceCapabilities`

- Source: `cmd/tossctl/optimization.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| data directory and clock | resolved active profile directory and non-nil clock | production assembly | open error returns zero capabilities |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | performance DB open fails | no capability escapes | return error | failure test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `performance.OpenReadOnly`, `optimizationevidence.New` | open an existing derived DB without create/migrate/WAL mutation and reduce it to reads | one attempt, no retry | performance read-only + cmd focused tests |

## State mutations and fallbacks

- Opens only an existing derived DB with SQLite `mode=ro`/`query_only`; missing/old/new DB returns an error and stores no partial capability.
- Stores narrow read interfaces plus close closure; `Collect` and `Prune` never enter the console assembly.

## Safety conclusion

- Safe edit boundary: read capability construction; high-risk impact: no trading authority.
