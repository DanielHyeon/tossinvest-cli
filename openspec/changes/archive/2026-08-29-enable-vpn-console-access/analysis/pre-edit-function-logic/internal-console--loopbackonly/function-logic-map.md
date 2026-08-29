# Function Logic Map: `loopbackOnly`

- Source: `internal/console/console.go`
- Qualified function: `loopbackOnly`
- Phase: pre-edit
- AST evidence: `ast.json` (`aec87df0eb373e7771cf18420a98de1dc480c98d32d42db6ee6ccda4b230d378`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| listener/request/options/receiver state | current loopback, session, CSRF, restart contracts | `internal/console/console.go` and operator-console spec | refuse before handler or preserve existing lifecycle; never widen account authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | pre-edit `if` at `internal/console/console.go:492` — if ln == nil { | existing branch side effects only | existing return/error contract | remote-boundary RED tests in tasks 2.1–2.4 |
| B2 | pre-edit `if` at `internal/console/console.go:496` — if !ok { | existing branch side effects only | existing return/error contract | remote-boundary RED tests in tasks 2.1–2.4 |
| B3 | pre-edit `if` at `internal/console/console.go:499` — if tcp.IP == nil \|\| !tcp.IP.IsLoopback() { | existing branch side effects only | existing return/error contract | remote-boundary RED tests in tasks 2.1–2.4 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `fmt.Errorf`, `ln.Addr`, `tcp.IP.IsLoopback` | current listener/auth/restart lifecycle | errors remain fail-closed and listener shutdown ordering is preserved | CodeGraph + `ast.json` |

## State mutations and fallbacks

- Current local session/CSRF state and listener/relaunch mutations are the compatibility baseline.
- Remote mode must be additive and fully configured; partial configuration must not fall back to an exposed listener.

## Safety conclusion

- Safe edit boundary: listener/auth transport only; direct order capability, verify nonce, LIVE gate and engine interlock remain unchanged.
- High-risk impact: yes; RED/GREEN branch mapping and post-edit AST refresh are mandatory.
