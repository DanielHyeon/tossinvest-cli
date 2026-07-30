# Function Logic Map: `loadRemoteAccessToken`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (`6fd28a2cbf68752b65fc15fb3691ca8aff1d4bc1301a73aca5ef2f46f1024104`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | regular non-symlink file, owner-only permissions, at most 4096 bytes, trimmed token at least 32 bytes | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 410 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B2 | existing if branch at line 413 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B3 | existing if branch at line 416 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B4 | existing if branch at line 419 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B5 | existing if branch at line 423 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |
| B6 | existing if branch at line 427 | only the branch's documented state transition | existing return/error contract | `TestRemoteAccessTokenFileMustBePrivateAndLong` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| os.Lstat, os.ReadFile | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- reject unsafe metadata or size before any listener/account construction.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: reject unsafe metadata or size before any listener/account construction.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
