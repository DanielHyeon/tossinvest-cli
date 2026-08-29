# Function Logic Map: `Console.URL`

- Source: `internal/console/console.go`
- AST evidence: `ast.json` (`85d2bb460f96627d062ed9cfbccfd64ca13ad3de1dee21d0af3d3d70e8e70178`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | trusted remote URL is the public root; authenticated remote URL is /login; native local retains the process session query | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 435 | only the branch's documented state transition | existing return/error contract | `TestTrustedNetworkConsoleNeedsNoApplicationSession` |
| B2 | existing if branch at line 436 | only the branch's documented state transition | existing return/error contract | `TestTrustedNetworkConsoleNeedsNoApplicationSession` |
| B3 | existing if branch at line 442 | only the branch's documented state transition | existing return/error contract | `TestTrustedNetworkConsoleNeedsNoApplicationSession` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Addr, strings.TrimSuffix, fmt.Sprintf | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- send trusted-network browsers directly to the console without weakening either compatibility authentication path.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: send trusted-network browsers directly to the console without weakening either compatibility authentication path.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
