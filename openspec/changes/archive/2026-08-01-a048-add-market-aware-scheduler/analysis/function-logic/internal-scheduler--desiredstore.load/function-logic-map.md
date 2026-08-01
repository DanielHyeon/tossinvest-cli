# Function Logic Map: `DesiredStore.Load`

- Source: `internal/scheduler/desired.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| context | active or canceled | caller | canceled returns without file read |
| file | absent defaults or strict revisioned JSON | desired-state file | decode/validation errors fail closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | context already canceled | none | context error | canceled-context test |
| success | active context | delegates strict read at one UTC instant | desired/default or error | load/round-trip tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `DesiredStore.loadAt` | strict duplicate/unknown/trailing/value validation with one clock instant | absent file returns OFF defaults | CodeGraph + AST |

## State mutations and fallbacks

- Read-only wrapper; no lock is required because Save installs by atomic rename.

## Safety conclusion

- Safe edit boundary: desired-state read only.
- High-risk impact: yes, because invalid persisted ON must fail rather than default permissively.
