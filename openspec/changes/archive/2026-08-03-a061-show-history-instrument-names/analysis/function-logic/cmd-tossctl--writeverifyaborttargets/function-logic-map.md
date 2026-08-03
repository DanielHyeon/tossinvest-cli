# Function Logic Map: `writeVerifyAbortTargets`

- Source: `cmd/tossctl/verify.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| writer | command output | Cobra | write failures follow existing best-effort CLI output behavior |
| record path/targets | local evidence projection and outstanding tool-owned artifacts | `loadVerifyRecord`, `AbortTargets` | empty set is stated explicitly |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | target set is empty | print honest empty state | return | abort empty-record tests |
| B2 | iterate targets | print kind/id/symbol only | continue | abort list tests |
| B3 | target has `HeldUntil` | add awaited-verdict explanation | continue | held-chain tests |
| tail | all targets printed | none | return | list/refreshed-target tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `fmt.Fprintf/Fprintln` | render the exact local target projection | no retry | AST B1-B3 |

## State mutations and fallbacks

- Output-only helper; it does not acquire authority, read credentials, or mutate evidence/account state.

## Safety conclusion

- Safe edit boundary: factor duplicated target rendering while preserving exact artifact fields.
- High-risk impact: no.
