# Function Logic Map: `validateLadderRecoveryOutput`

- Source: `internal/exitpolicy/recovery.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| ladder snapshot + exact rung table | ratchet level NONE; active rung in `[-1,len)`; ladder-only action; numeric order level equals rung | StockOS-derived ladder state machine | `ErrRecoveryIdentity` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | ratchet level or rung bounds invalid | none | refuse | rung-bound table |
| B2 | action belongs to another policy | none | refuse | foreign-action case |
| B3 | orderable level is nonnumeric or differs from active rung | none | refuse | wrong-level case |
| B4 | valid pre-first-rung stop uses level `-1` | none | accept | pre-first-rung recovery tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strconv.Atoi` | exact decimal rung identity | conversion failure refuses | CodeGraph + AST |

## State mutations and fallbacks

- Pure validation; `NoRung=-1` is a valid ladder state, values below it are not.

## Safety conclusion

- Safe edit boundary: ladder-specific recovered output.
- High-risk impact: yes; rung bounds define protective lines and proposal identity.
