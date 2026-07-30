# Function Logic Map: `Console.acceptHandoff`

- Source: `internal/console/restart.go`
- Qualified function: `Console.acceptHandoff`
- Phase: pre-edit
- AST evidence: `ast.json` (`cdfd8e277e636952e4687f5b48d520a980f85d61524bce3a9e23baa4b3f236dd`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| listener/request/options/receiver state | current loopback, session, CSRF, restart contracts | `internal/console/restart.go` and operator-console spec | refuse before handler or preserve existing lifecycle; never widen account authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | pre-edit `if` at `internal/console/restart.go:213` — if token == "" { | existing branch side effects only | existing return/error contract | remote-boundary RED tests in tasks 2.1–2.4 |
| B2 | pre-edit `if` at `internal/console/restart.go:216` — if c.opts.Handoff == nil { | existing branch side effects only | existing return/error contract | remote-boundary RED tests in tasks 2.1–2.4 |
| B3 | pre-edit `if` at `internal/console/restart.go:219` — if err := c.opts.Handoff.Consume(token, c.now()); err != nil { | existing branch side effects only | existing return/error contract | remote-boundary RED tests in tasks 2.1–2.4 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Get`, `r.URL.Query`, `c.opts.Handoff.Consume`, `c.now`, `c.refuse`, `c.grantSession` | current listener/auth/restart lifecycle | errors remain fail-closed and listener shutdown ordering is preserved | CodeGraph + `ast.json` |

## State mutations and fallbacks

- Current local session/CSRF state and listener/relaunch mutations are the compatibility baseline.
- Remote mode must be additive and fully configured; partial configuration must not fall back to an exposed listener.

## Safety conclusion

- Safe edit boundary: listener/auth transport only; direct order capability, verify nonce, LIVE gate and engine interlock remain unchanged.
- High-risk impact: yes; RED/GREEN branch mapping and post-edit AST refresh are mandatory.
