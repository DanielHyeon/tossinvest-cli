# Function Logic Map: `strategyFirstLegAdmissionBridge.admit`

- Source: `internal/app/engine/strategy_first_leg_admission.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| accepted result | exact accepted, pure KR/US leg 1 result | `strategyflow.Result` | `STRATEGY_RESULT_INVALID` |
| authority loader | engine-private loader returning Guardian issuance input | production assembly | `AUTHORITY_UNAVAILABLE/COLLECTION_FAILED` |
| Guardian issuer | exactly precheck + atomic issue methods | shared account `RiskGuardian` | `AUTHORITY_MISMATCH/ATOMIC_ADMISSION_FAILED` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | invalid result | none | result invalid | invalid/tamper tests |
| B2 | bridge/loader/issuer nil | none | authority unavailable | dormant paired test |
| B3 | loader fails | none | collection failed | paired failure test |
| B4 | loaded issuance mismatches accepted result | no Guardian call | authority mismatch | cross-market/tamper tests |
| B5 | Guardian precheck refuses | no atomic issue | authority mismatch | issuer spy test |
| B6 | Guardian issue refuses | issuer controls atomic boundary | atomic admission failed | issuer spy test |
| B7 | both calls succeed | one opaque receipt | admitted | paired KR/US test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `validateStrategyFirstLegResult` | fail closed before authority | synchronous, no retry | CodeGraph + AST |
| loader `collectStrategyFirstLegAuthority` | collect package-owned entry/CAS metadata | context-aware, one call | CodeGraph + AST |
| Guardian `PrecheckQFinalCampaignFirstLeg` | seal q_final/projection | no mutation | CodeGraph + AST |
| Guardian `IssuePrecheckedQFinalCampaignFirstLeg` | perform one atomic issue | no fallback path | CodeGraph + AST |

## State mutations and fallbacks

- Bridge owns no Journal, lease, Gateway, broker, activation or toggle capability. Production remains dormant when loader or issuer is absent.

## Safety conclusion

- Safe edit boundary: replace raw journal recorder with the two-method Guardian capability only.
- High-risk impact: yes — entry admission seam; paired zero-call and static capability tests required.
