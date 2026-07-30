# Function Logic Map: `newConsoleCmd`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (`15885682dd9ba6f9e67f8e2e2c9e81428db488c52603f04d55cbeedba484b1d0`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | only the reviewed local port plus complete remote-mode file/path/network flags and one explicit access mode | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | straight-line path | documented state transition | documented return | `TestConsoleOffersOnlyTheCompleteRemoteAccessFlagSet` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Cobra flag registration and runConsole | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- accept trusted-network as a boolean decision but no token value, session, insecure transport, nonce, or automatic approval flag.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: accept trusted-network as a boolean decision but no token value, session, insecure transport, nonce, or automatic approval flag.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
