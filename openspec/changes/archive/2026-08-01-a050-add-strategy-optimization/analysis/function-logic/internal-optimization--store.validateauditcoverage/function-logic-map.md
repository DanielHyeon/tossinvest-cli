# Function Logic Map: `Store.validateAuditCoverage`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| durable ledger | every post-initial snapshot has one application; every application has exactly the candidate change count of matching audit rows | snapshot/application/candidate/audit tables | any gap/malformed changes/group metadata mismatch is corrupt coverage |
| scope | entire ledger, independent of newest-page read limit | aggregate SQL | old deletion remains detectable |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | aggregate validation query fails | none | wrapped error | DB fault coverage |
| B2 | snapshot/application counts differ or invalid group exists | none | corrupt coverage error | deleted application/audit/partial multi-change tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| bounded aggregate SQL | validates full ledger cardinality and snapshot metadata agreement | one read/no page fallback | deletion tests |

## State mutations and fallbacks

- Read-only aggregate check. Append-only triggers are primary prevention; this detects offline deletion/tamper across all history before returning a page.

## Safety conclusion

- Safe edit boundary: full durable audit coverage validation.
- High-risk impact: yes; missing audit/application rows must not be hidden outside the read limit.
