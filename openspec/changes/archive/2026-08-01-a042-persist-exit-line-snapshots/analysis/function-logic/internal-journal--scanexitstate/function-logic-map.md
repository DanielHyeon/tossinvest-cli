# Function Logic Map: `scanExitState`

- Source: `internal/journal/apply_hook.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| one exit-state row | base fields plus nullable v10 tuple/status/effective JSON | schema v10 | SQL errors typed; semantic corruption returned separately |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B4 | no row, scan error, nullable rung, completion/semantic classification | populate exact stored values only | typed result | legacy/partial/full tuple tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `scanExitStateResult` / snapshot validator | distinguish all-NULL legacy, seed, full, partial-corrupt | no registry lookup or default | CodeGraph + AST |

## State mutations and fallbacks

- Removed LADDER NULL→`default_v1` inference; unknown remains unknown.

## Safety conclusion

- Safe edit boundary: typed scan/classification only.
- High-risk impact: yes.
