# Function Logic Map: `holdVerifyRateBudgetIntent`

- Source: `cmd/tossctl/verify.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| context | supervised CLI/console operation | caller | cancellation stops marker refresher |
| active profile | default data dir or explicit config dir | `verifyRateBudgetPath(root)` | path error returns before marker creation |
| execution exclusion | already held by every caller | entrypoint invariant | guarantees exactly one remove-on-release marker owner |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | profile budget path cannot resolve | none | return error | profile path tests |
| B2 | required intent marker cannot be created/refreshed | no lease or broker construction | wrapped error | unpublishable-intent regression test |
| tail | marker publishes | create/refresh required marker beside profile record | release closure | A061 contention tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `verifyRateBudgetPath` | bind intent to the active profile | pure | AST B1 |
| `runlock.Hold` | make verification priority visible to optional metadata/soak | publication failure is fatal for A061 admission; release always callable after success | contention and failure tests |

## State mutations and fallbacks

- The marker is created before the kernel budget wait and released after the operation.
- The caller-held execution flock prevents concurrent marker owners from deleting each other's file.

## Safety conclusion

- Safe edit boundary: strict profile-scoped admission signal only; no broker, record, or account mutation.
- High-risk impact: yes, because ordering controls live verification priority.
