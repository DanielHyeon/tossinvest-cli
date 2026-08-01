# Function Logic Map: `migrateControlPointerDigest`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| legacy control version | zero or points to existing fully digest-verified snapshot | v2 control/snapshot tables | invalid pointer/snapshot aborts v3 migration |
| new control digest | digest(version, verified snapshot digest), or digest(0, empty) | v3 control contract | transaction update error aborts migration |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | control pointer read fails | none | wrapped error | migration rollback coverage |
| B2 | legacy version is non-zero | read referenced snapshot | continue | v2 migration test |
| B3 | referenced snapshot missing/corrupt | none | wrapped validation error | corrupt migration test |
| B4 | control digest update fails | transaction-local | wrapped error | rollback coverage |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `scanSnapshot` | authenticate pointed legacy snapshot before binding | strict full digest; no fallback | v2 migration tests |
| `digestControlPointer` | create v3 pointer digest | deterministic | pointer tamper tests |

## State mutations and fallbacks

- Writes only the new digest in the migration transaction. It does not alter current version or choose another snapshot.

## Safety conclusion

- Safe edit boundary: atomic v2-to-v3 pointer binding.
- High-risk impact: yes; migration must not legitimize a corrupt/rolled-back pointer.
