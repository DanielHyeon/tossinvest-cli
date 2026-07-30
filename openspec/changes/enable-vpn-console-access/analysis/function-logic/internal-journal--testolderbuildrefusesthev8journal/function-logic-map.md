# Function Logic Map: `TestOlderBuildRefusesTheV8Journal`

- Source: `internal/journal/migration_v8_test.go`
- AST evidence: `ast.json` (`47390b93f0a39f2a46256ea58f99f024192dce4c9953c39906e44aeded5ceb09`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | validated caller inputs | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 232 | only the branch's documented state transition | existing return/error contract | `TestTestOlderBuildRefusesTheV8Journal` |
| B2 | existing if branch at line 242 | only the branch's documented state transition | existing return/error contract | `TestTestOlderBuildRefusesTheV8Journal` |
| B3 | existing if branch at line 245 | only the branch's documented state transition | existing return/error contract | `TestTestOlderBuildRefusesTheV8Journal` |
| B4 | existing if branch at line 248 | only the branch's documented state transition | existing return/error contract | `TestTestOlderBuildRefusesTheV8Journal` |
| B5 | existing if branch at line 255 | only the branch's documented state transition | existing return/error contract | `TestTestOlderBuildRefusesTheV8Journal` |
| B6 | existing if branch at line 258 | only the branch's documented state transition | existing return/error contract | `TestTestOlderBuildRefusesTheV8Journal` |

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
