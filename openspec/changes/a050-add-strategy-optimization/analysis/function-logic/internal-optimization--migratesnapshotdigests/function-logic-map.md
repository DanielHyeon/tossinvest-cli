# Function Logic Map: `migrateSnapshotDigests`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| migration transaction | schema v1 snapshots before v2 digest coverage | `Store.migrate` | any read/validation/close/iteration/update error aborts transaction |
| snapshot rows | structurally valid immutable metadata; old digest not trusted during conversion | persisted v1 rows | corrupt row refuses migration |
| new digest | covers every immutable snapshot metadata field | `digestSnapshot` v2 contract | updated only inside transaction before triggers reinstalled |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | snapshot query fails | none | wrapped error | migration rollback test |
| B2 | iterate legacy snapshot rows | local slice | continue | legacy migration test |
| B3 | structural scan fails | closes rows | wrapped validation error | corrupt legacy snapshot test |
| B4 | rows close fails | none | wrapped error | DB error coverage |
| B5 | iterator reports error | none | wrapped error | DB error coverage |
| B6 | iterate structurally validated snapshots | transaction updates only | continue | digest migration test |
| B7 | stored legacy digest missing or differs from frozen v1 digest | none | invalid legacy digest error | legacy tamper test |
| B8 | v2 digest update fails | transaction-local | wrapped version error | rollback test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `scanSnapshotUnchecked` | validates structure while intentionally ignoring legacy digest | error aborts all migration | corruption tests |
| `digestSnapshotV1`, `digestSnapshot` | authenticate frozen legacy row, then derive v2 full metadata digest | deterministic; mismatch aborts before update | migration and metadata tamper coverage |
| transaction update | rewrites digest only during controlled migration | whole transaction rolls back on any error | migration tests |

## State mutations and fallbacks

- Updates only snapshot digest fields inside the migration transaction after the old update trigger is temporarily removed; exact append-only triggers are restored and verified before commit.
- No invalid row is skipped and no partial digest migration can commit.

## Safety conclusion

- Safe edit boundary: atomic v1-to-v2 immutable snapshot digest upgrade.
- High-risk impact: yes; snapshot history integrity underpins CAS/rollback/audit.
