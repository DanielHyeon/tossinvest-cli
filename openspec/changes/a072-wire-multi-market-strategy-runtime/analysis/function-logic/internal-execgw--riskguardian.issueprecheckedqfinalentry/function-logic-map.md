# Function Logic Map: `RiskGuardian.IssuePrecheckedQFinalEntry`

- Source: `internal/execgw/riskguardian_qfinal.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| opaque precheck + context | nonzero sealed decision, collector, transaction, current FX/snapshot evidence | prior Precheck + Guardian clock + journal recollection | refuse before decision insert or rollback journal transaction |
| decision envelope | exact Guardian policy plus the same opaque FX pair/source/version/digest evaluated during precheck | sealed precheck, not caller JSON | envelope mismatch/refusal; zero first-leg authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | incomplete precheck or invalid policy version | none | typed/error refusal | existing tests |
| B3 | collector failure | recollection attempt only | error; transaction absent | existing journal tests |
| B4-B5 | exposure/usage failure | none committed | refusal/error | existing tests |
| B6 | atomic journal issue failure | journal transaction rollback | issuance refusal | existing tests |
| added | opaque FX expired/tampered at final clock | none committed | currency unresolved | final freshness test |
| added | account-base rebind differs from sealed precheck or reservation currency is not base | none committed | typed FX/snapshot refusal | paired KR/US issue RED |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| opaque FX ReserveAt | last-moment authority revalidation | exact Guardian clock | focused final-expiry test |
| `risk.BindAccountBaseFX` | rebind same opaque evidence at issue clock | pair/source/version/digest must match sealed precheck | paired KR/US issue RED |
| RecordQFinalDecisionAndReserveWithRecollection | atomic decision/reservations | retry/recollection contract preserved | journal implementation |

## State mutations and fallbacks

- Rebuilds public reserve FX DTO from retained opaque authority immediately before journal calculation.
- Target issue stamps a versioned account-base FX binding in decision `limits_json`; aggregate and all
  five monetary reservations must use that exact base currency/evidence digest.

## Safety conclusion

- Safe edit boundary: no stale precheck can mint authority; existing last-moment journal barrier
  remains atomic and receives one exact account-base envelope.
- High-risk impact: yes — atomic Guardian issuance.
