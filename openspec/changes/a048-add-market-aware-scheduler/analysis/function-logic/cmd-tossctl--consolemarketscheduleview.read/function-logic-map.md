# Function Logic Map: `consoleMarketScheduleView.Read`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| local status reader | non-nil production projection | runConsole assembly | reader error yields no partial console value |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | local reader returns error | none | empty console reading + error | safe-error page test |
| success | status available | field-by-field value copy | console reading | provenance/default tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `consoleMarketScheduleReader.Read` | isolate scheduler/official packages from console importer boundary | error propagated | AST + importer test |

## State mutations and fallbacks

- Pure adapter. It keeps `console.go` as the sole production importer of `internal/console`.

## Safety conclusion

- Safe edit boundary: read model conversion only.
- High-risk impact: low; no authority or mutation.
