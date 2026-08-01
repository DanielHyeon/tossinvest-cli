# Function Logic Map: `openPrivateDescriptor`

- Source: `internal/positionpolicyrpc/client.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| descriptor path | exact endpoint filename in validated control directory | engine descriptor locator | fail closed |
| opener | O_NOFOLLOW production opener or deterministic test seam | platform implementation | opened inode must match before/current path |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B4 | filename, directory, lstat, or pre-open descriptor validation fails | none | error | insecure filesystem tests |
| B5-B6 | no-follow open or fstat fails | close if opened | error | descriptor contract |
| B7 | path/inode differs across lstat-open-fstat-lstat | close | replacement error | inode replacement test |
| B8 | opened inode mode/owner/link invalid | close | error | symlink/hardlink tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ValidateControlDirectory` | validate parent ownership/mode/traversal | any uncertainty fails | AST |
| `Lstat`/`Stat`/`SameFile` | detect symlink and path replacement | all identities must match | AST |
| `openDescriptorNoFollow` | prevent final-component symlink traversal | Unix O_NOFOLLOW; unsupported platforms fail ownership verification | AST |

## State mutations and fallbacks

- The returned file is the same exact private single-link inode observed before and after open.

## Safety conclusion

- Safe edit boundary: preserve pre/open/post identity checks and close on every validation failure.
- High-risk impact: yes
