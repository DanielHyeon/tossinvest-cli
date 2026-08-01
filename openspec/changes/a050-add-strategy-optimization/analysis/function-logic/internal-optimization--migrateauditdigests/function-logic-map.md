# Function Logic Map: `migrateAuditDigests`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| legacy audit rows | structurally valid and exactly corroborated by their snapshot and candidate change | v2 audit/snapshot/candidate tables | any corrupt/unmatched event aborts entire migration |
| new digest | domain-separated full event digest | v3 audit contract | transaction update failure aborts all rows |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | audit query fails | none | wrapped error | migration rollback coverage |
| B2 | iterate legacy rows | local buffer | continue | multi-row migration test |
| B3 | structural scan fails | closes rows | wrapped error | corrupt legacy audit test |
| B4 | close fails | none | wrapped error | DB fault review |
| B5 | iterator fails | none | wrapped error | DB fault review |
| B6 | iterate buffered events | transaction updates only | continue | migration test |
| B7 | snapshot/candidate corroboration fails | none | error | corrupt legacy audit test |
| B8 | digest update fails | transaction-local | wrapped event error | rollback coverage |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `scanAuditEventUnchecked` | structural validation without trusting empty legacy digest | any error aborts | corruption tests |
| `validateLegacyAuditEvent` | proves event against immutable snapshot and candidate | exact match required | corrupt legacy test |
| `digestAuditEvent` | derives v3 full event digest | deterministic | tamper test |

## State mutations and fallbacks

- Buffers/validates all rows before updating. Updates occur only in the migration transaction while audit update trigger is temporarily removed and later reinstalled/verified.
- No corrupt event is skipped or re-signed.

## Safety conclusion

- Safe edit boundary: atomic v2-to-v3 audit integrity migration.
- High-risk impact: yes; audit history must remain evidentiary.
