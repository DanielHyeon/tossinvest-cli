# Function Logic Map: `AdoptionRequest.record`

- Source: `internal/journal/adoption.go`
- AST evidence: `ast.json` (`6adc78fcb71ddfc90ee929f0df2acc005105677d93a61a96c26932bb7a90dcf4`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | trimmed identity, decimal prices/quantity, RFC3339 time and valid common policy ID | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing switch branch at line 367 | only the branch's documented state transition | existing return/error contract | `TestAdoptionRecordPolicyID` |
| B2 | existing case branch at line 368 | only the branch's documented state transition | existing return/error contract | `TestAdoptionRecordPolicyID` |
| B3 | existing case branch at line 371 | only the branch's documented state transition | existing return/error contract | `TestAdoptionRecordPolicyID` |
| B4 | existing case branch at line 374 | only the branch's documented state transition | existing return/error contract | `TestAdoptionRecordPolicyID` |
| B5 | existing if branch at line 379 | only the branch's documented state transition | existing return/error contract | `TestAdoptionRecordPolicyID` |
| B6 | existing if branch at line 380 | only the branch's documented state transition | existing return/error contract | `TestAdoptionRecordPolicyID` |
| B7 | existing if branch at line 387 | only the branch's documented state transition | existing return/error contract | `TestAdoptionRecordPolicyID` |
| B8 | existing if branch at line 390 | only the branch's documented state transition | existing return/error contract | `TestAdoptionRecordPolicyID` |
| B9 | existing if branch at line 394 | only the branch's documented state transition | existing return/error contract | `TestAdoptionRecordPolicyID` |
| B10 | existing if branch at line 397 | only the branch's documented state transition | existing return/error contract | `TestAdoptionRecordPolicyID` |
| B11 | existing if branch at line 405 | only the branch's documented state transition | existing return/error contract | `TestAdoptionRecordPolicyID` |
| B12 | existing if branch at line 409 | only the branch's documented state transition | existing return/error contract | `TestAdoptionRecordPolicyID` |
| B13 | existing if branch at line 416 | only the branch's documented state transition | existing return/error contract | `TestAdoptionRecordPolicyID` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| riskcalc decimal validation, sha256 digest construction | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- include policy ID in deterministic adoption preimage/digest.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: include policy ID in deterministic adoption preimage/digest.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
