# Function Logic Map: `Journal.RecordDecision`

- Source: `internal/journal/decision.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
| --- | --- | --- | --- |
| `DecisionRequest` | fully validated identity, account, class, preimage, nonce, time window | `DecisionRequest.build` | invalid request error; no insert |
| journal DB | open migrated SQLite database | `journal.Open` | insert error; no success value |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
| --- | --- | --- | --- | --- |
| B1 | `req.build()` rejects any invariant | none | zero decision + validation error | `TestRecordDecisionRefusals` and class/preimage tests |
| B2 | insert fails, including duplicate nonce/ID | database rejects row | zero decision + wrapped DB error | nonce/durability tests |
| Scenario | build and insert succeed | one durable decision row | canonical `Decision` | `TestRecordAndLookupDecision` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
| --- | --- | --- | --- |
| `DecisionRequest.build` | validate and derive risk hash/client key/canonical preimage | fail closed; no defaults for missing authority fields | CodeGraph + AST |
| `insertDecisionRow` | one shared insert for DB and transactional issuance paths | caller context controls cancellation; error is returned | CodeGraph + AST |

## State mutations and fallbacks

- Inserts exactly one decision; no retry or fallback mutates a partial row.
- Exposure-raising production issuance uses the atomic decision+reservation
  transaction rather than this bare method.
- a047 provenance should be additive and immutably bound, never smuggled into
  mutable/recomputed decision fields.

## Safety conclusion

- Safe edit boundary: avoid editing this core recorder if a separate immutable
  provenance table/reference can be committed atomically by issuance.
- High-risk impact: yes.
