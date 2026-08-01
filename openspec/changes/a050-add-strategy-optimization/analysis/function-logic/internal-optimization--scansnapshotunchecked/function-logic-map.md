# Function Logic Map: `scanSnapshotUnchecked`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| scanner row | exact snapshot column order | snapshot SELECT statements | scan error propagated |
| structural metadata | valid JSON maps, boolean ints 0/1, positive ordered versions, non-blank actor/audit | snapshot schema/lifecycle | corrupt snapshot error |
| created time | canonical non-zero RFC3339Nano UTC | persisted row | corrupt timestamp error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | SQL scan fails | local output only | error | DB scan coverage |
| B2 | JSON/boolean/version/identity invariant fails | local output only | corrupt snapshot | snapshot corruption matrix |
| B3 | canonical timestamp parse fails | local output only | corrupt snapshot timestamp | timestamp corruption test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| scanner `Scan` | loads exact persisted metadata | one call | store query bindings |
| `json.Unmarshal` | decodes desired/effective maps without defaults | any error fails | corruption tests |
| `parseStoredTime` | enforces canonical time | strict/no fallback | timestamp test |

## State mutations and fallbacks

- Builds a local snapshot only. “Unchecked” skips only digest comparison for controlled migration; it still validates all structure, identities, booleans, versions, and time.

## Safety conclusion

- Safe edit boundary: structural snapshot decoding used by verified read and migration.
- High-risk impact: yes; malformed persisted metadata must never enter lifecycle logic.
