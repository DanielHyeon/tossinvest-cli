# Function Logic Map: `remoteAccessOptions`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (`15885682dd9ba6f9e67f8e2e2c9e81428db488c52603f04d55cbeedba484b1d0`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | zero flags mean native local mode; remote flags require exactly one of trusted-network or a private token file | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 425 | only the branch's documented state transition | existing return/error contract | `TestTrustedNetworkRemoteAccessDoesNotNeedATokenFile` |
| B2 | existing if branch at line 432 | only the branch's documented state transition | existing return/error contract | `TestTrustedNetworkRemoteAccessDoesNotNeedATokenFile` |
| B3 | existing if branch at line 435 | only the branch's documented state transition | existing return/error contract | `TestTrustedNetworkRemoteAccessDoesNotNeedATokenFile` |
| B4 | existing if branch at line 438 | only the branch's documented state transition | existing return/error contract | `TestTrustedNetworkRemoteAccessDoesNotNeedATokenFile` |
| B5 | existing if branch at line 442 | only the branch's documented state transition | existing return/error contract | `TestTrustedNetworkRemoteAccessDoesNotNeedATokenFile` |
| B6 | existing if branch at line 445 | only the branch's documented state transition | existing return/error contract | `TestTrustedNetworkRemoteAccessDoesNotNeedATokenFile` |
| B7 | existing if branch at line 459 | only the branch's documented state transition | existing return/error contract | `TestTrustedNetworkRemoteAccessDoesNotNeedATokenFile` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| loadRemoteAccessToken, openAuditLog, RecordAction | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- carry the explicit trusted-network decision without a credential and reject ambiguous or implicit access mode.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: carry the explicit trusted-network decision without a credential and reject ambiguous or implicit access mode.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
