# Function Logic Map: `Store.audit`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| durable audit ledger | full snapshot/application/change cardinality is intact; every row is structurally valid and digest-authenticated | control DB | any deletion, mismatch, corruption, or query error aborts View read |
| page projection | newest-first, bounded at 1000 only after full-ledger validation | audit table | old corruption cannot hide beyond page limit |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | full-ledger coverage validation fails | none | error | deleted-row tests |
| B2 | bounded query fails | none | error | DB fault coverage |
| B3-B4 | iterate/scan rows; structure invalid | none | corrupt event error | timestamp/corruption tests |
| B5 | event digest missing/mismatched | none | corrupt digest error | valid-looking tamper test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `validateAuditCoverage` | checks entire ledger before paging | aggregate error/mismatch fails closed | deletion tests |
| `scanAuditEventUnchecked`, `digestAuditEvent` | structural and cryptographic row verification | every row required; no fallback | tamper tests |

## State mutations and fallbacks

- Read-only. Audit rows are never repaired, skipped, or re-signed; append-only triggers remain primary write-time protection.

## Safety conclusion

- Safe edit boundary: private audit read integrity.
- High-risk impact: yes; audit evidence integrity is safety-sensitive.
