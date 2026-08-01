# Function Logic Map: `Journal.AppendStrategyExecutionLink`

- Source: `internal/journal/strategy_lineage.go`
- CodeGraph callers/callees: future fill/position/close lineage writers; no a047 runtime caller
- AST: generated after implementation

## Inputs and invariants

| Input/state | Range | Source of truth | Failure behavior |
|---|---|---|---|
| link | account + attempt + allowed kind + external ref | official outcome journal | reject |
| replay | same semantic unique key may repeat after restart | first durable row | preserve first timestamp and succeed |
| collision | same external key bound to another attempt/account | unique index | typed collision |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Test |
|---|---|---|---|---|
| B1 | malformed kind/key | none | invalid error | input table |
| B2 | attempt/account missing or mismatched | rollback | binding error | account-scope test |
| B3 | first link | one append row | commit | reverse lookup test |
| B4 | exact replay at later time | no mutation | success | delayed replay test |
| B5 | unique-key collision | rollback | typed collision | collision test |

## Calls and live bindings

| Callee | Contract | Failure path | Evidence |
|---|---|---|---|
| SQLite INSERT OR IGNORE + reverse lookup | semantic idempotency independent of timestamp | collision/DB error | lineage tests |

## State mutations and fallbacks

- First call appends one immutable row; exact replay preserves its first timestamp. There is no fallback to another account or attempt.

## Safety conclusion

- Append-only journal transaction; no broker call.
