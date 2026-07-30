# Function Logic Map: `TestConsoleOffersOnlyTheExplicitRemoteFlagSet`

- Source: `cmd/tossctl/console_test.go`
- AST evidence: `ast.json` (`34e8ca36af56210a38de32a4ac5864f1477259ec28ca581d164ba135bcf9d627`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | validated caller inputs | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 71 | only the branch's documented state transition | existing return/error contract | `TestTestConsoleOffersOnlyTheExplicitRemoteFlagSet` |
| B2 | existing if branch at line 74 | only the branch's documented state transition | existing return/error contract | `TestTestConsoleOffersOnlyTheExplicitRemoteFlagSet` |
| B3 | existing range branch at line 91 | only the branch's documented state transition | existing return/error contract | `TestTestConsoleOffersOnlyTheExplicitRemoteFlagSet` |
| B4 | existing if branch at line 92 | only the branch's documented state transition | existing return/error contract | `TestTestConsoleOffersOnlyTheExplicitRemoteFlagSet` |
| B5 | existing if branch at line 95 | only the branch's documented state transition | existing return/error contract | `TestTestConsoleOffersOnlyTheExplicitRemoteFlagSet` |
| B6 | existing if branch at line 105 | only the branch's documented state transition | existing return/error contract | `TestTestConsoleOffersOnlyTheExplicitRemoteFlagSet` |

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
