# Function Logic Map: `digestSnapshotV1`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| legacy snapshot | structurally validated schema-v1 snapshot | migration-only caller | deterministic legacy digest used only to authenticate before v2 upgrade |
| covered legacy fields | versions, desired/effective maps, effective-entry flag, activation manifest | frozen v1 historical algorithm | must remain byte-compatible with already persisted v1 digests |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | branchless happy path reproduces frozen v1 JSON/SHA-256 algorithm | local allocation only | legacy hex digest | legacy migration/tamper test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| JSON marshal and SHA-256 | verify persisted v1 row before replacing digest | fixed supported types; deterministic | migration tests |

## State mutations and fallbacks

- Pure compatibility helper called only during version<2 migration. New snapshots and post-migration reads use the full `digestSnapshot` algorithm.
- It does not bless a row by itself; migration compares the stored legacy digest first and aborts on mismatch.

## Safety conclusion

- Safe edit boundary: frozen legacy digest verification only.
- High-risk impact: yes; changing this algorithm could reject valid history or re-sign tampered legacy rows.
