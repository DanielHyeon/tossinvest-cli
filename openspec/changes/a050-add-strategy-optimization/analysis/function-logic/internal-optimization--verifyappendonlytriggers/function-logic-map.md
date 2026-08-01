# Function Logic Map: `verifyAppendOnlyTriggers`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| transaction/manifest | active migration transaction and fixed trigger manifest | store schema contract | missing/read-error/drift fails closed unless explicitly pre-install `allowMissing` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate every required trigger | read sqlite schema only | continue | append-only test |
| B2 | trigger absent and pre-install missing is allowed | none | continue | legacy migration test |
| B3 | trigger absent when required | none | missing-trigger error | drift test |
| B4 | schema query error | none | wrapped error | migration error coverage |
| B5 | canonical actual definition differs | none | unexpected-definition error | same-name no-op/drift test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| sqlite schema query | loads persisted SQL by exact trigger name | one query each/no retry | drift tests |
| `canonicalTriggerSQL` | ignores harmless whitespace/semicolon/IF-NOT-EXISTS variation only | pure | canonical definition tests |

## State mutations and fallbacks

- Read-only verification. `allowMissing` is used only before install during migration; post-install verification requires every exact definition.

## Safety conclusion

- Safe edit boundary: fail-closed trigger integrity verification.
- High-risk impact: yes; same-name no-op triggers must not bypass append-only controls.
