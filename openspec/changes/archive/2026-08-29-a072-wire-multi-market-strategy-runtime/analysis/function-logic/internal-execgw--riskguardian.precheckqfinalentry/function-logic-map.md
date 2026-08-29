# Function Logic Map: `RiskGuardian.PrecheckQFinalEntry`

- Source: `internal/execgw/riskguardian_qfinal.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| request + Guardian | exact policy/digest, canonical market/currency/symbol, positive candidate, exact owner, opaque FX current at Guardian clock | RiskGuardian state + opaque officialfx authority | typed q_final refusal; no journal/broker call |
| account-base FX | KR identity or US official quote-to-account conversion, exact Guardian policy base and market quote | `risk.BindAccountBaseFX` over retained opaque evidence | stale/forged/wrong-pair/cross-market refusal; no collector/write |
| account state | cash in quote currency; exposure/loss/equity in account base currency | authoritative caller snapshot later recollected at issue | unit mismatch refuses before q_final authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B8 | nil/stale/scope/quantity/Guardian-currency/owner/reserve-currency failures | none | typed refusal | existing q_final tests |
| B9 | public FX DTO/source-label substitution or missing opaque authority | none | currency unresolved | new authority bypass tests |
| B10-B14 | Guardian cap, admission or ordinary Guardian chain failure | none | typed refusal | existing focused tests |
| added B15 | market quote differs from Guardian account-base currency | no longer an automatic refusal when exact account-base FX binds | continue through account-base sizing | paired KR/US account-base RED |
| added B16 | frozen FX missing/stale/forged/wrong pair/cross-market | none | typed currency unresolved | paired KR/US authority matrix |
| added B17 | quote cash or base exposure/account units are inconsistent | none | typed Guardian-chain refusal | paired KR/US unit matrix |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| opaque FX ReserveAt | verify seal, exact pair and freshness at Guardian clock | no fallback | new officialfx authority tests |
| risk.StrategyEntryQuantity | existing Guardian cap | exact decimal/error | existing tests |
| `risk.BindAccountBaseFX` | turn opaque official evidence into the only sealed arithmetic authority | exact market/policy/current-at validation; no scalar fallback | account-base risk paired suite |
| `risk.AccountBaseStrategyEntryQuantity` | size quote prices against account-base risk/notional limits | floor quantity once | paired KR/US q_final RED |
| riskbucket.CalculateAdmission | five-bucket q_final | pure typed refusal | current source |
| evaluateChain/EntryExposureValue | existing Guardian safety chain | fail closed | current source |

## State mutations and fallbacks

- Deep-copies caller slices, overwrites QCandidate, EvaluatedAt, QExistingGuardian and FX DTO from opaque authority.
- Target precheck retains the same opaque evidence and the sealed `risk.AccountBaseFX`; no public
  rate, haircut, source label or caller evaluation time becomes authority.

## Safety conclusion

- Safe edit boundary: remove only the quote==Guardian-limit restriction; replace it with mandatory
  exact account-base FX binding for both KR identity and US official conversion. Caller FX fields are
  never authority and cannot survive overwrite.
- High-risk impact: yes — Guardian sizing; no mutation in this phase.
