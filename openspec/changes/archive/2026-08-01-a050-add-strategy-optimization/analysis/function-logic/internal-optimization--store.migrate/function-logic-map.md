# Function Logic Map: `Store.migrate`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| SQLite user version and DDL | version 0..3; all additive columns, legacy verification/digests, triggers and version update commit together | control DB | rollback and refuse Open |
| legacy history | v1 snapshot digest authentic; v2 pointer references verified snapshot; every legacy audit event corroborates snapshot/candidate | immutable tables | corrupt legacy state is never re-signed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | begin transaction fails | none | wrapped error | DB fault path |
| B2 | version read fails | rollback | wrapped error | DB fault path |
| B3 | future schema | no DDL | refuse Open | `TestOpenRefusesNewerSchemaAndSecuresFiles` |
| B4-B18 | schema/column initialization and pre-install trigger integrity | transaction-local | first error aborts | schema/trigger tests |
| B19-B21 | v1 snapshot digest verification and v2 expansion | transaction-local | corrupt legacy digest aborts | v1 migration tests |
| B22-B25 | v2 control pointer and audit corroboration/digest migration | transaction-local | corrupt pointer/event aborts | v2 migration tests |
| B26-B29 | exact triggers, user version 3 and commit | transaction-local until commit | any error rolls back | migration/trigger tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| transaction DDL | creates schema v3 atomically | no retry; Open fails closed | AST |
| snapshot/pointer/audit migration helpers | authenticate legacy state before expanding integrity coverage | strict; no row skip/re-sign fallback | migration tests |
| trigger install/verify | restores exact append-only definitions after controlled digest updates | post-install exact comparison required | trigger drift tests |

## State mutations and fallbacks

- Only the optimization-private schema changes. All migration writes are transactional and no corrupt legacy row is skipped or repaired.

## Safety conclusion

- Safe edit boundary: versioned control-store migration.
- High-risk impact: yes; schema migration binds durable setting/audit authority.
