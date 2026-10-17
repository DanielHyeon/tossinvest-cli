# Function Logic Map: `seedVerifyRecord`

- Source: `cmd/tossctl/verify_test.go`; unchanged base fixture pulled into the changed test file.
- AST evidence: `ast.json` (`revision: base`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| fixture record path | isolated test path | caller | recorder open/append failure fails test |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | recorder open error | none | test fatal | existing CLI fixture tests |
| B2 | fixture entries loop | test record append | visits all fixture entries | existing CLI fixture tests |
| B3 | append error | none | test fatal | existing CLI fixture tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `verifylive.OpenRecorder` | isolated fixture record | fatal on error | AST B1 |
| `rec.Append` | fixture entries | fatal on error | AST B2/B3 |

## State mutations and fallbacks

- Writes only a temporary test record. No account, broker, configuration, or production state is accessed.

## Safety conclusion

- Base fixture unchanged; map required because its containing test file changed.
