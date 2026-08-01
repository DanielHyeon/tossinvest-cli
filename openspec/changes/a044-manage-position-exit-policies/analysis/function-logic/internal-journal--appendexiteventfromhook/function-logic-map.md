# Function Logic Map: `appendExitEventFromHook`

- Source: `internal/journal/exit_state.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| fill/completion event | exact position in ApplyTx | fill hook | rollback on lookup/insert error |
| lifecycle generation | current exit_states value, legacy default 1 | same transaction | no unbound event |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | generation-bound insert or row confirmation fails | transaction rollback | wrapped/typed error | fill lifecycle tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ApplyTx.Exec` | INSERT ... SELECT current exit generation | same fill transaction | AST |

## State mutations and fallbacks

- PROPOSAL_FILLED and COMPLETED inherit the generation of the exit row being completed.

## Safety conclusion

- Safe edit boundary: bind hook events to lifecycle generation without exposing raw sql.Tx.
- High-risk impact: yes
