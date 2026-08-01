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
| B1 | exact AST `if` at source line 737: `if journalHash := HashPreimage(decision.RiskPreimage); journalHash != decision.RiskHash {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B2 | exact AST `if` at source line 741: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B3 | exact AST `if` at source line 745: `if !ok {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B4 | exact AST `if` at source line 751: `if err := decoder.Decode(&signalRecord); err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B5 | exact AST `if` at source line 754: `if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B6 | exact AST `if` at source line 758: `if err != nil \|\| string(canonicalPayload) != lineage.DecisionPayload {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B7 | exact AST `if` at source line 766: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B8 | exact AST `if` at source line 771: `if lineage.DecisionPayloadDigest != wantPayloadDigest \|\| signalRecord.Identity != wantIdentity \|\| signalRecord.Identity != lineage.DecisionIdentity \|\|` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |

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
