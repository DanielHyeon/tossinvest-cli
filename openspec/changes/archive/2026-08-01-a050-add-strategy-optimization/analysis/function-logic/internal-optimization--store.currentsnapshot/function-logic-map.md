# Function Logic Map: `Store.currentSnapshot`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| control row | singleton version plus domain-separated digest | optimization control table | query failure propagated |
| referenced snapshot | exists and passes full snapshot digest validation | append-only snapshot table | snapshot error propagated |
| pointer binding | control digest equals digest(version, verified snapshot digest) | `digestControlPointer` | corrupt pointer error before any caller state change |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | control row read fails | none | error | DB error coverage |
| B2 | referenced snapshot missing/corrupt | none | error | snapshot corruption coverage |
| B3 | pointer digest mismatches version/snapshot digest | none | corrupt pointer error | rollback tamper test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Store.snapshot` | loads and digest-verifies pointed snapshot | one read/no fallback version | snapshot tests |
| `digestControlPointer` | binds pointer to verified immutable state | deterministic | control tamper tests |

## State mutations and fallbacks

- Read-only. It never scans for a “latest” snapshot or repairs the pointer, so rollback/tamper cannot be laundered.

## Safety conclusion

- Safe edit boundary: authenticated current-state lookup.
- High-risk impact: yes; Read, Preview, Apply and conflict recovery depend on this pointer.
