# Function Logic Map: `loadRiskBucketState`

- Source: `internal/journal/risk_bucket.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| queryer/key | live DB/transaction and canonical owner key | caller | delegated loader returns typed replay/state errors |
| lifecycle mode | active only for all existing callers | wrapper constant `false` | released rows remain unreadable through public APIs |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| none | active-state wrapper delegates to lifecycle-aware replay | read-only | exact delegated result | full journal replay suite |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `loadRiskBucketStateLifecycle(..., false)` | preserve public active-owner semantics while sharing exact digest replay with receipt validation | no fallback; returns underlying error | AST and full journal tests |

## State mutations and fallbacks

- Read-only; no mutation or fallback.
- `loadReleasedRiskBucketState` is a new package-private sibling used only to verify a sealed release receipt against current rows.

## Safety conclusion

- Safe edit boundary: preserve the existing function as an active-only wrapper and move its body unchanged behind an explicit lifecycle flag.
- High-risk impact: yes — release replay can now detect late fills or other post-release state drift without exposing released owners to normal readers.
