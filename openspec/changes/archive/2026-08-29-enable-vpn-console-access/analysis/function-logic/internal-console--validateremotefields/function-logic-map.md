# Function Logic Map: `validateRemoteFields`

- Source: `internal/console/remote.go`
- AST evidence: `ast.json` (`4b20d9985799183daff6f70301c013b68fa6bb95351bbf09724d032df1baf365`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | complete remote transport fields, audit recorder and exactly one access mode | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 136 | only the branch's documented state transition | existing return/error contract | `TestTrustedNetworkAndTokenAuthenticationCannotBeCombined` |
| B2 | existing if branch at line 142 | only the branch's documented state transition | existing return/error contract | `TestTrustedNetworkAndTokenAuthenticationCannotBeCombined` |
| B3 | existing if branch at line 145 | only the branch's documented state transition | existing return/error contract | `TestTrustedNetworkAndTokenAuthenticationCannotBeCombined` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| strings.TrimSpace | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- fail before listener construction on missing or conflicting access mode.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: fail before listener construction on missing or conflicting access mode.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
