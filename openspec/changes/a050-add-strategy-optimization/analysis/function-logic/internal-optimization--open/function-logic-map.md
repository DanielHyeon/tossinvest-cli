# Function Logic Map: `Open`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| path, actor, parent directory, SQLite DB | Nonempty path and actor; private directory; supported schema | `Options`, filesystem, `PRAGMA user_version` | Fail closed before returning a usable store |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | path or actor is blank | none | descriptive error | `TestOpenRejectsInvalidOptions` |
| B3 | directory creation/permission fails | none | wrapped error | `TestOpenSecuresParentAndSQLiteFiles` |
| B4-B8 | DB open/ping/permission/migration fails | close DB if opened | wrapped error | `TestOpenRefusesNewerSchema` |
| B9-B10 | default dependencies or init | creates initial immutable snapshot | error or initialized Store | lifecycle tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `os.MkdirAll`, `os.Chmod`, `sql.Open/Ping` | private durable control DB | errors close resources; no retry beyond SQLite busy timeout | AST B3-B8 |
| `Store.init` | atomic schema migration/initialization | failure closes DB and prevents use | CodeGraph impact + AST B10 |

## State mutations and fallbacks

- Creates only the private control DB and initial snapshot. It has no broker, order, lane, gate, journal, or LIVE mutation.
- Security boundary: parent is 0700 and database sidecars are 0600 before the Store is returned.

## Safety conclusion

- Safe edit boundary: optimization-private SQLite storage only.
- High-risk impact: no; applying settings must not create trading authority.
