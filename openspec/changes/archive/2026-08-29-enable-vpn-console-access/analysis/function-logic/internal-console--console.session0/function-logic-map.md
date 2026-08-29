# Function Logic Map: `Console.session0`

- Source: `internal/console/console.go`
- AST evidence: `ast.json` (`85d2bb460f96627d062ed9cfbccfd64ca13ad3de1dee21d0af3d3d70e8e70178`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | trusted-network is explicit and already passed remote security; authenticated remote/local session exchange remains unchanged | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 751 | only the branch's documented state transition | existing return/error contract | `TestTrustedNetworkStillRejectsWrongPeerOriginAndCSRF` |
| B2 | existing if branch at line 752 | only the branch's documented state transition | existing return/error contract | `TestTrustedNetworkStillRejectsWrongPeerOriginAndCSRF` |
| B3 | existing if branch at line 756 | only the branch's documented state transition | existing return/error contract | `TestTrustedNetworkStillRejectsWrongPeerOriginAndCSRF` |
| B4 | existing if branch at line 760 | only the branch's documented state transition | existing return/error contract | `TestTrustedNetworkStillRejectsWrongPeerOriginAndCSRF` |
| B5 | existing if branch at line 766 | only the branch's documented state transition | existing return/error contract | `TestTrustedNetworkStillRejectsWrongPeerOriginAndCSRF` |
| B6 | existing if branch at line 772 | only the branch's documented state transition | existing return/error contract | `TestTrustedNetworkStillRejectsWrongPeerOriginAndCSRF` |
| B7 | existing if branch at line 780 | only the branch's documented state transition | existing return/error contract | `TestTrustedNetworkStillRejectsWrongPeerOriginAndCSRF` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| remote.hasSession, acceptHandoff, grantSession, hasSessionCookie | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- bypass only the application session in trusted mode; never bypass peer, Host, Origin or CSRF gates.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: bypass only the application session in trusted mode; never bypass peer, Host, Origin or CSRF gates.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
