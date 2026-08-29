# Function Logic Map: `TestSchemaTablesAndColumns`

- Source: `internal/journal/schema_test.go`
- AST evidence: `ast.json` (`f168a7b83293e52443453b19c389ec3cb3740a2356b739381f172e4b55c4904b`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | validated caller inputs | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 112 | only the branch's documented state transition | existing return/error contract | `TestTestSchemaTablesAndColumns` |
| B2 | existing for branch at line 116 | only the branch's documented state transition | existing return/error contract | `TestTestSchemaTablesAndColumns` |
| B3 | existing if branch at line 118 | only the branch's documented state transition | existing return/error contract | `TestTestSchemaTablesAndColumns` |
| B4 | existing if branch at line 123 | only the branch's documented state transition | existing return/error contract | `TestTestSchemaTablesAndColumns` |
| B5 | existing if branch at line 127 | only the branch's documented state transition | existing return/error contract | `TestTestSchemaTablesAndColumns` |
| B6 | existing range branch at line 223 | only the branch's documented state transition | existing return/error contract | `TestTestSchemaTablesAndColumns` |
| B7 | existing if branch at line 226 | only the branch's documented state transition | existing return/error contract | `TestTestSchemaTablesAndColumns` |

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
