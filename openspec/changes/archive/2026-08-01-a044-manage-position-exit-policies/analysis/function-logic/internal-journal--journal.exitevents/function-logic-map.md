# Function Logic Map: `Journal.ExitEvents`

- Source: `internal/journal/exit_state.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| position id | trimmed, non-empty | caller | query returns no matching rows |\n| stored events | append-only generation-bound rows | journal | corruption/query error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | query fails | none | wrapped error | exit event suite |\n| B2-B3 | scan each row | append in memory only | scan error | lifecycle event test |\n| B4 | iterator fails | none | error | exit event suite |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `scanExitEvent` | hydrate and validate event generation/evidence | fail closed | AST |

## State mutations and fallbacks

- Read-only history query; it never rewrites lifecycle attribution.

## Safety conclusion

- Safe edit boundary: append generation to the select/scan contract while preserving event order.
- High-risk impact: yes
