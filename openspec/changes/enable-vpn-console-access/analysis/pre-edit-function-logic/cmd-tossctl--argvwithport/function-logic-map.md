# Function Logic Map: `argvWithPort`

- Source: `cmd/tossctl/console.go`
- Qualified function: `argvWithPort`
- Phase: pre-edit
- AST evidence: `ast.json` (`ef133dc61d797dff9fadf273cb2ac7bd66c9f6c0c404fcdfe9a93195838e60a6`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| listener/request/options/receiver state | current loopback, session, CSRF, restart contracts | `cmd/tossctl/console.go` and operator-console spec | refuse before handler or preserve existing lifecycle; never widen account authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | pre-edit `range` at `cmd/tossctl/console.go:760` — for i, a := range args { | existing branch side effects only | existing return/error contract | remote-boundary RED tests in tasks 2.1–2.4 |
| B2 | pre-edit `switch` at `cmd/tossctl/console.go:761` — switch { | existing branch side effects only | existing return/error contract | remote-boundary RED tests in tasks 2.1–2.4 |
| B3 | pre-edit `case` at `cmd/tossctl/console.go:762` — case skipNext: | existing branch side effects only | existing return/error contract | remote-boundary RED tests in tasks 2.1–2.4 |
| B4 | pre-edit `case` at `cmd/tossctl/console.go:764` — case i == 0: | existing branch side effects only | existing return/error contract | remote-boundary RED tests in tasks 2.1–2.4 |
| B5 | pre-edit `case` at `cmd/tossctl/console.go:766` — case a == consolePortFlag: | existing branch side effects only | existing return/error contract | remote-boundary RED tests in tasks 2.1–2.4 |
| B6 | pre-edit `case` at `cmd/tossctl/console.go:768` — case strings.HasPrefix(a, consolePortFlag+"="): | existing branch side effects only | existing return/error contract | remote-boundary RED tests in tasks 2.1–2.4 |
| B7 | pre-edit `case` at `cmd/tossctl/console.go:769` — default: | existing branch side effects only | existing return/error contract | remote-boundary RED tests in tasks 2.1–2.4 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `make`, `len`, `append`, `strings.HasPrefix`, `strconv.Itoa` | current listener/auth/restart lifecycle | errors remain fail-closed and listener shutdown ordering is preserved | CodeGraph + `ast.json` |

## State mutations and fallbacks

- Current local session/CSRF state and listener/relaunch mutations are the compatibility baseline.
- Remote mode must be additive and fully configured; partial configuration must not fall back to an exposed listener.

## Safety conclusion

- Safe edit boundary: listener/auth transport only; direct order capability, verify nonce, LIVE gate and engine interlock remain unchanged.
- High-risk impact: yes; RED/GREEN branch mapping and post-edit AST refresh are mandatory.
