# Function Logic Map: `Console.grantSession`

- Source: `internal/console/restart.go`
- AST evidence: `ast.json` (`930b9598b7346c4e66d435338f3132f422e37dd7d8db9bc8132e4b1ed3b603b0`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | remote handoff has valid peer and durable audit; local cookie exchange remains unchanged | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 234 | only the branch's documented state transition | existing return/error contract | `TestRemoteHandoffIssuesANewAuditedRemoteSession` |
| B2 | existing if branch at line 236 | only the branch's documented state transition | existing return/error contract | `TestRemoteHandoffIssuesANewAuditedRemoteSession` |
| B3 | existing if branch at line 241 | only the branch's documented state transition | existing return/error contract | `TestRemoteHandoffIssuesANewAuditedRemoteSession` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| remote.record, remote.issueSession, http.SetCookie, http.Redirect | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- audit failure prevents remote session issuance; remote cookie is distinct and Secure.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: audit failure prevents remote session issuance; remote cookie is distinct and Secure.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
