# Function Logic Map: `Console.session0`

- Source: `internal/console/console.go`
- AST evidence: `ast.json` (`da1cbb194372c0f20b926357a65085ebb20021744f530209782e971d0357c254`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | remote sessions are server-side IP/UA-bound and expiring; local session exchange is unchanged | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 728 | only the branch's documented state transition | existing return/error contract | `TestRemoteQuerySessionCredentialIsNotAccepted` |
| B2 | existing if branch at line 729 | only the branch's documented state transition | existing return/error contract | `TestRemoteQuerySessionCredentialIsNotAccepted` |
| B3 | existing if branch at line 733 | only the branch's documented state transition | existing return/error contract | `TestRemoteQuerySessionCredentialIsNotAccepted` |
| B4 | existing if branch at line 739 | only the branch's documented state transition | existing return/error contract | `TestRemoteQuerySessionCredentialIsNotAccepted` |
| B5 | existing if branch at line 745 | only the branch's documented state transition | existing return/error contract | `TestRemoteQuerySessionCredentialIsNotAccepted` |
| B6 | existing if branch at line 753 | only the branch's documented state transition | existing return/error contract | `TestRemoteQuerySessionCredentialIsNotAccepted` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| remote.hasSession, acceptHandoff, grantSession, hasSessionCookie | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- remote query-session credentials are never accepted and failures redirect only to /login.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: remote query-session credentials are never accepted and failures redirect only to /login.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
