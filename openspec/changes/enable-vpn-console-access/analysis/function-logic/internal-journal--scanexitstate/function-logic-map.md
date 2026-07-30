# Function Logic Map: `scanExitState`

- Source: `internal/journal/apply_hook.go`
- AST evidence: `ast.json` (`d3edb0b0a6bc08316d569e329e132bd5fc0458c8ffe276a877d816ca773c99ad`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | query column order exactly matches scan targets | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 553 | only the branch's documented state transition | existing return/error contract | `TestLegacyLadderPolicyID` |
| B2 | existing if branch at line 556 | only the branch's documented state transition | existing return/error contract | `TestLegacyLadderPolicyID` |
| B3 | existing if branch at line 560 | only the branch's documented state transition | existing return/error contract | `TestLegacyLadderPolicyID` |
| B4 | existing if branch at line 563 | only the branch's documented state transition | existing return/error contract | `TestLegacyLadderPolicyID` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| row.Scan | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- map nullable legacy LADDER ID to default_v1 without rewriting rows.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: map nullable legacy LADDER ID to default_v1 without rewriting rows.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
