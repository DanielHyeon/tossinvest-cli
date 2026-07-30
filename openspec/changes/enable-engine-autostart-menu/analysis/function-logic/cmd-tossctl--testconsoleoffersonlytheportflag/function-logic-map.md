# Function Logic Map: `TestConsoleOffersOnlyThePortFlag`

- Source: `cmd/tossctl/console_test.go`
- AST evidence: `ast.json` (`85d9f3fcee393ead51e81a16bc898ad15bca9164cbd695f0c65e245ef5d3a22c`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | validated caller inputs | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 71 | only the branch's documented state transition | existing return/error contract | `TestTestConsoleOffersOnlyThePortFlag` |
| B2 | existing if branch at line 74 | only the branch's documented state transition | existing return/error contract | `TestTestConsoleOffersOnlyThePortFlag` |
| B3 | existing range branch at line 91 | only the branch's documented state transition | existing return/error contract | `TestTestConsoleOffersOnlyThePortFlag` |
| B4 | existing if branch at line 92 | only the branch's documented state transition | existing return/error contract | `TestTestConsoleOffersOnlyThePortFlag` |
| B5 | existing if branch at line 95 | only the branch's documented state transition | existing return/error contract | `TestTestConsoleOffersOnlyThePortFlag` |
| B6 | existing if branch at line 104 | only the branch's documented state transition | existing return/error contract | `TestTestConsoleOffersOnlyThePortFlag` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST-listed callees | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- preserve existing fail-closed behavior.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: preserve existing fail-closed behavior.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
