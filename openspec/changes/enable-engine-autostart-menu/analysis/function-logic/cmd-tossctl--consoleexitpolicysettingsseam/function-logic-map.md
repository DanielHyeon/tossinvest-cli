# Function Logic Map: `consoleExitPolicySettingsSeam`

- Source: `cmd/tossctl/console.go`
- Qualified function: `consoleExitPolicySettingsSeam`
- AST evidence: `ast.json` (`ef133dc61d797dff9fadf273cb2ac7bd66c9f6c0c404fcdfe9a93195838e60a6`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| declared parameters and receiver state | types plus persisted policy/config constraints | `cmd/tossctl/console.go` signature, config schema, journal schema, immutable policy registry | validation errors propagate; unknown policy/state refuses instead of widening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` at `cmd/tossctl/console.go:360` — if s := newExitPolicySettingsSeam(root); s != nil { | only mutations visible in this branch and its callees | existing return/error contract | `TestExitPolicySeamSavesAuditsAndPreservesUnrelatedConfig`; console characterization tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newExitPolicySettingsSeam` | preserve the function's validation, persistence, routing, or evaluation contract | errors remain fail-closed; no retry or authority expansion is introduced here | CodeGraph + `ast.json` |

## State mutations and fallbacks

- AST records 1 assignment(s) and 2 return(s); branch rows bind every control-flow site to regression evidence.
- Missing/unknown policy data follows the documented legacy compatibility or explicit refusal path; it never changes LIVE, trading, or order capability.

## Safety conclusion

- Safe edit boundary: policy selection/snapshot/routing only; existing stop urgency, cancel-first ordering, session+CSRF checks, and journal atomicity remain binding.
- High-risk impact: no; current AST hash and affected-package tests are required.
