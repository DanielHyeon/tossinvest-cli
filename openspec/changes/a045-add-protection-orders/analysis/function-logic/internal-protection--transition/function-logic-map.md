# Function Logic Map: `Transition`

- Source: `internal/protection/domain.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function inputs | A persisted-valid saga and one typed event with monotone timestamp. | Current HEAD + OpenSpec | Fail closed with typed error/decision |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1+ | B1 invalid input/time; B2 each state/event guard; B3 weaker/invalid replace; B4 unknown mutation → IN_DOUBT; B5 output currently returned without revalidation. | Copies input then changes state-specific fields. Existing flaw: stale fields can make an invalid output. | Typed refusal or validated result | See branch map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Current callees | `Saga.Validate`, `invalidTransition`; pure, no broker or repository side effect. | No implicit retry; errors propagate fail-closed | CodeGraph + AST |

## State mutations and fallbacks

- Copies input then changes state-specific fields. Existing flaw: stale fields can make an invalid output.

## Safety conclusion

- Safe edit boundary: Clear/set state-owned fields deliberately and validate the output before return.
- High-risk impact: yes; dormant logic only, no broker mutation or WIRED binding.
