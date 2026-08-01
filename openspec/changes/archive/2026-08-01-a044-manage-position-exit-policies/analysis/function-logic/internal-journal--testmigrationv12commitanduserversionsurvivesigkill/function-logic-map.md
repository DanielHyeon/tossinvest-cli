# Function Logic Map: `TestMigrationV12CommitAndUserVersionSurviveSIGKILL`

- Source: `internal/journal/migration_v12_crash_linux_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| crash child | production Open migrates v11 to v12 then receives SIGKILL before Close | real test subprocess | must leave committed v12 |
| raw reopen | read-only SQLite without migration | on-disk DB/WAL | observe user_version/artifacts directly |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | child branch verifies v12 then kills; parent seeds v11 | process boundary | fatal on mismatch | crash test |
| B4-B9 | raw user_version/artifacts and normal row recovery | read only then production reopen | fatal on mismatch | crash test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `runCrashChild`/`kill` | simulate power loss after migration commit | requires SIGKILL | AST |
| raw `sql.Open(mode=ro)` | avoid masking v11 by auto-migrating during assertion | no writes | AST |

## State mutations and fallbacks

- The test distinguishes a committed v12 from a reopen that merely finishes an incomplete migration.

## Safety conclusion

- Safe edit boundary: raw user_version/table checks must precede production reopen.
- High-risk impact: yes
