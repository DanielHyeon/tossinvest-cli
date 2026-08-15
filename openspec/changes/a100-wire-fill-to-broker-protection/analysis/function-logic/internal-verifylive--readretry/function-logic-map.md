# Function Logic Map: `readRetry`

- Source: `internal/verifylive/retry.go` (77-117)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| read closure, endpoint/request summary, retry budget | fixed extra attempts | caller and shared retry constants | final error returned after bounded retries |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | retry loop | logs every call | continue | retry attempt sequence |
| B2-B3 | success/non-rate error | return result/error | return | existing behavior |
| B4-B5 | 429/sleep failure | increments backoff, waits | return on sleep failure | M0 observer sees each attempt |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| closure, `sr.logCall`, `ReadRetryBackoff`, sleep | retain existing bounded read retry | only 429 retries | CodeGraph + AST |

## State mutations and fallbacks

- M0 must not let a later successful retry erase a critical-window failed attempt; observer receives every log boundary.

## Safety conclusion

- Safe edit boundary: observation hook only; retry timings unchanged.
- High-risk impact: yes — causal receipt depends on failures remaining visible.
