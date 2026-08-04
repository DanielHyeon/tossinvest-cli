# Function Logic Map: `verifyRateBudgetPath`

- Source: `cmd/tossctl/verify.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| root profile | default data directory or explicit `--config-dir` | `engineJournalDir(root)` | resolution error returns no path |
| record override | arbitrary evidence path | verify command option | intentionally ignored for rate-budget isolation |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | active profile journal directory cannot be resolved | none | wrapped error | existing path failure tests |
| tail | directory resolves | none | sibling `openapi-rate-budget.lock` path | A061 override-isolation and lease tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `engineJournalDir` | reuse the engine/update/verify profile boundary | pure path resolution | AST B1 |
| `filepath.Join` | name one stable cross-process lock file | pure | tail |

## State mutations and fallbacks

- Pure path derivation; neither the filesystem nor profile configuration is mutated.
- A `--record` override cannot create a second rate-budget domain for the same active profile.

## Safety conclusion

- Safe edit boundary: derive one profile-scoped lease location only.
- High-risk impact: yes, because divergent paths would permit verification and metadata to overlap.
