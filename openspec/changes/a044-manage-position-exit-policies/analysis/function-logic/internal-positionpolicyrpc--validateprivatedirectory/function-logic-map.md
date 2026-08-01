# Function Logic Map: `validatePrivateDirectory`

- Source: `internal/positionpolicyrpc/client.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| path | non-empty real directory with no symlink traversal | engine/control descriptor layout | error |
| exact flag | control leaf=true; engine parent=false | caller | exact 0700 or no group/other write |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B5 | empty, traversal, lstat, non-directory, or wrong owner | none | error | insecure filesystem tests |
| B6 | exact control leaf is not 0700 | none | error | control mode test |
| B7 | engine parent is group/other writable | none | error | parent mode test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `rejectSymlinkTraversal` | reject symlinks in any path component | EvalSymlinks mismatch/error fails | AST |
| `validateOwnerAndLinks` | require effective-user ownership | unsupported platforms fail closed | AST |

## State mutations and fallbacks

- Engine directory may be readable/executable by others for compatibility but never writable; dedicated control leaf is exactly 0700.

## Safety conclusion

- Safe edit boundary: never weaken owner, symlink, write-bit, or exact control-leaf checks.
- High-risk impact: yes
