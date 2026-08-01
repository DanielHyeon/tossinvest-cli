# Function Logic Map: `secureSQLiteFiles`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| database and WAL/SHM paths | existing SQLite files are 0600 | filesystem | non-not-exist error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | chmod database/sidecar error | none | return error | permission test |
| B2 | absent optional sidecar | none | continue | permission test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `os.Chmod` | restrict database artifacts | no retry | AST |

## State mutations and fallbacks

- Does not change DB content or any trading authority.

## Safety conclusion

- Safe edit boundary: file permissions.
- High-risk impact: no.
