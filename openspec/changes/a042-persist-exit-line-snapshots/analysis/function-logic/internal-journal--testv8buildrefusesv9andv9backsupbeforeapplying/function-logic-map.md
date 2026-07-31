# Function Logic Map: `TestV8BuildRefusesV9AndV9BacksUpBeforeApplying`

- Source: `internal/journal/migration_v9_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| v8 fixture and v8 reader plan | valid historical journal | migration helpers | fatal assertion on mismatch |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B5 | migrate to v9, inspect backup, prove v8 reader refusal | temp files only | fatal assertion | migration_v9 test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Open/backupsIn | verify forward-only schema and backup | bounded local filesystem operations | AST |

## State mutations and fallbacks

- Test-only branch changed to stop at v9 instead of current v10.

## Safety conclusion

- Safe edit boundary: historical migration test only.
- High-risk impact: no production mutation path change.
