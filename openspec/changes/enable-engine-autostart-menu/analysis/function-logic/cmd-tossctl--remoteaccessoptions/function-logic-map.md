# Function Logic Map: `remoteAccessOptions`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (`6fd28a2cbf68752b65fc15fb3691ca8aff1d4bc1301a73aca5ef2f46f1024104`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | zero flags mean local mode; any remote flag requires a private token file and later full validation | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 374 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B2 | existing if branch at line 380 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B3 | existing if branch at line 383 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B4 | existing if branch at line 387 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B5 | existing if branch at line 399 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| loadRemoteAccessToken, openAuditLog, RecordAction | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- keep credential bytes out of argv/banner/audit and fail closed when audit storage is unavailable.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: keep credential bytes out of argv/banner/audit and fail closed when audit storage is unavailable.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
