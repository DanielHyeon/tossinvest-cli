# Function Logic Map: `writePositionPolicyDescriptor`

- Source: `internal/app/engine/position_policy_transport.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| descriptor/path | fixed endpoint filename under validated private control directory | server | error and remove only this staged/published inode |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B5 | marshal, control-dir validation, temp creation/open/stat | private temp only | wrapped error | writer and insecure filesystem tests |
| B6-B10 | body write, closed inode/control-dir revalidation, rename | fsynced 0600 staged inode then atomic rename | no publication on failure | writer tests |
| B11-B12 | post-publish failure cleanup | remove only if path still names staged inode | preserve replacement | inode contract |
| B13-B17 | published mode/owner/link/inode checks and directory fsync/close | durable endpoint publication | wrapped error with safe rollback | endpoint integration test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ValidateControlDirectory` | verify both parent and exact 0700 leaf twice | any uncertainty refuses publication | AST |
| `CreateTemp`/`ValidatePrivateOpenFile`/`SameFile` | bind publication to one owner-only 0600 inode | replacement/hardlink/symlink fails | AST |
| `Rename`/directory `Sync` | atomically and durably publish | post-publish error rolls back only matching inode | AST |

## State mutations and fallbacks

- Helper preserves chmod/write/short-write/sync/close errors. Temporary is removed on all failures; published rollback never removes an attacker replacement.

## Safety conclusion

- Safe edit boundary: preserve exact 0700/0600 owner/link/inode checks, atomic rename, directory fsync, and inode-scoped rollback.
- High-risk impact: yes
