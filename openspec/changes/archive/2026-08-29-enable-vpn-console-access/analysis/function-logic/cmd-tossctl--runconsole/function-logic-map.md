# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (`15885682dd9ba6f9e67f8e2e2c9e81428db488c52603f04d55cbeedba484b1d0`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | remote flags are resolved before account/config wiring; local mode remains the zero value | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 195 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B2 | existing if branch at line 204 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B3 | existing if branch at line 209 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B4 | existing if branch at line 213 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B5 | existing if branch at line 217 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B6 | existing if branch at line 221 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B7 | existing if branch at line 226 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B8 | existing if branch at line 239 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B9 | existing else branch at line 242 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B10 | existing if branch at line 250 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B11 | existing else branch at line 253 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B12 | existing if branch at line 253 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B13 | existing else branch at line 255 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B14 | existing if branch at line 257 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B15 | existing else branch at line 265 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B16 | existing if branch at line 259 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B17 | existing else branch at line 261 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B18 | existing if branch at line 268 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B19 | existing if branch at line 272 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B20 | existing else branch at line 274 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B21 | existing if branch at line 281 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B22 | existing if branch at line 284 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B23 | existing if branch at line 300 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B24 | existing if branch at line 307 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| remoteAccessOptions, console.ListenAndServe, config-backed seams, audit recorder | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- wire the all-or-nothing remote transport only; no verify/order approval or broker capability is added.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: wire the all-or-nothing remote transport only; no verify/order approval or broker capability is added.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
