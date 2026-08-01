# Function Logic Map: `verifyStrategyRiskBinding`

- Source: `internal/journal/strategy_lineage.go`
- CodeGraph callers/callees: `Journal.RecordStrategyDecisionAndReserve` before transaction
- AST: generated after implementation

## Inputs and invariants

| Input/state | Range | Source of truth | Failure behavior |
|---|---|---|---|
| RiskIntent | hash-valid canonical exposure-raising record | Guardian-issued decision | reject before SQL |
| decision payload | one strict canonical JSON `DecisionRecord` and exact SHA-256 | opaque lane decision serializer | reject before SQL |
| identity | SHA-256 of all record fields with Identity empty | canonical DecisionRecord algorithm | reject before SQL |
| denormalized lineage | exact projection of payload plus risk quantity/policy | journal query schema | reject before SQL |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Test |
|---|---|---|---|---|
| B1 | RiskIntent hash/type invalid | none | risk binding error | issuance binding tests |
| B2 | JSON unknown/trailing/noncanonical or digest mismatch | none | payload error | strict payload table |
| B3 | full-record identity mismatch | none | identity error | non-denormalized field mutation |
| B4 | any denormalized/risk field mismatch | none | exact binding error | exhaustive mutation table |
| Success | all exact | none | nil | production issuance success |

## Calls and live bindings

| Callee | Contract | Failure path | Evidence |
|---|---|---|---|
| `ParsePreimage` | typed canonical RiskIntent | fail closed | journal tests |
| JSON decoder/remarshal | exactly one known canonical object | fail closed | payload tests |
| SHA-256 | payload digest + full record identity | fail closed | payload tests |

## State mutations and fallbacks

- Pure validation before a transaction begins. It performs no SQL/network mutation and has no permissive decode fallback.

## Safety conclusion

- Pure pre-transaction validation. No database or network side effects.
